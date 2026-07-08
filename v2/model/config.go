// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package model

type DomainConfig struct {
	ID         string
	ConfigID   string
	DomainName string
	RR         string
	IPType     IPType
	LastIP     string
}

type Config struct {
	ID              string
	AccessKeyId     string
	AccessKeySecret string
	Provider        ProviderType
	Interval        int
}
