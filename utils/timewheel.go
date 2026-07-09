// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package utils

import (
	"sync"
	"time"
)

type TaskID string

type Task struct {
	ID              TaskID
	internalID      int64 // 隐性唯一ID，用于防止ABA问题
	Interval        time.Duration
	remainingRounds int // 剩余需要经过的轮数
	Handler         func()
}

var taskInternalIDCounter int64
var taskInternalIDMu sync.Mutex

func getNextInternalID() int64 {
	taskInternalIDMu.Lock()
	defer taskInternalIDMu.Unlock()
	taskInternalIDCounter++
	return taskInternalIDCounter
}

type TimeWheel struct {
	sync.RWMutex
	slots      []*taskSlot
	currentIdx int
	interval   time.Duration // tick精度
	ticker     *time.Ticker
	tasks      map[TaskID]*Task
	quitChan   chan struct{}
}

type taskSlot struct {
	mu    sync.Mutex
	tasks []*Task
}

// NewTimeWheel 创建时间轮（默认60个槽位）
// precision: tick精度（如time.Second表示秒级精度）
func NewTimeWheel(precision time.Duration) *TimeWheel {
	return NewTimeWheelWithSlots(precision, 60)
}

// NewTimeWheelWithSlots 创建时间轮（支持自定义槽位数）
// precision: tick精度（如time.Second表示秒级精度）
// slotCount: 槽位数，建议值：60(默认)、3600(秒级精细分布)、86400(分钟级精细分布)
func NewTimeWheelWithSlots(precision time.Duration, slotCount int) *TimeWheel {
	if slotCount < 1 {
		slotCount = 60
	}

	slots := make([]*taskSlot, slotCount)
	for i := range slots {
		slots[i] = &taskSlot{}
	}

	return &TimeWheel{
		slots:    slots,
		interval: precision,
		tasks:    make(map[TaskID]*Task),
		quitChan: make(chan struct{}),
	}
}

func (tw *TimeWheel) Start() {
	tw.Lock()
	if tw.ticker != nil {
		tw.Unlock()
		return
	}
	tw.ticker = time.NewTicker(tw.interval)
	tw.quitChan = make(chan struct{})
	tw.Unlock()

	go tw.run()
}

func (tw *TimeWheel) Stop() {
	tw.Lock()
	if tw.ticker == nil {
		tw.Unlock()
		return
	}
	tw.ticker.Stop()
	tw.ticker = nil
	quitChan := tw.quitChan
	tw.quitChan = make(chan struct{})
	tw.Unlock()

	close(quitChan)
}

func (tw *TimeWheel) run() {
	for {
		select {
		case <-tw.ticker.C:
			tw.tick()
		case <-tw.quitChan:
			return
		}
	}
}

func (tw *TimeWheel) tick() {
	tw.Lock()
	slot := tw.slots[tw.currentIdx]
	tw.currentIdx = (tw.currentIdx + 1) % len(tw.slots)
	tw.Unlock()

	slot.mu.Lock()
	tasksToExecute := make([]*Task, 0)
	tasksToKeep := make([]*Task, 0)

	for _, task := range slot.tasks {
		tw.RLock()
		existing, exists := tw.tasks[task.ID]
		tw.RUnlock()

		if !exists || existing.internalID != task.internalID {
			continue
		}

		if task.remainingRounds > 0 {
			task.remainingRounds--
			tasksToKeep = append(tasksToKeep, task)
		} else {
			tasksToExecute = append(tasksToExecute, task)
		}
	}
	slot.tasks = tasksToKeep
	slot.mu.Unlock()

	for _, task := range tasksToExecute {
		go task.Handler()
		tw.rescheduleTask(task)
	}
}

func (tw *TimeWheel) rescheduleTask(task *Task) {
	tw.Lock()
	defer tw.Unlock()

	if _, exists := tw.tasks[task.ID]; !exists {
		return
	}

	tw.addTaskToSlot(task)
}

func (tw *TimeWheel) AddTask(task *Task) error {
	tw.Lock()
	defer tw.Unlock()

	if _, exists := tw.tasks[task.ID]; exists {
		return nil
	}

	task.internalID = getNextInternalID()
	tw.tasks[task.ID] = task
	tw.addTaskToSlot(task)

	return nil
}

func (tw *TimeWheel) UpdateTask(task *Task) error {
	tw.Lock()
	defer tw.Unlock()

	existing, exists := tw.tasks[task.ID]
	if !exists {
		return nil
	}

	existing.Interval = task.Interval
	existing.Handler = task.Handler

	return nil
}

func (tw *TimeWheel) RemoveTask(taskID TaskID) {
	tw.Lock()
	delete(tw.tasks, taskID)
	tw.Unlock()
}

func (tw *TimeWheel) addTaskToSlot(task *Task) {
	totalTicks := int(task.Interval / tw.interval)
	slotCount := len(tw.slots)

	task.remainingRounds = totalTicks / slotCount
	slotIdx := (tw.currentIdx + totalTicks%slotCount) % slotCount

	tw.slots[slotIdx].mu.Lock()
	tw.slots[slotIdx].tasks = append(tw.slots[slotIdx].tasks, task)
	tw.slots[slotIdx].mu.Unlock()
}
