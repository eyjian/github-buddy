## Context

GitHub-Buddy 是一个全新的跨平台命令行工具项目，旨在通过智能维护 GitHub 域名-IP 映射并修改系统 hosts 文件，解决开发者执行 `go mod tidy` 和 `git clone` 时的网络访问问题。

当前状态：项目从零开始，尚无现有代码库。目标用户为日常使用 Go/Git 的开发者，操作系统覆盖 Windows 10+、macOS 12+、主流 Linux 发行版。

约束条件：
- 工具为单二进制文件，无额外运行时依赖
- 需要管理员/root 权限修改 hosts 文件
- 不做代理/VPN 功能，仅通过 hosts 优化
- 遵循 MIT 开源协议

## Goals / Non-Goals

**Goals:**
- 提供可靠的 GitHub IP 检测机制（ICMP + TCP 端口检测），筛选出最优 IP
- 实现安全的跨平台 hosts 文件管理，支持备份/回滚/增量更新
- 提供简洁的 CLI 命令（init/update/rollback/status），开箱即用
- 实现后台自动更新机制，减少人工干预
- 确保工具稳定性（失败率＜1%），自动切换备选 IP

**Non-Goals:**
- 不提供代理/翻墙功能
- 不处理 GitHub 以外的域名（如 golang.org、goproxy.io）
- 不提供图形界面（仅 CLI）
- 不支持多用户管理或企业级部署
- 不负责 Git/Go 工具链本身的安装和配置

## Decisions

### 1. 编程语言：Go

**选择**：使用 Go 开发
**理由**：
- 天然跨平台编译，`GOOS/GOARCH` 一条命令即可交叉编译 Windows/macOS/Linux 二进制
- 编译产物为静态链接的单二进制文件，零运行时依赖，满足"下载即用"需求
- 标准库内置 `net/http`、`net`、`os` 等模块，网络编程和文件操作无需第三方依赖
- goroutine 天然并发模型，并发 IP 检测实现简洁高效
- Go 生态拥有成熟的 CLI 工具链（cobra、viper 等），Docker/Kubernetes/Hugo 等知名 CLI 均用 Go 编写
- 目标用户群体（Go/Git 开发者）与 Go 生态高度契合

**备选方案**：
- C++17：性能优异，但跨平台编译复杂、依赖管理困难、开发效率低
- Rust：内存安全但学习曲线高，生态成熟度略逊于 Go 的 CLI 领域

### 2. 项目结构：Go Module

**选择**：使用 Go Module 管理项目依赖
**理由**：
- Go 官方内置的依赖管理方案，`go mod tidy` 一键管理
- 项目结构遵循 Go 社区标准布局：`cmd/`（入口）、`internal/`（内部包）、`pkg/`（可导出包）
- 无需额外构建工具（如 CMake/Makefile），`go build` 即可编译

### 3. IP 数据源：GitHub520 + 多源聚合

**选择**：主要从 GitHub520 项目获取候选 IP，支持多数据源扩展
**理由**：
- GitHub520 是公开、活跃维护的 GitHub IP 列表源
- 支持 JSON/文本格式解析，易于程序处理
- 设计为可扩展架构，后续可接入其他 IP 数据源

**备选方案**：
- 仅依赖 DNS 解析：不够可靠，可能返回被封锁的 IP
- 自建 IP 探测服务：维护成本高，不符合轻量级定位

### 4. IP 检测策略：goroutine 并发多维度检测

**选择**：使用 goroutine + channel 并发检测，同时检查 ICMP（ping）、TCP 443、TCP 22
**理由**：
- goroutine 轻量级并发模型，可轻松启动数百个并发检测任务
- 并发检测可在＜10s 内完成全量 IP 检测
- 三维度检测确保 IP 真正可用（仅 ping 通不代表 HTTPS/SSH 可用）
- 质量评分机制（延迟 + 丢包率 + 端口状态）自动选择最优 IP

### 5. Hosts 文件管理：标记区块 + 增量更新

**选择**：使用 `# GitHub-Buddy Auto-Generated Start/End` 注释标记隔离工具管理区块
**理由**：
- 不干扰用户手动添加的 hosts 配置
- 明确标识工具管理的内容，便于调试和手动清理
- 增量更新仅替换标记区块内的内容

### 6. 数据存储：本地文件系统 (~/.github-buddy/)

**选择**：将配置、缓存、日志存储在用户主目录下的 `.github-buddy/` 目录
**理由**：
- 符合 Unix/XDG 惯例，不污染系统目录
- 日志存于 `~/.github-buddy/logs/`，缓存存于 `~/.github-buddy/cache/`
- 配置文件为 JSON 格式，方便用户手动查看和修改

### 7. 网络与检测库：Go 标准库

**选择**：使用 Go 标准库 `net/http`（HTTP 请求）、`net`（TCP 端口检测）、`os/exec`（ICMP ping）
**理由**：
- Go 标准库的 `net/http` 功能完备，无需引入 libcurl 等外部依赖
- `net.DialTimeout` 直接实现 TCP 端口检测，简洁高效
- ICMP ping 通过调用系统 `ping` 命令实现跨平台兼容（避免 raw socket 权限问题）
- 零外部依赖，编译产物更小更纯净

### 8. 命令行框架：cobra

**选择**：使用 cobra 库进行命令行参数解析
**理由**：
- Go 生态最流行的 CLI 框架，Docker/Kubernetes/Hugo 等均使用
- 支持子命令模式，完美匹配 init/update/rollback/status 命令结构
- 自动生成帮助文档和 shell 补全
- 搭配 viper 可实现配置文件自动加载

### 9. 日志库：zerolog

**选择**：使用 zerolog 作为日志框架
**理由**：
- 高性能、零内存分配的结构化日志库
- 支持多输出后端（控制台 + 文件）
- 支持日志级别控制（Debug/Info/Warn/Error）
- JSON 格式日志输出，便于后续分析

## Risks / Trade-offs

### [ICMP 权限限制] → 降级为 TCP-only 检测
- **风险**：部分系统（尤其是容器环境）禁止非 root 用户发送 ICMP 包
- **缓解**：通过调用系统 `ping` 命令而非 raw socket 规避权限问题；仍无法执行时自动降级为仅 TCP 端口检测，并在日志中提示

### [IP 数据源不可用] → 使用本地缓存
- **风险**：GitHub520 等数据源网络不可达时无法获取新 IP
- **缓解**：工具内置一份默认 IP 列表作为兜底；优先使用本地缓存；多数据源 failover

### [hosts 文件权限问题] → 明确错误提示
- **风险**：用户未以管理员身份运行，无法修改 hosts
- **缓解**：检测权限并给出明确操作指引（Linux: "请使用 sudo 运行"，Windows: "请以管理员身份运行"）

### [跨平台兼容性] → Go 编译标签 + 平台抽象层
- **风险**：不同 OS 的 hosts 路径、提权方式差异大
- **缓解**：利用 Go 的编译标签（build tags）和 `runtime.GOOS` 实现平台特定逻辑隔离，将 OS 差异封装在 `internal/platform/` 包中

### [自动更新可靠性] → 守护进程 vs 启动时检查
- **风险**：后台守护进程实现复杂，且在不同 OS 上注册方式不同
- **trade-off**：v1 版本采用"启动时检查 + 缓存过期"策略，暂不实现持久化守护进程，简化实现
