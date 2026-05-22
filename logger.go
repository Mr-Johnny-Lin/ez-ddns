// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package main

import (
	"fmt"
	"time"

	"ez-ddns/dao"
)

type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelError
)

var currentLogLevel LogLevel = LogLevelInfo
var serviceStartTime time.Time

func InitLogger() {
	serviceStartTime = time.Now()
	
	settings, err := dao.System.Get("system")
	if err == nil && settings != nil {
		if logLevelStr, ok := settings.Data["log_level"].(string); ok {
			switch logLevelStr {
			case "debug":
				currentLogLevel = LogLevelDebug
			case "info":
				currentLogLevel = LogLevelInfo
			case "error":
				currentLogLevel = LogLevelError
			}
		}
	}
}

func SetLogLevel(level LogLevel) {
	currentLogLevel = level
	
	logLevelStr := "info"
	switch level {
	case LogLevelDebug:
		logLevelStr = "debug"
	case LogLevelInfo:
		logLevelStr = "info"
	case LogLevelError:
		logLevelStr = "error"
	}
	
	err := dao.System.UpdateField("system", "log_level", logLevelStr)
	if err != nil {
		fmt.Printf("保存日志等级失败: %v\n", err)
	}
}

func Debug(format string, args ...interface{}) {
	if currentLogLevel <= LogLevelDebug {
		fmt.Printf("[DEBUG] [%s] %s\n", time.Now().Format("2006-01-02 15:04:05"), fmt.Sprintf(format, args...))
	}
}

func Info(format string, args ...interface{}) {
	if currentLogLevel <= LogLevelInfo {
		fmt.Printf("[INFO] [%s] %s\n", time.Now().Format("2006-01-02 15:04:05"), fmt.Sprintf(format, args...))
	}
}

func Error(format string, args ...interface{}) {
	if currentLogLevel <= LogLevelError {
		fmt.Printf("[ERROR] [%s] %s\n", time.Now().Format("2006-01-02 15:04:05"), fmt.Sprintf(format, args...))
	}
}

func GetServiceUptime() time.Duration {
	return time.Since(serviceStartTime)
}

func UpdateSystemStatus() error {
	status := map[string]interface{}{
		"uptime":         GetServiceUptime().String(),
		"last_check":     time.Now().Format("2006-01-02 15:04:05"),
		"log_level": func() string {
			switch currentLogLevel {
			case LogLevelDebug:
				return "debug"
			case LogLevelInfo:
				return "info"
			case LogLevelError:
				return "error"
			}
			return "info"
		}(),
	}
	
	return dao.System.Set("status", status)
}