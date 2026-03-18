## Why

用户执行 `github-buddy update` 或 `github-buddy init` 更新 hosts 文件后，操作系统的 DNS 缓存仍保留旧的域名-IP 映射，导致 `git clone`、`go mod tidy` 等操作可能继续使用失效的旧 IP，用户需要手动执行平台特定的 DNS 缓存刷新命令。这增加了使用门槛，尤其对不熟悉系统命令的开发者不友好。

## What Changes

- **新增系统 DNS 缓存刷新功能**：在 `init` 和 `update` 命令成功更新 hosts 文件后，自动调用平台对应的 DNS 缓存刷新命令
  - Linux (systemd): `systemd-resolve --flush-caches` 或 `resolvectl flush-caches`
  - macOS: `dscacheutil -flushcache && killall -HUP mDNSResponder`
  - Windows: `ipconfig /flushdns`
- **输出友好提示**：刷新成功后输出 "✅ 已自动刷新系统 DNS 缓存"；刷新失败时输出警告提示（非致命错误，不中断主流程），并提示用户可手动刷新
- **浏览器缓存提示**：无论刷新是否成功，均在最后输出提示："如仍有问题，请手动清除浏览器 DNS 缓存（Chrome: chrome://net-internals/#dns）"

## Capabilities

### New Capabilities
- `dns-cache-flush`: 系统 DNS 缓存刷新模块，封装跨平台的 DNS 缓存刷新逻辑和用户友好提示输出

### Modified Capabilities
- `cli-commands`: `init` 和 `update` 命令在 hosts 更新成功后新增调用 DNS 缓存刷新步骤

## Impact

- **新增文件**: `internal/platform/dns_flush.go` — DNS 缓存刷新的跨平台实现
- **修改文件**: `cmd/github-buddy/cmd_init.go` — init 命令新增 DNS 缓存刷新调用
- **修改文件**: `cmd/github-buddy/cmd_update.go` — update 命令新增 DNS 缓存刷新调用
- **依赖**: 无新增外部依赖，仅使用 Go 标准库 `os/exec` 执行系统命令
- **权限**: DNS 缓存刷新命令通常需要与 hosts 修改相同的管理员权限（sudo/管理员），因此不引入额外权限要求
