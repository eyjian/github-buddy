package main

import (
	"strings"
	"testing"

	"github.com/eyjian/github-buddy/internal/detector"
)

func TestGitHubBrowserSummaryLines_BrowserOK(t *testing.T) {
	lines := githubBrowserSummaryLines(&detector.IPEntry{
		IP:              "20.205.243.166",
		HTTPSStatusCode: 200,
		HTTPSLatency:    88,
		BasicHTTPS:      true,
		HTTPS:           true,
		BrowserOK:       true,
	})

	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "✓ 浏览器可用") {
		t.Fatalf("摘要中应包含浏览器可用标签，实际输出: %s", joined)
	}
}

func TestGitHubBrowserSummaryLines_Degraded(t *testing.T) {
	lines := githubBrowserSummaryLines(&detector.IPEntry{
		IP:              "20.205.243.166",
		HTTPSStatusCode: 200,
		HTTPSLatency:    88,
		Port443:         true,
		BasicHTTPS:      true,
		BrowserOK:       false,
		BrowserReason:   "缺少 GitHub 首页关键身份特征",
	})

	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, "⚠ 仅基础 HTTPS") {
		t.Fatalf("摘要中应包含基础 HTTPS 降级标签，实际输出: %s", joined)
	}
	if !strings.Contains(joined, "首页特征校验未通过") {
		t.Fatalf("摘要中应包含首页特征校验失败提示，实际输出: %s", joined)
	}
}

func TestStatusLabel_GitHubDegradedWhenOnlyBasicHTTPSPasses(t *testing.T) {
	label := statusLabel("github.com", true, false, detector.HTTPSCheckResult{BasicOK: true, BrowserOK: false})
	if label != "⚠ 仅基础HTTPS" {
		t.Fatalf("github.com 在仅基础 HTTPS 通过时应标记为降级，实际: %s", label)
	}
}

func TestStatusLabel_NonGitHubUsesHTTPSNormal(t *testing.T) {
	label := statusLabel("raw.githubusercontent.com", true, false, detector.HTTPSCheckResult{BasicOK: true})
	if label != "✓ HTTPS正常" {
		t.Fatalf("非 github.com 域名在基础 HTTPS 正常时应标记为 HTTPS 正常，实际: %s", label)
	}
}

func TestStatusLabel_GitHubBrowserOKWins(t *testing.T) {
	label := statusLabel("github.com", true, true, detector.HTTPSCheckResult{BasicOK: true, BrowserOK: true, Port443OK: true})
	if label != "✓ 浏览器正常" {
		t.Fatalf("github.com 在浏览器级校验通过时应标记为浏览器正常，实际: %s", label)
	}
}

func TestStatusLabel_GitHubOnlyPort443ShowsDegraded(t *testing.T) {
	label := statusLabel("github.com", true, false, detector.HTTPSCheckResult{Port443OK: true})
	if label != "⚠ 仅443连通" {
		t.Fatalf("github.com 在仅 443 连通时应标记为仅443连通，实际: %s", label)
	}
}
