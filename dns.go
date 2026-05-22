// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package main

import (
	"fmt"
	"net"

	"ez-ddns/dao"
	"ez-ddns/model"
)

func checkLocalDNS(configID string, currentIP string, domainConfig model.DomainConfig) bool {
	domain := fmt.Sprintf("%s.%s", domainConfig.RR, domainConfig.DomainName)
	Debug("配置ID[%s] - 本地DNS检测: %s", configID, domain)

	ips, err := net.LookupIP(domain)
	if err != nil {
		Debug("配置ID[%s] - 本地DNS解析失败: %v", configID, err)
		return false
	}

	for _, ip := range ips {
		if ipStr := ip.String(); ipStr == currentIP {
			return true
		}
	}

	Debug("配置ID[%s] - 本地DNS解析结果: %v，与当前公网IP %s 不一致", configID, ips, currentIP)
	return false
}

func recordExists(configID string, records []DNSRecord, currentIP string) bool {
	for _, record := range records {
		Debug("配置ID[%s] - 检查记录: %s.%s -> %s", configID, record.RR, record.Domain, record.Value)
		if record.Value == currentIP {
			return true
		}
	}
	return false
}

func findRecordsToDelete(configID string, records []DNSRecord, lastIP string) []DNSRecord {
	if lastIP != "" {
		for _, record := range records {
			if record.Value == lastIP {
				return []DNSRecord{record}
			}
		}
	}

	if len(records) > 0 {
		Debug("配置ID[%s] - 未找到旧IP记录，将删除所有解析记录", configID)
		return records
	}

	return nil
}

func deleteDomainRecords(configID string, provider DNSProvider, records []DNSRecord) error {
	var lastErr error
	for _, record := range records {
		Debug("配置ID[%s] - 删除记录 %s.%s -> %s...", configID, record.RR, record.Domain, record.Value)

		if err := provider.DeleteDomainRecord(record.ID); err != nil {
			Error("配置ID[%s] - 删除记录 %s.%s -> %s 失败: %v", configID, record.RR, record.Domain, record.Value, err)
			lastErr = err
		} else {
			Info("配置ID[%s] - 删除记录 %s.%s -> %s 成功！", configID, record.RR, record.Domain, record.Value)
		}
	}
	return lastErr
}

func updateDomainRecords(configID string, provider DNSProvider, domainConfig model.DomainConfig, records []DNSRecord, currentIP string) bool {
	Debug("配置ID[%s] - 当前IP不在解析记录中，开始更新...", configID)

	deleteRecords := findRecordsToDelete(configID, records, domainConfig.LastIP)

	if err := deleteDomainRecords(configID, provider, deleteRecords); err != nil {
		Error("配置ID[%s] - 删除记录过程出现错误: %v", configID, err)
	}

	if err := provider.AddDomainRecord(domainConfig, currentIP); err != nil {
		Error("配置ID[%s] - 添加新记录失败: %v", configID, err)
		return false
	}

	Info("配置ID[%s] - DNS记录更新成功！%s.%s -> %s", configID, domainConfig.RR, domainConfig.DomainName, currentIP)
	return true
}

func handleConfig(configID string) {
	config, err := dao.Config.ReadByID(configID)
	if err != nil {
		Error("配置ID[%s] - 读取配置失败: %v", configID, err)
		return
	}
	if config == nil {
		Error("配置ID[%s] - 配置不存在", configID)
		return
	}

	provider, err := NewDNSProvider(*config)
	if err != nil {
		Error("配置ID[%s] - 初始化DNS提供商失败: %v", configID, err)
		return
	}

	asyncHandleAllDomains(*config, provider)
}

func asyncHandleAllDomains(config model.Config, provider DNSProvider) {
	configID := config.ID
	domains, err := dao.Domain.ReadByConfigID(configID)
	if err != nil {
		Error("配置ID[%s] - 查询域名记录失败: %v", configID, err)
		return
	}

	if len(domains) == 0 {
		Debug("配置ID[%s] - 未找到域名记录", configID)
		return
	}

	for i := range domains {
		go func(domainConfig model.DomainConfig) {
			resolver := NewIPResolver(domainConfig.IPType)
			handleDomain(configID, domainConfig, resolver, provider)
		}(domains[i])
	}
}

func handleDomain(configID string, domainConfig model.DomainConfig, resolver IPResolver, provider DNSProvider) {
	Debug("配置ID[%s] - 开始DNS检测..", configID)

	currentIP, err := resolver.GetPublicIP()
	if err != nil {
		Error("配置ID[%s] - 获取公网IP(%s)失败: %v", configID, resolver.GetIPType(), err)
		return
	}
	Debug("配置ID[%s] - 当前公网 IP(%s): %s", configID, resolver.GetIPType(), currentIP)

	if checkLocalDNS(configID, currentIP, domainConfig) {
		Debug("配置ID[%s] - 本地DNS检测通过，无需更新", configID)
		return
	}

	records, err := provider.GetDomainRecords(domainConfig)
	if err != nil {
		Error("配置ID[%s] - 查询DNS记录失败: %v", configID, err)
		return
	}

	if recordExists(configID, records, currentIP) {
		Debug("配置ID[%s] - DNS记录已正确指向当前IP，无需更新", configID)
		return
	}

	if updateDomainRecords(configID, provider, domainConfig, records, currentIP) {
		domainConfig.LastIP = currentIP
		if err := dao.Domain.Update(&domainConfig); err != nil {
			Error("配置ID[%s] - 更新数据库中LastIP失败: %v", configID, err)
		}
	}
}