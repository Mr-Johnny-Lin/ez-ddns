// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package core

import (
	"context"
	"fmt"
	"net"

	"ez-ddns/v2/dao"
	"ez-ddns/v2/dnsproviders"
	"ez-ddns/v2/model"
	"ez-ddns/v2/utils"
)

type DDNSService struct {
	configRepo   dao.ConfigRepository
	domainRepo   dao.DomainRepository
	logger       utils.Logger
	providerFact dnsproviders.ProviderFactory
}

func NewDDNSService(configRepo dao.ConfigRepository, domainRepo dao.DomainRepository, logger utils.Logger, providerFact dnsproviders.ProviderFactory) *DDNSService {
	return &DDNSService{
		configRepo:   configRepo,
		domainRepo:   domainRepo,
		logger:       logger,
		providerFact: providerFact,
	}
}

func (s *DDNSService) HandleConfig(configID string) {
	s.HandleConfigWithContext(context.Background(), configID)
}

func (s *DDNSService) HandleConfigWithContext(ctx context.Context, configID string) {
	config, err := s.configRepo.ReadByID(ctx, configID)
	if err != nil {
		s.logger.Error("配置ID[%s] - 读取配置失败: %v", configID, err)
		return
	}
	if config == nil {
		s.logger.Error("配置ID[%s] - 配置不存在", configID)
		return
	}

	provider, err := s.providerFact(*config)
	if err != nil {
		s.logger.Error("配置ID[%s] - 初始化DNS服务提供商失败: %v", configID, err)
		return
	}

	s.asyncHandleAllDomains(ctx, *config, provider)
}

func (s *DDNSService) asyncHandleAllDomains(ctx context.Context, config model.Config, provider dnsproviders.DNSProvider) {
	configID := config.ID
	domains, err := s.domainRepo.ReadByConfigID(ctx, configID)
	if err != nil {
		s.logger.Error("配置ID[%s] - 查询域名记录失败: %v", configID, err)
		return
	}

	if len(domains) == 0 {
		s.logger.Debug("配置ID[%s] - 未找到域名记录", configID)
		return
	}

	for i := range domains {
		go func(domainConfig model.DomainConfig) {
			s.handleDomain(ctx, configID, domainConfig, provider)
		}(domains[i])
	}
}

func (s *DDNSService) handleDomain(ctx context.Context, configID string, domainConfig model.DomainConfig, provider dnsproviders.DNSProvider) {
	s.logger.Debug("配置ID[%s] - 开始DNS检测", configID)

	resolver := NewIPResolver(domainConfig.IPType)
	currentIP, err := resolver.GetPublicIP()
	if err != nil {
		s.logger.Error("配置ID[%s] - 获取公网IP(%s)失败: %v", configID, resolver.GetIPType(), err)
		return
	}
	s.logger.Debug("配置ID[%s] - 当前公网 IP(%s): %s", configID, resolver.GetIPType(), currentIP)

	if s.checkLocalDNS(configID, currentIP, domainConfig) {
		s.logger.Debug("配置ID[%s] - 本地DNS检测通过，无需更新", configID)
		return
	}

	records, err := provider.GetDomainRecords(domainConfig)
	if err != nil {
		s.logger.Error("配置ID[%s] - 查询DNS记录失败: %v", configID, err)
		return
	}

	if s.recordExists(configID, records, currentIP) {
		s.logger.Debug("配置ID[%s] - DNS记录已正确指向当前IP，无需更新", configID)
		return
	}

	if s.updateDomainRecords(configID, provider, domainConfig, records, currentIP) {
		domainConfig.LastIP = currentIP
		if err := s.domainRepo.Update(ctx, &domainConfig); err != nil {
			s.logger.Error("配置ID[%s] - 更新数据库中LastIP失败: %v", configID, err)
		}
	}
}

func (s *DDNSService) checkLocalDNS(configID string, currentIP string, domainConfig model.DomainConfig) bool {
	domain := fmt.Sprintf("%s.%s", domainConfig.RR, domainConfig.DomainName)
	s.logger.Debug("配置ID[%s] - 本地DNS检测: %s", configID, domain)

	ips, err := net.LookupIP(domain)
	if err != nil {
		s.logger.Debug("配置ID[%s] - 本地DNS解析失败: %v", configID, err)
		return false
	}

	for _, ip := range ips {
		if ipStr := ip.String(); ipStr == currentIP {
			return true
		}
	}

	s.logger.Debug("配置ID[%s] - 本地DNS解析结果: %v，与当前公网IP %s 不一致", configID, ips, currentIP)
	return false
}

func (s *DDNSService) recordExists(configID string, records []dnsproviders.DNSRecord, currentIP string) bool {
	for _, record := range records {
		s.logger.Debug("配置ID[%s] - 检查记录: %s.%s -> %s", configID, record.RR, record.Domain, record.Value)
		if record.Value == currentIP {
			return true
		}
	}
	return false
}

func (s *DDNSService) findRecordsToDelete(configID string, records []dnsproviders.DNSRecord, lastIP string) []dnsproviders.DNSRecord {
	if lastIP != "" {
		for _, record := range records {
			if record.Value == lastIP {
				return []dnsproviders.DNSRecord{record}
			}
		}
	}

	if len(records) > 0 {
		s.logger.Debug("配置ID[%s] - 未找到旧IP记录，将删除所有解析记录", configID)
		return records
	}

	return nil
}

func (s *DDNSService) deleteDomainRecords(configID string, provider dnsproviders.DNSProvider, records []dnsproviders.DNSRecord) error {
	var lastErr error
	for _, record := range records {
		s.logger.Debug("配置ID[%s] - 删除记录 %s.%s -> %s...", configID, record.RR, record.Domain, record.Value)

		if err := provider.DeleteDomainRecord(record.ID); err != nil {
			s.logger.Error("配置ID[%s] - 删除记录 %s.%s -> %s 失败: %v", configID, record.RR, record.Domain, record.Value, err)
			lastErr = err
		} else {
			s.logger.Info("配置ID[%s] - 删除记录 %s.%s -> %s 成功！", configID, record.RR, record.Domain, record.Value)
		}
	}
	return lastErr
}

func (s *DDNSService) updateDomainRecords(configID string, provider dnsproviders.DNSProvider, domainConfig model.DomainConfig, records []dnsproviders.DNSRecord, currentIP string) bool {
	s.logger.Debug("配置ID[%s] - 当前IP不在解析记录中，开始更新...", configID)

	deleteRecords := s.findRecordsToDelete(configID, records, domainConfig.LastIP)

	if err := s.deleteDomainRecords(configID, provider, deleteRecords); err != nil {
		s.logger.Error("配置ID[%s] - 删除记录过程出现错误: %v", configID, err)
	}

	if err := provider.AddDomainRecord(domainConfig, currentIP); err != nil {
		s.logger.Error("配置ID[%s] - 添加新记录失败: %v", configID, err)
		return false
	}

	s.logger.Info("配置ID[%s] - DNS记录更新成功！%s.%s -> %s", configID, domainConfig.RR, domainConfig.DomainName, currentIP)
	return true
}
