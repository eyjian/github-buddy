# Auto Update Spec

## Overview

定时自动更新模块，负责后台检测最优 IP 变化并自动更新 hosts 文件，减少人工干预。

## Requirements

### REQ-1: 缓存机制
- 将最近一次 IP 检测结果缓存到 `~/.github-buddy/cache/ip_cache.json`
- 缓存内容包括：域名-IP 映射、检测时间戳、IP 质量评分
- 工具启动时优先使用缓存数据，避免每次都触发网络检测
- 缓存过期时间默认为 6 小时，可通过配置文件调整

### REQ-2: 启动时自动检查
- 每次执行 `init`、`update`、`status` 命令时，检查缓存是否过期
- 缓存过期时，后台异步触发 IP 检测更新
- 检测到更优 IP 时自动更新 hosts 文件
- 缓存未过期时，直接使用缓存数据，命令执行耗时＜1s

### REQ-3: 手动强制更新
- `github-buddy update` 命令强制触发一次完整 IP 检测
- 忽略缓存过期时间，立即执行多维度检测
- 检测结果写入缓存并更新 hosts 文件
- 输出更新结果摘要（更新了哪些域名的 IP、新旧 IP 对比）

### REQ-4: 配置管理
- 配置文件存储在 `~/.github-buddy/config.json`
- 可配置项包括：
  - `update_interval`: 自动检查间隔（默认 6 小时）
  - `data_sources`: IP 数据源列表及优先级
  - `domains`: 需要维护的域名列表（可自定义扩展）
- 首次 `init` 时生成默认配置文件

### REQ-5: 更新日志
- 每次自动/手动更新均记录日志到 `~/.github-buddy/logs/`
- 日志内容包括：更新时间、检测结果、IP 变化详情、更新是否成功
- 日志文件按日期轮转，单个日志文件不超过 10MB

## Acceptance Criteria

- [ ] IP 检测结果正确缓存到 `~/.github-buddy/cache/ip_cache.json`
- [ ] 缓存未过期时命令执行耗时＜1s
- [ ] 缓存过期后自动触发 IP 检测更新
- [ ] `update` 命令强制执行完整检测并更新 hosts
- [ ] 更新结果显示新旧 IP 对比信息
- [ ] `~/.github-buddy/config.json` 包含可配置的更新间隔和数据源
- [ ] 更新操作均有日志记录
- [ ] 日志文件按日期轮转，不超过 10MB
