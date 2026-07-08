// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"

	"ez-ddns/v2/core"
	"ez-ddns/v2/dao/cache"
	"ez-ddns/v2/dao/caching"
	"ez-ddns/v2/dao/db"
	"ez-ddns/v2/dnsproviders"
	"ez-ddns/v2/model"
	"ez-ddns/v2/utils"
	"ez-ddns/v2/web"
	"ez-ddns/v2/web/service"
)

var baseURL = "http://localhost:8080"
var apiKey = ""

func SetServer(server string) {
	baseURL = server
}

func SetAPIKey(key string) {
	apiKey = key
}

func PrintHelp() {
	fmt.Println(`ez-ddns - 动态DNS更新服务

使用方法: ez-ddns [命令] [参数]

命令列表:
  start     启动DDNS HTTP服务
  config    管理DNS提供商配置
  domain    管理域名记录
  system    管理系统配置
  ddns      DDNS操作（触发、启动、停止）
  help      显示此帮助信息

-------------------------------------------------------------------------------
start 命令
-------------------------------------------------------------------------------
用途: 启动DDNS服务端

参数:
  -port       HTTP服务端口（默认: 8080）
  -api-key    API密钥，用于启用认证保护
  -auto-ddns  启动时自动开启定时DDNS更新

示例:
  ez-ddns start
  ez-ddns start -port 8080 -api-key my-secret-key
  ez-ddns start -port 8080 -auto-ddns

-------------------------------------------------------------------------------
config 命令
-------------------------------------------------------------------------------
用途: 管理DNS提供商配置（如阿里云AccessKey）

子命令:
  create    创建新配置
  get       根据ID获取单个配置
  list      获取所有配置列表
  update    更新指定配置
  delete    删除指定配置

参数:
  -f string   指定配置文件路径（create/update子命令必需）

示例:
  ez-ddns config create -f config.json      # 创建配置
  ez-ddns config get my-config-id           # 获取单个配置
  ez-ddns config list                       # 获取所有配置
  ez-ddns config update my-config-id -f new-config.json  # 更新配置
  ez-ddns config delete my-config-id        # 删除配置

配置文件格式 (config.json):
{
  "ID": "my-aliyun",
  "AccessKeyId": "your-key",
  "AccessKeySecret": "secret",
  "Provider": "aliyun",
  "Interval": 300,
  "Domains": []
}
配置项说明:
  ID              - 配置唯一标识
  AccessKeyId     - DNS服务商的AccessKey ID
  AccessKeySecret - DNS服务商的AccessKey Secret
  Provider        - 服务商类型 (aliyun)
  Interval        - DDNS更新间隔（秒）
  Domains         - 关联的域名列表

-------------------------------------------------------------------------------
domain 命令
-------------------------------------------------------------------------------
用途: 管理域名记录（需关联到某个config）

子命令:
  create    创建新域名记录
  list      获取指定配置下的所有域名
  update    更新指定域名记录
  delete    删除指定域名记录

参数:
  -f string   指定域名配置文件路径（create/update子命令必需）

示例:
  ez-ddns domain create -f domain.json      # 创建域名
  ez-ddns domain list config-id             # 列出配置下的域名
  ez-ddns domain update domain-id -f new-domain.json  # 更新域名
  ez-ddns domain delete domain-id           # 删除域名

域名配置文件格式 (domain.json):
{
  "ID": "domain-001",
  "DomainName": "example.com",
  "RR": "@",
  "IPType": "ipv4",
  "LastIP": "192.168.1.100"
}
配置项说明:
  ID         - 域名记录唯一标识
  DomainName - 域名（如 example.com）
  RR         - 主机记录（@表示根域名，www表示二级域名）
  IPType     - IP类型 (ipv4/ipv6)
  LastIP     - 最后更新的IP地址

-------------------------------------------------------------------------------
system 命令
-------------------------------------------------------------------------------
用途: 管理系统级配置

子命令:
  get       获取系统配置
  update    更新系统配置

参数:
  -f string   指定系统配置文件路径（update子命令必需）

示例:
  ez-ddns system get                        # 获取系统配置
  ez-ddns system update -f system.json      # 更新系统配置

系统配置文件格式 (system.json):
{
  "LogLevel": "info"
}
配置项说明:
  LogLevel - 日志级别 (debug/info/error)

-------------------------------------------------------------------------------
ddns 命令
-------------------------------------------------------------------------------
用途: 手动触发或控制自动DDNS更新

子命令:
  trigger   立即触发指定配置的DDNS更新
  start     启动自动DDNS定时更新服务
  stop      停止自动DDNS定时更新服务

示例:
  ez-ddns ddns trigger config-id            # 触发DDNS更新
  ez-ddns ddns start                        # 启动自动DDNS服务
  ez-ddns ddns stop                         # 停止自动DDNS服务

-------------------------------------------------------------------------------
远程操作参数
-------------------------------------------------------------------------------
适用于: config / domain / system / ddns

  -server string   远程服务器地址（默认: http://localhost:8080）
  -api-key string  服务器API密钥（如果服务器启用了认证）

示例:
  ez-ddns config list -server http://192.168.1.100:8080
  ez-ddns domain list my-config -api-key secret123
  ez-ddns config create -f config.json -server http://remote:8080

-------------------------------------------------------------------------------
完整使用示例
-------------------------------------------------------------------------------
1. 启动服务（带自动DDNS和认证）:
   ez-ddns start -port 8080 -auto-ddns -api-key secure123

2. 创建DNS配置:
   ez-ddns config create -f config.json

3. 添加域名:
   ez-ddns domain create -f domain.json

4. 更新域名配置:
   ez-ddns domain update domain-id -f updated-domain.json

5. 手动触发DDNS更新:
   ez-ddns ddns trigger my-config-id

6. 远程管理:
   ez-ddns config list -server http://remote:8080 -api-key xxx
`)
}

func StartServer(port, apiKey string, autoDDNS bool) {
	fmt.Printf("启动DDNS服务，端口: %s...\n", port)

	sqliteDB, err := db.NewSQLiteDB("ez-ddns.db")
	if err != nil {
		fmt.Printf("初始化数据库失败: %v\n", err)
		os.Exit(1)
	}

	configDB := db.NewConfigDB(sqliteDB)
	domainDB := db.NewDomainDB(sqliteDB)
	systemDB := db.NewSystemDB(sqliteDB)

	configCache := cache.NewConfigCache()
	domainCache := cache.NewDomainCache()
	systemCache := cache.NewSystemCache()

	systemCaching := caching.NewSystemCaching(systemCache, systemDB)

	systemConfig, err := systemCaching.Read(context.Background())
	if err != nil {
		fmt.Printf("读取系统配置失败: %v\n", err)
		os.Exit(1)
	}

	logger := utils.NewLoggerWithProvider(func() utils.LogLevel {
		config, err := systemCaching.Read(context.Background())
		if err != nil {
			return utils.LogLevelInfo
		}
		return utils.ParseLogLevel(config.LogLevel)
	})
	fmt.Printf("日志级别: %s\n", systemConfig.LogLevel)

	var autoDDNSService *core.AutoDDNSService

	configCaching := caching.NewConfigCaching(configCache, configDB, func() {
		if autoDDNSService != nil {
			autoDDNSService.NotifyConfigChange()
		}
	})

	domainCaching := caching.NewDomainCaching(domainCache, domainDB, func() {
		if autoDDNSService != nil {
			autoDDNSService.NotifyConfigChange()
		}
	})

	ddnsService := core.NewDDNSService(configCaching, domainCaching, logger, dnsproviders.NewDNSProvider)

	autoDDNSService = core.NewAutoDDNSService(configCaching, domainCaching, ddnsService)

	handlers := web.NewAPIHandlers(service.NewConfigService(configCaching, domainCaching, sqliteDB), domainCaching, systemCaching, ddnsService, autoDDNSService)

	serverConfig := &web.ServerConfig{
		Addr:   ":" + port,
		APIKey: apiKey,
	}

	server := web.NewServer(serverConfig, handlers)

	if autoDDNS {
		fmt.Println("启动自动DDNS服务...")
		if err := autoDDNSService.Start(); err != nil {
			fmt.Printf("启动自动DDNS服务失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("自动DDNS服务启动成功")
	}

	fmt.Printf("DDNS服务已启动，监听端口: %s\n", port)
	if apiKey != "" {
		fmt.Println("API认证已启用")
	}

	if err := server.Start(); err != nil {
		fmt.Printf("启动HTTP服务失败: %v\n", err)
		os.Exit(1)
	}
}

func HandleConfigCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("请指定子命令: create, get, list, update, delete")
		return
	}

	subCmd := args[0]
	switch subCmd {
	case "create":
		createConfig(args[1:])
	case "get":
		getConfig(args[1:])
	case "list":
		listConfigs()
	case "update":
		updateConfig(args[1:])
	case "delete":
		deleteConfig(args[1:])
	default:
		fmt.Printf("未知子命令: %s\n", subCmd)
	}
}

func HandleDomainCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("请指定子命令: create, list, update, delete")
		return
	}

	subCmd := args[0]
	switch subCmd {
	case "create":
		createDomain(args[1:])
	case "list":
		listDomains(args[1:])
	case "update":
		updateDomain(args[1:])
	case "delete":
		deleteDomain(args[1:])
	default:
		fmt.Printf("未知子命令: %s\n", subCmd)
	}
}

func HandleSystemCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("请指定子命令: get, update")
		return
	}

	subCmd := args[0]
	switch subCmd {
	case "get":
		getSystem()
	case "update":
		updateSystem(args[1:])
	default:
		fmt.Printf("未知子命令: %s\n", subCmd)
	}
}

func HandleDDNSCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("请指定子命令: trigger, start, stop")
		return
	}

	subCmd := args[0]
	switch subCmd {
	case "trigger":
		triggerDDNS(args[1:])
	case "start":
		startAutoDDNS()
	case "stop":
		stopAutoDDNS()
	default:
		fmt.Printf("未知子命令: %s\n", subCmd)
	}
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type SuccessResponse struct {
	Message string `json:"message"`
}

func sendRequest(method, path string, body interface{}) ([]byte, int, error) {
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, 0, err
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	req, err := http.NewRequest(method, baseURL+path, reqBody)
	if err != nil {
		return nil, 0, err
	}

	if apiKey != "" {
		token := utils.GenerateAuthToken(apiKey)
		req.Header.Set("Authorization", "Bearer "+token)
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}

	return data, resp.StatusCode, nil
}

func handleResponse(data []byte, statusCode int) error {
	if statusCode >= 400 {
		var errResp ErrorResponse
		if err := json.Unmarshal(data, &errResp); err == nil && errResp.Error != "" {
			return fmt.Errorf("错误: %s", errResp.Error)
		}
		return fmt.Errorf("请求失败 (HTTP %d): %s", statusCode, string(data))
	}

	return nil
}

func printJSONResponse(data []byte) {
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, data, "", "  "); err != nil {
		fmt.Println(string(data))
		return
	}
	fmt.Println(prettyJSON.String())
}

func createConfig(args []string) {
	if len(args) < 1 {
		fmt.Println("请指定配置文件路径: -f config.json")
		return
	}

	var filePath string
	for i := 0; i < len(args); i++ {
		if args[i] == "-f" && i+1 < len(args) {
			filePath = args[i+1]
			break
		}
	}

	if filePath == "" {
		fmt.Println("请指定配置文件路径: -f config.json")
		return
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("读取文件失败: %v\n", err)
		return
	}

	var config model.Config
	if err := json.Unmarshal(data, &config); err != nil {
		fmt.Printf("解析配置文件失败: %v\n", err)
		return
	}

	resp, statusCode, err := sendRequest("POST", "/api/config", config)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	if err := handleResponse(resp, statusCode); err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	fmt.Println("配置创建成功:")
	printJSONResponse(resp)
}

func getConfig(args []string) {
	if len(args) < 1 {
		fmt.Println("请指定配置ID")
		return
	}

	configID := args[0]
	resp, statusCode, err := sendRequest("GET", "/api/config/"+configID, nil)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	if err := handleResponse(resp, statusCode); err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	printJSONResponse(resp)
}

func listConfigs() {
	resp, statusCode, err := sendRequest("GET", "/api/configs", nil)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	if err := handleResponse(resp, statusCode); err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	printJSONResponse(resp)
}

func updateConfig(args []string) {
	if len(args) < 2 {
		fmt.Println("请指定配置ID和配置文件: <id> -f config.json")
		return
	}

	configID := args[0]
	var filePath string

	for i := 1; i < len(args); i++ {
		if args[i] == "-f" && i+1 < len(args) {
			filePath = args[i+1]
			break
		}
	}

	if filePath == "" {
		fmt.Println("请指定配置文件路径: -f config.json")
		return
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("读取文件失败: %v\n", err)
		return
	}

	var config model.Config
	if err := json.Unmarshal(data, &config); err != nil {
		fmt.Printf("解析配置文件失败: %v\n", err)
		return
	}
	config.ID = configID

	resp, statusCode, err := sendRequest("PUT", "/api/config/"+configID, config)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	if err := handleResponse(resp, statusCode); err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	fmt.Println("配置更新成功:")
	printJSONResponse(resp)
}

func deleteConfig(args []string) {
	if len(args) < 1 {
		fmt.Println("请指定配置ID")
		return
	}

	configID := args[0]
	resp, statusCode, err := sendRequest("DELETE", "/api/config/"+configID, nil)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	if err := handleResponse(resp, statusCode); err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	if len(resp) > 0 {
		printJSONResponse(resp)
	} else {
		fmt.Println("配置删除成功")
	}
}

func createDomain(args []string) {
	if len(args) < 1 {
		fmt.Println("请指定域名配置文件路径: -f domain.json")
		return
	}

	var filePath string
	for i := 0; i < len(args); i++ {
		if args[i] == "-f" && i+1 < len(args) {
			filePath = args[i+1]
			break
		}
	}

	if filePath == "" {
		fmt.Println("请指定域名配置文件路径: -f domain.json")
		return
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("读取文件失败: %v\n", err)
		return
	}

	var domain model.DomainConfig
	if err := json.Unmarshal(data, &domain); err != nil {
		fmt.Printf("解析域名配置文件失败: %v\n", err)
		return
	}

	resp, statusCode, err := sendRequest("POST", "/api/domain", domain)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	if err := handleResponse(resp, statusCode); err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	fmt.Println("域名创建成功:")
	printJSONResponse(resp)
}

func listDomains(args []string) {
	if len(args) < 1 {
		fmt.Println("请指定配置ID")
		return
	}

	configID := args[0]
	resp, statusCode, err := sendRequest("GET", "/api/domains/"+configID, nil)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	if err := handleResponse(resp, statusCode); err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	printJSONResponse(resp)
}

func updateDomain(args []string) {
	if len(args) < 2 {
		fmt.Println("请指定域名ID和配置文件: <id> -f domain.json")
		return
	}

	domainID := args[0]
	var filePath string

	for i := 1; i < len(args); i++ {
		if args[i] == "-f" && i+1 < len(args) {
			filePath = args[i+1]
			break
		}
	}

	if filePath == "" {
		fmt.Println("请指定域名配置文件路径: -f domain.json")
		return
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("读取文件失败: %v\n", err)
		return
	}

	var domain model.DomainConfig
	if err := json.Unmarshal(data, &domain); err != nil {
		fmt.Printf("解析域名配置文件失败: %v\n", err)
		return
	}
	domain.ID = domainID

	resp, statusCode, err := sendRequest("PUT", "/api/domain/"+domainID, domain)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	if err := handleResponse(resp, statusCode); err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	fmt.Println("域名更新成功:")
	printJSONResponse(resp)
}

func deleteDomain(args []string) {
	if len(args) < 1 {
		fmt.Println("请指定域名ID")
		return
	}

	domainID := args[0]
	resp, statusCode, err := sendRequest("DELETE", "/api/domain/"+domainID, nil)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	if err := handleResponse(resp, statusCode); err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	if len(resp) > 0 {
		printJSONResponse(resp)
	} else {
		fmt.Println("域名删除成功")
	}
}

func getSystem() {
	resp, statusCode, err := sendRequest("GET", "/api/system", nil)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	if err := handleResponse(resp, statusCode); err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	printJSONResponse(resp)
}

func updateSystem(args []string) {
	if len(args) < 1 {
		fmt.Println("请指定系统配置文件路径: -f system.json")
		return
	}

	var filePath string
	for i := 0; i < len(args); i++ {
		if args[i] == "-f" && i+1 < len(args) {
			filePath = args[i+1]
			break
		}
	}

	if filePath == "" {
		fmt.Println("请指定系统配置文件路径: -f system.json")
		return
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("读取文件失败: %v\n", err)
		return
	}

	var system model.SystemConfig
	if err := json.Unmarshal(data, &system); err != nil {
		fmt.Printf("解析系统配置文件失败: %v\n", err)
		return
	}

	resp, statusCode, err := sendRequest("PUT", "/api/system", system)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	if err := handleResponse(resp, statusCode); err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	fmt.Println("系统配置更新成功:")
	printJSONResponse(resp)
}

func triggerDDNS(args []string) {
	if len(args) < 1 {
		fmt.Println("请指定配置ID")
		return
	}

	configID := args[0]
	resp, statusCode, err := sendRequest("POST", "/api/ddns/"+configID, nil)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	if err := handleResponse(resp, statusCode); err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	printJSONResponse(resp)
}

func startAutoDDNS() {
	resp, statusCode, err := sendRequest("POST", "/api/auto-ddns/start", nil)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	if err := handleResponse(resp, statusCode); err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	printJSONResponse(resp)
}

func stopAutoDDNS() {
	resp, statusCode, err := sendRequest("POST", "/api/auto-ddns/stop", nil)
	if err != nil {
		fmt.Printf("请求失败: %v\n", err)
		return
	}

	if err := handleResponse(resp, statusCode); err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	printJSONResponse(resp)
}
