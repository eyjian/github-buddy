# HTTPS Verification Spec

## Overview

HTTPS 应用层验证模块，通过真正的 HTTPS 请求（含 TLS 证书校验）验证候选 IP 的实际可用性，解决 TCP 端口连通但 HTTPS 不可用的问题。

## Requirements

### REQ-1: TLS 握手验证
- 对候选 IP 的 443 端口发起 TLS 握手，设置 ServerName 为目标域名
- 验证服务端证书是否匹配目标域名（使用系统 CA 证书池）
- TLS 握手成功 = 该 IP 确实为目标域名提供 HTTPS 服务
- 单次验证超时 5 秒

### REQ-2: HTTP 响应验证
- TLS 握手成功后，发起 HTTP HEAD 请求验证 HTTP 层可用性
- 状态码 < 500 视为可用（允许 3xx 重定向、4xx 认证等正常响应）
- 记录 HTTPS 完整延迟（TLS 握手 + HTTP 往返）

### REQ-3: 评分权重调整
- HTTPS 验证结果纳入评分体系，权重最高（40%）
- 新权重分配：HTTPS(40%) + 延迟(30%) + 丢包率(15%) + 端口连通(15%)
- HTTPS 验证通过得 100 分，未通过得 0 分
- 未通过 HTTPS 验证的 IP 总分上限被大幅压低

### REQ-4: 降级策略
- HTTPS 验证与 TCP/ICMP 检测并发执行
- 如果所有候选 IP 的 HTTPS 验证均失败，回退到 TCP-only 评分模式
- 确保在极端网络环境下仍能选出"最不差"的 IP

### REQ-5: 检测结果记录
- IPEntry 新增 HTTPS 验证状态和 HTTPS 延迟字段
- 日志中记录每个 IP 的 HTTPS 验证结果
- DetectResult 中标记是否使用了 HTTPS 验证

## Acceptance Criteria

- [ ] 能对候选 IP 发起正确的 TLS 握手（ServerName 匹配目标域名）
- [ ] TLS 证书不匹配时验证失败
- [ ] HTTPS 验证通过的 IP 评分显著高于仅 TCP 连通的 IP
- [ ] 所有 HTTPS 验证失败时降级到 TCP-only 模式，不会返回空结果
- [ ] HTTPS 验证超时（5 秒）不影响总体检测流程
- [ ] 单元测试覆盖评分权重调整后的场景
