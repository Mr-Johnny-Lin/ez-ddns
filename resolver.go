// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"ez-ddns/model"
)

// IPResolver 定义IP解析策略接口
type IPResolver interface {
	GetPublicIP() (string, error)
	GetIPType() model.IPType
}

// IPv4Resolver IPv4解析策略实现
type IPv4Resolver struct{}

func (r *IPv4Resolver) GetPublicIP() (string, error) {
	services := []string{
		"https://api.ipify.org",
		"https://ipv4.icanhazip.com",
		"https://api.ipify.org/?format=text",
		"https://ip4.seeip.org",
		"https://ipv4.myexternalip.com/raw",
	}

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
		if parsedIP != nil && parsedIP.To4() != nil {
			return ipStr, nil
		}
	}

	return "", fmt.Errorf("所有 IPv4 查询服务均失败，最后错误: %v", lastErr)
}

func (r *IPv4Resolver) GetIPType() model.IPType {
	return model.IPTypeIPv4
}

// IPv6Resolver IPv6解析策略实现
type IPv6Resolver struct{}

func (r *IPv6Resolver) GetPublicIP() (string, error) {
	services := []string{
		"https://api6.ipify.org",
		"https://ipv6.icanhazip.com",
		"https://ip6.seeip.org",
		"https://ipv6.myexternalip.com/raw",
		"https://ifconfig.me/ip",
	}

	client := &http.Client{
		Transport: &http.Transport{
			ForceAttemptHTTP2: false,
			DialContext: (&net.Dialer{
				LocalAddr: &net.TCPAddr{IP: net.IPv6zero, Port: 0},
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
		if parsedIP != nil && parsedIP.To16() != nil {
			return ipStr, nil
		}
	}

	return "", fmt.Errorf("所有 IPv6 查询服务均失败，最后错误: %v", lastErr)
}

func (r *IPv6Resolver) GetIPType() model.IPType {
	return model.IPTypeIPv6
}

// NewIPResolver 根据配置创建IP解析器
func NewIPResolver(ipType model.IPType) IPResolver {
	if ipType == model.IPTypeIPv6 {
		return &IPv6Resolver{}
	}
	return &IPv4Resolver{}
}

// NewIPResolvers 根据配置创建IP解析器列表
func NewIPResolvers(ipType model.IPType) []IPResolver {
	switch ipType {
	case model.IPTypeIPv4:
		return []IPResolver{&IPv4Resolver{}}
	case model.IPTypeIPv6:
		return []IPResolver{&IPv6Resolver{}}
	default:
		return []IPResolver{&IPv4Resolver{}}
	}
}