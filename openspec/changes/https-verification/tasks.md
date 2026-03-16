# Tasks - HTTPS Verification

## Task 1: 新增 HTTPS 验证器 (`internal/detector/https.go`)

**文件**: `internal/detector/https.go` (新建)

- [x] 定义 `HTTPSCheckResult` 结构体（IP、Domain、OK、Latency、Error）
- [x] 定义 `HTTPSChecker` 结构体（timeout）
- [x] 实现 `NewHTTPSChecker(timeout)` 构造函数
- [x] 实现 `Check(ctx, ip, domain)` 方法：
  - 创建自定义 `tls.Config`，设置 `ServerName` 为目标域名
  - 创建自定义 `http.Transport`，通过 `DialTLSContext` 指定连接 `ip:443`
  - 发起 HTTP HEAD 请求，Host 设为目标域名
  - 验证 TLS 握手成功 + 状态码 < 500
  - 记录完整延迟
- [x] 实现 `CheckIPs(ctx, ips, domain)` 并发批量检测方法

## Task 2: 修改 IP 数据结构 (`internal/detector/types.go`)

**文件**: `internal/detector/types.go` (修改)

- [x] `IPEntry` 新增 `HTTPS bool` 字段（HTTPS 验证是否通过）
- [x] `IPEntry` 新增 `HTTPSLatency float64` 字段（HTTPS 验证延迟 ms）
- [x] `DetectResult` 新增 `HTTPSUsed bool` 字段（是否使用了 HTTPS 验证）

## Task 3: 修改检测流程 (`internal/detector/detector.go`)

**文件**: `internal/detector/detector.go` (修改)

- [x] `Detector` 结构体新增 `httpsChecker *HTTPSChecker` 字段
- [x] `NewDetector` 中初始化 `httpsChecker`（超时 5 秒）
- [x] `detectIP` 方法中新增 HTTPS 验证 goroutine（与 TCP/ICMP 并发）
- [x] `DetectAll` 结果中标记 `HTTPSUsed = true`

## Task 4: 修改评分算法 (`internal/detector/scorer.go`)

**文件**: `internal/detector/scorer.go` (修改)

- [x] 新增评分权重常量 `httpsWeight = 0.4`
- [x] 调整现有权重：`latencyWeight = 0.3`, `lossWeight = 0.15`, `portWeight = 0.15`
- [x] 新增 `calcHTTPSScore(https bool)` 函数
- [x] `ScoreIP` 函数中纳入 HTTPS 评分维度
- [x] 新增降级逻辑：如果所有 IP 的 HTTPS 均未通过，使用旧权重体系

## Task 5: 更新测试用例 (`internal/detector/detector_test.go`)

**文件**: `internal/detector/detector_test.go` (修改)

- [x] 新增 `TestScoreIP_HTTPSPass` 测试：HTTPS 通过的 IP 评分应显著高于未通过的
- [x] 新增 `TestScoreIP_HTTPSFail` 测试：HTTPS 未通过的 IP 评分应低于阈值
- [x] 更新现有评分测试用例以适配新的权重体系

## Task 6: 更新 ip-detection spec

**文件**: `openspec/changes/github-buddy/specs/ip-detection/spec.md` (修改)

- [x] REQ-2 新增 HTTPS 验证维度
- [x] REQ-3 更新评分权重说明
- [x] 更新 Acceptance Criteria
