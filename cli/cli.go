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

	"ez-ddns/core"
	"ez-ddns/core/dnsproviders"
	"ez-ddns/dao/cache"
	"ez-ddns/dao/caching"
	"ez-ddns/dao/db"
	"ez-ddns/utils"
	"ez-ddns/web"
	"ez-ddns/web/dto"
	"ez-ddns/web/service"
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
	fmt.Println(helpText)
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

	configCaching := caching.NewConfigCaching(configCache, configDB)
	domainCaching := caching.NewDomainCaching(domainCache, domainDB)

	ddnsService := core.NewDDNSService(configCaching, domainCaching, dnsproviders.NewDNSProvider)

	autoDDNSService = core.NewAutoDDNSService(configCaching, domainCaching, ddnsService)

	configService := service.NewConfigService(configCaching, domainCaching, sqliteDB, autoDDNSService)

	handlers := web.NewAPIHandlers(configService, domainCaching, systemCaching, ddnsService, autoDDNSService)

	serverConfig := &web.ServerConfig{
		Addr:   ":" + port,
		APIKey: apiKey,
		Logger: logger,
	}

	server := web.NewServer(serverConfig, handlers)

	if autoDDNS {
		fmt.Println("启动自动DDNS服务...")
		if err := autoDDNSService.Start(utils.ContextWithLogger(context.Background(), logger)); err != nil {
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
		listConfigs(args[1:])
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
	var result interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		fmt.Println(string(data))
		return
	}
	printKeyValue(result, "")
}

func printKeyValue(data interface{}, prefix string) {
	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			if value == nil {
				continue
			}
			switch val := value.(type) {
			case map[string]interface{}:
				fmt.Printf("%s%s:\n", prefix, key)
				printKeyValue(val, prefix+"  ")
			case []interface{}:
				if len(val) == 0 {
					continue
				}
				fmt.Printf("%s%s:\n", prefix, key)
				for _, item := range val {
					fmt.Printf("%s  -\n", prefix)
					printKeyValue(item, prefix+"    ")
				}
			default:
				fmt.Printf("%s%s: %v\n", prefix, key, value)
			}
		}
	case []interface{}:
		for i, item := range v {
			if i > 0 {
				fmt.Println()
			}
			printKeyValue(item, "")
		}
	default:
		fmt.Println(data)
	}
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

	var config dto.ConfigDTO
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

func listConfigs(args []string) {
	var path string
	if len(args) > 0 {
		path = "/api/config/" + args[0]
	} else {
		path = "/api/configs"
	}

	resp, statusCode, err := sendRequest("GET", path, nil)
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

	var config dto.ConfigDTO
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
	if len(args) < 2 {
		fmt.Println("请指定配置ID和域名配置文件路径: <configID> -f domain.json")
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
		fmt.Println("请指定域名配置文件路径: -f domain.json")
		return
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		fmt.Printf("读取文件失败: %v\n", err)
		return
	}

	var domain dto.DomainDTO
	if err := json.Unmarshal(data, &domain); err != nil {
		fmt.Printf("解析域名配置文件失败: %v\n", err)
		return
	}

	domain.ConfigID = configID

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

	var domain dto.DomainDTO
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

	var system dto.SystemDTO
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
