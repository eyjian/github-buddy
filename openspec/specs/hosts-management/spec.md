# Hosts Management Spec

## Overview

跨平台 hosts 文件管理模块，负责安全地读写系统 hosts 文件，实现 GitHub 域名-IP 映射的增量更新。

## Requirements

### REQ-1: 系统自动识别
- 自动检测当前操作系统类型（Windows/macOS/Linux）
- 根据系统类型确定 hosts 文件路径：
  - Linux/macOS: `/etc/hosts`
  - Windows: `C:\Windows\System32\drivers\etc\hosts`

### REQ-2: 自动提权
- **Linux/macOS**：检测当前用户是否有 hosts 文件写权限，无权限时提示使用 `sudo` 运行
- **Windows**：检测是否以管理员身份运行，未提权时终止执行并给出明确指引："请以管理员身份运行命令提示符"
- 权限不足时不做任何修改，仅给出提示后安全退出

### REQ-3: 标记区块管理
- 工具修改的内容使用专属注释标记隔离：
  ```
  # GitHub-Buddy Auto-Generated Start
  <IP-域名映射行>
  # GitHub-Buddy Auto-Generated End
  ```
- 更新时仅替换 Start/End 标记之间的内容
- 不修改标记区块外的任何 hosts 配置

### REQ-4: 冲突处理
- 读取现有 hosts 中是否已存在 GitHub 域名配置（标记区块外）
- 发现冲突时提示用户选择：「保留最优（工具推荐）」或「保留原有」
- 默认推荐使用工具检测到的最优 IP

### REQ-5: 增量更新
- 仅修改 GitHub 相关域名映射，不覆盖或删除 hosts 中其他配置
- hosts 文件中标记区块不存在时，追加到文件末尾
- 确保 hosts 文件格式正确（每行一个 IP-域名映射，保留原有空行和注释）

### REQ-6: hosts 文件写入原子性
- 写入前先写入临时文件，验证内容正确后再替换原文件
- 写入失败时保持原 hosts 文件不变

## Acceptance Criteria

- [ ] 正确识别 Windows/macOS/Linux 三个平台的 hosts 文件路径
- [ ] 无权限时给出明确错误提示并安全退出
- [ ] 使用 `# GitHub-Buddy Auto-Generated Start/End` 标记区隔工具管理内容
- [ ] 更新仅影响标记区块内容，不修改用户手动配置
- [ ] 发现域名冲突时提示用户选择处理方式
- [ ] hosts 文件写入失败时不损坏原文件
- [ ] 在 Windows/macOS/Linux 上均能正确读写 hosts 文件
