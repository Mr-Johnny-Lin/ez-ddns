// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package core

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"ez-ddns/model"
	"ez-ddns/utils"
)

type IPResolver interface {
	GetPublicIP(ctx context.Context) (string, error)
	GetIPType() model.IPType
}

type IPv4Resolver struct{}

func (r *IPv4Resolver) GetPublicIP(ctx context.Context) (string, error) {
	logger := utils.LoggerFromContext(ctx)
	logger.Debug("开始获取IPv4公网IP")

	ip, err := getPublicIPFromServices(ctx, []string{
		"https://api.ipify.org",
		"https://ipv4.icanhazip.com",
		"https://api.ipify.org/?format=text",
		"https://ip4.seeip.org",
		"https://ipv4.myexternalip.com/raw",
	})

	if err != nil {
		logger.Error("获取IPv4公网IP失败: %v", err)
	} else {
		logger.Debug("成功获取IPv4公网IP: %s", ip)
	}

	return ip, err
}

func (r *IPv4Resolver) GetIPType() model.IPType {
	return model.IPTypeIPv4
}

type IPv6Resolver struct{}

func (r *IPv6Resolver) GetPublicIP(ctx context.Context) (string, error) {
	logger := utils.LoggerFromContext(ctx)
	logger.Debug("开始获取IPv6公网IP")

	ip, err := getPublicIPFromServices(ctx, []string{
		"https://api6.ipify.org",
		"https://ipv6.icanhazip.com",
		"https://ip6.seeip.org",
		"https://ipv6.myexternalip.com/raw",
		"https://ifconfig.me/ip",
	})

	if err != nil {
		logger.Error("获取IPv6公网IP失败: %v", err)
	} else {
		logger.Debug("成功获取IPv6公网IP: %s", ip)
	}

	return ip, err
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

func getPublicIPFromServices(ctx context.Context, services []string) (string, error) {
	logger := utils.LoggerFromContext(ctx)

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
		logger.Debug("尝试从 %s 获取IP", service)
		resp, err := client.Get(service)
		if err != nil {
			lastErr = err
			logger.Debug("从 %s 获取IP失败: %v", service, err)
			continue
		}
		defer resp.Body.Close()

		ip, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = err
			logger.Debug("读取 %s 响应失败: %v", service, err)
			continue
		}

		ipStr := string(ip)
		ipStr = strings.TrimSpace(ipStr)

		parsedIP := net.ParseIP(ipStr)
		if parsedIP != nil {
			return ipStr, nil
		}

		logger.Debug("%s 返回的IP格式无效: %s", service, ipStr)
	}

	return "", fmt.Errorf("所有 IP 查询服务均失败，最后错误: %v", lastErr)
}
