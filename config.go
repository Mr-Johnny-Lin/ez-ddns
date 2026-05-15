// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
)

const configFilePath = "ez-ddns-config.json"

type Config struct {
	AccessKeyId     string `json:"accessKeyId"`
	AccessKeySecret string `json:"accessKeySecret"`
	DomainName      string `json:"domainName"`
	RR              string `json:"rr"`
	Interval        int    `json:"interval"`
	LastIP          string `json:"lastIP"`
	LogLevel        string `json:"logLevel"`
}

func LoadAndValidateConfig() (*Config, error) {
	config, err := loadConfigFromFile()
	if err != nil {
		return nil, fmt.Errorf("加载配置文件失败: %v", err)
	}

	overrideWithEnv(config)

	if err := validateConfig(config); err != nil {
		return nil, fmt.Errorf("配置校验失败: %v", err)
	}

	if err := saveConfigToFile(config); err != nil {
		return nil, fmt.Errorf("保存配置文件失败: %v", err)
	}

	return config, nil
}

func loadConfigFromFile() (*Config, error) {
	config := &Config{
		Interval: 180,
		LogLevel: "info",
	}

	file, err := os.Open(configFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return config, nil
		}
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return nil, err
	}

	return config, nil
}

func overrideWithEnv(config *Config) {
	if envVal := os.Getenv("ALIBABA_CLOUD_ACCESS_KEY_ID"); envVal != "" {
		config.AccessKeyId = envVal
	}

	if envVal := os.Getenv("ALIBABA_CLOUD_ACCESS_KEY_SECRET"); envVal != "" {
		config.AccessKeySecret = envVal
	}

	if envVal := os.Getenv("DDNS_DOMAIN_NAME"); envVal != "" {
		config.DomainName = envVal
	}

	if envVal := os.Getenv("DDNS_RR"); envVal != "" {
		config.RR = envVal
	}

	if envVal := os.Getenv("DDNS_INTERVAL"); envVal != "" {
		if parsedInterval, err := strconv.Atoi(envVal); err == nil && parsedInterval > 0 {
			config.Interval = parsedInterval
		}
	}

	if envVal := os.Getenv("DDNS_LOG_LEVEL"); envVal != "" {
		config.LogLevel = envVal
	}
}

func validateConfig(config *Config) error {
	if config.AccessKeyId == "" {
		return fmt.Errorf("请设置环境变量 ALIBABA_CLOUD_ACCESS_KEY_ID 或在配置文件中配置 accessKeyId")
	}

	if config.AccessKeySecret == "" {
		return fmt.Errorf("请设置环境变量 ALIBABA_CLOUD_ACCESS_KEY_SECRET 或在配置文件中配置 accessKeySecret")
	}

	if config.DomainName == "" {
		return fmt.Errorf("请设置环境变量 DDNS_DOMAIN_NAME 或在配置文件中配置 domainName")
	}

	if config.RR == "" {
		return fmt.Errorf("请设置环境变量 DDNS_RR 或在配置文件中配置 rr")
	}

	if config.Interval <= 0 {
		return fmt.Errorf("interval 必须大于 0")
	}

	return nil
}

func saveConfigToFile(config *Config) error {
	file, err := os.Create(configFilePath)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	return encoder.Encode(config)
}

func UpdateLastIP(config *Config, ip string) error {
	if config.LastIP == ip {
		return nil
	}
	config.LastIP = ip
	return saveConfigToFile(config)
}