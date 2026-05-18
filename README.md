# ez-ddns

基于阿里云 DNS API 的动态域名解析（DDNS）工具，自动检测公网 IP 变化并更新 DNS 记录。

## 功能特性

- ✅ 自动获取当前公网 IP
- ✅ 本地DNS预检查：更新前先通过DNS解析验证，减少运营商解析服务API调用
- ✅ 智能判断：无记录时新增，有记录且 IP 变化时更新
- ✅ 支持多条解析记录管理
- ✅ 持续轮询监控，可配置检查间隔（默认3分钟）
- ✅ 支持配置文件和环境变量两种配置方式（环境变量优先）
- ✅ 支持日志级别配置（DEBUG/INFO/ERROR）
- ✅ 自动保存上次 IP 记录，避免重复更新
- ✅ 支持 Windows 和 Linux 平台
- ✅ 完整的 GitLab CI/CD 自动化构建

## 配置方式

### 方式一：配置文件

创建 `ez-ddns-config.json` 文件：

```json
{
  "accessKeyId": "your-access-key-id",
  "accessKeySecret": "your-access-key-secret",
  "domainName": "example.com",
  "rr": "home",
  "interval": 180,
  "lastIP": "",
  "logLevel": "info"
}
```

### 方式二：环境变量

程序优先读取环境变量，如果环境变量为空则使用配置文件中的值。

| 变量名 | 说明 | 示例 |
|--------|------|------|
| `ALIBABA_CLOUD_ACCESS_KEY_ID` | 阿里云 AccessKey ID | `LTAI5t...` |
| `ALIBABA_CLOUD_ACCESS_KEY_SECRET` | 阿里云 AccessKey Secret | `xxxxx...` |
| `DDNS_DOMAIN_NAME` | 主域名 | `example.com` |
| `DDNS_RR` | 子域名/主机记录 | `home`, `www`, `@` |
| `DDNS_INTERVAL` | 轮询间隔（秒） | `180` |
| `DDNS_LOG_LEVEL` | 日志级别 | `debug`, `info` |

### 配置项说明

| 配置项 | 说明 | 默认值 |
|--------|------|--------|
| `accessKeyId` | 阿里云 AccessKey ID | 必填 |
| `accessKeySecret` | 阿里云 AccessKey Secret | 必填 |
| `domainName` | 主域名 | 必填 |
| `rr` | 子域名/主机记录 | `@` |
| `interval` | 轮询间隔（秒） | `180` |
| `lastIP` | 上次记录的 IP（自动更新） | 空 |
| `logLevel` | 日志级别 | `info` |

### 日志级别

| 级别 | 说明 | 输出内容 |
|------|------|----------|
| `debug` | 调试模式 | 详细的调试信息，适合开发调试 |
| `info` | 信息模式 | 关键操作信息，适合生产环境 |

## 使用方法

### Windows (PowerShell)

```powershell
# 使用配置文件
.\ez-ddns-windows-amd64.exe

# 或使用环境变量覆盖
$env:ALIBABA_CLOUD_ACCESS_KEY_ID = "your-access-key-id"
$env:ALIBABA_CLOUD_ACCESS_KEY_SECRET = "your-access-key-secret"
$env:DDNS_DOMAIN_NAME = "example.com"
$env:DDNS_RR = "home"
$env:DDNS_INTERVAL = "180"
$env:DDNS_LOG_LEVEL = "info"

.\ez-ddns-windows-amd64.exe
```

### Linux

```bash
# 使用配置文件
./ez-ddns-linux-amd64

# 或使用环境变量覆盖
export ALIBABA_CLOUD_ACCESS_KEY_ID="your-access-key-id"
export ALIBABA_CLOUD_ACCESS_KEY_SECRET="your-access-key-secret"
export DDNS_DOMAIN_NAME="example.com"
export DDNS_RR="home"
export DDNS_INTERVAL="180"
export DDNS_LOG_LEVEL="info"

chmod +x ez-ddns-linux-amd64
./ez-ddns-linux-amd64
```

### 后台运行（Linux）

```bash
nohup ./ez-ddns-linux-amd64 > ddns.log 2>&1 &
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
Environment=ALIBABA_CLOUD_ACCESS_KEY_ID=your-access-key-id
Environment=ALIBABA_CLOUD_ACCESS_KEY_SECRET=your-access-key-secret
Environment=DDNS_DOMAIN_NAME=example.com
Environment=DDNS_RR=home
Environment=DDNS_INTERVAL=180
Environment=DDNS_LOG_LEVEL=info
WorkingDirectory=/opt/ez-ddns
ExecStart=/opt/ez-ddns/ez-ddns-linux-amd64
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

## GitLab CI/CD 构建

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

#### 提交示例

```bash
# 新功能
git commit -m "feat: 添加Linux systemd服务支持"

# 修复bug
git commit -m "fix: 修复环境变量解析错误"

# 多个变更
git commit -m "feat: 支持自定义轮询间隔

- 添加DDNS_INTERVAL环境变量
- 默认间隔3分钟
- 支持动态调整"
```

### 构建产物

- `ez-ddns-windows-amd64.exe` - Windows 64位版本
- `ez-ddns-linux-amd64` - Linux 64位版本

### 下载构建产物

1. 进入 GitLab 项目
2. 导航到 **CI/CD → Pipelines**
3. 点击最新的流水线
4. 在 **Jobs** 标签页下载 artifacts

### 发布 Release（打标签时）

当创建 Git 标签时，会自动：
1. 提取从上一个标签以来的所有提交记录
2. 按类型分类（新功能、修复、文档等）
3. 生成格式化的 Changelog
4. 创建 GitLab Release 并附加构建产物

```bash
# 创建标签
git tag v1.0.0

# 推送标签到远程仓库（触发CI/CD）
git push origin v1.0.0
```

Release 会自动包含：
- 📋 分类的更新日志（基于提交规范）
- 🪟 Windows AMD64 可执行文件
- 🐧 Linux AMD64 可执行文件

#### Release 示例输出

```
## 🚀 更新内容

### ✨ 新功能
- feat: 添加轮询间隔配置 (a1b2c3d)
- feat: 支持多平台构建 (e4f5g6h)

### 🐛 修复
- fix: 修复IP获取超时问题 (i7j8k9l)

### 📝 文档
- docs: 更新使用说明 (m0n1o2p)
```

### 查看历史版本

在 GitLab 项目的 **Deployments → Releases** 页面可以查看所有发布的版本和更新日志。

## 开发

### 本地调试

在 VSCode 中按 `F5` 启动调试（需先配置 `.vscode/launch.json` 中的环境变量）。

### 手动构建

```bash
# Windows
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o ez-ddns-windows-amd64.exe .

# Linux
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o ez-ddns-linux-amd64 .
```

## 工作原理

DDNS 更新流程：

1. **获取公网IP**：通过多个公共服务（如 ipify.org、icanhazip.com）获取当前公网 IP
2. **本地DNS预检查**：通过系统DNS解析域名，验证当前IP是否已正确指向（减少阿里云API调用）
3. **查询阿里云记录**：如果DNS预检查未通过，查询阿里云DNS解析记录
4. **记录比对**：检查阿里云记录中是否已有当前公网IP
5. **执行更新**：删除旧记录并添加新的A记录

## 注意事项

1. **安全提示**：不要将 AccessKey 硬编码在代码中或提交到版本控制系统
2. **权限要求**：确保 AccessKey 具有阿里云 DNS 的管理权限（AliyunDNSFullAccess）
3. **API 限制**：注意阿里云 API 的调用频率限制，建议轮询间隔不低于 60 秒
4. **日志监控**：建议定期检查运行日志，确保服务正常运行
5. **IP 检测**：程序通过多个公共 IP 服务获取公网 IP，请确保网络可达
6. **域名配置**：`DDNS_RR` 设置为 `@` 表示主域名，其他值表示子域名（如 `www`、`home` 等）
7. **多记录支持**：程序会检查所有匹配的解析记录，如果当前 IP 已存在则跳过更新
8. **DNS缓存**：本地DNS预检查可能受DNS缓存影响，若刚更新过记录需等待缓存过期

## 原创声明

本项目由 **林展毅 (Johnny Lin)** 原创开发。

欢迎进行二次开发和修改，但请务必保留原作者署名，且未经许可不得以本项目名义进行商业宣传。

## License

Apache License 2.0