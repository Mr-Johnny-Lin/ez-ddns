// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package dto

type ConfigDTO struct {
	ID              string       `json:"ID"`
	AccessKeyId     string       `json:"AccessKeyId"`
	AccessKeySecret string       `json:"AccessKeySecret"`
	Provider        string       `json:"Provider"`
	Interval        int          `json:"Interval"`
	Domains         []*DomainDTO `json:"Domains"`
}

type DomainDTO struct {
	ID         string `json:"ID,omitempty"`
	ConfigID   string `json:"ConfigID,omitempty"`
	DomainName string `json:"DomainName"`
	RR         string `json:"RR"`
	IPType     string `json:"IPType"`
}
