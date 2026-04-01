package main

import (
	"fmt"
	"strings"

	"github.com/eyjian/github-buddy/internal/detector"
)

func bestDomainEntry(result *detector.DetectResult, domain string) *detector.IPEntry {
	if result == nil {
		return nil
	}
	domainIPs, ok := result.Domains[domain]
	if !ok || domainIPs == nil {
		return nil
	}
	return domainIPs.BestIP
}

func printGitHubBrowserSummary(result *detector.DetectResult) {
	fmt.Println()
	for _, line := range githubBrowserSummaryLines(bestDomainEntry(result, "github.com")) {
		fmt.Println(line)
	}
}

func githubBrowserSummaryLines(entry *detector.IPEntry) []string {
	if entry == nil {
		return []string{"⚠️  github.com 首页验证: 未检测到可用候选 IP"}
	}

	lines := []string{
		fmt.Sprintf("🌐 github.com 首页验证: %s", githubSummaryLabel(entry)),
		fmt.Sprintf("   选中 IP: %s | HTTPS状态码: %s | 延迟: %s", entry.IP, formatStatusCode(entry.HTTPSStatusCode), formatEntryLatency(entry)),
	}

	switch {
	case entry.BrowserOK:
		lines = append(lines, "   结果: 浏览器首页特征校验已通过，通常无需再次切换 IP")
	case entry.BasicHTTPS:
		lines = append(lines,
			fmt.Sprintf("   警告: 当前 IP 仅满足基础 HTTPS，可连上但首页特征校验未通过：%s", browserReason(entry.BrowserReason)),
			"   建议: 如浏览器仍打不开，请重新执行 update 选择更优 IP，并清理浏览器 DNS 缓存与 socket 连接池",
		)
	case entry.Port443:
		lines = append(lines,
			"   警告: 当前 IP 只有 443 端口可达，尚未通过基础 HTTPS，更无法保证浏览器首页可打开",
			"   建议: 重新执行 update，优先选择通过浏览器级校验的 IP",
		)
	case entry.Port22:
		lines = append(lines,
			"   警告: 当前 IP 仅 SSH(22) 端口可达，443 端口不通，浏览器无法访问 github.com",
			"   建议: 重新执行 update，当前候选池中可能暂无优质 IP，可稍后再试",
		)
	default:
		lines = append(lines,
			"   警告: 当前 IP 完全不可用（443 和 22 端口均不通），建议重新执行 update 或回滚 hosts",
		)
	}

	lines = append(lines,
		"   浏览器排障: Chrome/Edge 除了清 DNS，还要执行 net-internals 的 Flush socket pools；必要时请重启浏览器",
	)

	return lines
}

func githubSummaryLabel(entry *detector.IPEntry) string {
	switch {
	case entry == nil:
		return "✗ 未检测到结果"
	case entry.BrowserOK:
		return "✓ 浏览器可用"
	case entry.BasicHTTPS:
		return "⚠ 仅基础 HTTPS"
	case entry.Port443:
		return "⚠ 仅 443 端口可达"
	case entry.Port22:
		return "⚠ 仅 SSH 可达"
	default:
		return "✗ 不可用"
	}
}

func statusLabel(domain string, port443OK, port22OK bool, httpsRes detector.HTTPSCheckResult) string {
	if domain == "github.com" {
		switch {
		case httpsRes.BrowserOK:
			return "✓ 浏览器正常"
		case httpsRes.BasicOK:
			return "⚠ 仅基础HTTPS"
		case port443OK:
			return "⚠ 仅443连通"
		case port22OK:
			return "⚠ 仅SSH连通"
		default:
			return "✗ 不可用"
		}
	}

	switch {
	case httpsRes.BasicOK:
		return "✓ HTTPS正常"
	case port443OK && port22OK:
		return "⚠ 仅端口通"
	case port443OK:
		return "⚠ 仅443连通"
	case port22OK:
		return "⚠ 仅SSH连通"
	default:
		return "✗ 不可用"
	}
}

func mark(ok bool) string {
	if ok {
		return "✓"
	}
	return "✗"
}

func browserMark(domain string, browserOK bool) string {
	if domain != "github.com" {
		return "-"
	}
	if browserOK {
		return "✓"
	}
	return "✗"
}

func browserReason(reason string) string {
	reason = strings.TrimSpace(reason)
	if reason == "" {
		return "原因未记录"
	}
	return reason
}

func formatEntryLatency(entry *detector.IPEntry) string {
	if entry == nil {
		return "-"
	}
	if entry.HTTPSLatency > 0 {
		return fmt.Sprintf("%.0fms", entry.HTTPSLatency)
	}
	if entry.Latency > 0 {
		return fmt.Sprintf("%.0fms", entry.Latency)
	}
	return "-"
}

func formatStatusCode(code int) string {
	if code <= 0 {
		return "-"
	}
	return fmt.Sprintf("%d", code)
}
