## MODIFIED Requirements

### Requirement: HTTP 响应验证
- TLS 握手成功后，系统 SHALL 发起 HTTP GET 请求验证 HTTP 层可用性
- 状态码 < 500 视为基础 HTTP 可用（允许 3xx 重定向、4xx 认证等正常响应）
- 对 `github.com`，系统 MUST 在基础 HTTP 可用之外追加浏览器导向的首页成功判定，包括响应内容特征、最小内容完整性或等价的成功信号校验
- 对除 `github.com` 外的其他 GitHub 域名，系统 MAY 继续使用现有基础 HTTP 成功判定
- 系统 SHALL 记录 HTTPS 完整延迟（TLS 握手 + HTTP 往返）

#### Scenario: github.com passes browser-oriented HTTP validation
- **WHEN** `github.com` 的候选 IP 返回通过 TLS 校验的 HTTPS 响应，且首页响应满足 GitHub 成功特征
- **THEN** 系统 SHALL 将该 IP 视为 HTTPS 验证通过

#### Scenario: github.com fails browser-oriented HTTP validation
- **WHEN** `github.com` 的候选 IP 仅满足基础 HTTP 可达，但首页响应不满足 GitHub 成功特征
- **THEN** 系统 SHALL 将该 IP 视为 HTTPS 验证失败

#### Scenario: non-homepage domains keep basic HTTP validation
- **WHEN** 目标域名不是 `github.com`
- **THEN** 系统 SHALL 继续基于 TLS 握手成功和基础 HTTP 响应成功判定 HTTPS 可用性
