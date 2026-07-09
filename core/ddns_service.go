// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package core

import (
	"context"
	"fmt"
	"net"

	"ez-ddns/core/dnsproviders"
	"ez-ddns/dao"
	"ez-ddns/model"
	"ez-ddns/utils"
)

type DDNSService struct {
	configRepo   dao.ConfigRepository
	domainRepo   dao.DomainRepository
	providerFact dnsproviders.ProviderFactory
}

func NewDDNSService(configRepo dao.ConfigRepository, domainRepo dao.DomainRepository, providerFact dnsproviders.ProviderFactory) *DDNSService {
	return &DDNSService{
		configRepo:   configRepo,
		domainRepo:   domainRepo,
		providerFact: providerFact,
	}
}

func (s *DDNSService) HandleConfig(ctx context.Context, configID string) {
	s.HandleConfigWithContext(ctx, configID)
}

func (s *DDNSService) HandleConfigWithContext(ctx context.Context, configID string) {
	logger := utils.LoggerFromContext(ctx)

	config, err := s.configRepo.ReadByID(ctx, configID)
	if err != nil {
		logger.Error("配置ID[%s] - 读取配置失败: %v", configID, err)
		return
	}
	if config == nil {
		logger.Error("配置ID[%s] - 配置不存在", configID)
		return
	}

	provider, err := s.providerFact(*config)
	if err != nil {
		logger.Error("配置ID[%s] - 初始化DNS服务提供商失败: %v", configID, err)
		return
	}

	s.asyncHandleAllDomains(ctx, *config, provider)
}

func (s *DDNSService) asyncHandleAllDomains(ctx context.Context, config model.Config, provider dnsproviders.DNSProvider) {
	logger := utils.LoggerFromContext(ctx)

	configID := config.ID
	domains, err := s.domainRepo.ReadByConfigID(ctx, configID)
	if err != nil {
		logger.Error("配置ID[%s] - 查询域名记录失败: %v", configID, err)
		return
	}

	if len(domains) == 0 {
		logger.Debug("配置ID[%s] - 未找到域名记录", configID)
		return
	}

	for i := range domains {
		go func(domainConfig model.DomainConfig) {
			s.handleDomain(ctx, configID, domainConfig, provider)
		}(domains[i])
	}
}

func (s *DDNSService) handleDomain(ctx context.Context, configID string, domainConfig model.DomainConfig, provider dnsproviders.DNSProvider) {
	logger := utils.LoggerFromContext(ctx)

	logger.Debug("配置ID[%s] - 开始DNS检测", configID)

	resolver := NewIPResolver(domainConfig.IPType)
	currentIP, err := resolver.GetPublicIP(ctx)
	if err != nil {
		logger.Error("配置ID[%s] - 获取公网IP(%s)失败: %v", configID, resolver.GetIPType(), err)
		return
	}
	logger.Debug("配置ID[%s] - 当前公网 IP(%s): %s", configID, resolver.GetIPType(), currentIP)

	if s.checkLocalDNS(ctx, configID, currentIP, domainConfig) {
		logger.Debug("配置ID[%s] - 本地DNS检测通过，无需更新", configID)
		return
	}

	records, err := provider.GetDomainRecords(ctx, domainConfig)
	if err != nil {
		logger.Error("配置ID[%s] - 查询DNS记录失败: %v", configID, err)
		return
	}

	if s.recordExists(ctx, configID, records, currentIP) {
		logger.Debug("配置ID[%s] - DNS记录已正确指向当前IP，无需更新", configID)
		return
	}

	if s.updateDomainRecords(ctx, configID, provider, domainConfig, records, currentIP) {
		domainConfig.LastIP = currentIP
		if err := s.domainRepo.Update(ctx, &domainConfig); err != nil {
			logger.Error("配置ID[%s] - 更新数据库中LastIP失败: %v", configID, err)
		}
	}
}

func (s *DDNSService) checkLocalDNS(ctx context.Context, configID string, currentIP string, domainConfig model.DomainConfig) bool {
	logger := utils.LoggerFromContext(ctx)

	domain := fmt.Sprintf("%s.%s", domainConfig.RR, domainConfig.DomainName)
	logger.Debug("配置ID[%s] - 本地DNS检测: %s", configID, domain)

	ips, err := net.LookupIP(domain)
	if err != nil {
		logger.Debug("配置ID[%s] - 本地DNS解析失败: %v", configID, err)
		return false
	}

	for _, ip := range ips {
		if ipStr := ip.String(); ipStr == currentIP {
			return true
		}
	}

	logger.Debug("配置ID[%s] - 本地DNS解析结果: %v，与当前公网IP %s 不一致", configID, ips, currentIP)
	return false
}

func (s *DDNSService) recordExists(ctx context.Context, configID string, records []dnsproviders.DNSRecord, currentIP string) bool {
	logger := utils.LoggerFromContext(ctx)

	for _, record := range records {
		logger.Debug("配置ID[%s] - 检查记录: %s.%s -> %s", configID, record.RR, record.Domain, record.Value)
		if record.Value == currentIP {
			return true
		}
	}
	return false
}

func (s *DDNSService) deleteOldIPRecord(ctx context.Context, configID string, provider dnsproviders.DNSProvider, records []dnsproviders.DNSRecord, lastIP string) error {
	if len(records) == 0 {
		return nil
	}

	logger := utils.LoggerFromContext(ctx)

	if lastIP != "" {
		for _, record := range records {
			if record.Value == lastIP {
				logger.Debug("配置ID[%s] - 删除记录 %s.%s -> %s...", configID, record.RR, record.Domain, record.Value)

				if err := provider.DeleteDomainRecord(ctx, record.ID); err != nil {
					logger.Error("配置ID[%s] - 删除记录 %s.%s -> %s 失败: %v", configID, record.RR, record.Domain, record.Value, err)
					return err
				}

				logger.Info("配置ID[%s] - 删除记录 %s.%s -> %s 成功！", configID, record.RR, record.Domain, record.Value)
				return nil
			}
		}
	}

	logger.Debug("配置ID[%s] - 删除所有远程记录重新同步", configID)
	var lastErr error
	for _, record := range records {
		logger.Debug("配置ID[%s] - 删除记录 %s.%s -> %s...", configID, record.RR, record.Domain, record.Value)

		if err := provider.DeleteDomainRecord(ctx, record.ID); err != nil {
			logger.Error("配置ID[%s] - 删除记录 %s.%s -> %s 失败: %v", configID, record.RR, record.Domain, record.Value, err)
			lastErr = err
		} else {
			logger.Info("配置ID[%s] - 删除记录 %s.%s -> %s 成功！", configID, record.RR, record.Domain, record.Value)
		}
	}
	return lastErr
}

func (s *DDNSService) updateDomainRecords(ctx context.Context, configID string, provider dnsproviders.DNSProvider, domainConfig model.DomainConfig, records []dnsproviders.DNSRecord, currentIP string) bool {
	logger := utils.LoggerFromContext(ctx)

	logger.Debug("配置ID[%s] - 当前IP不在解析记录中，开始更新...", configID)

	if err := s.deleteOldIPRecord(ctx, configID, provider, records, domainConfig.LastIP); err != nil {
		logger.Error("配置ID[%s] - 删除记录过程出现错误: %v", configID, err)
	}

	if err := provider.AddDomainRecord(ctx, domainConfig, currentIP); err != nil {
		logger.Error("配置ID[%s] - 添加新记录失败: %v", configID, err)
		return false
	}

	logger.Info("配置ID[%s] - DNS记录更新成功！%s.%s -> %s", configID, domainConfig.RR, domainConfig.DomainName, currentIP)
	return true
}
