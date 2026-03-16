# IP Detection Spec

## Overview

智能 IP 检测与质量筛选模块，负责从数据源获取 GitHub 候选 IP，并通过多维度检测筛选出最优 IP。

## Requirements

### REQ-1: 候选 IP 获取
- 从 GitHub520 等公开数据源获取 GitHub 域名对应的候选 IP 列表
- 支持多数据源配置，主数据源不可用时自动 failover 到备用源
- 内置一份默认 IP 列表作为兜底，确保首次运行或网络完全不通时仍可工作
- 维护的目标域名清单包括：
  - github.com
  - ssh.github.com
  - gist.github.com
  - raw.githubusercontent.com
  - api.github.com
  - assets-cdn.github.com

### REQ-2: 多维度 IP 检测
- **ICMP 检测**：对候选 IP 执行 ping 操作，获取延迟和丢包率
- **TCP 443 端口检测**：验证 HTTPS 端口连通性
- **TCP 22 端口检测**：验证 SSH 端口连通性
- 三种检测并发执行，单次全量检测耗时＜10s
- ICMP 不可用时（权限限制），自动降级为仅 TCP 端口检测

### REQ-3: IP 质量评分与筛选
- 计算每个 IP 的质量评分，综合考虑：
  - 平均延迟（权重最高，目标＜50ms）
  - 丢包率（目标 = 0%）
  - 端口连通状态（443 和 22 均通为满分）
- 按评分排序，选择最优 IP 作为主 IP
- 保留 2-3 个备选 IP，用于兜底切换

### REQ-4: IP 兜底切换
- 主 IP 失效时（检测到端口不通或延迟超标），自动切换到评分次高的备选 IP
- 切换无需用户干预，切换后记录日志

## Acceptance Criteria

- [ ] 能成功从 GitHub520 数据源获取候选 IP 列表
- [ ] 数据源不可达时使用内置默认 IP 列表
- [ ] 同时检测 ICMP、TCP 443、TCP 22 三个维度
- [ ] ICMP 不可用时自动降级为 TCP-only 检测
- [ ] 单次全量 IP 检测耗时＜10s
- [ ] 筛选出的最优 IP 延迟＜50ms、丢包率=0、443/22 端口全通
- [ ] 保留至少 2 个备选 IP
- [ ] 主 IP 失效后自动切换到备选 IP，无需人工干预
