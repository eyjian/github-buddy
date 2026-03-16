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

### 1. 编程语言：C++17

**选择**：使用 C++17 开发
**理由**：
- 可编译为单二进制文件，无运行时依赖，满足"下载即用"需求
- 高性能网络操作（并发 IP 检测需要高效的 socket 编程）
- 跨平台编译支持良好（通过 CMake 管理）
- 用户规则指定为 C++ 开发

**备选方案**：
- Go：天然跨平台编译，但用户规则指定 C++
- Rust：内存安全但学习曲线高

### 2. 构建系统：CMake

**选择**：使用 CMake 作为构建系统
**理由**：
- 业界标准的 C++ 跨平台构建工具
- 支持 Windows（MSVC）、macOS（Clang）、Linux（GCC）多编译器
- 丰富的第三方库集成支持

### 3. IP 数据源：GitHub520 + 多源聚合

**选择**：主要从 GitHub520 项目获取候选 IP，支持多数据源扩展
**理由**：
- GitHub520 是公开、活跃维护的 GitHub IP 列表源
- 支持 JSON/文本格式解析，易于程序处理
- 设计为可扩展架构，后续可接入其他 IP 数据源

**备选方案**：
- 仅依赖 DNS 解析：不够可靠，可能返回被封锁的 IP
- 自建 IP 探测服务：维护成本高，不符合轻量级定位

### 4. IP 检测策略：并发多维度检测

**选择**：使用 std::thread/std::async 并发检测，同时检查 ICMP（ping）、TCP 443、TCP 22
**理由**：
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

### 7. 网络库：平台原生 socket API + libcurl

**选择**：底层用平台原生 socket 做 TCP 端口检测，用 libcurl 做 HTTP 数据获取
**理由**：
- 原生 socket 对 TCP 端口检测最直接、最高效
- libcurl 是成熟的跨平台 HTTP 客户端库，用于从数据源获取 IP 列表
- ICMP ping 使用平台原生实现（Linux/macOS: raw socket，Windows: IcmpSendEcho）

### 8. 命令行解析：CLI11

**选择**：使用 CLI11 库进行命令行参数解析
**理由**：
- 头文件库（header-only），无需额外编译
- 支持子命令模式，完美匹配 init/update/rollback/status 命令结构
- 自动生成帮助文档

### 9. 日志库：spdlog

**选择**：使用 spdlog 作为日志框架
**理由**：
- 高性能、头文件优先的 C++ 日志库
- 支持多后端（控制台 + 文件轮转）
- 格式化输出，支持日志级别控制

## Risks / Trade-offs

### [ICMP 权限限制] → 降级为 TCP-only 检测
- **风险**：部分系统（尤其是容器环境）禁止非 root 用户发送 ICMP 包
- **缓解**：检测到 ICMP 不可用时，自动降级为仅 TCP 端口检测，并在日志中提示

### [IP 数据源不可用] → 使用本地缓存
- **风险**：GitHub520 等数据源网络不可达时无法获取新 IP
- **缓解**：工具内置一份默认 IP 列表作为兜底；优先使用本地缓存；多数据源 failover

### [hosts 文件权限问题] → 明确错误提示
- **风险**：用户未以管理员身份运行，无法修改 hosts
- **缓解**：检测权限并给出明确操作指引（Linux: "请使用 sudo 运行"，Windows: "请以管理员身份运行"）

### [跨平台兼容性] → 充分的平台抽象层
- **风险**：不同 OS 的 hosts 路径、提权方式、socket API 差异大
- **缓解**：设计平台抽象层（Platform Abstraction Layer），将 OS 特定逻辑隔离在独立模块中

### [自动更新可靠性] → 守护进程 vs 启动时检查
- **风险**：后台守护进程实现复杂，且在不同 OS 上注册方式不同
- **trade-off**：v1 版本采用"启动时检查 + 缓存过期"策略，暂不实现持久化守护进程，简化实现
