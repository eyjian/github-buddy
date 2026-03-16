## 1. 项目基础搭建

- [ ] 1.1 创建 CMake 项目结构：CMakeLists.txt、src/、include/、tests/ 目录，设置 C++17 标准
- [ ] 1.2 集成第三方库依赖：CLI11（命令行解析）、spdlog（日志）、nlohmann/json（JSON 处理）、libcurl（HTTP 客户端）
- [ ] 1.3 创建平台抽象层基础框架：include/platform/ 目录，定义 PlatformDetector 类（OS 类型识别、hosts 路径、home 目录）
- [ ] 1.4 创建用户数据目录管理模块：~/.github-buddy/ 目录结构初始化（config/cache/logs/backups）

## 2. IP 检测模块（ip-detection）

- [ ] 2.1 实现 IP 数据源获取：从 GitHub520 获取候选 IP 列表，解析为域名-IP 映射结构体
- [ ] 2.2 实现内置默认 IP 列表：硬编码一份兜底 IP 列表，数据源不可达时使用
- [ ] 2.3 实现 ICMP ping 检测：跨平台实现（Linux/macOS: raw socket，Windows: IcmpSendEcho），获取延迟和丢包率
- [ ] 2.4 实现 TCP 端口检测：并发检测 443（HTTPS）和 22（SSH）端口连通性
- [ ] 2.5 实现 IP 质量评分算法：综合延迟、丢包率、端口状态计算评分，排序筛选最优 IP 和备选 IP
- [ ] 2.6 实现 ICMP 降级机制：检测 ICMP 不可用时自动降级为 TCP-only 模式
- [ ] 2.7 实现 IP 兜底切换逻辑：主 IP 失效时自动切换到备选 IP

## 3. Hosts 文件管理模块（hosts-management）

- [ ] 3.1 实现 hosts 文件读取器：解析 hosts 文件内容为结构化数据（IP-域名映射列表）
- [ ] 3.2 实现标记区块管理：识别/创建/更新 `# GitHub-Buddy Auto-Generated Start/End` 标记区块
- [ ] 3.3 实现 hosts 文件写入器：原子写入（先写临时文件再替换），保持原有格式和注释不变
- [ ] 3.4 实现权限检测与提权提示：检测 hosts 文件写权限，无权限时根据平台给出操作指引
- [ ] 3.5 实现域名冲突检测与处理：扫描标记区块外的 GitHub 域名配置，提示用户选择处理方式

## 4. 备份与回滚模块（backup-rollback）

- [ ] 4.1 实现 hosts 自动备份功能：备份到 hosts 同目录 hosts.bak + ~/.github-buddy/backups/ 带时间戳备份
- [ ] 4.2 实现备份文件管理：历史备份最多保留 10 份，超出删除最旧备份
- [ ] 4.3 实现备份完整性校验：备份时记录 SHA256 校验和，回滚时验证
- [ ] 4.4 实现回滚功能：从备份恢复 hosts 文件，支持确认提示和 --force 参数

## 5. 自动更新模块（auto-update）

- [ ] 5.1 实现 IP 缓存管理：读写 ~/.github-buddy/cache/ip_cache.json，包含域名-IP 映射、时间戳、评分
- [ ] 5.2 实现缓存过期检查：根据配置的 update_interval 判断缓存是否过期
- [ ] 5.3 实现配置文件管理：读写 ~/.github-buddy/config.json，支持 update_interval、data_sources、domains 配置
- [ ] 5.4 实现启动时自动检查逻辑：缓存过期时异步触发 IP 检测更新

## 6. CLI 命令实现（cli-commands）

- [ ] 6.1 实现 `github-buddy init` 命令：初始化目录、生成配置、备份 hosts、首次检测、更新 hosts
- [ ] 6.2 实现 `github-buddy update` 命令：强制检测、备份、更新 hosts、输出新旧 IP 对比表
- [ ] 6.3 实现 `github-buddy rollback` 命令：检查备份、确认提示、恢复 hosts
- [ ] 6.4 实现 `github-buddy status` 命令：读取当前 IP、实时检测、表格输出状态信息
- [ ] 6.5 实现全局参数：--version、--help、--verbose、--force
- [ ] 6.6 实现统一错误处理：权限不足/网络不通/未初始化/备份不存在等场景的友好提示

## 7. 日志系统

- [ ] 7.1 配置 spdlog：设置控制台输出 + 文件轮转双后端，日志存储到 ~/.github-buddy/logs/
- [ ] 7.2 实现日志级别控制：默认 INFO，--verbose 时输出 DEBUG 到控制台
- [ ] 7.3 实现日志轮转：按日期轮转，单文件不超过 10MB

## 8. 测试

- [ ] 8.1 编写 IP 检测模块单元测试：数据源解析、评分算法、降级机制
- [ ] 8.2 编写 hosts 管理模块单元测试：文件解析、标记区块管理、冲突检测
- [ ] 8.3 编写备份回滚模块单元测试：备份创建、完整性校验、回滚恢复
- [ ] 8.4 编写集成测试：完整的 init→update→status→rollback 流程测试
- [ ] 8.5 编写跨平台兼容性测试：验证 Windows/macOS/Linux 路径和权限处理

## 9. 构建与分发

- [ ] 9.1 配置 CMake 跨平台编译：支持 GCC/Clang/MSVC，生成单二进制文件
- [ ] 9.2 编写 CI/CD 配置：自动构建 Windows/macOS/Linux 三平台二进制
- [ ] 9.3 编写项目 README：安装说明、使用指南、命令参考
