// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package dnsproviders

import (
	"fmt"

	"ez-ddns/v2/model"

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

	addReq := &alidns.AddDomainRecordRequest{
		DomainName: tea.String(domainConfig.DomainName),
		RR:         tea.String(domainConfig.RR),
		Type:       tea.String(recordType),
		Value:      tea.String(ip),
	}

	_, err := p.client.AddDomainRecord(addReq)
	if err != nil {
		return fmt.Errorf("新增DNS记录失败: %v", err)
	}

	return nil
}

func (p *AliDNSProvider) DeleteDomainRecord(recordID string) error {
	deleteReq := &alidns.DeleteDomainRecordRequest{
		RecordId: tea.String(recordID),
	}

	_, err := p.client.DeleteDomainRecord(deleteReq)
	if err != nil {
		return fmt.Errorf("删除DNS记录失败: %v", err)
	}

	return nil
}
