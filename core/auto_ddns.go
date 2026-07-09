// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package core

import (
	"context"
	"sync"
	"time"

	"ez-ddns/dao"
	"ez-ddns/model"
	"ez-ddns/utils"
)

type AutoDDNSService struct {
	sync.RWMutex
	timeWheel   *utils.TimeWheel
	configRepo  dao.ConfigRepository
	domainRepo  dao.DomainRepository
	ddnsService *DDNSService
	updateChan  chan struct{}
	quitChan    chan struct{}
	wg          sync.WaitGroup
}

func NewAutoDDNSService(configRepo dao.ConfigRepository, domainRepo dao.DomainRepository, ddnsService *DDNSService) *AutoDDNSService {
	return &AutoDDNSService{
		timeWheel:   utils.NewTimeWheel(time.Second), // 秒级精度，自动支持任意时间间隔
		configRepo:  configRepo,
		domainRepo:  domainRepo,
		ddnsService: ddnsService,
		updateChan:  make(chan struct{}, 1), // 缓冲区为1实现去重通知：多个配置变更合并为一次处理
	}
}

func (s *AutoDDNSService) Start(ctx context.Context) error {
	s.Lock()
	if s.quitChan != nil {
		s.Unlock()
		return nil
	}
	s.quitChan = make(chan struct{})
	s.Unlock()

	err := s.loadAllTasks(ctx)
	if err != nil {
		return err
	}

	s.timeWheel.Start()

	s.wg.Add(1)
	go s.watchConfigChanges(ctx)

	logger := utils.LoggerFromContext(ctx)
	logger.Info("AutoDDNS服务已启动")
	return nil
}

func (s *AutoDDNSService) Stop(ctx context.Context) {
	s.Lock()
	if s.quitChan == nil {
		s.Unlock()
		return
	}
	quitChan := s.quitChan
	s.quitChan = nil
	s.Unlock()

	s.timeWheel.Stop()
	close(quitChan)
	s.wg.Wait()

	logger := utils.LoggerFromContext(ctx)
	logger.Info("AutoDDNS服务已停止")
}

func (s *AutoDDNSService) loadAllTasks(ctx context.Context) error {
	configs, err := s.configRepo.ReadAll(ctx)
	if err != nil {
		logger := utils.LoggerFromContext(ctx)
		logger.Error("加载配置失败: %v", err)
		return err
	}

	logger := utils.LoggerFromContext(ctx)
	logger.Debug("已加载 %d 个配置", len(configs))

	for _, config := range configs {
		s.scheduleConfigTasks(ctx, &config)
	}

	return nil
}

func (s *AutoDDNSService) scheduleConfigTasks(ctx context.Context, config *model.Config) {
	taskID := utils.TaskID("config_" + config.ID)

	interval := time.Duration(config.Interval) * time.Second

	task := &utils.Task{
		ID:       taskID,
		Interval: interval,
		Handler: func() {
			s.ddnsService.HandleConfig(ctx, config.ID)
		},
	}

	s.timeWheel.RemoveTask(taskID)
	s.timeWheel.AddTask(task)

	logger := utils.LoggerFromContext(ctx)
	logger.Debug("已为配置 [%s] 调度定时任务，间隔 %d 秒", config.ID, config.Interval)
}

func (s *AutoDDNSService) watchConfigChanges(ctx context.Context) {
	defer s.wg.Done()

	for {
		select {
		case <-s.updateChan:
			logger := utils.LoggerFromContext(ctx)
			logger.Debug("收到配置变更通知，开始加载所有任务")
			s.loadAllTasks(ctx)
			logger.Debug("所有任务已加载")
		case <-s.quitChan:
			return
		}
	}
}

func (s *AutoDDNSService) NotifyConfigChange() {
	select {
	case s.updateChan <- struct{}{}:
	default:
	}
}
