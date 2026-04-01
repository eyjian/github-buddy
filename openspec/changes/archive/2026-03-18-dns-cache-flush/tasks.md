# Tasks - DNS Cache Flush

## 1. 新增 DNS 缓存刷新模块

**文件**: `internal/platform/dns_flush.go` (新建)

- [x] 1.1 定义 `FlushResult` 结构体（Success bool, Message string, Error error）
- [x] 1.2 实现 `FlushDNSCache(ctx context.Context, osType string) *FlushResult` 函数，按平台分支调用对应的刷新命令
- [x] 1.3 实现 Linux 刷新逻辑：先尝试 `systemd-resolve --flush-caches`，失败则 fallback 到 `resolvectl flush-caches`
- [x] 1.4 实现 macOS 刷新逻辑：执行 `dscacheutil -flushcache` 和 `killall -HUP mDNSResponder`
- [x] 1.5 实现 Windows 刷新逻辑：执行 `ipconfig /flushdns`
- [x] 1.6 为所有外部命令执行设置 5 秒超时（使用 `context.WithTimeout`）

## 2. 新增用户提示输出函数

**文件**: `internal/platform/dns_flush.go` (续)

- [x] 2.1 实现 `PrintFlushResult(result *FlushResult, osType string)` 函数，根据刷新结果输出友好提示
- [x] 2.2 刷新成功时输出 "✅ 已自动刷新系统 DNS 缓存"
- [x] 2.3 刷新失败时输出 "⚠️ 自动刷新系统 DNS 缓存失败" 并附带当前平台的手动刷新命令
- [x] 2.4 无论成功或失败，始终输出浏览器缓存清除提示（Chrome: chrome://net-internals/#dns）

## 3. 修改 init 命令集成 DNS 刷新

**文件**: `cmd/github-buddy/cmd_init.go` (修改)

- [x] 3.1 在 hosts 文件写入成功后、输出结果摘要前，调用 `platform.FlushDNSCache()` 并传入 context 和 OS 类型
- [x] 3.2 调用 `platform.PrintFlushResult()` 输出刷新结果提示
- [x] 3.3 DNS 刷新失败时记录 Warning 级别日志，不影响 init 命令返回值

## 4. 修改 update 命令集成 DNS 刷新

**文件**: `cmd/github-buddy/cmd_update.go` (修改)

- [x] 4.1 在 hosts 文件写入成功后、输出更新结果前，调用 `platform.FlushDNSCache()` 并传入 context 和 OS 类型
- [x] 4.2 调用 `platform.PrintFlushResult()` 输出刷新结果提示
- [x] 4.3 DNS 刷新失败时记录 Warning 级别日志，不影响 update 命令返回值

## 5. 单元测试

**文件**: `internal/platform/dns_flush_test.go` (新建)

- [x] 5.1 测试 `FlushDNSCache` 在不同 OS 类型参数下调用正确的命令
- [x] 5.2 测试 `PrintFlushResult` 在成功/失败场景下输出正确的提示文本
- [x] 5.3 测试超时场景：模拟命令执行超时返回错误
