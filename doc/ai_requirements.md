# 变更提议：开发 GitHub-Buddy 工具解决 go mod/git 访问 GitHub 网络问题
## 1. 变更名称
GitHub-Buddy - 基于 hosts 优化的 GitHub 网络访问工具

## 2. 变更目标
开发一款轻量级工具，通过智能维护 GitHub 域名-IP 映射并修改系统 hosts 文件，彻底解决执行 `go mod tidy` 和 `git clone` 时访问 GitHub 的网络不通、卡顿、超时问题，确保：
- 核心操作（go mod tidy/git clone）100% 网络可通；
- 访问延迟＜50ms，无明显卡顿/丢包；
- 工具易用、跨平台、无操作风险。

## 3. 背景与核心痛点
### 3.1 业务背景
开发者执行 `go mod tidy`（拉取 Go 模块）和 `git clone`（克隆 GitHub 仓库）时，因 GitHub 域名解析问题频繁出现：
- 网络不通：连接超时、拒绝访问；
- 访问卡顿：拉取速度慢（＜100KB/s）、中途断连；
- 手动改 hosts 痛点：仅改 github.com 无效、IP 易失效、跨平台操作复杂、易误改原有配置。

### 3.2 核心痛点
1. 域名覆盖不全：仅修改 github.com 无法解决 go mod/git 依赖的子域名问题；
2. IP 质量无保障：仅 ping 通的 IP 可能端口不通/延迟高；
3. 操作风险高：直接修改 hosts 易丢失原有配置，无回滚机制；
4. 跨平台适配差：仅支持 Linux/macOS，无 Windows 适配；
5. IP 需手动维护：GitHub IP 动态变化，手动改 hosts 频繁失效。

## 4. 核心功能需求
### 4.1 域名覆盖（核心）
工具默认维护 GitHub 全量核心域名清单（覆盖 go mod/git 场景），包括：
- 基础域名：github.com、ssh.github.com、gist.github.com；
- Go 模块专属：raw.githubusercontent.com、api.github.com；
- 静态资源：assets-cdn.github.com。

### 4.2 智能 IP 检测与筛选
1. 多维度检测：对候选 IP 同时检测「连通性（ping）」「443 端口（HTTPS）」「22 端口（SSH）」；
2. 质量筛选：自动计算 IP 平均延迟、丢包率，筛选出「延迟＜50ms + 丢包率=0 + 端口全通」的最优 IP；
3. 兜底机制：保留 2-3 个备选 IP，最优 IP 失效时自动切换。

### 4.3 跨平台 hosts 管理
1. 系统自动识别：
   - Linux/macOS：操作 `/etc/hosts`；
   - Windows：操作 `C:\Windows\System32\drivers\etc\hosts`；
2. 自动提权：
   - Linux/macOS：触发 `sudo` 获取修改 hosts 权限；
   - Windows：提示「以管理员身份运行」，无权限时终止并给出指引；
3. 冲突处理：
   - 读取现有 hosts 中 GitHub 域名配置，提示用户选择「保留最优」或「覆盖」；
   - 工具修改的内容添加专属注释标记 `# GitHub-Buddy Auto-Generated Start/End`，仅更新该区块内容，不干扰用户手动配置。

### 4.4 安全保障
1. 自动备份：首次运行/每次修改前，将原有 hosts 备份为 `hosts.bak`（同目录）；
2. 一键回滚：提供 `github-buddy rollback` 命令，快速恢复到备份的原始 hosts 状态；
3. 增量更新：仅修改 GitHub 相关域名映射，不覆盖/删除 hosts 中其他配置。

### 4.5 自动更新
1. 定时检测：默认每 6 小时后台检测最优 IP，仅当检测到更优 IP 时更新 hosts；
2. 手动触发：提供 `github-buddy update` 命令，支持用户强制更新 IP 列表；
3. 缓存机制：缓存最近一次最优 IP 列表，启动时直接使用缓存，后台异步更新。

### 4.6 基础命令
| 命令                | 功能说明                     |
|---------------------|------------------------------|
| `github-buddy init` | 初始化工具，备份 hosts 并检测最优 IP |
| `github-buddy update` | 手动强制更新 IP 并修改 hosts |
| `github-buddy rollback` | 回滚到原始 hosts 配置        |
| `github-buddy status` | 查看当前 IP 状态、延迟、端口连通性 |

## 5. 非功能需求
### 5.1 性能
- IP 检测耗时＜10s（单次全量检测）；
- 工具启动/命令执行耗时＜1s（缓存生效时）；
- `go mod tidy`/`git clone` 速度≥1MB/s（最优 IP 生效后）。

### 5.2 兼容性
- 操作系统：Windows 10+/macOS 12+/Linux（CentOS/Ubuntu/Debian）；
- Go 版本：兼容 Go 1.18+（主流开发版本）；
- Git 版本：兼容 Git 2.30+。

### 5.3 易用性
- 无额外依赖：工具为单二进制文件，下载即可运行；
- 错误提示：操作失败时给出明确指引（如“无管理员权限，请以管理员运行”）；
- 日志记录：关键操作（IP 检测、hosts 修改、回滚）记录到 `~/.github-buddy/logs`，方便排查问题。

### 5.4 稳定性
- 工具运行失败率＜1%；
- IP 失效时自动切换备选 IP，无人工干预下恢复时间＜1min。

## 6. 验收标准（可量化、可验证）
| 验收项                          | 验证方式                                                                 |
|---------------------------------|--------------------------------------------------------------------------|
| 域名覆盖完整性                  | 检查工具默认域名清单是否包含 4.1 中所有域名                               |
| IP 质量达标                     | 执行 `github-buddy status`，验证最优 IP 延迟＜50ms、丢包率=0、443/22 端口通 |
| 跨平台适配                      | 在 Windows/macOS/Ubuntu 分别执行 init/update/rollback 命令，验证 hosts 正常修改 |
| 安全保障                        | 修改 hosts 后删除工具专属注释块，执行 rollback 验证配置恢复                |
| go mod tidy 可用性              | 清空 Go 模块缓存后执行 `go mod tidy`，验证 100% 拉取成功、无超时          |
| git clone 可用性                | 克隆大仓库（如 https://github.com/golang/go.git），验证速度≥1MB/s、无断连  |
| 自动更新有效性                  | 手动修改 hosts 为失效 IP，等待 6 小时后验证工具自动更新为最优 IP          |

## 7. 实现范围（明确边界）
### 7.1 包含范围
- GitHub 核心域名的 IP 检测与 hosts 管理；
- 跨平台 hosts 操作与权限处理；
- 备份、回滚、自动更新机制；
- 基础命令行交互。

### 7.2 排除范围
- 代理功能（仅通过 hosts 优化，不做代理/翻墙）；
- 其他海外域名（如 golang.org、goproxy.io 等）；
- 图形化界面（仅命令行工具）；
- 私有化部署/企业级功能（如多用户管理）。

## 8. 依赖与约束
### 8.1 技术依赖
- IP 数据源：从公开、可靠的 GitHub IP 列表源（如 GitHub520）获取候选 IP；
- 网络检测：基于 ICMP（ping）、TCP 端口扫描实现 IP 质量检测。

### 8.2 约束
- 需遵守系统权限规则，不绕过管理员权限修改 hosts；
- 不收集用户敏感数据（仅本地存储 IP 列表、日志）；
- 工具为开源项目，代码需托管在 GitHub/Gitee，遵循 MIT 协议。
