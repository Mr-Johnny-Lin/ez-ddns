// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package main

import (
	"fmt"

	"ez-ddns/model"

	alidns "github.com/alibabacloud-go/alidns-20150109/v5/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/tea"
)

type DNSRecord struct {
	ID     string
	Domain string
	RR     string
	Type   string
	Value  string
}

type DNSProvider interface {
	GetDomainRecords(domainConfig model.DomainConfig) ([]DNSRecord, error)
	AddDomainRecord(domainConfig model.DomainConfig, ip string) error
	DeleteDomainRecord(recordID string) error
}

type AliDNSProvider struct {
	client *alidns.Client
}

func NewDNSProvider(config model.Config) (DNSProvider, error) {
	switch config.Provider {
	case model.ProviderTypeAliyun:
		return NewAliDNSProvider(config)
	default:
		return nil, fmt.Errorf("不支持的运营商类型: %s", config.Provider)
	}
}

func NewAliDNSProvider(config model.Config) (*AliDNSProvider, error) {
	clientConfig := &openapi.Config{
		AccessKeyId:     &config.AccessKeyId,
		AccessKeySecret: &config.AccessKeySecret,
	}
	clientConfig.Endpoint = tea.String("alidns.aliyuncs.com")

	client, err := alidns.NewClient(clientConfig)
	if err != nil {
		return nil, fmt.Errorf("初始化阿里云DNS客户端失败: %v", err)
	}

	return &AliDNSProvider{client: client}, nil
}

func (p *AliDNSProvider) GetDomainRecords(domainConfig model.DomainConfig) ([]DNSRecord, error) {
	recordType := p.getRecordType(domainConfig.IPType)

	describeReq := &alidns.DescribeDomainRecordsRequest{
		DomainName: tea.String(domainConfig.DomainName),
		RRKeyWord:  tea.String(domainConfig.RR),
		Type:       tea.String(recordType),
	}

	describeResp, err := p.client.DescribeDomainRecords(describeReq)
	if err != nil {
		return nil, err
	}

	records := describeResp.Body.DomainRecords.Record
	result := make([]DNSRecord, 0, len(records))
	for _, record := range records {
		result = append(result, DNSRecord{
			ID:     *record.RecordId,
			Domain: *record.DomainName,
			RR:     *record.RR,
			Type:   *record.Type,
			Value:  *record.Value,
		})
	}

	Debug("远程云解析查询到 %d 条%s类型解析记录", len(result), recordType)
	return result, nil
}

func (p *AliDNSProvider) getRecordType(ipType model.IPType) string {
	if ipType == model.IPTypeIPv6 {
		return "AAAA"
	}
	return "A"
}

func (p *AliDNSProvider) AddDomainRecord(domainConfig model.DomainConfig, ip string) error {
	recordType := p.getRecordType(domainConfig.IPType)

	Debug("新增%s解析记录 %s.%s -> %s...", recordType, domainConfig.RR, domainConfig.DomainName, ip)

	addReq := &alidns.AddDomainRecordRequest{
		DomainName: tea.String(domainConfig.DomainName),
		RR:         tea.String(domainConfig.RR),
		Type:       tea.String(recordType),
		Value:      tea.String(ip),
	}

	_, err := p.client.AddDomainRecord(addReq)
	if err != nil {
		return fmt.Errorf("新增记录失败: %v", err)
	}

	Info("新增%s记录 %s.%s -> %s 成功！", recordType, domainConfig.RR, domainConfig.DomainName, ip)
	return nil
}

func (p *AliDNSProvider) DeleteDomainRecord(recordID string) error {
	deleteReq := &alidns.DeleteDomainRecordRequest{
		RecordId: tea.String(recordID),
	}

	_, err := p.client.DeleteDomainRecord(deleteReq)
	if err != nil {
		return fmt.Errorf("删除记录失败: %v", err)
	}

	return nil
}
