// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package utils

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

type Logger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Error(format string, args ...interface{})
	SetLevel(level LogLevel)
	GetLevel() LogLevel
}

type ConsoleLogger struct {
	levelProvider func() LogLevel
}

func NewLogger() Logger {
	return NewLoggerWithLevel("info")
}

func NewLoggerWithLevel(levelStr string) Logger {
	level := ParseLogLevel(levelStr)
	return &ConsoleLogger{
		levelProvider: func() LogLevel {
			return level
		},
	}
}

func NewLoggerWithProvider(levelProvider func() LogLevel) Logger {
	return &ConsoleLogger{
		levelProvider: levelProvider,
	}
}

func (l *ConsoleLogger) getLevel() LogLevel {
	return l.levelProvider()
}

func (l *ConsoleLogger) Debug(format string, args ...interface{}) {
	if l.getLevel() <= LogLevelDebug {
		fmt.Printf("[DEBUG] [%s] %s\n", time.Now().Format("2006-01-02 15:04:05"), fmt.Sprintf(format, args...))
	}
}

func (l *ConsoleLogger) Info(format string, args ...interface{}) {
	if l.getLevel() <= LogLevelInfo {
		fmt.Printf("[INFO] [%s] %s\n", time.Now().Format("2006-01-02 15:04:05"), fmt.Sprintf(format, args...))
	}
}

func (l *ConsoleLogger) Error(format string, args ...interface{}) {
	if l.getLevel() <= LogLevelError {
		fmt.Printf("[ERROR] [%s] %s\n", time.Now().Format("2006-01-02 15:04:05"), fmt.Sprintf(format, args...))
	}
}

func (l *ConsoleLogger) SetLevel(level LogLevel) {
	l.levelProvider = func() LogLevel {
		return level
	}
}

func (l *ConsoleLogger) GetLevel() LogLevel {
	return l.getLevel()
}

func ParseLogLevel(levelStr string) LogLevel {
	switch levelStr {
	case "debug":
		return LogLevelDebug
	case "info":
		return LogLevelInfo
	case "error":
		return LogLevelError
	default:
		return LogLevelInfo
	}
}

func LogLevelToString(level LogLevel) string {
	switch level {
	case LogLevelDebug:
		return "debug"
	case LogLevelInfo:
		return "info"
	case LogLevelError:
		return "error"
	default:
		return "info"
	}
}
