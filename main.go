// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package main

import (
	"fmt"
	"time"
)

func main() {
	config, err := LoadAndValidateConfig()
	if err != nil {
		fmt.Printf("错误: %v\n", err)
		return
	}

	if config.LogLevel == "debug" {
		SetLogLevel(LogLevelDebug)
	} else {
		SetLogLevel(LogLevelInfo)
	}

	provider, err := NewDNSProvider(config)
	if err != nil {
		Error("初始化DNS运营商失败: %v", err)
		return
	}

	Info("DDNS 服务启动,轮询间隔: %d 秒, 运营商: %s, 域名数量: %d", config.Interval, config.Provider, len(config.Domains))
	Info("按 Ctrl+C 停止服务")

	for {
		updateChan := make(chan bool)

		for i := range config.Domains {
			go func(domainConfig *DomainConfig) {
				resolver := NewIPResolver(domainConfig.IPType)
				lastIP := domainConfig.LastIP
				updateDNS(provider, domainConfig, resolver)
				updateChan <- lastIP != domainConfig.LastIP
			}(&config.Domains[i])
		}

		needSave := false
		for range config.Domains {
			if <-updateChan {
				needSave = true
			}
		}
		close(updateChan)

		if needSave {
			if err := saveConfigToFile(config); err != nil {
				Error("保存配置失败: %v", err)
			}
		}

		time.Sleep(time.Duration(config.Interval) * time.Second)
	}
}

func updateDNS(provider DNSProvider, domainConfig *DomainConfig, resolver IPResolver) {
	Debug("开始DDNS检查...")

	currentIP, err := resolver.GetPublicIP()
	if err != nil {
		Error("获取公网IP(%s)失败: %v", resolver.GetIPType(), err)
		return
	}
	Debug("当前公网 IP(%s): %s", resolver.GetIPType(), currentIP)

	if checkLocalDNS(currentIP, *domainConfig) {
		Debug("本地DNS检测通过，无需更新")
		return
	}

	records, err := provider.GetDomainRecords(*domainConfig)
	if err != nil {
		Error("查询DNS记录失败: %v", err)
		return
	}

	if recordExists(records, currentIP) {
		Debug("DNS记录已正确指向当前IP，无需更新")
		return
	}

	if updateDomainRecords(provider, *domainConfig, records, currentIP) {
		domainConfig.LastIP = currentIP
	}
}