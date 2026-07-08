// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package main

import (
	"flag"
	"fmt"
	"os"

	"ez-ddns/v2/cli"
)

func main() {
	if len(os.Args) < 2 {
		cli.PrintHelp()
		os.Exit(0)
	}

	command := os.Args[1]

	switch command {
	case "start":
		handleStartCommand()
	case "config":
		handleClientCommand(os.Args[2:], cli.HandleConfigCommand)
	case "domain":
		handleClientCommand(os.Args[2:], cli.HandleDomainCommand)
	case "system":
		handleClientCommand(os.Args[2:], cli.HandleSystemCommand)
	case "ddns":
		handleClientCommand(os.Args[2:], cli.HandleDDNSCommand)
	case "help":
		cli.PrintHelp()
	default:
		fmt.Printf("未知命令: %s\n", command)
		cli.PrintHelp()
		os.Exit(1)
	}
}

func handleStartCommand() {
	startCmd := flag.NewFlagSet("start", flag.ExitOnError)

	autoDDNS := startCmd.Bool("auto-ddns", false, "启动时同时启动自动DDNS服务")
	port := startCmd.String("port", "8080", "HTTP服务端口")
	apiKey := startCmd.String("api-key", "", "API密钥（用于认证）")

	err := startCmd.Parse(os.Args[2:])
	if err != nil {
		fmt.Printf("参数解析失败: %v\n", err)
		os.Exit(1)
	}

	cli.StartServer(*port, *apiKey, *autoDDNS)
}

func handleClientCommand(args []string, handler func([]string)) {
	flags := map[string]*string{
		"-server":  nil,
		"-api-key": nil,
	}

	remainingArgs := extractFlags(args, flags)

	server := "http://localhost:8080"
	if flags["-server"] != nil {
		server = *flags["-server"]
	}

	apiKey := ""
	if flags["-api-key"] != nil {
		apiKey = *flags["-api-key"]
	}

	cli.SetServer(server)
	cli.SetAPIKey(apiKey)

	handler(remainingArgs)
}

func extractFlags(args []string, flags map[string]*string) []string {
	remainingArgs := []string{}
	i := 0
	for i < len(args) {
		if _, ok := flags[args[i]]; ok {
			if i+1 < len(args) {
				val := args[i+1]
				flags[args[i]] = &val
				i += 2
			} else {
				fmt.Printf("错误: %s 需要指定值\n", args[i])
				os.Exit(1)
			}
		} else {
			remainingArgs = append(remainingArgs, args[i])
			i++
		}
	}
	return remainingArgs
}
