# ez-ddns

基于阿里云 DNS API 的轻量级动态域名解析（DDNS）工具，自动检测公网 IP 变化并更新 DNS 记录。

## 功能特性

- ✅ 自动获取当前公网 IP（支持 IPv4/IPv6）
- ✅ 支持 IPv4（A记录）、IPv6（AAAA记录）两种模式
- ✅ 支持多域名配置（每个域名独立配置IP类型）
- ✅ 本地DNS预检查：更新前先通过DNS解析验证，减少运营商API调用
- ✅ 智能判断：无记录时新增，有记录且 IP 变化时更新
- ✅ 支持多条解析记录管理
- ✅ 持续轮询监控，可配置检查间隔（默认3分钟）
- ✅ 支持日志级别配置（DEBUG/INFO/ERROR）
- ✅ 自动保存上次 IP 记录，避免重复更新
- ✅ 支持 Windows 和 Linux 平台
- ✅ 完整的 CI/CD 自动化构建（GitHub Actions）
- ✅ **CLI命令管理**：通过命令行管理配置和域名
- ✅ **SQLite数据持久化**：配置存储在本地数据库中

## 快速开始

### 启动服务

```bash
# 启动DDNS服务（默认日志级别info）
ez-ddns start

# 启动服务并设置调试日志
ez-ddns start --loglevel=debug
```

### 配置管理

```bash
# 创建配置
ez-ddns config create <config-id> --access-key-id=<key> --access-key-secret=<secret> [--provider=aliyun] [--interval=180]

# 示例
ez-ddns config create home --access-key-id=LTAI5t... --access-key-secret=xxx --interval=300

# 查看所有配置
ez-ddns config list

# 更新配置
ez-ddns config update <config-id> --interval=600

# 删除配置
ez-ddns config delete <config-id>
```

### 域名管理

```bash
# 添加域名到配置
ez-ddns domain add <config-id> --domain-name=<domain> --rr=<record> [--ip-type=ipv4]

# 示例 - 添加IPv4记录
ez-ddns domain add home --domain-name=example.com --rr=www --ip-type=ipv4

# 示例 - 添加IPv6记录
ez-ddns domain add home --domain-name=example.com --rr=www --ip-type=ipv6

# 查看配置的域名列表
ez-ddns domain list <config-id>

# 删除域名
ez-ddns domain remove <config-id> --domain-name=<domain> --rr=<record> --ip-type=ipv4

# 查看系统状态
ez-ddns status

# 设置日志级别
ez-ddns loglevel debug
```

### CLI命令完整帮助

```bash
ez-ddns help
```

## 配置项说明

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `access-key-id` | 阿里云 AccessKey ID | 必填 |
| `access-key-secret` | 阿里云 AccessKey Secret | 必填 |
| `provider` | DNS运营商类型 | `aliyun` |
| `interval` | 轮询间隔（秒） | `180`（3分钟） |
| `loglevel` | 日志级别：`debug`、`info`、`error` | `info` |
| `ip-type` | IP类型：`ipv4`、`ipv6` | `ipv4` |

## 使用方法

### Windows (PowerShell)

```powershell
# 创建配置
.\ez-ddns-windows-amd64.exe config create home --access-key-id=LTAI5t... --access-key-secret=xxx

# 添加域名
.\ez-ddns-windows-amd64.exe domain add home --domain-name=example.com --rr=www --ip-type=ipv4

# 启动服务
.\ez-ddns-windows-amd64.exe start --loglevel=debug
```

### Linux

```bash
# 创建配置
./ez-ddns-linux-amd64 config create home --access-key-id=LTAI5t... --access-key-secret=xxx

# 添加域名
./ez-ddns-linux-amd64 domain add home --domain-name=example.com --rr=www --ip-type=ipv4

# 启动服务
./ez-ddns-linux-amd64 start
```

### 后台运行（Linux）

```bash
nohup ./ez-ddns-linux-amd64 start > ddns.log 2>&1 &
```

查看日志：
```bash
tail -f ddns.log
```

### 使用 systemd 服务（推荐）

创建服务文件 `/etc/systemd/system/ddns.service`：

```ini
[Unit]
Description=EZ DDNS
After=network.target

[Service]
Type=simple
WorkingDirectory=/opt/ez-ddns
ExecStart=/opt/ez-ddns/ez-ddns-linux-amd64 start --loglevel=info
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

启动服务：

```bash
sudo systemctl daemon-reload
sudo systemctl enable ddns
sudo systemctl start ddns
sudo systemctl status ddns
```

> **注意**：配置和域名通过CLI命令管理，无需在systemd服务中设置环境变量。

## 命令参考

### 全局命令

| 命令 | 说明 |
|------|------|
| `ez-ddns start [--loglevel=<level>]` | 启动DDNS服务 |
| `ez-ddns status` | 查看系统状态 |
| `ez-ddns loglevel <debug/info/error>` | 设置日志级别 |
| `ez-ddns help` | 显示帮助信息 |

### 配置管理命令

| 命令 | 说明 |
|------|------|
| `ez-ddns config list` | 列出所有配置 |
| `ez-ddns config create <id> --access-key-id=<key> --access-key-secret=<secret> [--interval=<seconds>] [--provider=<provider>]` | 创建配置 |
| `ez-ddns config update <id> [--access-key-id=<key>] [--access-key-secret=<secret>] [--interval=<seconds>] [--provider=<provider>]` | 更新配置 |
| `ez-ddns config delete <id>` | 删除配置 |

### 域名管理命令

| 命令 | 说明 |
|------|------|
| `ez-ddns domain list <config-id>` | 列出指定配置的域名 |
| `ez-ddns domain add <config-id> --domain-name=<domain> --rr=<record> [--ip-type=<ipv4\|ipv6>]` | 添加域名 |
| `ez-ddns domain remove <config-id> --domain-name=<domain> --rr=<record> [--ip-type=<ipv4\|ipv6>]` | 删除域名 |

## GitHub Actions 构建

项目配置了自动化构建流程，每次推送代码都会自动构建。

### 提交规范（Conventional Commits）

为了自动生成清晰的更新日志，建议使用以下提交规范：

```
<type>: <description>

[optional body]

[optional footer]
```

#### Type 类型说明

| Type | 说明 | 示例 |
|------|------|------|
| `feat` | 新功能 | `feat: 添加轮询间隔配置` |
| `fix` | 修复bug | `fix: 修复IP获取失败问题` |
| `docs` | 文档变更 | `docs: 更新README说明` |
| `style` | 代码格式（不影响功能） | `style: 格式化代码` |
| `refactor` | 重构 | `refactor: 优化DNS更新逻辑` |
| `perf` | 性能优化 | `perf: 减少API调用次数` |
| `test` | 测试相关 | `test: 添加单元测试` |
| `chore` | 构建过程或辅助工具变动 | `chore: 更新依赖版本` |
| `ci` | CI/CD配置变更 | `ci: 优化构建流程` |
| `build` | 构建系统变更 | `build: 更新Go版本` |

### 构建产物

- `ez-ddns-windows-amd64.exe` - Windows 64位版本
- `ez-ddns-linux-amd64` - Linux 64位版本

## 开发

### 技术栈

- **语言**: Go 1.25.3
- **数据库**: SQLite
- **DNS SDK**: 阿里云 alidns-20150109/v5
- **定时器**: timingwheel

### 项目结构

```
ez-ddns/
├── dao/                    # 数据访问层
│   ├── config_dao.go       # 配置数据访问
│   ├── domain_dao.go       # 域名数据访问
│   ├── system_dao.go       # 系统设置数据访问
│   └── dao.go              # DAO初始化
├── model/                  # 数据模型
│   └── types.go            # 类型定义
├── cli_handler.go          # CLI命令处理
├── ddns_service.go         # DDNS服务核心逻辑
├── dns.go                  # DNS解析工具
├── provider.go             # DNS提供商接口
├── resolver.go             # IP解析工具
├── task_scheduler.go       # 任务调度器
├── logger.go               # 日志管理
├── main.go                 # 主入口
└── go.mod                  # 依赖管理
```

### 本地调试

在 VSCode 中按 `F5` 启动调试。

### 手动构建

```bash
# Windows
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o ez-ddns-windows-amd64.exe .

# Linux
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ez-ddns-linux-amd64 .
```

## 工作原理

DDNS 更新流程：

1. **获取公网IP**：通过多个公共服务获取当前公网 IP
2. **本地DNS预检查**：通过系统DNS解析域名，验证当前IP是否已正确指向（减少阿里云API调用）
3. **查询阿里云记录**：如果DNS预检查未通过，查询阿里云DNS解析记录
4. **记录比对**：检查阿里云记录中是否已有当前公网IP
5. **执行更新**：删除旧记录并添加新的A/AAAA记录

## 注意事项

1. **安全提示**：不要将 AccessKey 硬编码在代码中或提交到版本控制系统
2. **权限要求**：确保 AccessKey 具有阿里云 DNS 的管理权限（AliyunDNSFullAccess）
3. **API 限制**：注意阿里云 API 的调用频率限制，建议轮询间隔不低于 60 秒
4. **日志监控**：建议定期检查运行日志，确保服务正常运行
5. **IP 检测**：程序通过多个公共 IP 服务获取公网 IP，请确保网络可达
6. **域名配置**：`RR` 设置为 `@` 表示主域名，其他值表示子域名（如 `www`、`home` 等）
7. **多记录支持**：程序会检查所有匹配的解析记录，如果当前 IP 已存在则跳过更新
8. **DNS缓存**：本地DNS预检查可能受DNS缓存影响，若刚更新过记录需等待缓存过期
9. **双栈配置**：如需同时支持IPv4和IPv6，请为同一域名添加两个记录（分别指定ipv4和ipv6）

## 原创声明

本项目由 **林展毅 (Johnny Lin)** 原创开发。

欢迎进行二次开发和修改，但请务必保留原作者署名，且未经许可不得以本项目名义进行商业宣传。

## License

Apache License 2.0