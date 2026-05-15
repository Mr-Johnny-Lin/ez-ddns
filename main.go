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
	"time"

	alidns "github.com/alibabacloud-go/alidns-20150109/v5/client"
	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	"github.com/alibabacloud-go/tea/tea"
)

func getPublicIP() (string, error) {
	services := []string{
		"https://api.ipify.org",
		"https://ifconfig.me/ip",
		"https://icanhazip.com",
		"https://ident.me",
		"https://ipecho.net/plain",
	}

	var lastErr error
	for _, service := range services {
		resp, err := http.Get(service)
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

		if net.ParseIP(ipStr) != nil {
			return ipStr, nil
		}
	}

	return "", fmt.Errorf("所有 IP 查询服务均失败，最后错误: %v", lastErr)
}

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

	Info("DDNS 服务启动,轮询间隔: %d 秒", config.Interval)
	Info("按 Ctrl+C 停止服务")

	clientConfig := &openapi.Config{AccessKeyId: &config.AccessKeyId, AccessKeySecret: &config.AccessKeySecret}
	clientConfig.Endpoint = tea.String("alidns.aliyuncs.com")
	client, err := alidns.NewClient(clientConfig)
	if err != nil {
		Error("初始化客户端失败: %v", err)
		return
	}

	for {
		updateDNS(client, config)
		time.Sleep(time.Duration(config.Interval) * time.Second)
	}
}

func updateDNS(client *alidns.Client, config *Config) {
	Debug("开始检查...")

	currentIP, err := getPublicIP()
	if err != nil {
		Error("获取公网IP失败: %v", err)
		return
	}
	Debug("当前公网 IP: %s", currentIP)

	describeReq := &alidns.DescribeDomainRecordsRequest{
		DomainName: tea.String(config.DomainName),
		RRKeyWord:  tea.String(config.RR),
	}
	describeResp, err := client.DescribeDomainRecords(describeReq)
	if err != nil {
		Error("查询记录失败: %v", err)
		return
	}

	records := describeResp.Body.DomainRecords.Record
	Debug("找到 %d 条解析记录", len(records))

	for _, record := range records {
		Debug("检查记录: %s.%s -> %s", *record.RR, config.DomainName, *record.Value)
		if *record.Value == currentIP {
			Debug("当前 IP 已存在于解析记录中，无需更新。")
			if err := UpdateLastIP(config, currentIP); err != nil {
				Error("保存配置失败: %v", err)
			}
			return
		}
	}

	Debug("当前 IP 不在解析记录中，需要更新...")

	var deleteRecords []*alidns.DescribeDomainRecordsResponseBodyDomainRecordsRecord

	if config.LastIP != "" {
		for _, record := range records {
			if *record.Value == config.LastIP {
				deleteRecords = append(deleteRecords, record)
				break
			}
		}
	}

	if len(deleteRecords) == 0 && len(records) > 0 {
		deleteRecords = records
		Debug("未找到旧 IP 记录，清空所有解析记录")
	}

	for _, record := range deleteRecords {
		Debug("删除记录 %s.%s -> %s...", *record.RR, config.DomainName, *record.Value)
		deleteReq := &alidns.DeleteDomainRecordRequest{
			RecordId: tea.String(*record.RecordId),
		}
		_, err = client.DeleteDomainRecord(deleteReq)
		if err != nil {
			Error("删除记录 %s.%s -> %s 失败: %v", *record.RR, config.DomainName, *record.Value, err)
		} else {
			Info("删除记录 %s.%s -> %s 成功！", *record.RR, config.DomainName, *record.Value)
		}
	}

	Debug("新增解析记录 %s.%s -> %s...", config.RR, config.DomainName, currentIP)
	addReq := &alidns.AddDomainRecordRequest{
		DomainName: tea.String(config.DomainName),
		RR:         tea.String(config.RR),
		Type:       tea.String("A"),
		Value:      tea.String(currentIP),
	}
	_, err = client.AddDomainRecord(addReq)
	if err != nil {
		Error("新增记录 %s.%s -> %s 失败: %v", config.RR, config.DomainName, currentIP, err)
	} else {
		Info("新增记录 %s.%s -> %s 成功！", config.RR, config.DomainName, currentIP)
		if err := UpdateLastIP(config, currentIP); err != nil {
			Error("保存配置失败: %v", err)
		}
	}
}