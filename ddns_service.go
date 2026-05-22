// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package main

import (
	"time"

	"ez-ddns/dao"

	"github.com/RussellLuo/timingwheel"
)

const refreshInterval = 1 * time.Minute
const updateSystemStatusInterval = 5 * time.Minute

var taskScheduler *TaskScheduler[string]
var refreshChan = make(chan struct{}, 10)

func StartDDNSService() {
	InitLogger()
	Info("启动 DDNS 服务...")

	tw := timingwheel.NewTimingWheel(time.Second, 65536)
	tw.Start()
	defer tw.Stop()

	taskScheduler = NewTaskScheduler[string](tw)
	refreshTasks()

	tw.ScheduleFunc(&everyScheduler{Interval: refreshInterval}, refreshTasks)
	updateSystemStatusTask()
	tw.ScheduleFunc(&everyScheduler{Interval: updateSystemStatusInterval}, updateSystemStatusTask)

	go listenConfigChanges()

	Info("DDNS 服务已启动，按 Ctrl+C 退出")
	select {}
}

func listenConfigChanges() {
	for range refreshChan {
		Debug("收到配置变更通知，刷新任务...")
		refreshTasks()
	}
}

func NotifyConfigChange() {
	if taskScheduler == nil {
		Debug("DDNS 服务尚未启动，忽略配置变更通知")
		return
	}

	select {
	case refreshChan <- struct{}{}:
		Debug("配置变更通知已发送")
	default:
		Debug("配置变更通知通道已满，忽略")
	}
}

func updateSystemStatusTask() {
	if err := UpdateSystemStatus(); err != nil {
		Error("更新系统状态失败: %v", err)
	}
}

func refreshTasks() {
	Debug("开始刷新配置和任务...")

	configs, err := dao.Config.ReadAll()
	if err != nil {
		Error("刷新配置失败: %v", err)
		return
	}

	tasks := make([]SyncTask[string], 0, len(configs))
	for _, config := range configs {
		tasks = append(tasks, SyncTask[string]{
			Key:      config.ID,
			Interval: time.Duration(config.Interval) * time.Second,
			Executor: handleConfig,
		})
		Debug("配置ID[%s] - 任务已刷新", config.ID)
	}

	taskScheduler.SyncTasks(tasks)

	Debug("配置和任务刷新完成")
}