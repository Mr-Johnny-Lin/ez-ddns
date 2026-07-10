// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package dnsproviders

import (
	"context"
	"fmt"

	"ez-ddns/model"
	"ez-ddns/utils"

	alidns "github.com/alibabacloud-go/alidns-20150109/v5/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/tea"
)

type AliDNSProvider struct {
	client *alidns.Client
}

func NewAliDNSProvider(config model.Config) (DNSProvider, error) {
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

func (p *AliDNSProvider) GetDomainRecords(ctx context.Context, domainConfig model.DomainConfig) ([]DNSRecord, error) {
	logger := utils.LoggerFromContext(ctx)
	domain := fmt.Sprintf("%s.%s", domainConfig.RR, domainConfig.DomainName)

	recordType := p.getRecordType(domainConfig.IPType)

	describeReq := &alidns.DescribeDomainRecordsRequest{
		DomainName: tea.String(domainConfig.DomainName),
		KeyWord:    tea.String(domainConfig.RR),
		Type:       tea.String(recordType),
		SearchMode: tea.String("EXACT"),
	}

	describeResp, err := p.client.DescribeDomainRecords(describeReq)
	if err != nil {
		logger.Error("服务提供商[阿里云] 域名[%s] - 查询域名记录失败: %v", domain, err)
		return nil, err
	}

	records := describeResp.Body.DomainRecords.Record
	result := make([]DNSRecord, 0, len(records))
	var ipMapping string
	for _, record := range records {
		result = append(result, DNSRecord{
			ID:     *record.RecordId,
			Domain: *record.DomainName,
			RR:     *record.RR,
			Type:   *record.Type,
			Value:  *record.Value,
		})
		if ipMapping != "" {
			ipMapping += ","
		}
		ipMapping += *record.Value
	}

	logger.Debug("服务提供商[阿里云] 域名[%s] - 查询到 %d 条记录 [%s]", domain, len(result), ipMapping)
	return result, nil
}

func (p *AliDNSProvider) getRecordType(ipType model.IPType) string {
	if ipType == model.IPTypeIPv6 {
		return "AAAA"
	}
	return "A"
}

func (p *AliDNSProvider) AddDomainRecord(ctx context.Context, domainConfig model.DomainConfig, ip string) error {
	logger := utils.LoggerFromContext(ctx)
	domain := fmt.Sprintf("%s.%s", domainConfig.RR, domainConfig.DomainName)

	recordType := p.getRecordType(domainConfig.IPType)

	addReq := &alidns.AddDomainRecordRequest{
		DomainName: tea.String(domainConfig.DomainName),
		RR:         tea.String(domainConfig.RR),
		Type:       tea.String(recordType),
		Value:      tea.String(ip),
	}

	_, err := p.client.AddDomainRecord(addReq)
	if err != nil {
		logger.Error("服务提供商[阿里云] 域名[%s] - 添加DNS记录失败: %v", domain, err)
		return fmt.Errorf("新增DNS记录失败: %v", err)
	}

	logger.Debug("服务提供商[阿里云] 域名[%s] - 添加DNS记录成功 -> %s", domain, ip)
	return nil
}

func (p *AliDNSProvider) DeleteDomainRecord(ctx context.Context, recordID string) error {
	logger := utils.LoggerFromContext(ctx)

	deleteReq := &alidns.DeleteDomainRecordRequest{
		RecordId: tea.String(recordID),
	}

	_, err := p.client.DeleteDomainRecord(deleteReq)
	if err != nil {
		logger.Error("服务提供商[阿里云] - 删除DNS记录(%s)失败: %v", recordID, err)
		return fmt.Errorf("删除DNS记录失败: %v", err)
	}

	logger.Debug("服务提供商[阿里云] - 删除DNS记录(%s)成功", recordID)
	return nil
}
