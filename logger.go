// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package main

import (
	"fmt"
	"time"
)

type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelError
)

var currentLogLevel LogLevel = LogLevelInfo

func SetLogLevel(level LogLevel) {
	currentLogLevel = level
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