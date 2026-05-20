// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package main

import (
	"fmt"
	"time"

	alidns "github.com/alibabacloud-go/alidns-20150109/v5/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/tea"
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

	ipResolvers := NewIPResolvers(config.IPType)
	
	ipTypes := make([]string, 0, len(ipResolvers))
	for _, resolver := range ipResolvers {
		ipTypes = append(ipTypes, resolver.GetIPType())
	}
	
	Info("DDNS 服务启动,轮询间隔: %d 秒, IP类型: %v", config.Interval, ipTypes)
	Info("按 Ctrl+C 停止服务")

	clientConfig := &openapi.Config{AccessKeyId: &config.AccessKeyId, AccessKeySecret: &config.AccessKeySecret}
	clientConfig.Endpoint = tea.String("alidns.aliyuncs.com")
	client, err := alidns.NewClient(clientConfig)
	if err != nil {
		Error("初始化客户端失败: %v", err)
		return
	}

	for {
		for _, resolver := range ipResolvers {
			updateDNS(client, config, resolver)
		}
		time.Sleep(time.Duration(config.Interval) * time.Second)
	}
}

// updateDNS DDNS主更新流程，协调各步骤执行
func updateDNS(client *alidns.Client, config *Config, resolver IPResolver) {
	Debug("开始DDNS检查...")

	// 步骤1: 获取当前公网IP
	currentIP, err := resolver.GetPublicIP()
	if err != nil {
		Error("获取公网IP(%s)失败: %v", resolver.GetIPType(), err)
		return
	}
	Debug("当前公网 IP(%s): %s", resolver.GetIPType(), currentIP)

	// 步骤2: 本地DNS预检查（减少运营商API调用）
	if checkLocalDNS(currentIP, config) {
		Debug("本地DNS检测通过，无需更新")
		return
	}

	// 步骤3: 查询阿里云DNS记录
	records, err := getDomainRecords(client, config, resolver)
	if err != nil {
		Error("查询阿里云记录失败: %v", err)
		return
	}

	// 步骤4: 检查是否已存在正确记录
	if recordExists(records, currentIP) {
		Debug("阿里云记录已正确指向当前IP，无需更新")
		return
	}

	// 步骤5: 执行更新操作
	updateDomainRecords(client, config, resolver, records, currentIP)
}