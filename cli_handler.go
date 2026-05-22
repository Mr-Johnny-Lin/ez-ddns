// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"ez-ddns/dao"
	"ez-ddns/model"
)

type CLIHandler struct {
	db  *dao.ConfigDAO
	dm  *dao.DomainDAO
	sys *dao.SystemSettingsDAO
}

func NewCLIHandler(db *dao.ConfigDAO, dm *dao.DomainDAO, sys *dao.SystemSettingsDAO) *CLIHandler {
	return &CLIHandler{db: db, dm: dm, sys: sys}
}

func PrintHelp() {
	fmt.Println(`EZ-DDNS - 轻量级动态DNS服务

Usage:
  ez-ddns <command> [options]

Commands:
  start        启动DDNS服务
  status       查看系统状态
  loglevel     设置日志等级 (debug/info/error)
  config       配置管理
    list       列出所有配置
    create     创建配置
    update     更新配置
    delete     删除配置
  domain       域名管理
    list       列出域名
    add        添加域名
    remove     删除域名
  help         显示帮助信息

Config Commands:
  ez-ddns config create <id> --access-key-id=<key> --access-key-secret=<secret> [--interval=<seconds>] [--provider=<provider>]
  ez-ddns config update <id> [--access-key-id=<key>] [--access-key-secret=<secret>] [--interval=<seconds>] [--provider=<provider>]
  ez-ddns config delete <id>
  ez-ddns config list

Domain Commands:
  ez-ddns domain add <config-id> --domain-name=<domain> --rr=<record> [--ip-type=<ipv4|ipv6>]
  ez-ddns domain remove <config-id> --domain-name=<domain> --rr=<record> [--ip-type=<ipv4|ipv6>]
  ez-ddns domain list <config-id>

System Commands:
  ez-ddns status        查看系统运行状态
  ez-ddns loglevel      设置日志等级 (debug/info/error)

Start Command:
  ez-ddns start [--loglevel=<level>]

Options:
  --access-key-id      阿里云AccessKey ID
  --access-key-secret  阿里云AccessKey Secret
  --provider           DNS提供商 (默认: aliyun)
  --interval           更新间隔秒数 (默认: 180)
  --domain-name        域名
  --rr                 记录前缀
  --ip-type            IP类型 (默认: ipv4)
  --loglevel           日志等级 (debug/info/error, 默认: info)

Examples:
  ez-ddns config create myconfig --access-key-id=xxx --access-key-secret=xxx
  ez-ddns config update myconfig --interval=300
  ez-ddns domain add myconfig --domain-name=example.com --rr=www
  ez-ddns loglevel debug
  ez-ddns start --loglevel=debug
  ez-ddns status`)
}

func ParseArgs() (string, string, map[string]string) {
	if len(os.Args) < 2 {
		return "", "", nil
	}

	command := os.Args[1]

	var subcommand string
	args := make(map[string]string)

	idx := 2
	if len(os.Args) > 2 && !strings.HasPrefix(os.Args[2], "--") {
		subcommand = os.Args[2]
		idx = 3
	}

	for i := idx; i < len(os.Args); i++ {
		arg := os.Args[i]
		if strings.HasPrefix(arg, "--") {
			parts := strings.SplitN(arg[2:], "=", 2)
			key := parts[0]
			value := ""
			if len(parts) == 2 {
				value = parts[1]
			}
			args[key] = value
		} else {
			if _, exists := args["_arg"]; !exists {
				args["_arg"] = arg
			} else {
				args["_arg2"] = arg
			}
		}
	}

	return command, subcommand, args
}

func (s *CLIHandler) HandleCommand(command, subcommand string, args map[string]string) error {
	switch strings.ToLower(command) {
	case "help":
		PrintHelp()
		return nil

	case "status":
		return s.handleStatus()

	case "start":
		return s.handleStart(args)

	case "config":
		return s.handleConfig(subcommand, args)

	case "domain":
		return s.handleDomain(subcommand, args)

	case "loglevel":
		return s.handleLogLevel(subcommand)

	default:
		return fmt.Errorf("未知命令: %s\n使用 'ez-ddns help' 查看帮助", command)
	}
}

func (s *CLIHandler) handleStatus() error {
	settings, err := s.sys.Get("status")
	if err != nil {
		return fmt.Errorf("获取系统状态失败: %v", err)
	}

	fmt.Println("系统状态信息:")
	if settings != nil {
		if uptime, ok := settings.Data["uptime"]; ok {
			fmt.Printf("  运行时长: %s\n", uptime)
		}
		if lastCheck, ok := settings.Data["last_check"]; ok {
			fmt.Printf("  最后检查: %s\n", lastCheck)
		}
		if logLevel, ok := settings.Data["log_level"]; ok {
			fmt.Printf("  日志等级: %s\n", logLevel)
		}
		fmt.Printf("  记录时间: %s\n", settings.UpdatedAt)
	} else {
		fmt.Println("  服务未运行或状态信息未更新")
	}

	sysSettings, err := s.sys.Get("system")
	if err == nil && sysSettings != nil {
		fmt.Println("\n系统配置:")
		if logLevel, ok := sysSettings.Data["log_level"]; ok {
			fmt.Printf("  日志等级: %s\n", logLevel)
		}
	}

	return nil
}

func (s *CLIHandler) handleStart(args map[string]string) error {
	if loglevel, exists := args["loglevel"]; exists {
		switch strings.ToLower(loglevel) {
		case "debug":
			SetLogLevel(LogLevelDebug)
		case "info":
			SetLogLevel(LogLevelInfo)
		case "error":
			SetLogLevel(LogLevelError)
		default:
			return fmt.Errorf("无效的日志等级: %s\n可用等级: debug, info, error", loglevel)
		}
	}

	StartDDNSService()
	return nil
}

func (s *CLIHandler) handleConfig(subcommand string, args map[string]string) error {
	switch strings.ToLower(subcommand) {
	case "list":
		return s.listConfigs()

	case "create":
		return s.createConfig(args)

	case "update":
		return s.updateConfig(args)

	case "delete":
		return s.deleteConfig(args)

	default:
		return fmt.Errorf("未知配置命令: %s\n使用 'ez-ddns help' 查看帮助", subcommand)
	}
}

func (s *CLIHandler) handleDomain(subcommand string, args map[string]string) error {
	switch strings.ToLower(subcommand) {
	case "list":
		return s.listDomains(args)

	case "add":
		return s.addDomain(args)

	case "remove":
		return s.removeDomain(args)

	default:
		return fmt.Errorf("未知域名命令: %s\n使用 'ez-ddns help' 查看帮助", subcommand)
	}
}

func (s *CLIHandler) handleLogLevel(level string) error {
	switch strings.ToLower(level) {
	case "debug":
		SetLogLevel(LogLevelDebug)
		fmt.Println("日志等级已设置为: debug")
	case "info":
		SetLogLevel(LogLevelInfo)
		fmt.Println("日志等级已设置为: info")
	case "error":
		SetLogLevel(LogLevelError)
		fmt.Println("日志等级已设置为: error")
	default:
		return fmt.Errorf("无效的日志等级: %s\n可用等级: debug, info, error", level)
	}
	return nil
}

func (s *CLIHandler) listConfigs() error {
	configs, err := s.db.ReadAll()
	if err != nil {
		return fmt.Errorf("查询配置失败: %v", err)
	}

	if len(configs) == 0 {
		fmt.Println("暂无配置")
		return nil
	}

	for _, cfg := range configs {
		fmt.Printf("配置ID: %s\n", cfg.ID)
		fmt.Printf("  Provider: %s\n", cfg.Provider)
		fmt.Printf("  Interval: %ds\n", cfg.Interval)
		fmt.Printf("  AccessKeyId: %s\n", maskString(cfg.AccessKeyId))
		fmt.Printf("  AccessKeySecret: %s\n", maskString(cfg.AccessKeySecret))

		domains, _ := s.dm.ReadByConfigID(cfg.ID)
		if len(domains) > 0 {
			fmt.Println("  Domains:")
			for _, d := range domains {
				fmt.Printf("    - %s.%s (%s)\n", d.RR, d.DomainName, d.IPType)
			}
		}
		fmt.Println()
	}

	return nil
}

func (s *CLIHandler) createConfig(args map[string]string) error {
	configID := args["_arg"]
	if configID == "" {
		return fmt.Errorf("请指定配置ID\nUsage: ez-ddns config create <id> --access-key-id=<key> --access-key-secret=<secret>")
	}

	accessKeyId := args["access-key-id"]
	if accessKeyId == "" {
		return fmt.Errorf("请指定AccessKey ID (--access-key-id)")
	}

	accessKeySecret := args["access-key-secret"]
	if accessKeySecret == "" {
		return fmt.Errorf("请指定AccessKey Secret (--access-key-secret)")
	}

	existing, _ := s.db.ReadByID(configID)
	if existing != nil {
		return fmt.Errorf("配置ID[%s]已存在", configID)
	}

	provider := args["provider"]
	if provider == "" {
		provider = "aliyun"
	}

	interval := 180
	if val, exists := args["interval"]; exists {
		var err error
		interval, err = parseInt(val, 180)
		if err != nil {
			return fmt.Errorf("interval 必须是数字")
		}
	}

	cfg := &model.Config{
		ID:              configID,
		AccessKeyId:     accessKeyId,
		AccessKeySecret: accessKeySecret,
		Provider:        model.ParseProviderType(provider),
		Interval:        interval,
	}

	if err := s.db.Create(cfg); err != nil {
		return fmt.Errorf("创建配置失败: %v", err)
	}

	fmt.Printf("配置ID[%s]创建成功\n", configID)
	return nil
}

func (s *CLIHandler) updateConfig(args map[string]string) error {
	configID := args["_arg"]
	if configID == "" {
		return fmt.Errorf("请指定配置ID\nUsage: ez-ddns config update <id> [options]")
	}

	existing, err := s.db.ReadByID(configID)
	if err != nil {
		return fmt.Errorf("读取配置失败: %v", err)
	}
	if existing == nil {
		return fmt.Errorf("配置ID[%s]不存在", configID)
	}

	if val, exists := args["access-key-id"]; exists && val != "" {
		existing.AccessKeyId = val
	}
	if val, exists := args["access-key-secret"]; exists && val != "" {
		existing.AccessKeySecret = val
	}
	if val, exists := args["provider"]; exists && val != "" {
		existing.Provider = model.ParseProviderType(val)
	}
	if val, exists := args["interval"]; exists && val != "" {
		interval, err := parseInt(val, 180)
		if err != nil {
			return fmt.Errorf("interval 必须是数字")
		}
		existing.Interval = interval
	}

	if err := s.db.Update(existing); err != nil {
		return fmt.Errorf("更新配置失败: %v", err)
	}

	fmt.Printf("配置ID[%s]更新成功\n", configID)
	return nil
}

func (s *CLIHandler) deleteConfig(args map[string]string) error {
	configID := args["_arg"]
	if configID == "" {
		return fmt.Errorf("请指定配置ID\nUsage: ez-ddns config delete <id>")
	}

	existing, err := s.db.ReadByID(configID)
	if err != nil {
		return fmt.Errorf("读取配置失败: %v", err)
	}
	if existing == nil {
		return fmt.Errorf("配置ID[%s]不存在", configID)
	}

	if err := s.db.Delete(configID); err != nil {
		return fmt.Errorf("删除配置失败: %v", err)
	}

	fmt.Printf("配置ID[%s]删除成功\n", configID)
	return nil
}

func (s *CLIHandler) listDomains(args map[string]string) error {
	configID := args["_arg"]
	if configID == "" {
		return fmt.Errorf("请指定配置ID\nUsage: ez-ddns domain list <config-id>")
	}

	domains, err := s.dm.ReadByConfigID(configID)
	if err != nil {
		return fmt.Errorf("查询域名失败: %v", err)
	}

	if len(domains) == 0 {
		fmt.Printf("配置ID[%s]暂无域名\n", configID)
		return nil
	}

	fmt.Printf("配置ID[%s]的域名列表:\n", configID)
	for _, d := range domains {
		fmt.Printf("  - %s.%s (%s) [ID: %s]\n", d.RR, d.DomainName, d.IPType, d.ID)
	}

	return nil
}

func (s *CLIHandler) addDomain(args map[string]string) error {
	configID := args["_arg"]
	if configID == "" {
		return fmt.Errorf("请指定配置ID\nUsage: ez-ddns domain add <config-id> --domain-name=<domain> --rr=<record>")
	}

	domainName := args["domain-name"]
	if domainName == "" {
		return fmt.Errorf("请指定域名 (--domain-name)")
	}

	rr := args["rr"]
	if rr == "" {
		return fmt.Errorf("请指定记录前缀 (--rr)")
	}

	ipType := args["ip-type"]
	if ipType == "" {
		ipType = "ipv4"
	}

	existing, err := s.db.ReadByID(configID)
	if err != nil {
		return fmt.Errorf("读取配置失败: %v", err)
	}
	if existing == nil {
		return fmt.Errorf("配置ID[%s]不存在", configID)
	}

	existingDomain, err := s.dm.ReadByUniqueKey(domainName, rr, ipType)
	if err != nil {
		return fmt.Errorf("检查域名记录失败: %v", err)
	}
	if existingDomain != nil {
		return fmt.Errorf("域名记录已存在: %s.%s (%s)", rr, domainName, ipType)
	}

	domain := model.DomainConfig{
		ID:         generateID(),
		DomainName: domainName,
		RR:         rr,
		IPType:     model.ParseIPType(ipType),
	}

	if err := s.dm.Create(configID, &domain); err != nil {
		return fmt.Errorf("添加域名失败: %v", err)
	}

	fmt.Printf("配置ID[%s] - 域名 %s.%s (%s) 添加成功\n", configID, rr, domainName, ipType)
	return nil
}

func (s *CLIHandler) removeDomain(args map[string]string) error {
	configID := args["_arg"]
	if configID == "" {
		return fmt.Errorf("请指定配置ID\nUsage: ez-ddns domain remove <config-id> --domain-name=<domain> --rr=<record>")
	}

	domainName := args["domain-name"]
	if domainName == "" {
		return fmt.Errorf("请指定域名 (--domain-name)")
	}

	rr := args["rr"]
	if rr == "" {
		return fmt.Errorf("请指定记录前缀 (--rr)")
	}

	ipType := args["ip-type"]
	if ipType == "" {
		ipType = "ipv4"
	}

	domain, err := s.dm.ReadByUniqueKey(domainName, rr, ipType)
	if err != nil {
		return fmt.Errorf("查询域名失败: %v", err)
	}
	if domain == nil {
		return fmt.Errorf("域名 %s.%s (%s) 不存在", rr, domainName, ipType)
	}

	if err := s.dm.Delete(domain.ID); err != nil {
		return fmt.Errorf("删除域名失败: %v", err)
	}

	fmt.Printf("配置ID[%s] - 域名 %s.%s (%s) 删除成功\n", configID, rr, domainName, ipType)
	return nil
}

func maskString(s string) string {
	if len(s) <= 4 {
		return "****"
	}
	return s[:2] + "****" + s[len(s)-2:]
}

func generateID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

func parseInt(s string, defaultValue int) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	if err != nil {
		return defaultValue, err
	}
	return result, nil
}