package detector

import (
	"math"
	"sort"
)

// 评分权重配置
const (
	latencyWeight  = 0.5  // 延迟权重
	lossWeight     = 0.3  // 丢包率权重
	portWeight     = 0.2  // 端口连通性权重
	maxLatencyMS   = 500  // 最大可接受延迟（毫秒），超过此值得 0 分
	targetLatency  = 50.0 // 目标延迟（毫秒），低于此值得满分
)

// ScoreIP 计算单个 IP 的质量评分（0-100）
func ScoreIP(entry *IPEntry) float64 {
	// 1. 延迟评分（0-100）
	latencyScore := calcLatencyScore(entry.Latency)

	// 2. 丢包率评分（0-100）
	lossScore := calcLossScore(entry.LossRate)

	// 3. 端口连通性评分（0-100）
	portScore := calcPortScore(entry.Port443, entry.Port22)

	// 加权计算总分
	score := latencyScore*latencyWeight + lossScore*lossWeight + portScore*portWeight

	return math.Round(score*100) / 100
}

// calcLatencyScore 计算延迟评分
func calcLatencyScore(latency float64) float64 {
	if latency <= 0 {
		return 0 // 无法检测到延迟，给 0 分
	}
	if latency <= targetLatency {
		return 100
	}
	if latency >= maxLatencyMS {
		return 0
	}
	// 线性递减
	return 100 * (1 - (latency-targetLatency)/(maxLatencyMS-targetLatency))
}

// calcLossScore 计算丢包率评分
func calcLossScore(lossRate float64) float64 {
	if lossRate <= 0 {
		return 100
	}
	if lossRate >= 1.0 {
		return 0
	}
	// 丢包率越高分数越低，非线性递减
	return 100 * (1 - lossRate) * (1 - lossRate)
}

// calcPortScore 计算端口连通性评分
func calcPortScore(port443, port22 bool) float64 {
	score := 0.0
	if port443 {
		score += 70 // 443 端口（HTTPS）权重更高
	}
	if port22 {
		score += 30 // 22 端口（SSH）
	}
	return score
}

// RankIPs 对 IP 列表按评分排序，返回排序后的列表
// 评分高的排在前面
func RankIPs(entries []IPEntry) []IPEntry {
	// 先计算评分
	for i := range entries {
		entries[i].Score = ScoreIP(&entries[i])
		entries[i].Available = entries[i].Port443 || entries[i].Port22
	}

	// 按评分降序排序
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Score > entries[j].Score
	})

	return entries
}

// SelectBestIPs 从排序后的 IP 列表中选择最优 IP 和备选 IP
// 返回最优 IP 和最多 backupCount 个备选 IP
func SelectBestIPs(entries []IPEntry, backupCount int) (best *IPEntry, backups []IPEntry) {
	if len(entries) == 0 {
		return nil, nil
	}

	ranked := RankIPs(entries)

	// 过滤出可用的 IP
	var available []IPEntry
	for _, e := range ranked {
		if e.Available {
			available = append(available, e)
		}
	}

	if len(available) == 0 {
		// 没有可用 IP，返回评分最高的作为"最佳猜测"
		first := ranked[0]
		return &first, nil
	}

	first := available[0]
	best = &first

	// 选择备选 IP
	if len(available) > 1 {
		end := backupCount + 1
		if end > len(available) {
			end = len(available)
		}
		backups = available[1:end]
	}

	return best, backups
}
