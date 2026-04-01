## Why

当前 `github-buddy` 在更新 hosts 后，已经能够提升命令行场景下的 GitHub 连通率，但仍可能出现“hosts 已更新、浏览器解析 IP 也正确、`github.com` 依然打不开”的情况。根因主要在于现有 IP 筛选更偏向端口与基础 HTTPS 可达，而不够贴近浏览器真实访问链路，同时候选 IP 仍依赖单一成功数据源，降低了选中浏览器可用 IP 的概率。

## What Changes

- 将 `github.com` 的 HTTPS 验证升级为更贴近浏览器实际访问的校验：除 TLS 握手和状态码外，增加响应内容/响应特征校验，避免将“能连上但不是 GitHub 正常首页”的 IP 误判为可用
- 改造候选 IP 获取策略：从“首个成功数据源即返回”升级为“聚合多个成功数据源并去重”，提升候选池质量与稳定性
- 对候选 IP 增加合法性过滤，跳过明显无效、保留地址、私网地址或不适合作为公网 GitHub 映射的 IP
- 优化 `status` 命令输出，增加 HTTPS/浏览器级验证结果，帮助用户区分“端口通”与“浏览器可打开”
- 优化 `init`/`update` 命令结果提示，在完成 hosts 更新后输出关键浏览器访问域名的验证摘要与失败提示

## Capabilities

### New Capabilities
- `browser-reachability`: 浏览器导向的 GitHub 首页可达性验证能力，覆盖更贴近浏览器真实访问的成功判定与用户可见诊断信息

### Modified Capabilities
- `https-verification`: 将 HTTPS 校验从“基础应用层可达”增强为“浏览器首页可用性优先”的验证策略
- `ip-detection`: 将候选 IP 获取从单源 failover 调整为多源聚合，并增加候选 IP 合法性过滤与更稳健的最优 IP 选择
- `cli-commands`: 调整 `status`、`init`、`update` 的输出，使其展示 HTTPS/浏览器验证结果与更明确的诊断提示

## Impact

- **修改代码**：`internal/detector/`、`cmd/github-buddy/`、`internal/config/` 相关逻辑与测试
- **用户可见行为**：`status` 输出列会变化，`init`/`update` 将增加浏览器可达性验证摘要
- **风险控制**：不新增外部运行时依赖，继续仅使用 Go 标准库与现有命令行环境
- **测试方式**：增加单元测试，并允许结合 `curl`/`wget` 进行浏览器场景回归验证
