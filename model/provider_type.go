// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package model

type ProviderType string

const (
	ProviderTypeAliyun ProviderType = "aliyun"
)

func (t ProviderType) String() string {
	return string(t)
}

func (t ProviderType) IsValid() bool {
	return t == ProviderTypeAliyun
}

func ParseProviderType(s string) ProviderType {
	switch s {
	case string(ProviderTypeAliyun):
		return ProviderTypeAliyun
	default:
		return ProviderTypeAliyun
	}
}
