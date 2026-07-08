// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package core

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"ez-ddns/v2/model"
)

type IPResolver interface {
	GetPublicIP() (string, error)
	GetIPType() model.IPType
}

type IPv4Resolver struct{}

func (r *IPv4Resolver) GetPublicIP() (string, error) {
	return getPublicIPFromServices([]string{
		"https://api.ipify.org",
		"https://ipv4.icanhazip.com",
		"https://api.ipify.org/?format=text",
		"https://ip4.seeip.org",
		"https://ipv4.myexternalip.com/raw",
	})
}

func (r *IPv4Resolver) GetIPType() model.IPType {
	return model.IPTypeIPv4
}

type IPv6Resolver struct{}

func (r *IPv6Resolver) GetPublicIP() (string, error) {
	return getPublicIPFromServices([]string{
		"https://api6.ipify.org",
		"https://ipv6.icanhazip.com",
		"https://ip6.seeip.org",
		"https://ipv6.myexternalip.com/raw",
		"https://ifconfig.me/ip",
	})
}

func (r *IPv6Resolver) GetIPType() model.IPType {
	return model.IPTypeIPv6
}

func NewIPResolver(ipType model.IPType) IPResolver {
	if ipType == model.IPTypeIPv6 {
		return &IPv6Resolver{}
	}
	return &IPv4Resolver{}
}

func getPublicIPFromServices(services []string) (string, error) {
	client := &http.Client{
		Transport: &http.Transport{
			ForceAttemptHTTP2: false,
			DialContext: (&net.Dialer{
				LocalAddr: &net.TCPAddr{IP: net.IPv4zero, Port: 0},
			}).DialContext,
		},
	}

	var lastErr error
	for _, service := range services {
		resp, err := client.Get(service)
		if err != nil {
			lastErr = err
			continue
		}
		defer resp.Body.Close()

		ip, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = err
			continue
		}

		ipStr := string(ip)
		ipStr = strings.TrimSpace(ipStr)

		parsedIP := net.ParseIP(ipStr)
		if parsedIP != nil {
			return ipStr, nil
		}
	}

	return "", fmt.Errorf("所有 IP 查询服务均失败，最后错误: %v", lastErr)
}
