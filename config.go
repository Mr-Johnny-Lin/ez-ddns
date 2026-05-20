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

type DomainConfig struct {
	DomainName string `json:"domainName"`
	RR         string `json:"rr"`
	IPType     IPType `json:"ipType"`
	LastIP     string `json:"lastIP"`
}

type Config struct {
	AccessKeyId     string         `json:"accessKeyId"`
	AccessKeySecret string         `json:"accessKeySecret"`
	Provider        ProviderType   `json:"provider"`
	Interval        int            `json:"interval"`
	LogLevel        string         `json:"logLevel"`
	Domains         []DomainConfig `json:"domains"`
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
		Provider: ProviderTypeAliyun,
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

	if envVal := os.Getenv("DDNS_INTERVAL"); envVal != "" {
		if parsedInterval, err := strconv.Atoi(envVal); err == nil && parsedInterval > 0 {
			config.Interval = parsedInterval
		}
	}

	if envVal := os.Getenv("DDNS_LOG_LEVEL"); envVal != "" {
		config.LogLevel = envVal
	}

	if envVal := os.Getenv("DDNS_PROVIDER"); envVal != "" {
		config.Provider = ParseProviderType(envVal)
	}

	overrideDomainsWithEnv(config)
}

func overrideDomainsWithEnv(config *Config) {
	for i := 0; ; i++ {
		domainName := os.Getenv(fmt.Sprintf("DDNS_DOMAIN_%d_DOMAIN_NAME", i))
		rr := os.Getenv(fmt.Sprintf("DDNS_DOMAIN_%d_RR", i))
		ipType := os.Getenv(fmt.Sprintf("DDNS_DOMAIN_%d_IP_TYPE", i))

		if domainName == "" && rr == "" && ipType == "" {
			if i == 0 && len(config.Domains) == 0 {
				continue
			}
			break
		}

		if i < len(config.Domains) {
			if domainName != "" {
				config.Domains[i].DomainName = domainName
			}
			if rr != "" {
				config.Domains[i].RR = rr
			}
			if ipType != "" {
				config.Domains[i].IPType = ParseIPType(ipType)
			}
		} else {
			config.Domains = append(config.Domains, DomainConfig{
				DomainName: domainName,
				RR:         rr,
				IPType:     ParseIPType(ipType),
			})
		}
	}
}

func validateConfig(config *Config) error {
	if config.AccessKeyId == "" {
		return fmt.Errorf("请设置环境变量 ALIBABA_CLOUD_ACCESS_KEY_ID 或在配置文件中配置 accessKeyId")
	}

	if config.AccessKeySecret == "" {
		return fmt.Errorf("请设置环境变量 ALIBABA_CLOUD_ACCESS_KEY_SECRET 或在配置文件中配置 accessKeySecret")
	}

	if !config.Provider.IsValid() {
		return fmt.Errorf("provider 必须为 %s", ProviderTypeAliyun)
	}

	if len(config.Domains) == 0 {
		return fmt.Errorf("请在配置文件中至少配置一个域名")
	}

	for i, domain := range config.Domains {
		if domain.DomainName == "" {
			return fmt.Errorf("第 %d 个域名配置中 domainName 不能为空", i+1)
		}
		if domain.RR == "" {
			return fmt.Errorf("第 %d 个域名配置中 rr 不能为空", i+1)
		}
		if !domain.IPType.IsValid() {
			return fmt.Errorf("第 %d 个域名配置中 ipType 必须为 %s 或 %s", i+1, IPTypeIPv4, IPTypeIPv6)
		}
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