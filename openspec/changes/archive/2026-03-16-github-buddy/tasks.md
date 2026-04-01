## 1. 项目基础搭建

- [x] 1.1 初始化 Go Module 项目结构：`go mod init`，创建 `cmd/github-buddy/`（入口）、`internal/`（内部包）、`pkg/`（可导出包）目录
- [x] 1.2 集成第三方库依赖：cobra（命令行框架）、zerolog（日志）、配置用标准库 `encoding/json`
- [x] 1.3 创建平台抽象层：`internal/platform/` 包，利用 Go 编译标签实现跨平台逻辑（OS 类型识别、hosts 路径、home 目录）
- [x] 1.4 创建用户数据目录管理模块：`internal/storage/` 包，负责 `~/.github-buddy/` 目录结构初始化（config/cache/logs/backups）

## 2. IP 检测模块（ip-detection）

- [x] 2.1 实现 IP 数据源获取：`internal/detector/source.go`，使用 `net/http` 从 GitHub520 获取候选 IP 列表，解析为域名-IP 映射
- [x] 2.2 实现内置默认 IP 列表：`internal/detector/defaults.go`，硬编码一份兜底 IP 列表，数据源不可达时使用
- [x] 2.3 实现 ICMP ping 检测：`internal/detector/ping.go`，通过 `os/exec` 调用系统 ping 命令，跨平台获取延迟和丢包率
- [x] 2.4 实现 TCP 端口检测：`internal/detector/tcp.go`，使用 `net.DialTimeout` 并发检测 443（HTTPS）和 22（SSH）端口连通性
- [x] 2.5 实现 IP 质量评分算法：`internal/detector/scorer.go`，综合延迟、丢包率、端口状态计算评分，排序筛选最优 IP 和备选 IP
- [x] 2.6 实现 ICMP 降级机制：ping 命令执行失败时自动降级为 TCP-only 模式
- [x] 2.7 实现 IP 兜底切换逻辑：主 IP 失效时自动切换到备选 IP

## 3. Hosts 文件管理模块（hosts-management）

- [x] 3.1 实现 hosts 文件读取器：`internal/hosts/reader.go`，解析 hosts 文件为结构化数据（IP-域名映射切片）
- [x] 3.2 实现标记区块管理：`internal/hosts/block.go`，识别/创建/更新 `# GitHub-Buddy Auto-Generated Start/End` 标记区块
- [x] 3.3 实现 hosts 文件写入器：`internal/hosts/writer.go`，原子写入（先写临时文件再 `os.Rename` 替换），保持原有格式和注释
- [x] 3.4 实现权限检测与提权提示：`internal/platform/` 包中检测 hosts 文件写权限，无权限时根据平台给出操作指引
- [x] 3.5 实现域名冲突检测与处理：扫描标记区块外的 GitHub 域名配置，提示用户选择处理方式

## 4. 备份与回滚模块（backup-rollback）

- [x] 4.1 实现 hosts 自动备份功能：`internal/backup/backup.go`，备份到 hosts 同目录 `hosts.bak` + `~/.github-buddy/backups/` 带时间戳备份
- [x] 4.2 实现备份文件管理：历史备份最多保留 10 份，超出删除最旧备份
- [x] 4.3 实现备份完整性校验：备份时使用 `crypto/sha256` 记录校验和，回滚时验证
- [x] 4.4 实现回滚功能：从备份恢复 hosts 文件，支持确认提示和 `--force` 参数

## 5. 自动更新模块（auto-update）

- [x] 5.1 实现 IP 缓存管理：`internal/cache/cache.go`，读写 `~/.github-buddy/cache/ip_cache.json`，使用 `encoding/json` 序列化
- [x] 5.2 实现缓存过期检查：根据配置的 `update_interval` 判断缓存是否过期
- [x] 5.3 实现配置文件管理：`internal/config/config.go`，读写 `~/.github-buddy/config.json`，支持 update_interval、data_sources、domains 配置
- [x] 5.4 实现启动时自动检查逻辑：缓存过期时使用 goroutine 异步触发 IP 检测更新

## 6. CLI 命令实现（cli-commands）

- [x] 6.1 实现 `github-buddy init` 命令：`cmd/github-buddy/cmd_init.go`，初始化目录、生成配置、备份 hosts、首次检测、更新 hosts
- [x] 6.2 实现 `github-buddy update` 命令：`cmd/github-buddy/cmd_update.go`，强制检测、备份、更新 hosts、输出新旧 IP 对比表
- [x] 6.3 实现 `github-buddy rollback` 命令：`cmd/github-buddy/cmd_rollback.go`，检查备份、确认提示、恢复 hosts
- [x] 6.4 实现 `github-buddy status` 命令：`cmd/github-buddy/cmd_status.go`，读取当前 IP、实时检测、表格输出状态信息
- [x] 6.5 实现全局参数：`--version`、`--help`、`--verbose`、`--force`（通过 cobra 的 PersistentFlags）
- [x] 6.6 实现统一错误处理：权限不足/网络不通/未初始化/备份不存在等场景的友好提示

## 7. 日志系统

- [x] 7.1 配置 zerolog：`internal/logger/logger.go`，设置控制台输出（彩色美化）+ 文件输出双后端，日志存储到 `~/.github-buddy/logs/`
- [x] 7.2 实现日志级别控制：默认 INFO，`--verbose` 时输出 DEBUG 到控制台
- [x] 7.3 实现日志轮转：使用 `lumberjack` 库实现按大小轮转，单文件不超过 10MB

## 8. 测试

- [x] 8.1 编写 IP 检测模块单元测试：`internal/detector/*_test.go`，覆盖数据源解析、评分算法、降级机制
- [x] 8.2 编写 hosts 管理模块单元测试：`internal/hosts/*_test.go`，覆盖文件解析、标记区块管理、冲突检测
- [x] 8.3 编写备份回滚模块单元测试：`internal/backup/*_test.go`，覆盖备份创建、完整性校验、回滚恢复
- [x] 8.4 编写集成测试：完整的 init→update→status→rollback 流程测试
- [x] 8.5 编写跨平台兼容性测试：验证 Windows/macOS/Linux 路径和权限处理

## 9. 构建与分发

- [x] 9.1 编写 Makefile：封装 `go build` 交叉编译命令，支持 `make build-all` 一键生成三平台二进制
- [x] 9.2 编写 CI/CD 配置（GitHub Actions）：自动构建 Windows/macOS/Linux 三平台二进制，自动发布 Release
- [x] 9.3 编写 Goreleaser 配置：自动化版本发布、changelog 生成、多平台二进制打包
