// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package main

import (
	"sync"
	"time"

	"github.com/RussellLuo/timingwheel"
)

type TaskAction int

const (
	TaskAdd TaskAction = iota
	TaskDelete
)

type TaskExecutor[K comparable] func(key K)

type TaskRequest[K comparable] struct {
	Action   TaskAction
	Key      K
	Interval time.Duration
	Executor TaskExecutor[K]
	Version  string
}

type SyncTask[K comparable] struct {
	Key      K
	Interval time.Duration
	Executor TaskExecutor[K]
	Version  string
}

type ScheduledTask[K comparable] struct {
	Timer    *timingwheel.Timer
	Interval time.Duration
	Version  string
}

type TaskScheduler[K comparable] struct {
	tw             *timingwheel.TimingWheel
	taskChannel    chan TaskRequest[K]
	scheduledTasks map[K]*ScheduledTask[K]
	mu             sync.RWMutex
}

func NewTaskScheduler[K comparable](tw *timingwheel.TimingWheel) *TaskScheduler[K] {
	ts := &TaskScheduler[K]{
		tw:             tw,
		taskChannel:    make(chan TaskRequest[K], 100),
		scheduledTasks: make(map[K]*ScheduledTask[K]),
	}
	go ts.taskSubscriber()
	return ts
}

func (ts *TaskScheduler[K]) SubscribeTask(req TaskRequest[K]) {
	ts.taskChannel <- req
}

func (ts *TaskScheduler[K]) SyncTasks(tasks []SyncTask[K]) {
	currentKeys := make(map[K]bool)

	for _, task := range tasks {
		currentKeys[task.Key] = true
		ts.taskChannel <- TaskRequest[K]{
			Action:   TaskAdd,
			Key:      task.Key,
			Interval: task.Interval,
			Executor: task.Executor,
			Version:  task.Version,
		}
	}

	ts.mu.Lock()
	for key := range ts.scheduledTasks {
		if !currentKeys[key] {
			ts.taskChannel <- TaskRequest[K]{
				Action: TaskDelete,
				Key:    key,
			}
		}
	}
	ts.mu.Unlock()
}

func (ts *TaskScheduler[K]) taskSubscriber() {
	for req := range ts.taskChannel {
		switch req.Action {
		case TaskAdd:
			ts.addOrUpdateTask(req)
		case TaskDelete:
			ts.deleteTask(req.Key)
		}
	}
}

func (ts *TaskScheduler[K]) addOrUpdateTask(req TaskRequest[K]) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if existingTask, exists := ts.scheduledTasks[req.Key]; exists {
		if existingTask.Interval == req.Interval && existingTask.Version == req.Version {
			return
		}
		existingTask.Timer.Stop()
	}

	timer := ts.tw.ScheduleFunc(&everyScheduler{Interval: req.Interval}, func() {
		if req.Executor != nil {
			req.Executor(req.Key)
		}
	})

	ts.scheduledTasks[req.Key] = &ScheduledTask[K]{
		Timer:    timer,
		Interval: req.Interval,
		Version:  req.Version,
	}
}

func (ts *TaskScheduler[K]) deleteTask(key K) {
	ts.mu.Lock()
	defer ts.mu.Unlock()

	if task, exists := ts.scheduledTasks[key]; exists {
		task.Timer.Stop()
		delete(ts.scheduledTasks, key)
	}
}

type everyScheduler struct {
	Interval time.Duration
}

func (s *everyScheduler) Next(prev time.Time) time.Time {
	return prev.Add(s.Interval)
}