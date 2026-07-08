// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package core

import (
	"context"
	"sync"
	"time"

	"ez-ddns/v2/dao"
	"ez-ddns/v2/model"
	"ez-ddns/v2/utils"
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

func (s *AutoDDNSService) Start() error {
	s.Lock()
	if s.quitChan != nil {
		s.Unlock()
		return nil
	}
	s.quitChan = make(chan struct{})
	s.Unlock()

	err := s.loadAllTasks()
	if err != nil {
		return err
	}

	s.timeWheel.Start()

	s.wg.Add(1)
	go s.watchConfigChanges()

	return nil
}

func (s *AutoDDNSService) Stop() {
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
}

func (s *AutoDDNSService) loadAllTasks() error {
	configs, err := s.configRepo.ReadAll(context.Background())
	if err != nil {
		return err
	}

	for _, config := range configs {
		s.scheduleConfigTasks(&config)
	}

	return nil
}

func (s *AutoDDNSService) scheduleConfigTasks(config *model.Config) {
	s.Lock()
	defer s.Unlock()

	taskID := utils.TaskID("config_" + config.ID)

	interval := time.Duration(config.Interval) * time.Second

	task := &utils.Task{
		ID:       taskID,
		Interval: interval,
		Handler: func() {
			s.ddnsService.HandleConfig(config.ID)
		},
	}

	s.timeWheel.RemoveTask(taskID)
	s.timeWheel.AddTask(task)
}

func (s *AutoDDNSService) watchConfigChanges() {
	defer s.wg.Done()

	for {
		select {
		case <-s.updateChan:
			s.refreshAllTasks()
		case <-s.quitChan:
			return
		}
	}
}

func (s *AutoDDNSService) refreshAllTasks() {
	s.Lock()
	defer s.Unlock()

	configs, err := s.configRepo.ReadAll(context.Background())
	if err != nil {
		return
	}

	for _, config := range configs {
		s.scheduleConfigTasks(&config)
	}
}

func (s *AutoDDNSService) NotifyConfigChange() {
	select {
	case s.updateChan <- struct{}{}:
	default:
	}
}
