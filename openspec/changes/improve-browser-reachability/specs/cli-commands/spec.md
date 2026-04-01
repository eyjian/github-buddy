## MODIFIED Requirements

### Requirement: update 命令
- 命令：`github-buddy update`
- 功能：
  1. 强制执行完整 IP 检测（忽略缓存）
  2. 备份当前 hosts 文件
  3. 将最优 IP 写入 hosts 文件
  4. 更新缓存
  5. **自动刷新系统 DNS 缓存并输出提示**
  6. 输出更新结果（域名-新IP-旧IP-延迟 对比表）
  7. **输出 `github.com` 浏览器级验证摘要或明确诊断提示**

#### Scenario: update shows browser-oriented github.com summary
- **WHEN** `update` 命令成功完成 hosts 更新
- **THEN** 系统 SHALL 在结果摘要中展示 `github.com` 的 HTTPS/浏览器级验证结果或等价诊断信息

#### Scenario: update warns when browser-oriented validation is weak
- **WHEN** `github.com` 被选中的最优 IP 仅满足基础连通但未通过浏览器导向验证
- **THEN** 系统 SHALL 输出明确警告，提示该 IP 可能仍不足以保证浏览器打开首页

### Requirement: status 命令
- 命令：`github-buddy status`
- 功能：
  1. 读取当前 hosts 中的 GitHub 域名-IP 映射
  2. 对当前 IP 执行实时检测（延迟、端口连通性、HTTPS/浏览器级验证）
  3. 以表格形式输出状态信息，至少能够区分 TCP/端口可达与 HTTPS/浏览器级可用
  4. 显示缓存状态（上次更新时间、是否过期）
  5. 显示备份状态（最近备份时间）

#### Scenario: status distinguishes port reachability from browser reachability
- **WHEN** 用户执行 `github-buddy status`
- **THEN** 系统 SHALL 展示能够区分“443 端口连通”和“浏览器级首页可用”的状态信息

#### Scenario: status marks github.com as degraded when only TCP passes
- **WHEN** `github.com` 当前映射 IP 的 443 端口可达，但 HTTPS/浏览器级验证失败
- **THEN** 系统 SHALL 将其展示为降级或部分可用状态，而不是直接标记为正常
