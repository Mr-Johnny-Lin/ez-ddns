// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package model

type IPType string

const (
	IPTypeIPv4 IPType = "ipv4"
	IPTypeIPv6 IPType = "ipv6"
)

func (t IPType) String() string {
	return string(t)
}

func (t IPType) IsValid() bool {
	return t == IPTypeIPv4 || t == IPTypeIPv6
}

func ParseIPType(s string) IPType {
	switch s {
	case string(IPTypeIPv6):
		return IPTypeIPv6
	default:
		return IPTypeIPv4
	}
}
