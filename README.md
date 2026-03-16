
# GitHub-Buddy

[![Go](https://img.shields.io/badge/Go-1.21+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey)]()

**GitHub-Buddy** 是一个跨平台命令行工具，通过智能维护 GitHub 域名-IP 映射并修改系统 hosts 文件，解决开发者执行 `go mod tidy`、`git clone` 等操作时频繁遇到的 GitHub 网络访问问题。

> 🚫 本工具**不是**代理/VPN，仅通过优化 DNS 解析（hosts）来改善连接质量。

---

## ✨ 功能特性

- 🔍 **智能 IP 检测** — 多维度并发检测（ICMP Ping / TCP 443 / TCP 22），自动评分筛选最优 IP
- 🌐 **全量域名覆盖** — 维护 `github.com`、`raw.githubusercontent.com`、`api.github.com` 等 11 个核心域名
- 🖥️ **跨平台支持** — 一套代码覆盖 Windows 10+、macOS 12+、主流 Linux 发行版
- 🔒 **安全保障** — 自动备份 hosts 文件，支持一键回滚，SHA256 完整性校验
- 📦 **零依赖分发** — 静态编译的单二进制文件，下载即用
- 🔄 **自动更新** — 默认每 6 小时自动检查 IP 有效性，支持缓存和手动强制更新
- 🏷️ **隔离管理** — 使用注释标记区块隔离工具修改内容，不干扰用户手动配置

---

## 📥 安装

### 方式一：从源码编译（推荐）

确保已安装 [Go 1.21+](https://go.dev/dl/)：

```bash
go install github.com/eyjian/github-buddy/cmd/github-buddy@latest
```

### 方式二：手动编译

```bash
git clone https://github.com/eyjian/github-buddy.git
cd github-buddy
make build
```

编译产物为当前目录下的 `github-buddy` 可执行文件。

### 方式三：下载预编译二进制

前往 [Releases](https://github.com/eyjian/github-buddy/releases) 页面，下载对应平台的二进制文件。

---

## 🚀 快速开始

### 1. 初始化

首次使用时运行 `init` 命令，自动检测最优 IP 并更新 hosts：

```bash
# Linux / macOS 需要 sudo
sudo github-buddy init

# Windows（以管理员身份运行命令提示符）
github-buddy init
```

### 2. 查看状态

查看当前 hosts 中 GitHub 域名的映射状态和连通性：

```bash
github-buddy status
```

### 3. 手动更新

强制重新检测并更新 IP（自动备份当前 hosts）：

```bash
sudo github-buddy update
```

### 4. 回滚

如遇问题，一键恢复到上次备份的 hosts 文件：

```bash
sudo github-buddy rollback
```

---

## 📖 命令参考

| 命令 | 说明 |
|------|------|
| `github-buddy init` | 初始化工具，首次检测最优 IP 并更新 hosts |
| `github-buddy update` | 强制重新检测 IP，备份并更新 hosts |
| `github-buddy rollback` | 从最近的备份恢复 hosts 文件 |
| `github-buddy status` | 查看当前域名映射状态和实时连通性 |

### 全局参数

| 参数 | 说明 |
|------|------|
| `-v, --verbose` | 输出详细日志到控制台 |
| `-f, --force` | 跳过确认提示，强制执行 |
| `--version` | 显示版本信息 |
| `-h, --help` | 显示帮助信息 |

---

## ⚙️ 配置说明

工具的配置文件和数据存储在用户主目录下：

```
~/.github-buddy/
├── config.json     # 配置文件
├── cache/          # IP 检测缓存
├── backups/        # hosts 备份文件
└── logs/           # 运行日志
```

### 配置文件示例（`config.json`）

```json
{
  "update_interval_hours": 6,
  "data_sources": [
    {
      "name": "GitHub520",
      "url": "https://raw.hellogithub.com/hosts",
      "priority": 1,
      "enabled": true
    }
  ],
  "domains": [
    "github.com",
    "ssh.github.com",
    "gist.github.com",
    "raw.githubusercontent.com",
    "api.github.com",
    "assets-cdn.github.com",
    "github.global.ssl.fastly.net",
    "collector.github.com",
    "avatars.githubusercontent.com",
    "codeload.github.com"
  ]
}
```

### 配置项说明

| 字段 | 类型 | 默认值 | 说明 |
|------|------|--------|------|
| `update_interval_hours` | int | `6` | 自动检查 IP 的间隔（小时） |
| `data_sources` | array | GitHub520 | IP 数据源列表，支持多源扩展 |
| `domains` | array | 11 个域名 | 需要维护的 GitHub 域名清单 |

---

## 🔧 Hosts 文件管理

工具使用注释标记区块来隔离管理的内容，**不会修改标记区块外的任何内容**：

```
# 用户原有的配置不受影响
127.0.0.1  myapp.local

# GitHub-Buddy Auto-Generated Start
# 更新时间: 2024-01-01 12:00:00
140.82.121.3    github.com
185.199.108.133 raw.githubusercontent.com
140.82.121.5    api.github.com
# ... 更多域名映射
# GitHub-Buddy Auto-Generated End
```

---

## 🏗️ 项目结构

```
github-buddy/
├── cmd/
│   └── github-buddy/          # CLI 入口和命令定义
│       ├── main.go            # 程序入口
│       ├── root.go            # 根命令和全局参数
│       ├── cmd_init.go        # init 命令
│       ├── cmd_update.go      # update 命令
│       ├── cmd_rollback.go    # rollback 命令
│       └── cmd_status.go      # status 命令
├── internal/                  # 内部核心模块
│   ├── detector/              # IP 检测与评分
│   │   ├── detector.go        # 检测调度器
│   │   ├── source.go          # 数据源获取
│   │   ├── ping.go            # ICMP Ping 检测
│   │   ├── tcp.go             # TCP 端口检测
│   │   ├── scorer.go          # 质量评分
│   │   ├── defaults.go        # 兜底 IP 列表
│   │   └── types.go           # 数据类型定义
│   ├── hosts/                 # Hosts 文件管理
│   │   ├── reader.go          # 读取解析
│   │   ├── writer.go          # 写入更新
│   │   ├── block.go           # 标记区块管理
│   │   └── conflict.go        # 冲突检测
│   ├── backup/                # 备份与回滚
│   │   └── backup.go          # 备份/恢复/校验
│   ├── cache/                 # IP 缓存管理
│   │   └── cache.go           # 缓存读写和过期检查
│   ├── config/                # 配置管理
│   │   └── config.go          # 配置加载/保存/默认值
│   ├── logger/                # 日志系统
│   │   └── logger.go          # zerolog + lumberjack
│   ├── platform/              # 平台抽象层
│   │   └── platform.go        # OS 检测和路径适配
│   └── storage/               # 存储管理
│       └── storage.go         # 目录结构管理
├── Makefile                   # 构建脚本
├── go.mod                     # Go Module 依赖
└── go.sum                     # 依赖校验
```

---

## 🔨 构建指南

### 构建当前平台

```bash
make build
```

### 交叉编译全平台

```bash
make build-all
```

编译产物输出到 `dist/` 目录：

```
dist/
├── github-buddy-linux-amd64
├── github-buddy-linux-arm64
├── github-buddy-darwin-amd64
├── github-buddy-darwin-arm64
└── github-buddy-windows-amd64.exe
```

### 运行测试

```bash
# 详细输出
make test

# 简洁输出
make test-short
```

### 代码检查

```bash
make lint
```

---

## 🛡️ 安全说明

- **权限要求**：修改 hosts 文件需要管理员/root 权限，工具会自动检测并给出提权指引
- **自动备份**：每次修改 hosts 前自动创建带时间戳的备份文件，存储于 `~/.github-buddy/backups/`
- **完整性校验**：备份文件使用 SHA256 校验，回滚时自动验证文件完整性
- **隔离修改**：仅修改标记区块内的内容，绝不触碰用户手动配置的 hosts 条目
- **兜底机制**：内置默认 IP 列表，即使所有数据源不可达也能正常工作

---

## 🤝 贡献指南

欢迎贡献代码！请遵循以下步骤：

1. Fork 本仓库
2. 创建功能分支：`git checkout -b feature/your-feature`
3. 提交改动：`git commit -m "feat: 添加某功能"`
4. 推送分支：`git push origin feature/your-feature`
5. 创建 Pull Request

### 开发环境

- Go 1.21+
- Make（可选，用于构建脚本）

### 代码规范

- 使用 `gofmt` 格式化代码
- 使用 `go vet` 进行静态检查
- 新功能需附带单元测试

---

## 📋 技术栈

| 组件 | 技术选型 | 说明 |
|------|----------|------|
| 编程语言 | Go | 交叉编译、静态链接、goroutine 并发 |
| CLI 框架 | [cobra](https://github.com/spf13/cobra) | 子命令、自动帮助文档、shell 补全 |
| 日志框架 | [zerolog](https://github.com/rs/zerolog) | 零分配、结构化日志 |
| 日志轮转 | [lumberjack](https://github.com/natefinish/lumberjack) | 日志文件自动轮转和压缩 |
| 网络检测 | Go 标准库 `net` | TCP 端口检测、HTTP 请求 |
| 构建工具 | Make + Goreleaser | 交叉编译和自动化发布 |

---

## 📄 许可证

本项目采用 [MIT License](LICENSE) 开源协议。

---

## 🙏 致谢

- [GitHub520](https://github.com/521xueweihan/GitHub520) — 提供 GitHub 域名 IP 数据源
- [cobra](https://github.com/spf13/cobra) — 优秀的 Go CLI 框架
- [zerolog](https://github.com/rs/zerolog) — 高性能结构化日志库
