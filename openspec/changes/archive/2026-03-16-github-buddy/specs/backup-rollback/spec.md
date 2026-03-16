# Backup & Rollback Spec

## Overview

hosts 文件安全保障模块，提供自动备份和一键回滚功能，确保 hosts 文件修改操作可逆、安全。

## Requirements

### REQ-1: 自动备份
- 首次运行 `github-buddy init` 时，将当前 hosts 文件备份
- 每次执行 `update` 修改 hosts 前，自动备份当前 hosts 文件
- 备份文件存储位置：hosts 文件同目录下，文件名为 `hosts.bak`
- 同时在 `~/.github-buddy/backups/` 目录保留带时间戳的历史备份（如 `hosts.bak.20260316-113700`）
- 历史备份最多保留 10 份，超出时删除最旧的备份

### REQ-2: 一键回滚
- 提供 `github-buddy rollback` 命令
- 回滚操作将 hosts 文件恢复为最近一次备份的状态
- 回滚前确认提示："即将将 hosts 恢复为 <备份时间> 的版本，是否继续？[Y/n]"
- 支持 `--force` 参数跳过确认提示
- 回滚成功后输出确认信息并记录日志

### REQ-3: 备份完整性验证
- 备份时记录原始文件的 SHA256 校验和
- 回滚时验证备份文件完整性，校验和不匹配时发出警告
- 备份文件不存在或损坏时给出明确错误提示

### REQ-4: 增量更新保障
- 通过增量更新机制（仅修改标记区块），将意外损坏风险降到最低
- 即使不回滚，删除 `# GitHub-Buddy Auto-Generated Start/End` 标记区块也能手动恢复

## Acceptance Criteria

- [ ] `init` 命令执行时自动备份 hosts 文件
- [ ] 每次 `update` 前自动备份 hosts 文件
- [ ] `hosts.bak` 存于 hosts 同目录，历史备份存于 `~/.github-buddy/backups/`
- [ ] `rollback` 命令成功恢复 hosts 到备份状态
- [ ] `rollback --force` 跳过确认直接执行
- [ ] 备份文件损坏时给出明确错误提示
- [ ] 历史备份最多保留 10 份
