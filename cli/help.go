// Copyright (c) 2026 Johnny Lin (林展毅)
// 本项目由林展毅原创开发，欢迎二次开发，但请务必保留原作者署名
// 未经许可不得以本项目名义进行商业宣传

package cli

const helpText = `ez-ddns - 动态DNS更新服务

使用方法: ez-ddns [命令] [参数]

命令列表:
  start     启动DDNS HTTP服务
  config    管理DNS提供商配置
  domain    管理域名记录
  system    管理系统配置
  ddns      DDNS操作（触发、启动、停止）
  help      显示此帮助信息

-------------------------------------------------------------------------------
start 命令
-------------------------------------------------------------------------------
用途: 启动DDNS服务端

参数:
  -port       HTTP服务端口（默认: 8080）
  -api-key    API密钥，用于启用认证保护
  -auto-ddns  启动时自动开启定时DDNS更新

示例:
  ez-ddns start
  ez-ddns start -port 8080 -api-key my-secret-key
  ez-ddns start -port 8080 -auto-ddns

-------------------------------------------------------------------------------
config 命令
-------------------------------------------------------------------------------
用途: 管理DNS提供商配置（如阿里云AccessKey）

子命令:
  create    创建新配置
  get       根据ID获取单个配置（含域名详情）
  list      获取配置列表（不带参数）或单个配置详情（带配置ID）
  update    更新指定配置
  delete    删除指定配置

参数:
  -f string   指定配置文件路径（create/update子命令必需）

示例:
  ez-ddns config create -f config.json      # 创建配置
  ez-ddns config get my-config-id           # 获取单个配置详情
  ez-ddns config list                       # 获取所有配置列表（简短）
  ez-ddns config list my-config-id          # 获取单个配置详情（含域名）
  ez-ddns config update my-config-id -f new-config.json  # 更新配置
  ez-ddns config delete my-config-id        # 删除配置

配置文件格式 (config.json):
{
  "ID": "my-aliyun",
  "AccessKeyId": "your-key",
  "AccessKeySecret": "secret",
  "Provider": "aliyun",
  "Interval": 300,
  "Domains": [
    {
      "DomainName": "example.com",
      "RR": "www",
      "IPType": "ipv4"
    }
  ]
}
配置项说明:
  ID              - 配置唯一标识
  AccessKeyId     - DNS服务商的AccessKey ID（输出时自动脱敏）
  AccessKeySecret - DNS服务商的AccessKey Secret（输出时自动脱敏）
  Provider        - 服务商类型 (aliyun)
  Interval        - DDNS更新间隔（秒）
  Domains         - 关联的域名列表

-------------------------------------------------------------------------------
domain 命令
-------------------------------------------------------------------------------
用途: 管理域名记录（需关联到某个config）

子命令:
  create    创建新域名记录（需指定所属配置ID）
  update    更新指定域名记录
  delete    删除指定域名记录

参数:
  -f string   指定域名配置文件路径（create/update子命令必需）

示例:
  ez-ddns domain create config-id -f domain.json  # 创建域名（绑定到指定配置）
  ez-ddns domain update domain-id -f new-domain.json  # 更新域名
  ez-ddns domain delete domain-id           # 删除域名

域名配置文件格式 (domain.json):
{
  "DomainName": "example.com",
  "RR": "@",
  "IPType": "ipv4"
}
配置项说明:
  DomainName - 域名（如 example.com）
  RR         - 主机记录（@表示根域名，www表示二级域名）
  IPType     - IP类型 (ipv4/ipv6)

-------------------------------------------------------------------------------
system 命令
-------------------------------------------------------------------------------
用途: 管理系统级配置

子命令:
  get       获取系统配置
  update    更新系统配置

参数:
  -f string   指定系统配置文件路径（update子命令必需）

示例:
  ez-ddns system get                        # 获取系统配置
  ez-ddns system update -f system.json      # 更新系统配置

系统配置文件格式 (system.json):
{
  "LogLevel": "info"
}
配置项说明:
  LogLevel - 日志级别 (debug/info/error)，不区分大小写

-------------------------------------------------------------------------------
ddns 命令
-------------------------------------------------------------------------------
用途: 手动触发或控制自动DDNS更新

子命令:
  trigger   立即触发指定配置的DDNS更新
  start     启动自动DDNS定时更新服务
  stop      停止自动DDNS定时更新服务

示例:
  ez-ddns ddns trigger config-id            # 触发DDNS更新
  ez-ddns ddns start                        # 启动自动DDNS服务
  ez-ddns ddns stop                         # 停止自动DDNS服务

-------------------------------------------------------------------------------
远程操作参数
-------------------------------------------------------------------------------
适用于: config / domain / system / ddns

  -server string   远程服务器地址（默认: http://localhost:8080）
  -api-key string  服务器API密钥（如果服务器启用了认证）

示例:
  ez-ddns config list -server http://192.168.1.100:8080
  ez-ddns config list my-config -server http://remote:8080 -api-key xxx
  ez-ddns domain create -f domain.json -server http://remote:8080

-------------------------------------------------------------------------------
完整使用示例
-------------------------------------------------------------------------------
1. 启动服务（带自动DDNS和认证）:
   ez-ddns start -port 8080 -auto-ddns -api-key secure123

2. 创建DNS配置（包含域名）:
   ez-ddns config create -f config.json

3. 查看所有配置:
   ez-ddns config list

4. 查看单个配置详情（含域名）:
   ez-ddns config list my-config-id

5. 添加独立域名:
   ez-ddns domain create -f domain.json

6. 更新域名配置:
   ez-ddns domain update domain-id -f updated-domain.json

7. 手动触发DDNS更新:
   ez-ddns ddns trigger my-config-id

8. 远程管理:
   ez-ddns config list -server http://remote:8080 -api-key xxx
`
