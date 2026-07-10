// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package core

import (
	"context"
	"fmt"
	"net"
	"strings"

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

	ipMap, err := s.resolveIPs(ctx, domains)
	if err != nil {
		logger.Error("配置ID[%s] - 获取公网IP失败: %v", configID, err)
		return
	}

	for i := range domains {
		go func(domainConfig model.DomainConfig) {
			s.handleDomain(ctx, configID, domainConfig, provider, ipMap)
		}(domains[i])
	}
}

func (s *DDNSService) resolveIPs(ctx context.Context, domains []model.DomainConfig) (map[model.IPType]string, error) {
	logger := utils.LoggerFromContext(ctx)

	ipTypes := make(map[model.IPType]bool)
	for _, domain := range domains {
		ipTypes[domain.IPType] = true
	}

	ipMap := make(map[model.IPType]string)
	for ipType := range ipTypes {
		resolver := NewIPResolver(ipType)
		ip, err := resolver.GetPublicIP(ctx)
		if err != nil {
			logger.Error("获取公网IP(%s)失败: %v", ipType, err)
			return nil, err
		}
		ipMap[ipType] = ip
		logger.Debug("成功获取公网IP(%s): %s", ipType, ip)
	}

	return ipMap, nil
}

func (s *DDNSService) handleDomain(ctx context.Context, configID string, domainConfig model.DomainConfig, provider dnsproviders.DNSProvider, ipMap map[model.IPType]string) {
	logger := utils.LoggerFromContext(ctx)
	domain := fmt.Sprintf("%s.%s", domainConfig.RR, domainConfig.DomainName)

	logger.Debug("配置ID[%s] 域名[%s] - 开始DNS检测", configID, domain)

	currentIP, ok := ipMap[domainConfig.IPType]
	if !ok {
		logger.Error("配置ID[%s] 域名[%s] - 未找到对应的公网IP(%s)", configID, domain, domainConfig.IPType)
		return
	}
	logger.Debug("配置ID[%s] 域名[%s] - 当前公网 IP(%s): %s", configID, domain, domainConfig.IPType, currentIP)

	if s.checkLocalDNS(ctx, configID, currentIP, domainConfig) {
		logger.Debug("配置ID[%s] 域名[%s] - 本地DNS检测通过，无需更新", configID, domain)
		return
	}

	records, err := provider.GetDomainRecords(ctx, domainConfig)
	if err != nil {
		logger.Error("配置ID[%s] 域名[%s] - 查询DNS记录失败: %v", configID, domain, err)
		return
	}

	if s.recordExists(ctx, configID, records, currentIP) {
		logger.Debug("配置ID[%s] 域名[%s] - DNS记录已正确指向当前IP，无需更新", configID, domain)
		return
	}

	if s.updateDomainRecords(ctx, configID, provider, domainConfig, records, currentIP) {
		domainConfig.LastIP = currentIP
		if err := s.domainRepo.Update(ctx, &domainConfig); err != nil {
			logger.Error("配置ID[%s] 域名[%s] - 更新数据库中LastIP失败: %v", configID, domain, err)
		}
	}
}

func (s *DDNSService) checkLocalDNS(ctx context.Context, configID string, currentIP string, domainConfig model.DomainConfig) bool {
	logger := utils.LoggerFromContext(ctx)

	originalDomain := fmt.Sprintf("%s.%s", domainConfig.RR, domainConfig.DomainName)
	logger.Debug("配置ID[%s] 域名[%s] - 本地DNS检测", configID, originalDomain)

	rr := domainConfig.RR
	if strings.Contains(rr, "*") {
		rr = strings.ReplaceAll(rr, "*", "test")
	}
	resolveDomain := fmt.Sprintf("%s.%s", rr, domainConfig.DomainName)

	ips, err := net.LookupIP(resolveDomain)
	if err != nil {
		logger.Debug("配置ID[%s] 域名[%s] - 本地DNS解析失败: %v", configID, originalDomain, err)
		return false
	}

	for _, ip := range ips {
		if ipStr := ip.String(); ipStr == currentIP {
			return true
		}
	}

	logger.Debug("配置ID[%s] 域名[%s] - 本地DNS解析结果: %v，与当前公网IP %s 不一致", configID, originalDomain, ips, currentIP)
	return false
}

func (s *DDNSService) recordExists(ctx context.Context, configID string, records []dnsproviders.DNSRecord, currentIP string) bool {
	logger := utils.LoggerFromContext(ctx)

	for _, record := range records {
		domain := fmt.Sprintf("%s.%s", record.RR, record.Domain)
		logger.Debug("配置ID[%s] 域名[%s] - 检查记录 -> %s", configID, domain, record.Value)
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
			domain := fmt.Sprintf("%s.%s", record.RR, record.Domain)
			if record.Value == lastIP {
				logger.Debug("配置ID[%s] 域名[%s] - 删除记录 -> %s...", configID, domain, record.Value)

				if err := provider.DeleteDomainRecord(ctx, record.ID); err != nil {
					logger.Error("配置ID[%s] 域名[%s] - 删除记录 -> %s 失败: %v", configID, domain, record.Value, err)
					return err
				}

				logger.Info("配置ID[%s] 域名[%s] - 删除记录 -> %s 成功！", configID, domain, record.Value)
				return nil
			}
		}
	}

	var firstDomain string
	if len(records) > 0 {
		firstDomain = fmt.Sprintf("%s.%s", records[0].RR, records[0].Domain)
	}
	logger.Debug("配置ID[%s] 域名[%s] - 删除所有远程记录重新同步", configID, firstDomain)
	var lastErr error
	for _, record := range records {
		domain := fmt.Sprintf("%s.%s", record.RR, record.Domain)
		logger.Debug("配置ID[%s] 域名[%s] - 删除记录 -> %s...", configID, domain, record.Value)

		if err := provider.DeleteDomainRecord(ctx, record.ID); err != nil {
			logger.Error("配置ID[%s] 域名[%s] - 删除记录 -> %s 失败: %v", configID, domain, record.Value, err)
			lastErr = err
		} else {
			logger.Info("配置ID[%s] 域名[%s] - 删除记录 -> %s 成功！", configID, domain, record.Value)
		}
	}
	return lastErr
}

func (s *DDNSService) updateDomainRecords(ctx context.Context, configID string, provider dnsproviders.DNSProvider, domainConfig model.DomainConfig, records []dnsproviders.DNSRecord, currentIP string) bool {
	logger := utils.LoggerFromContext(ctx)
	domain := fmt.Sprintf("%s.%s", domainConfig.RR, domainConfig.DomainName)

	logger.Debug("配置ID[%s] 域名[%s] - 当前IP不在解析记录中，开始更新...", configID, domain)

	if err := s.deleteOldIPRecord(ctx, configID, provider, records, domainConfig.LastIP); err != nil {
		logger.Error("配置ID[%s] 域名[%s] - 删除记录过程出现错误: %v", configID, domain, err)
	}

	if err := provider.AddDomainRecord(ctx, domainConfig, currentIP); err != nil {
		logger.Error("配置ID[%s] 域名[%s] - 添加新记录失败: %v", configID, domain, err)
		return false
	}

	logger.Info("配置ID[%s] 域名[%s] - DNS记录更新成功！-> %s", configID, domain, currentIP)
	return true
}
