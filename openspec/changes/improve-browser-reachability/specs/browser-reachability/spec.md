## ADDED Requirements

### Requirement: Browser-oriented homepage validation for github.com
系统 SHALL 对 `github.com` 的候选 IP 执行浏览器导向的首页可用性验证。在 TLS 握手与 HTTP 请求成功之外，系统 MUST 校验响应满足 GitHub 首页的基本成功特征，避免将异常中间页、拦截页或明显不完整的响应误判为可用。

#### Scenario: github.com homepage validation succeeds
- **WHEN** 某个 `github.com` 候选 IP 返回了通过 TLS 证书校验的 HTTPS 响应，且响应特征满足 GitHub 首页成功判定
- **THEN** 系统 SHALL 将该 IP 标记为浏览器级验证通过

#### Scenario: github.com homepage validation rejects abnormal response
- **WHEN** 某个 `github.com` 候选 IP 虽然返回非 5xx 响应，但响应体缺少 GitHub 首页关键特征、响应过短或明显不属于 GitHub 首页
- **THEN** 系统 SHALL 将该 IP 标记为浏览器级验证失败

### Requirement: Browser reachability diagnostics
系统 SHALL 在用户可见输出中区分“基础端口/HTTPS 连通”与“浏览器级首页可用性”。对 `github.com` 的诊断结果 MUST 能帮助用户判断当前 IP 是否适合浏览器直接访问首页。

#### Scenario: status shows browser-level result
- **WHEN** 用户执行 `github-buddy status`
- **THEN** 系统 SHALL 在 `github.com` 的状态结果中展示浏览器级验证结果或等价状态文案

#### Scenario: update and init show browser-level summary
- **WHEN** `github-buddy init` 或 `github-buddy update` 完成 hosts 更新
- **THEN** 系统 SHALL 输出 `github.com` 浏览器级验证摘要或明确的后续诊断提示
