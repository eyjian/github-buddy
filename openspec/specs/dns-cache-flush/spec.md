# DNS Cache Flush Spec

## Overview

系统 DNS 缓存刷新模块，封装跨平台的 DNS 缓存刷新逻辑，在 hosts 文件更新后自动清除操作系统的 DNS 缓存，确保新的域名-IP 映射立即生效。

## ADDED Requirements

### Requirement: Cross-platform DNS cache flush
系统 SHALL 支持在 Linux、macOS 和 Windows 三个平台上自动刷新系统 DNS 缓存。

#### Scenario: Linux systemd 系统刷新成功
- **WHEN** 当前平台为 Linux 且 `systemd-resolve --flush-caches` 命令可用
- **THEN** 系统 SHALL 执行该命令刷新 DNS 缓存并返回成功

#### Scenario: Linux systemd-resolve 不可用时 fallback 到 resolvectl
- **WHEN** 当前平台为 Linux 且 `systemd-resolve` 执行失败
- **THEN** 系统 SHALL 尝试执行 `resolvectl flush-caches` 作为 fallback

#### Scenario: macOS 刷新成功
- **WHEN** 当前平台为 macOS
- **THEN** 系统 SHALL 同时执行 `dscacheutil -flushcache` 和 `killall -HUP mDNSResponder`

#### Scenario: Windows 刷新成功
- **WHEN** 当前平台为 Windows
- **THEN** 系统 SHALL 执行 `ipconfig /flushdns`

#### Scenario: 命令执行超时
- **WHEN** DNS 刷新命令执行超过 5 秒
- **THEN** 系统 SHALL 超时终止并返回超时错误

### Requirement: Non-fatal error handling
DNS 缓存刷新失败 SHALL NOT 导致命令整体失败。刷新失败时 SHALL 输出 Warning 级别日志。

#### Scenario: 刷新命令不存在
- **WHEN** 当前 Linux 系统无 systemd-resolve 和 resolvectl 命令
- **THEN** 系统 SHALL 输出警告提示并提供手动刷新命令指引，不中断主流程

#### Scenario: 刷新命令执行失败
- **WHEN** DNS 刷新命令返回非零退出码
- **THEN** 系统 SHALL 输出警告提示并提供手动刷新命令指引，不中断主流程

### Requirement: User-friendly output
刷新操作完成后 SHALL 输出友好的用户提示信息。

#### Scenario: 刷新成功后输出
- **WHEN** DNS 缓存刷新成功
- **THEN** 系统 SHALL 输出成功提示 "✅ 已自动刷新系统 DNS 缓存"，并附带浏览器缓存清除提示

#### Scenario: 刷新失败后输出
- **WHEN** DNS 缓存刷新失败
- **THEN** 系统 SHALL 输出警告提示，包含当前平台对应的手动刷新命令，并附带浏览器缓存清除提示

#### Scenario: 始终输出浏览器提示
- **WHEN** DNS 缓存刷新操作完成（无论成功或失败）
- **THEN** 系统 SHALL 始终输出浏览器 DNS 缓存清除提示（Chrome: chrome://net-internals/#dns）

## Acceptance Criteria

- [ ] Linux (systemd) 环境下成功刷新 DNS 缓存
- [ ] Linux 无 systemd 环境下输出友好的手动刷新指引
- [ ] macOS 环境下成功执行 dscacheutil + killall mDNSResponder
- [ ] Windows 环境下成功执行 ipconfig /flushdns
- [ ] 刷新失败不阻塞主命令流程
- [ ] 刷新成功输出 ✅ 提示，失败输出 ⚠️ 提示
- [ ] 始终输出浏览器缓存清除提示
- [ ] 命令执行设置 5 秒超时
