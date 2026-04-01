## MODIFIED Requirements

### Requirement: 候选 IP 获取
- 系统 SHALL 从多个公开数据源获取 GitHub 域名对应的候选 IP 列表
- 系统 SHALL 聚合所有成功数据源返回的候选 IP，而不是在首个成功数据源后停止
- 系统 SHALL 对聚合后的候选 IP 执行去重
- 系统 MUST 过滤非法 IP、私网地址、环回地址、链路本地地址、未指定地址、多播地址及其他明显不适合作为公网 GitHub 映射的地址
- 当所有外部数据源均失败时，系统 SHALL 回退到内置默认 IP 列表作为兜底
- 维护的目标域名清单 MUST 至少覆盖 `github.com`、`raw.githubusercontent.com`、`api.github.com`、`codeload.github.com`、`avatars.githubusercontent.com` 与 GitHub 页面所需的关键静态资源域名

#### Scenario: aggregate candidates from multiple successful sources
- **WHEN** 两个或以上数据源均成功返回候选 IP
- **THEN** 系统 SHALL 合并这些候选 IP，并在去重后进入检测流程

#### Scenario: ignore invalid candidate IPs
- **WHEN** 候选列表中包含私网地址、保留地址或格式非法的 IP
- **THEN** 系统 SHALL 在检测前跳过这些 IP

#### Scenario: fallback to built-in defaults only when all sources fail
- **WHEN** 所有外部数据源均不可用或均未返回有效候选 IP
- **THEN** 系统 SHALL 使用内置默认 IP 列表继续检测

### Requirement: IP 质量评分与筛选
- 系统 SHALL 优先选择同时满足浏览器导向 HTTPS 验证和基础网络连通性的 `github.com` 候选 IP
- 对 `github.com`，浏览器导向 HTTPS 验证失败的候选 IP MUST 在排序中明显落后于验证通过的候选 IP
- 对其他域名，系统 SHALL 继续基于 HTTPS、延迟、丢包率和端口连通性综合评分
- 当 `github.com` 所有候选 IP 的浏览器导向 HTTPS 验证均失败时，系统 MAY 回退到基础 HTTPS/TCP 结果选择“最不差”的候选 IP，以避免完全无结果

#### Scenario: browser-validated github.com IP ranks first
- **WHEN** `github.com` 候选集中同时存在浏览器导向验证通过与失败的 IP
- **THEN** 系统 SHALL 优先选择验证通过的 IP 作为最优 IP

#### Scenario: degraded selection still returns a best-effort github.com IP
- **WHEN** `github.com` 所有候选 IP 的浏览器导向验证均失败
- **THEN** 系统 SHALL 仍返回基于基础 HTTPS/TCP 评分的最佳候选作为兜底
