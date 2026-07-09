// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package dnsproviders

import (
	"context"
	"fmt"

	"ez-ddns/model"
)

type DNSRecord struct {
	ID     string
	Domain string
	RR     string
	Type   string
	Value  string
}

type DNSProvider interface {
	GetDomainRecords(ctx context.Context, domainConfig model.DomainConfig) ([]DNSRecord, error)
	AddDomainRecord(ctx context.Context, domainConfig model.DomainConfig, ip string) error
	DeleteDomainRecord(ctx context.Context, recordID string) error
}

type ProviderFactory func(config model.Config) (DNSProvider, error)

var providerRegistry = make(map[model.ProviderType]ProviderFactory)

func RegisterProvider(providerType model.ProviderType, factory ProviderFactory) {
	providerRegistry[providerType] = factory
}

func NewDNSProvider(config model.Config) (DNSProvider, error) {
	factory, exists := providerRegistry[config.Provider]
	if !exists {
		return nil, fmt.Errorf("不支持的DNS服务提供商类型: %s", config.Provider)
	}
	return factory(config)
}

func init() {
	RegisterProvider(model.ProviderTypeAliyun, NewAliDNSProvider)
}
