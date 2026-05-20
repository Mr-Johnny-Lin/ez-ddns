// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package main

import (
	"fmt"
	"net"
)

func checkLocalDNS(currentIP string, domainConfig DomainConfig) bool {
	domain := fmt.Sprintf("%s.%s", domainConfig.RR, domainConfig.DomainName)
	Debug("本地DNS检测: %s", domain)

	ips, err := net.LookupIP(domain)
	if err != nil {
		Debug("本地DNS解析失败: %v", err)
		return false
	}

	for _, ip := range ips {
		if ipStr := ip.String(); ipStr == currentIP {
			return true
		}
	}

	Debug("本地DNS解析结果: %v，与当前公网IP %s 不一致", ips, currentIP)
	return false
}

func recordExists(records []DNSRecord, currentIP string) bool {
	for _, record := range records {
		Debug("检查记录: %s.%s -> %s", record.RR, record.Domain, record.Value)
		if record.Value == currentIP {
			return true
		}
	}
	return false
}

func findRecordsToDelete(records []DNSRecord, lastIP string) []DNSRecord {
	if lastIP != "" {
		for _, record := range records {
			if record.Value == lastIP {
				return []DNSRecord{record}
			}
		}
	}

	if len(records) > 0 {
		Debug("未找到旧IP记录，将删除所有解析记录")
		return records
	}

	return nil
}

func deleteDomainRecords(provider DNSProvider, records []DNSRecord) error {
	var lastErr error
	for _, record := range records {
		Debug("删除记录 %s.%s -> %s...", record.RR, record.Domain, record.Value)

		if err := provider.DeleteDomainRecord(record.ID); err != nil {
			Error("删除记录 %s.%s -> %s 失败: %v", record.RR, record.Domain, record.Value, err)
			lastErr = err
		} else {
			Info("删除记录 %s.%s -> %s 成功！", record.RR, record.Domain, record.Value)
		}
	}
	return lastErr
}

func updateDomainRecords(provider DNSProvider, domainConfig DomainConfig, records []DNSRecord, currentIP string) bool {
	Debug("当前IP不在解析记录中，开始更新...")

	deleteRecords := findRecordsToDelete(records, domainConfig.LastIP)

	if err := deleteDomainRecords(provider, deleteRecords); err != nil {
		Error("删除记录过程出现错误: %v", err)
	}

	if err := provider.AddDomainRecord(domainConfig, currentIP); err != nil {
		Error("添加新记录失败: %v", err)
		return false
	}

	return true
}
