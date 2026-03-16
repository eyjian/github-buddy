# CLI Commands Spec

## Overview

命令行界面与交互模块，提供 init/update/rollback/status 四个核心命令及友好的用户交互体验。

## Requirements

### REQ-1: init 命令
- 命令：`github-buddy init`
- 功能：
  1. 创建 `~/.github-buddy/` 目录结构（config/cache/logs/backups）
  2. 生成默认配置文件 `config.json`
  3. 备份当前 hosts 文件
  4. 执行首次 IP 检测
  5. 将最优 IP 写入 hosts 文件
  6. 输出初始化结果摘要
- 已初始化时重复执行，提示"已初始化，可使用 update 更新 IP"

### REQ-2: update 命令
- 命令：`github-buddy update`
- 功能：
  1. 强制执行完整 IP 检测（忽略缓存）
  2. 备份当前 hosts 文件
  3. 将最优 IP 写入 hosts 文件
  4. 更新缓存
  5. 输出更新结果（域名-新IP-旧IP-延迟 对比表）

### REQ-3: rollback 命令
- 命令：`github-buddy rollback`
- 功能：
  1. 检查备份文件是否存在
  2. 提示用户确认回滚操作
  3. 将 hosts 恢复为备份版本
  4. 输出回滚结果
- 支持 `--force` 参数跳过确认

### REQ-4: status 命令
- 命令：`github-buddy status`
- 功能：
  1. 读取当前 hosts 中的 GitHub 域名-IP 映射
  2. 对当前 IP 执行实时检测（延迟、端口连通性）
  3. 以表格形式输出状态信息：
     ```
     域名                          IP              延迟    443端口  22端口  状态
     github.com                    140.82.121.4    12ms    ✓       ✓      ✓ 最优
     raw.githubusercontent.com     185.199.108.133 8ms     ✓       -      ✓ 正常
     ```
  4. 显示缓存状态（上次更新时间、是否过期）
  5. 显示备份状态（最近备份时间）

### REQ-5: 错误处理与提示
- 所有命令在执行失败时给出明确的错误信息和解决建议
- 常见错误场景及提示：
  - 权限不足："请使用 sudo 运行（Linux/macOS）"或"请以管理员身份运行（Windows）"
  - 网络不通："无法连接数据源，使用缓存中的 IP 列表"
  - 未初始化："请先执行 github-buddy init 进行初始化"
  - 备份不存在："无可用备份，无法执行回滚"

### REQ-6: 日志记录
- 关键操作（IP 检测、hosts 修改、回滚）记录到 `~/.github-buddy/logs/`
- 日志级别：INFO（正常操作）、WARN（降级/兜底）、ERROR（失败）
- 支持 `--verbose` 参数输出详细日志到控制台

### REQ-7: 单二进制文件分发
- 编译产物为单个可执行文件 `github-buddy`（Windows 下为 `github-buddy.exe`）
- 无额外运行时依赖，下载即可运行
- 支持通过 `github-buddy --version` 查看版本号
- 支持通过 `github-buddy --help` 查看帮助信息

## Acceptance Criteria

- [ ] `init` 命令创建完整目录结构并完成首次 IP 检测和 hosts 更新
- [ ] `update` 命令强制更新 IP 并显示新旧 IP 对比
- [ ] `rollback` 命令成功恢复 hosts 并支持 `--force` 参数
- [ ] `status` 命令以表格形式显示所有域名的 IP 状态、延迟和端口连通性
- [ ] 权限不足时给出平台相关的提权指引
- [ ] 网络不通时使用缓存数据，不阻塞命令执行
- [ ] `--verbose` 参数控制控制台日志详细程度
- [ ] `--version` 和 `--help` 参数正常工作
- [ ] 单二进制文件，无运行时依赖
