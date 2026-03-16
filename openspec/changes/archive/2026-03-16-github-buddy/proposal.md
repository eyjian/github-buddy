## Why

开发者执行 `go mod tidy` 和 `git clone` 时，因 GitHub 域名解析问题频繁出现网络不通、连接超时、拉取卡顿等问题。手动修改 hosts 存在域名覆盖不全、IP 易失效、跨平台操作复杂、易误改原有配置等痛点，亟需一款自动化工具来彻底解决。

## What Changes

- **新增命令行工具 `github-buddy`**：单二进制文件，跨平台（Windows/macOS/Linux）支持
- **智能 IP 检测与筛选**：多维度检测（ping/443端口/22端口），自动筛选延迟＜50ms、丢包率=0 的最优 IP
- **全量 GitHub 域名覆盖**：维护 github.com、ssh.github.com、raw.githubusercontent.com、api.github.com 等核心域名清单
- **跨平台 hosts 管理**：自动识别操作系统、自动提权、冲突处理，使用专属注释标记隔离工具修改区块
- **安全保障机制**：自动备份 hosts、一键回滚、增量更新（不干扰用户手动配置）
- **自动更新机制**：默认每 6 小时后台检测最优 IP，支持缓存和手动强制更新
- **四个核心命令**：`init`（初始化）、`update`（更新）、`rollback`（回滚）、`status`（状态查看）

## Capabilities

### New Capabilities
- `ip-detection`: 智能 IP 检测与质量筛选，包含多维度检测（ICMP/TCP 443/TCP 22）、延迟评估、丢包率计算，以及备选 IP 兜底切换机制
- `hosts-management`: 跨平台 hosts 文件管理，包含系统识别、自动提权、冲突处理、专属注释标记区块管理、增量更新
- `backup-rollback`: hosts 文件安全保障，包含自动备份、一键回滚恢复、备份文件管理
- `auto-update`: 定时自动更新机制，包含后台 IP 检测、缓存管理、手动触发更新
- `cli-commands`: 命令行界面与交互，包含 init/update/rollback/status 四个核心命令及错误提示、日志记录

### Modified Capabilities
（无，此为全新项目）

## Impact

- **新增代码**：整个 Go 项目代码库，包含 IP 检测模块、hosts 管理模块、命令行接口等
- **系统依赖**：需要操作系统管理员/root 权限来修改 hosts 文件
- **网络依赖**：依赖公开 GitHub IP 数据源（如 GitHub520）获取候选 IP 列表
- **文件系统影响**：修改系统 hosts 文件（Linux/macOS: `/etc/hosts`，Windows: `C:\Windows\System32\drivers\etc\hosts`）
- **用户数据目录**：在 `~/.github-buddy/` 下存储日志、缓存、配置
- **兼容性要求**：Windows 10+、macOS 12+、Linux（CentOS/Ubuntu/Debian）；Go 1.21+；Git 2.30+
