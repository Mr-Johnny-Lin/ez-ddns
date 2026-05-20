
// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package main

import (
	"fmt"
	"net"

	alidns "github.com/alibabacloud-go/alidns-20150109/v5/client"
	"github.com/alibabacloud-go/tea/tea"
)

// checkLocalDNS 本地DNS预检查
func checkLocalDNS(currentIP string, config *Config) bool {
	domain := fmt.Sprintf("%s.%s", config.RR, config.DomainName)
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

// getDomainRecords 查询阿里云域名解析记录
func getDomainRecords(client *alidns.Client, config *Config, resolver IPResolver) ([]*alidns.DescribeDomainRecordsResponseBodyDomainRecordsRecord, error) {
	recordType := resolver.GetRecordType()

	describeReq := &alidns.DescribeDomainRecordsRequest{
		DomainName: tea.String(config.DomainName),
		RRKeyWord:  tea.String(config.RR),
		Type:       tea.String(recordType),
	}

	describeResp, err := client.DescribeDomainRecords(describeReq)
	if err != nil {
		return nil, err
	}

	records := describeResp.Body.DomainRecords.Record
	Debug("远程云解析查询到 %d 条%s类型解析记录", len(records), recordType)

	return records, nil
}

// recordExists 检查记录列表中是否已存在指定IP的记录
func recordExists(records []*alidns.DescribeDomainRecordsResponseBodyDomainRecordsRecord, currentIP string) bool {
	for _, record := range records {
		Debug("检查记录: %s.%s -> %s", *record.RR, *record.DomainName, *record.Value)
		if *record.Value == currentIP {
			return true
		}
	}
	return false
}

// updateDomainRecords 执行域名记录更新
func updateDomainRecords(client *alidns.Client, config *Config, resolver IPResolver, records []*alidns.DescribeDomainRecordsResponseBodyDomainRecordsRecord, currentIP string) {
	Debug("当前IP不在解析记录中，开始更新...")

	deleteRecords := findRecordsToDelete(records, config.LastIP)

	if err := deleteDomainRecords(client, config, deleteRecords); err != nil {
		Error("删除记录过程出现错误: %v", err)
	}

	if err := addDomainRecord(client, config, resolver, currentIP); err != nil {
		Error("添加新记录失败: %v", err)
	}
}

// findRecordsToDelete 查找需要删除的记录
func findRecordsToDelete(records []*alidns.DescribeDomainRecordsResponseBodyDomainRecordsRecord, lastIP string) []*alidns.DescribeDomainRecordsResponseBodyDomainRecordsRecord {
	if lastIP != "" {
		for _, record := range records {
			if *record.Value == lastIP {
				return []*alidns.DescribeDomainRecordsResponseBodyDomainRecordsRecord{record}
			}
		}
	}

	if len(records) > 0 {
		Debug("未找到旧IP记录，将删除所有解析记录")
		return records
	}

	return nil
}

// deleteDomainRecords 批量删除域名记录
func deleteDomainRecords(client *alidns.Client, config *Config, records []*alidns.DescribeDomainRecordsResponseBodyDomainRecordsRecord) error {
	var lastErr error
	for _, record := range records {
		Debug("删除记录 %s.%s -> %s...", *record.RR, config.DomainName, *record.Value)

		deleteReq := &alidns.DeleteDomainRecordRequest{
			RecordId: tea.String(*record.RecordId),
		}

		_, err := client.DeleteDomainRecord(deleteReq)
		if err != nil {
			Error("删除记录 %s.%s -> %s 失败: %v", *record.RR, config.DomainName, *record.Value, err)
			lastErr = err
		} else {
			Info("删除记录 %s.%s -> %s 成功！", *record.RR, config.DomainName, *record.Value)
		}
	}
	return lastErr
}

// addDomainRecord 添加新的DNS记录到阿里云DNS
func addDomainRecord(client *alidns.Client, config *Config, resolver IPResolver, currentIP string) error {
	recordType := resolver.GetRecordType()

	Debug("新增%s解析记录 %s.%s -> %s...", recordType, config.RR, config.DomainName, currentIP)

	addReq := &alidns.AddDomainRecordRequest{
		DomainName: tea.String(config.DomainName),
		RR:         tea.String(config.RR),
		Type:       tea.String(recordType),
		Value:      tea.String(currentIP),
	}

	_, err := client.AddDomainRecord(addReq)
	if err != nil {
		return fmt.Errorf("新增记录失败: %v", err)
	}

	Info("新增%s记录 %s.%s -> %s 成功！", recordType, config.RR, config.DomainName, currentIP)

	if err := UpdateLastIP(config, currentIP); err != nil {
		Error("保存配置失败: %v", err)
	}

	return nil
}