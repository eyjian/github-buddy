## Why

当前 IP 检测仅使用 TCP 端口连通性（`net.DialContext("tcp", ...)` 连接 443/22 端口）+ ICMP ping 来评估 IP 质量。但在国内网络环境下，TCP 连接成功不代表 HTTPS 真正可用——被 DNS 污染或劫持的 IP 可能 TCP 握手成功，但 TLS 握手失败、证书不匹配或返回错误内容。这导致 `update` 命令选出的 IP 实际不可用，用户体验时好时坏。

需要在检测流程中加入 **HTTPS 应用层验证**，直接向候选 IP 发起真正的 HTTPS 请求（携带正确的 Host 头和 TLS SNI），验证 TLS 握手和 HTTP 响应是否正常，从而确保选出的 IP 真正可用。

## What Changes

- **新增 HTTPS 验证器**（`internal/detector/https.go`）：对候选 IP 发起 HTTPS GET 请求，验证 TLS 握手成功 + 证书匹配 + HTTP 响应状态码正常
- **修改 IP 检测流程**（`detector.go`）：在现有的 TCP/ICMP 检测基础上，增加 HTTPS 验证维度
- **修改 IP 数据结构**（`types.go`）：新增 `HTTPS` 布尔字段和 `HTTPSLatency` 延迟字段
- **修改评分算法**（`scorer.go`）：HTTPS 验证通过的 IP 获得最高权重加分，未通过 HTTPS 验证的 IP 大幅降分
- **更新测试用例**（`detector_test.go`）：补充 HTTPS 验证相关的单元测试

## Capabilities

### New Capabilities
- `https-verification`: HTTPS 应用层验证能力，通过真正的 HTTPS 请求（含 TLS 证书校验）验证候选 IP 的可用性

### Modified Capabilities
- `ip-detection`: IP 检测流程新增 HTTPS 验证维度，评分权重体系调整（HTTPS 验证权重最高）

## Impact

- **代码影响**：`internal/detector/` 包，新增 `https.go`，修改 `detector.go`、`types.go`、`scorer.go`、`detector_test.go`
- **依赖影响**：使用 Go 标准库 `crypto/tls` + `net/http`，无需引入新的第三方依赖
- **性能影响**：HTTPS 验证比 TCP connect 耗时更长（约 2-5 秒），但与其他检测并发执行，不影响总检测超时（30 秒上限）
- **兼容性**：无 breaking change，原有的 TCP/ICMP 检测仍然保留作为辅助维度
