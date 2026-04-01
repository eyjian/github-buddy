## Context

当前 `github-buddy` 的 `init` 和 `update` 命令更新 hosts 文件后，操作系统的 DNS 缓存中仍残留旧的域名-IP 映射。用户必须手动执行平台特定的 DNS 缓存刷新命令（如 `sudo dscacheutil -flushcache`），否则 `git clone`、`go mod tidy` 等操作可能继续使用旧 IP 导致连接失败。

现有 `internal/platform/platform.go` 已封装了跨平台检测逻辑（`Info` 结构体含 OS 字段），可以在此基础上扩展 DNS 缓存刷新功能。

## Goals / Non-Goals

**Goals:**
- 在 `init` 和 `update` 命令成功更新 hosts 后，自动刷新系统 DNS 缓存
- 刷新成功时输出友好的成功提示
- 刷新失败时输出警告提示（不中断主流程），提供手动刷新指引
- 始终提醒用户浏览器有独立 DNS 缓存，需单独处理

**Non-Goals:**
- 不自动清除浏览器 DNS 缓存（浏览器无标准化的外部清除接口）
- 不实现浏览器自动重启或 CDP 远程控制
- 不引入新的外部依赖或守护进程

## Decisions

### 决策 1: DNS 刷新实现位置 — 放在 `internal/platform/` 模块

**选择**: 在 `internal/platform/` 新增 `dns_flush.go` 文件

**理由**: platform 模块已经封装了 OS 类型检测和跨平台抽象，DNS 缓存刷新本质上是平台相关操作，放在此处符合职责内聚。

**替代方案**: 在 `cmd/` 层直接内联执行 — 不够内聚，无法复用。

### 决策 2: 刷新命令的平台适配策略

| 平台 | 命令 | 备注 |
|------|------|------|
| Linux (systemd) | `systemd-resolve --flush-caches`，失败则尝试 `resolvectl flush-caches` | 两个命令覆盖不同 systemd 版本 |
| macOS | `dscacheutil -flushcache` + `killall -HUP mDNSResponder` | 两条命令均需执行 |
| Windows | `ipconfig /flushdns` | 标准方式 |

**选择**: 使用 `os/exec` 执行外部命令，按平台分支处理。Linux 采用 fallback 策略（先 systemd-resolve，失败再 resolvectl）。

**理由**: 系统原生命令最可靠，无需额外依赖。fallback 策略兼容不同 Linux 发行版。

### 决策 3: 错误处理策略 — 非致命错误（Warning）

**选择**: DNS 缓存刷新失败不阻塞命令执行，仅输出 Warning 级别日志和用户提示。

**理由**:
- hosts 更新是核心操作，DNS 缓存刷新是锦上添花
- 部分 Linux 发行版可能不使用 systemd（如 Alpine），刷新命令不存在
- 用户可以根据提示手动刷新

**替代方案**: 刷新失败时返回错误退出码 — 过于严格，可能误导用户以为 hosts 更新失败。

### 决策 4: 用户提示输出格式

```
✅ 已自动刷新系统 DNS 缓存
💡 提示: 如仍有问题，请手动清除浏览器 DNS 缓存
   Chrome: 访问 chrome://net-internals/#dns 点击 "Clear host cache"
```

刷新失败时：
```
⚠️ 自动刷新系统 DNS 缓存失败，请手动执行:
   Linux:  sudo systemd-resolve --flush-caches
   macOS:  sudo dscacheutil -flushcache && sudo killall -HUP mDNSResponder
   Windows: ipconfig /flushdns
💡 提示: 如仍有问题，请手动清除浏览器 DNS 缓存
   Chrome: 访问 chrome://net-internals/#dns 点击 "Clear host cache"
```

## Risks / Trade-offs

| 风险 | 缓解措施 |
|------|----------|
| Linux 无 systemd 的发行版（如 Alpine）刷新命令不存在 | fallback 策略 + 失败输出手动命令指引 |
| DNS 刷新命令执行需要 root 权限 | init/update 本身已需要 sudo，无额外权限要求 |
| macOS 不同版本的 DNS 刷新命令不同 | `dscacheutil` + `killall mDNSResponder` 组合覆盖主流版本 |
| 刷新命令执行超时阻塞 CLI | 为 exec.Command 设置 5 秒超时 context |
