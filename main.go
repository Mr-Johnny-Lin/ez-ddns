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

// getPublicIP 获取当前公网IP地址，通过多个服务轮询确保可靠性
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

// updateDNS DDNS主更新流程，协调各步骤执行
func updateDNS(client *alidns.Client, config *Config) {
	Debug("开始DDNS检查...")

	// 步骤1: 获取当前公网IP
	currentIP, err := getPublicIP()
	if err != nil {
		Error("获取公网IP失败: %v", err)
		return
	}
	Debug("当前公网 IP: %s", currentIP)

	// 步骤2: 本地DNS预检查（减少运营商API调用）
	if checkLocalDNS(currentIP, config) {
		Debug("本地DNS检测通过，无需更新")
		return
	}

	// 步骤3: 查询阿里云DNS记录
	records, err := getDomainRecords(client, config)
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
	updateDomainRecords(client, config, records, currentIP)
}

// checkLocalDNS 本地DNS预检查，通过DNS解析验证当前IP是否已生效
// 返回true表示IP已正确解析，无需继续检查
func checkLocalDNS(currentIP string, config *Config) bool {
	domain := fmt.Sprintf("%s.%s", config.RR, config.DomainName)
	Debug("本地DNS检测: %s", domain)

	ips, err := net.LookupIP(domain)
	if err != nil {
		Debug("本地DNS解析失败: %v", err)
		return false
	}

	for _, ip := range ips {
		if ipStr := ip.String(); ipStr == currentIP {
			return true
		}
	}

	Debug("本地DNS解析结果: %v，与当前公网IP %s 不一致", ips, currentIP)
	return false
}

// getDomainRecords 查询阿里云域名解析记录
func getDomainRecords(client *alidns.Client, config *Config) ([]*alidns.DescribeDomainRecordsResponseBodyDomainRecordsRecord, error) {
	describeReq := &alidns.DescribeDomainRecordsRequest{
		DomainName: tea.String(config.DomainName),
		RRKeyWord:  tea.String(config.RR),
	}

	describeResp, err := client.DescribeDomainRecords(describeReq)
	if err != nil {
		return nil, err
	}

	records := describeResp.Body.DomainRecords.Record
	Debug("远程云解析查询到 %d 条解析记录", len(records))

	return records, nil
}

// recordExists 检查记录列表中是否已存在指定IP的记录
func recordExists(records []*alidns.DescribeDomainRecordsResponseBodyDomainRecordsRecord, currentIP string) bool {
	for _, record := range records {
		Debug("检查记录: %s.%s -> %s", *record.RR, *record.DomainName, *record.Value)
		if *record.Value == currentIP {
			return true
		}
	}
	return false
}

// updateDomainRecords 执行域名记录更新（删除旧记录 + 添加新记录）
func updateDomainRecords(client *alidns.Client, config *Config, records []*alidns.DescribeDomainRecordsResponseBodyDomainRecordsRecord, currentIP string) {
	Debug("当前IP不在解析记录中，开始更新...")

	// 步骤1: 确定需要删除的记录
	deleteRecords := findRecordsToDelete(records, config.LastIP)

	// 步骤2: 删除旧记录
	if err := deleteDomainRecords(client, config, deleteRecords); err != nil {
		Error("删除记录过程出现错误: %v", err)
	}

	// 步骤3: 添加新记录
	if err := addDomainRecord(client, config, currentIP); err != nil {
		Error("添加新记录失败: %v", err)
	}
}

// findRecordsToDelete 查找需要删除的记录
// 优先删除与LastIP匹配的记录，若无则删除所有记录
func findRecordsToDelete(records []*alidns.DescribeDomainRecordsResponseBodyDomainRecordsRecord, lastIP string) []*alidns.DescribeDomainRecordsResponseBodyDomainRecordsRecord {
	if lastIP != "" {
		for _, record := range records {
			if *record.Value == lastIP {
				return []*alidns.DescribeDomainRecordsResponseBodyDomainRecordsRecord{record}
			}
		}
	}

	if len(records) > 0 {
		Debug("未找到旧IP记录，将删除所有解析记录")
		return records
	}

	return nil
}

// deleteDomainRecords 批量删除域名记录
func deleteDomainRecords(client *alidns.Client, config *Config, records []*alidns.DescribeDomainRecordsResponseBodyDomainRecordsRecord) error {
	var lastErr error
	for _, record := range records {
		Debug("删除记录 %s.%s -> %s...", *record.RR, config.DomainName, *record.Value)

		deleteReq := &alidns.DeleteDomainRecordRequest{
			RecordId: tea.String(*record.RecordId),
		}

		_, err := client.DeleteDomainRecord(deleteReq)
		if err != nil {
			Error("删除记录 %s.%s -> %s 失败: %v", *record.RR, config.DomainName, *record.Value, err)
			lastErr = err
		} else {
			Info("删除记录 %s.%s -> %s 成功！", *record.RR, config.DomainName, *record.Value)
		}
	}
	return lastErr
}

// addDomainRecord 添加新的A记录到阿里云DNS
func addDomainRecord(client *alidns.Client, config *Config, currentIP string) error {
	Debug("新增解析记录 %s.%s -> %s...", config.RR, config.DomainName, currentIP)

	addReq := &alidns.AddDomainRecordRequest{
		DomainName: tea.String(config.DomainName),
		RR:         tea.String(config.RR),
		Type:       tea.String("A"),
		Value:      tea.String(currentIP),
	}

	_, err := client.AddDomainRecord(addReq)
	if err != nil {
		return fmt.Errorf("新增记录失败: %v", err)
	}

	Info("新增记录 %s.%s -> %s 成功！", config.RR, config.DomainName, currentIP)

	// 更新配置中的LastIP
	if err := UpdateLastIP(config, currentIP); err != nil {
		Error("保存配置失败: %v", err)
	}

	return nil
}
