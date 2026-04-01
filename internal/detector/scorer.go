package detector

import (
	"math"
	"sort"
)

// 评分权重配置（含 HTTPS 验证维度）
const (
	httpsWeight          = 0.4  // HTTPS 验证权重（最高优先级）
	latencyWeight        = 0.3  // 延迟权重
	lossWeight           = 0.15 // 丢包率权重
	portWeight           = 0.15 // 端口连通性权重
	maxLatencyMS         = 500  // 最大可接受延迟（毫秒），超过此值得 0 分
	targetLatency        = 50.0 // 目标延迟（毫秒），低于此值得满分
	githubBasicHTTPSHint = 35.0 // github.com 未通过浏览器校验但基础 HTTPS 可达时的保底分
)

// ScoreIP 计算单个 IP 的质量评分（0-100）
// 评分维度：HTTPS 验证 + 延迟 + 丢包率 + 端口连通
func ScoreIP(entry *IPEntry) float64 {
	if entry == nil {
		return 0
	}

	httpsScore := calcHTTPSScore(entry)

	latency := entry.Latency
	if entry.BasicHTTPS && entry.HTTPSLatency > 0 {
		latency = entry.HTTPSLatency
	}
	latencyScore := calcLatencyScore(latency)
	lossScore := calcLossScore(entry.LossRate)
	portScore := calcPortScore(entry.Port443, entry.Port22)

	score := httpsScore*httpsWeight + latencyScore*latencyWeight + lossScore*lossWeight + portScore*portWeight
	return math.Round(score*100) / 100
}

// calcHTTPSScore 计算 HTTPS 验证评分
func calcHTTPSScore(entry *IPEntry) float64 {
	if entry.Domain == "github.com" {
		switch {
		case entry.BrowserOK:
			return 100
		case entry.BasicHTTPS:
			return githubBasicHTTPSHint
		case entry.Port443:
			return 10
		default:
			return 0
		}
	}

	if entry.BasicHTTPS || entry.HTTPS {
		return 100
	}
	return 0
}

// calcLatencyScore 计算延迟评分
func calcLatencyScore(latency float64) float64 {
	if latency <= 0 {
		return 0
	}
	if latency <= targetLatency {
		return 100
	}
	if latency >= maxLatencyMS {
		return 0
	}
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
	return 100 * (1 - lossRate) * (1 - lossRate)
}

// calcPortScore 计算端口连通性评分
func calcPortScore(port443, port22 bool) float64 {
	score := 0.0
	if port443 {
		score += 70
	}
	if port22 {
		score += 30
	}
	return score
}

// RankIPs 对 IP 列表按评分排序，返回排序后的列表
func RankIPs(entries []IPEntry) []IPEntry {
	for i := range entries {
		entries[i].Score = ScoreIP(&entries[i])
		// github.com 域名要求至少 443 端口可达才算可用，仅 SSH 可达不足以支撑浏览器访问
		if entries[i].Domain == "github.com" {
			entries[i].Available = entries[i].HTTPS || entries[i].BasicHTTPS || entries[i].Port443
		} else {
			entries[i].Available = entries[i].HTTPS || entries[i].BasicHTTPS || entries[i].Port443 || entries[i].Port22
		}
	}

	sort.SliceStable(entries, func(i, j int) bool {
		if entries[i].Domain == "github.com" && entries[j].Domain == "github.com" {
			if entries[i].BrowserOK != entries[j].BrowserOK {
				return entries[i].BrowserOK
			}
			if entries[i].BasicHTTPS != entries[j].BasicHTTPS {
				return entries[i].BasicHTTPS
			}
		}

		if entries[i].Score != entries[j].Score {
			return entries[i].Score > entries[j].Score
		}
		if entries[i].Port443 != entries[j].Port443 {
			return entries[i].Port443
		}
		return betterLatency(entries[i].Latency, entries[j].Latency)
	})

	return entries
}

func betterLatency(a, b float64) bool {
	if a <= 0 {
		return false
	}
	if b <= 0 {
		return true
	}
	return a < b
}

// SelectBestIPs 从排序后的 IP 列表中选择最优 IP 和备选 IP
// 返回最优 IP 和最多 backupCount 个备选 IP
// 对 github.com 域名：如果没有任何 Available 的候选，返回 nil（不选择不可用的 IP）
// 对其他域名：如果没有 Available 的候选，退而求其次返回排名第一的 IP
func SelectBestIPs(entries []IPEntry, backupCount int) (best *IPEntry, backups []IPEntry) {
	if len(entries) == 0 {
		return nil, nil
	}

	ranked := RankIPs(entries)

	available := make([]IPEntry, 0, len(ranked))
	for _, e := range ranked {
		if e.Available {
			available = append(available, e)
		}
	}

	if len(available) == 0 {
		// github.com 域名：没有可用候选时不选择，避免写入不可用 IP 导致退化
		if len(ranked) > 0 && ranked[0].Domain == "github.com" {
			return nil, nil
		}
		// 其他域名：退而求其次返回排名第一的
		first := ranked[0]
		return &first, nil
	}

	first := available[0]
	best = &first

	if len(available) > 1 {
		end := backupCount + 1
		if end > len(available) {
			end = len(available)
		}
		backups = available[1:end]
	}

	return best, backups
}
