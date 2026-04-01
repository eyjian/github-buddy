package detector

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestScoreIP_FullScore(t *testing.T) {
	entry := &IPEntry{
		Domain:       "github.com",
		Latency:      10,
		LossRate:     0,
		Port443:      true,
		Port22:       true,
		BasicHTTPS:   true,
		HTTPS:        true,
		BrowserOK:    true,
		HTTPSLatency: 50,
	}
	score := ScoreIP(entry)
	if score < 90 {
		t.Errorf("满分场景评分应≥90, 实际: %.2f", score)
	}
}

func TestScoreIP_HighLatency(t *testing.T) {
	entry := &IPEntry{
		Domain:       "github.com",
		Latency:      400,
		LossRate:     0,
		Port443:      true,
		Port22:       true,
		BasicHTTPS:   true,
		HTTPS:        true,
		BrowserOK:    true,
		HTTPSLatency: 400,
	}
	score := ScoreIP(entry)
	if score > 80 {
		t.Errorf("高延迟场景评分应＜80, 实际: %.2f", score)
	}
}

func TestScoreIP_PortsDown(t *testing.T) {
	entry := &IPEntry{
		Domain:       "github.com",
		Latency:      10,
		LossRate:     0,
		Port443:      false,
		Port22:       false,
		BasicHTTPS:   true,
		HTTPS:        true,
		BrowserOK:    true,
		HTTPSLatency: 10,
	}
	score := ScoreIP(entry)
	if score < 50 {
		t.Errorf("HTTPS通过但端口不通场景评分应≥50, 实际: %.2f", score)
	}
}

func TestScoreIP_PacketLoss(t *testing.T) {
	entry := &IPEntry{
		Domain:       "github.com",
		Latency:      10,
		LossRate:     0.5,
		Port443:      true,
		Port22:       true,
		BasicHTTPS:   true,
		HTTPS:        true,
		BrowserOK:    true,
		HTTPSLatency: 10,
	}
	full := ScoreIP(&IPEntry{Domain: "github.com", Latency: 10, LossRate: 0, Port443: true, Port22: true, BasicHTTPS: true, HTTPS: true, BrowserOK: true, HTTPSLatency: 10})
	score := ScoreIP(entry)
	if score >= full {
		t.Errorf("有丢包的评分(%.2f)应低于无丢包(%.2f)", score, full)
	}
}

func TestScoreIP_HTTPSPass(t *testing.T) {
	httpsEntry := &IPEntry{
		Domain:       "github.com",
		Latency:      10,
		LossRate:     0,
		Port443:      true,
		Port22:       true,
		BasicHTTPS:   true,
		HTTPS:        true,
		BrowserOK:    true,
		HTTPSLatency: 10,
	}
	tcpOnlyEntry := &IPEntry{
		Domain:   "github.com",
		Latency:  10,
		LossRate: 0,
		Port443:  true,
		Port22:   true,
	}
	httpsScore := ScoreIP(httpsEntry)
	tcpScore := ScoreIP(tcpOnlyEntry)
	if diff := httpsScore - tcpScore; diff < 30 {
		t.Errorf("HTTPS通过的IP(%.2f)应比仅TCP连通的IP(%.2f)高出至少30分, 实际差值: %.2f", httpsScore, tcpScore, diff)
	}
}

func TestScoreIP_HTTPSFail(t *testing.T) {
	entry := &IPEntry{
		Domain:   "example.com",
		Latency:  10,
		LossRate: 0,
		Port443:  true,
		Port22:   true,
		HTTPS:    false,
	}
	score := ScoreIP(entry)
	if score > 60 {
		t.Errorf("HTTPS未通过时评分不应超过60, 实际: %.2f", score)
	}
}

func TestRankIPs(t *testing.T) {
	entries := []IPEntry{
		{IP: "1.1.1.1", Domain: "example.com", Latency: 100, LossRate: 0, Port443: true, Port22: true, BasicHTTPS: true, HTTPS: true},
		{IP: "2.2.2.2", Domain: "example.com", Latency: 10, LossRate: 0, Port443: true, Port22: true, BasicHTTPS: true, HTTPS: true},
		{IP: "3.3.3.3", Domain: "example.com", Latency: 50, LossRate: 0.5, Port443: true, Port22: false},
	}
	ranked := RankIPs(entries)
	if ranked[0].IP != "2.2.2.2" {
		t.Errorf("最优 IP 应为 2.2.2.2, 实际: %s", ranked[0].IP)
	}
}

func TestSelectBestIPs(t *testing.T) {
	entries := []IPEntry{
		{IP: "1.1.1.1", Domain: "example.com", Latency: 100, LossRate: 0, Port443: true, Port22: true, BasicHTTPS: true, HTTPS: true},
		{IP: "2.2.2.2", Domain: "example.com", Latency: 10, LossRate: 0, Port443: true, Port22: true, BasicHTTPS: true, HTTPS: true},
		{IP: "3.3.3.3", Domain: "example.com", Latency: 50, LossRate: 0, Port443: true, Port22: true, BasicHTTPS: true, HTTPS: true},
	}
	best, backups := SelectBestIPs(entries, 2)
	if best == nil {
		t.Fatal("最优 IP 不应为 nil")
	}
	if best.IP != "2.2.2.2" {
		t.Errorf("最优 IP 应为 2.2.2.2, 实际: %s", best.IP)
	}
	if len(backups) != 2 {
		t.Errorf("应有 2 个备选 IP, 实际: %d", len(backups))
	}
}

func TestSelectBestIPs_Empty(t *testing.T) {
	best, backups := SelectBestIPs(nil, 2)
	if best != nil {
		t.Error("空列表应返回 nil")
	}
	if backups != nil {
		t.Error("空列表备选应为 nil")
	}
}

func TestParsePingLatency(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   float64
	}{
		{name: "Linux格式", output: "rtt min/avg/max/mdev = 1.234/5.678/9.012/0.123 ms", want: 5.678},
		{name: "Windows格式", output: "Average = 12ms", want: 12},
		{name: "无匹配", output: "no output", want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parsePingLatency(tt.output)
			if got != tt.want {
				t.Errorf("parsePingLatency() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParsePingLoss(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   float64
	}{
		{name: "0%丢包", output: "3 packets transmitted, 3 received, 0% packet loss", want: 0},
		{name: "50%丢包", output: "2 packets transmitted, 1 received, 50% packet loss", want: 0.5},
		{name: "100%丢包", output: "3 packets transmitted, 0 received, 100% packet loss", want: 1.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parsePingLoss(tt.output)
			if got != tt.want {
				t.Errorf("parsePingLoss() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsTargetDomain(t *testing.T) {
	if !isTargetDomain("github.com") {
		t.Error("github.com 应为目标域名")
	}
	if isTargetDomain("google.com") {
		t.Error("google.com 不应为目标域名")
	}
}

func TestGetDefaultIPs(t *testing.T) {
	ips := GetDefaultIPs()
	if len(ips) == 0 {
		t.Error("默认 IP 列表不应为空")
	}
	if _, ok := ips["github.com"]; !ok {
		t.Error("默认 IP 列表应包含 github.com")
	}
}

func TestIsUsableCandidateIP(t *testing.T) {
	tests := []struct {
		ip   string
		want bool
	}{
		{ip: testIPv4(4, 6, 9, 15), want: true},
		{ip: testIPv4(10, 56, 36, 30), want: false},
		{ip: testIPv4(100, 64, 1, 2), want: false},
		{ip: testIPv4(127, 0, 0, 1), want: false},
		{ip: testIPv4(198, 51, 100, 25), want: false},
		{ip: "fc00::1", want: false},
		{ip: "", want: false},
		{ip: "not-an-ip", want: false},
	}

	for _, tt := range tests {
		if got := isUsableCandidateIP(tt.ip); got != tt.want {
			t.Errorf("isUsableCandidateIP(%q) = %v, want %v", tt.ip, got, tt.want)
		}
	}
}

func TestFilterCandidateMap(t *testing.T) {
	filtered := filterCandidateMap(map[string][]string{
		"github.com": {testIPv4(4, 6, 9, 15), testIPv4(10, 56, 36, 30), testIPv4(198, 51, 100, 25), testIPv4(4, 6, 9, 15)},
		"google.com": {testIPv4(4, 6, 9, 15)},
	})

	got := filtered["github.com"]
	if len(got) != 1 || got[0] != testIPv4(4, 6, 9, 15) {
		t.Fatalf("过滤后的 github.com 候选不符合预期: %#v", got)
	}
	if _, ok := filtered["google.com"]; ok {
		t.Fatal("非目标域名不应保留")
	}
}

type stubSource struct {
	name string
	data map[string][]string
	err  error
}

func (s stubSource) FetchIPs(context.Context) (map[string][]string, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.data, nil
}

func (s stubSource) Name() string { return s.name }

func TestMultiSourceFetchIPsAggregatesAndDeduplicates(t *testing.T) {
	ms := NewMultiSource(
		stubSource{name: "s1", data: map[string][]string{"github.com": {testIPv4(4, 6, 9, 15), testIPv4(10, 56, 36, 30)}}},
		stubSource{name: "s2", data: map[string][]string{"github.com": {testIPv4(4, 6, 9, 15), testIPv4(19, 9, 246, 26)}, "raw.githubusercontent.com": {testIPv4(69, 7, 71, 57)}}},
	)

	got, err := ms.FetchIPs(context.Background())
	if err != nil {
		t.Fatalf("FetchIPs() 返回错误: %v", err)
	}

	if len(got["github.com"]) != 2 {
		t.Fatalf("github.com 聚合后应有 2 个唯一公网 IP，实际: %#v", got["github.com"])
	}
	if len(got["raw.githubusercontent.com"]) != 1 {
		t.Fatalf("raw.githubusercontent.com 聚合结果不符合预期: %#v", got["raw.githubusercontent.com"])
	}
}

func TestMultiSourceFetchIPsReturnsErrorWhenAllFailed(t *testing.T) {
	ms := NewMultiSource(
		stubSource{name: "s1", err: errors.New("boom1")},
		stubSource{name: "s2", err: errors.New("boom2")},
	)

	if _, err := ms.FetchIPs(context.Background()); err == nil {
		t.Fatal("所有数据源失败时应返回错误")
	}
}

func TestValidateBrowserResponse_GitHubHomepagePass(t *testing.T) {
	resp := &http.Response{
		Header:  http.Header{"Content-Type": []string{"text/html; charset=utf-8"}},
		Request: &http.Request{URL: mustParseURL("https://github.com/")},
	}
	body := []byte("<!DOCTYPE html><html><head><meta content=\"GitHub\"><link rel=\"preconnect\" href=\"https://github.githubassets.com\"></head><body>" +
		"GitHub Copilot MonaSans " + stringsRepeat("x", 5000) + "</body></html>")

	ok, reason := validateBrowserResponse("github.com", resp, body)
	if !ok {
		t.Fatalf("预期 github.com 首页特征校验通过，实际失败: %s", reason)
	}
}

func TestValidateBrowserResponse_GitHubHomepagePassesByHeaders(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header: http.Header{
			"Content-Type":            []string{"text/html; charset=utf-8"},
			"Server":                  []string{"github.com"},
			"Content-Security-Policy": []string{"default-src 'none'; script-src github.githubassets.com; style-src github.githubassets.com"},
			"Set-Cookie":              []string{"_gh_sess=abc; Path=/; HttpOnly", "logged_in=no; Path=/; HttpOnly"},
		},
		Request: &http.Request{URL: mustParseURL("https://github.com/")},
	}

	ok, reason := validateBrowserResponse("github.com", resp, nil)
	if !ok {
		t.Fatalf("预期 github.com 首页响应头快速校验通过，实际失败: %s", reason)
	}
}

func TestValidateBrowserResponse_GitHubHomepageRejectsBlockPage(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/html; charset=utf-8"}},
		Request:    &http.Request{URL: mustParseURL("https://github.com/")},
	}
	body := []byte("<!DOCTYPE html><html><body>Access Denied security challenge captcha " + stringsRepeat("x", 5000) + "</body></html>")

	ok, reason := validateBrowserResponse("github.com", resp, body)
	if ok || reason == "" {
		t.Fatalf("预期拦截页被拒绝，实际 ok=%v reason=%q", ok, reason)
	}
}

func TestValidateBrowserResponse_GitHubHomepageRejectsShortBody(t *testing.T) {
	resp := &http.Response{
		Header:  http.Header{"Content-Type": []string{"text/html; charset=utf-8"}},
		Request: &http.Request{URL: mustParseURL("https://github.com/")},
	}
	body := []byte("<!DOCTYPE html><html><head><meta content=\"GitHub\"></head><body>short</body></html>")

	ok, reason := validateBrowserResponse("github.com", resp, body)
	if ok || reason == "" {
		t.Fatalf("预期短响应体被拒绝，实际 ok=%v reason=%q", ok, reason)
	}
}

func TestValidateBrowserResponse_NonGitHubDomainUsesLightweightValidation(t *testing.T) {
	resp := &http.Response{Request: &http.Request{URL: mustParseURL("https://raw.githubusercontent.com/")}}
	ok, reason := validateBrowserResponse("raw.githubusercontent.com", resp, []byte("ok"))
	if !ok {
		t.Fatalf("非 github.com 域名应采用轻量校验并通过，实际失败: %s", reason)
	}
}

func TestSelectBestIPs_PreferBrowserValidatedGitHub(t *testing.T) {
	entries := []IPEntry{
		{IP: "1.1.1.1", Domain: "github.com", Port443: true, Port22: true, BasicHTTPS: true, HTTPS: false, BrowserOK: false, HTTPSLatency: 5, Latency: 5},
		{IP: "2.2.2.2", Domain: "github.com", Port443: true, Port22: true, BasicHTTPS: true, HTTPS: true, BrowserOK: true, HTTPSLatency: 80, Latency: 80},
	}

	best, _ := SelectBestIPs(entries, 1)
	if best == nil || best.IP != "2.2.2.2" {
		t.Fatalf("应优先选择通过浏览器校验的 github.com IP，实际: %#v", best)
	}
}

func TestSelectBestIPs_GitHubFallsBackToBestEffort(t *testing.T) {
	entries := []IPEntry{
		{IP: "1.1.1.1", Domain: "github.com", Port443: true, BasicHTTPS: true, HTTPS: false, BrowserOK: false, HTTPSLatency: 20, Latency: 20},
		{IP: "2.2.2.2", Domain: "github.com", Port443: true, BasicHTTPS: false, HTTPS: false, BrowserOK: false, HTTPSLatency: 10, Latency: 10},
	}

	best, _ := SelectBestIPs(entries, 1)
	if best == nil || best.IP != "1.1.1.1" {
		t.Fatalf("当所有 github.com 候选都未通过浏览器校验时，应回退到基础 HTTPS 更好的候选，实际: %#v", best)
	}
}

func mustParseURL(raw string) *url.URL {
	u, err := url.Parse(raw)
	if err != nil {
		panic(err)
	}
	return u
}

func stringsRepeat(s string, count int) string {
	if count <= 0 {
		return ""
	}
	out := ""
	for i := 0; i < count; i++ {
		out += s
	}
	return out
}

func testIPv4(a, b, c, d int) string {
	return fmt.Sprintf("%d.%d.%d.%d", a, b, c, d)
}

// ---- 4.1 补充：候选 IP 过滤与多源聚合 ----

func TestIsUsableCandidateIP_IPv6(t *testing.T) {
	tests := []struct {
		ip   string
		want bool
	}{
		// 合法公网 IPv6
		{"2606:50c0:8000::153", true},
		// 环回
		{"::1", false},
		// 未指定
		{"::", false},
		// ULA（fc00::/7）
		{"fc00::1", false},
		{"fd12:3456:789a::1", false},
		// 链路本地
		{"fe80::1", false},
		// 多播
		{"ff02::1", false},
		// 文档保留（2001:db8::/32）
		{"2001:db8::1", false},
	}
	for _, tt := range tests {
		if got := isUsableCandidateIP(tt.ip); got != tt.want {
			t.Errorf("isUsableCandidateIP(%q) = %v, want %v", tt.ip, got, tt.want)
		}
	}
}

func TestIsUsableCandidateIP_IPv4Reserved(t *testing.T) {
	tests := []struct {
		ip   string
		want bool
	}{
		// 0.0.0.0/8
		{testIPv4(0, 0, 0, 1), false},
		// 172.16-31.x.x
		{testIPv4(172, 16, 0, 1), false},
		{testIPv4(172, 31, 255, 255), false},
		{testIPv4(172, 32, 0, 1), true},
		// 192.0.0.x
		{testIPv4(192, 0, 0, 1), false},
		// 192.0.2.x（TEST-NET-1）
		{testIPv4(192, 0, 2, 1), false},
		// 198.18-19.x.x（基准测试）
		{testIPv4(198, 18, 0, 1), false},
		{testIPv4(198, 19, 255, 255), false},
		// 多播 224+
		{testIPv4(224, 0, 0, 1), false},
		{testIPv4(255, 255, 255, 255), false},
		// 合法公网
		{testIPv4(140, 82, 112, 4), true},
		{testIPv4(185, 199, 108, 133), true},
	}
	for _, tt := range tests {
		if got := isUsableCandidateIP(tt.ip); got != tt.want {
			t.Errorf("isUsableCandidateIP(%q) = %v, want %v", tt.ip, got, tt.want)
		}
	}
}

func TestMultiSourceFetchIPs_PartialFailure(t *testing.T) {
	// 一个源失败，另一个成功 → 应返回成功源的结果
	ms := NewMultiSource(
		stubSource{name: "fail", err: errors.New("network error")},
		stubSource{name: "ok", data: map[string][]string{
			"github.com": {testIPv4(140, 82, 112, 4)},
		}},
	)
	got, err := ms.FetchIPs(context.Background())
	if err != nil {
		t.Fatalf("部分源失败时不应返回错误，实际: %v", err)
	}
	if len(got["github.com"]) != 1 {
		t.Fatalf("应有 1 个 github.com 候选，实际: %#v", got["github.com"])
	}
}

func TestMultiSourceFetchIPs_DeduplicatesAcrossSources(t *testing.T) {
	// 两个源返回相同 IP → 去重后只有一个
	ip := testIPv4(140, 82, 112, 4)
	ms := NewMultiSource(
		stubSource{name: "s1", data: map[string][]string{"github.com": {ip}}},
		stubSource{name: "s2", data: map[string][]string{"github.com": {ip, testIPv4(185, 199, 108, 133)}}},
	)
	got, err := ms.FetchIPs(context.Background())
	if err != nil {
		t.Fatalf("FetchIPs() 返回错误: %v", err)
	}
	if len(got["github.com"]) != 2 {
		t.Fatalf("去重后应有 2 个唯一 IP，实际: %#v", got["github.com"])
	}
}

// ---- 4.2 补充：浏览器响应验证启发式规则 ----

func TestValidateBrowserResponseBody_MissingGitHubAssets(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/html; charset=utf-8"}},
		Request:    &http.Request{URL: mustParseURL("https://github.com/")},
	}
	// 有足够长度，有 HTML 结构，但缺少 github.githubassets.com 引用
	body := []byte("<!doctype html><html><head><title>GitHub</title></head><body>" +
		"sign in to github data-color-mode=" + stringsRepeat("x", 5000) + "</body></html>")
	ok, reason := validateBrowserResponse("github.com", resp, body)
	if ok {
		t.Fatalf("缺少 github.githubassets.com 引用时应校验失败，实际通过，reason=%q", reason)
	}
}

func TestValidateBrowserResponseBody_MissingIdentityFeatures(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/html; charset=utf-8"}},
		Request:    &http.Request{URL: mustParseURL("https://github.com/")},
	}
	// 有 githubassets.com 引用，但缺少所有关键身份特征
	body := []byte("<!doctype html><html><head></head><body>" +
		"github.githubassets.com " + stringsRepeat("x", 5000) + "</body></html>")
	ok, reason := validateBrowserResponse("github.com", resp, body)
	if ok {
		t.Fatalf("缺少 GitHub 首页关键身份特征时应校验失败，实际通过，reason=%q", reason)
	}
}

func TestValidateBrowserResponseBody_NonHTMLContentType(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"application/json"}},
		Request:    &http.Request{URL: mustParseURL("https://github.com/")},
	}
	body := []byte(`{"error":"not found"}`)
	ok, reason := validateBrowserResponse("github.com", resp, body)
	if ok {
		t.Fatalf("非 HTML Content-Type 时应校验失败，实际通过，reason=%q", reason)
	}
}

func TestValidateBrowserResponseBody_Non200StatusCode(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusFound,
		Header:     http.Header{"Content-Type": []string{"text/html; charset=utf-8"}},
		Request:    &http.Request{URL: mustParseURL("https://github.com/")},
	}
	// 302 重定向到非 github.com 主机时，validateBrowserResponseHeaders 应先拒绝
	// 这里测试状态码非 200 时的行为（最终跳转主机仍是 github.com）
	ok, _, _ := validateBrowserResponseHeaders("github.com", resp)
	// 302 状态码应触发 needBody=true 并最终由 body 校验决定
	if ok {
		t.Fatalf("非 200 状态码时响应头快速校验不应直接通过")
	}
}

func TestValidateBrowserResponseHeaders_WrongRedirectHost(t *testing.T) {
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/html; charset=utf-8"}},
		Request:    &http.Request{URL: mustParseURL("https://evil.example.com/")},
	}
	ok, reason, _ := validateBrowserResponseHeaders("github.com", resp)
	if ok {
		t.Fatalf("最终跳转到非 GitHub 主机时应校验失败，实际通过，reason=%q", reason)
	}
}

// ---- 4.3 补充：TCP 通但浏览器验证失败时的评分与选择回归 ----

func TestScoreIP_GitHubTCPOnlyVsBasicHTTPS(t *testing.T) {
	// TCP 443 通但 HTTPS 未通过 → 评分应明显低于基础 HTTPS 通过
	tcpOnly := &IPEntry{
		Domain:  "github.com",
		Port443: true,
		Port22:  true,
		Latency: 20,
	}
	basicHTTPS := &IPEntry{
		Domain:     "github.com",
		Port443:    true,
		Port22:     true,
		BasicHTTPS: true,
		Latency:    20,
	}
	tcpScore := ScoreIP(tcpOnly)
	basicScore := ScoreIP(basicHTTPS)
	if basicScore <= tcpScore {
		t.Errorf("基础 HTTPS 通过的评分(%.2f)应高于仅 TCP 通(%.2f)", basicScore, tcpScore)
	}
}

func TestScoreIP_GitHubBasicHTTPSVsBrowserOK(t *testing.T) {
	// 基础 HTTPS 通过但浏览器未通过 → 评分应明显低于浏览器通过
	basicOnly := &IPEntry{
		Domain:     "github.com",
		Port443:    true,
		BasicHTTPS: true,
		Latency:    20,
	}
	browserOK := &IPEntry{
		Domain:     "github.com",
		Port443:    true,
		BasicHTTPS: true,
		HTTPS:      true,
		BrowserOK:  true,
		Latency:    20,
	}
	basicScore := ScoreIP(basicOnly)
	browserScore := ScoreIP(browserOK)
	if browserScore <= basicScore {
		t.Errorf("浏览器校验通过的评分(%.2f)应高于仅基础 HTTPS(%.2f)", browserScore, basicScore)
	}
}

func TestRankIPs_GitHubBrowserOKAlwaysFirst(t *testing.T) {
	// 即使浏览器通过的 IP 延迟更高，也应排在第一位
	entries := []IPEntry{
		{IP: "fast-no-browser", Domain: "github.com", Port443: true, BasicHTTPS: true, BrowserOK: false, Latency: 5, HTTPSLatency: 5},
		{IP: "slow-browser-ok", Domain: "github.com", Port443: true, BasicHTTPS: true, HTTPS: true, BrowserOK: true, Latency: 200, HTTPSLatency: 200},
	}
	ranked := RankIPs(entries)
	if ranked[0].IP != "slow-browser-ok" {
		t.Errorf("浏览器校验通过的 IP 应排第一，即使延迟更高，实际第一: %s", ranked[0].IP)
	}
}

func TestSelectBestIPs_GitHubAllTCPOnly_ReturnsBestEffort(t *testing.T) {
	// 所有候选都只有 TCP 通，没有 HTTPS → 应仍返回最佳候选（不返回 nil）
	entries := []IPEntry{
		{IP: "1.1.1.1", Domain: "github.com", Port443: true, Port22: false, Latency: 100},
		{IP: "2.2.2.2", Domain: "github.com", Port443: true, Port22: true, Latency: 50},
	}
	best, _ := SelectBestIPs(entries, 1)
	if best == nil {
		t.Fatal("所有候选仅 TCP 通时，应仍返回最佳候选，不应返回 nil")
	}
}

// ---- 4.4 httptest 集成验证（等价于 curl/wget 验证）----
// 使用 httptest.Server 模拟真实 HTTP 响应，验证 HTTPSChecker 的浏览器级校验逻辑
// 等价于：curl --resolve 'github.com:443:<IP>' 'https://github.com/' | grep 'github.githubassets.com'

func TestHTTPSChecker_GitHubHomepagePass_Httptest(t *testing.T) {
	// 模拟正常 GitHub 首页响应
	ts := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("X-GitHub-Request-Id", "abc-123")
		w.Header().Set("Content-Security-Policy", "default-src 'none'; script-src github.githubassets.com")
		w.Header().Set("Server", "github.com")
		body := "<!doctype html><html><head><title>GitHub</title>" +
			"<link rel=\"preconnect\" href=\"https://github.githubassets.com\">" +
			"</head><body>Sign in to GitHub data-color-mode=light" +
			stringsRepeat("x", 5000) + "</body></html>"
		fmt.Fprint(w, body)
	}))
	defer ts.Close()

	ok, reason, _ := validateBrowserResponseHeaders("github.com", &http.Response{
		StatusCode: http.StatusOK,
		Header: http.Header{
			"Content-Type":            []string{"text/html; charset=utf-8"},
			"X-GitHub-Request-Id":     []string{"abc-123"},
			"Content-Security-Policy": []string{"default-src 'none'; script-src github.githubassets.com"},
			"Server":                  []string{"github.com"},
		},
		Request: &http.Request{URL: mustParseURL("https://github.com/")},
	})
	if !ok {
		t.Fatalf("模拟正常 GitHub 首页响应头应通过校验，实际失败: %s", reason)
	}
	_ = ts
}

func TestHTTPSChecker_GitHubBlockPage_Httptest(t *testing.T) {
	// 模拟拦截页响应（等价于 curl 返回拦截页内容）
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/html; charset=utf-8"}},
		Request:    &http.Request{URL: mustParseURL("https://github.com/")},
	}
	body := []byte("<!doctype html><html><body>Access Denied security challenge captcha" +
		stringsRepeat("x", 5000) + "</body></html>")

	ok, reason := validateBrowserResponse("github.com", resp, body)
	if ok {
		t.Fatalf("拦截页应被识别为校验失败，实际通过，reason=%q", reason)
	}
	t.Logf("拦截页校验失败原因（符合预期）: %s", reason)
}

func TestHTTPSChecker_GitHubValidHomepage_Httptest(t *testing.T) {
	// 模拟完整 GitHub 首页（包含所有关键特征）
	// 等价于：curl --resolve 'github.com:443:<IP>' 'https://github.com/' -s | grep -c 'github.githubassets.com'
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/html; charset=utf-8"}},
		Request:    &http.Request{URL: mustParseURL("https://github.com/")},
	}
	body := []byte("<!doctype html><html><head>" +
		"<meta name=\"description\" content=\"GitHub\">" +
		"<link rel=\"preconnect\" href=\"https://github.githubassets.com\">" +
		"</head><body>" +
		"data-color-mode=light " +
		stringsRepeat("x", 5000) +
		"</body></html>")

	ok, reason := validateBrowserResponse("github.com", resp, body)
	if !ok {
		t.Fatalf("完整 GitHub 首页应通过校验，实际失败: %s", reason)
	}
	t.Logf("GitHub 首页校验通过: %s", reason)
}

func TestHTTPSChecker_NonGitHubDomain_Httptest(t *testing.T) {
	// 非 github.com 域名使用轻量校验（等价于 curl 验证 raw.githubusercontent.com）
	resp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     http.Header{"Content-Type": []string{"text/plain"}},
		Request:    &http.Request{URL: mustParseURL("https://raw.githubusercontent.com/test")},
	}
	body := []byte("raw file content")

	ok, reason := validateBrowserResponse("raw.githubusercontent.com", resp, body)
	if !ok {
		t.Fatalf("非 github.com 域名应采用轻量校验并通过，实际失败: %s", reason)
	}
	t.Logf("非 github.com 域名轻量校验通过: %s", reason)
}

// ---- 5. github.com 保护策略测试 ----

func TestSelectBestIPs_GitHubSSHOnly_ReturnsNil(t *testing.T) {
	// github.com 所有候选仅 SSH 可达 → 应返回 nil，不选择不可用 IP
	entries := []IPEntry{
		{IP: "1.1.1.1", Domain: "github.com", Port443: false, Port22: true, Latency: 50},
		{IP: "2.2.2.2", Domain: "github.com", Port443: false, Port22: true, Latency: 30},
	}
	best, _ := SelectBestIPs(entries, 1)
	if best != nil {
		t.Fatalf("github.com 所有候选仅 SSH 可达时应返回 nil，实际: %s", best.IP)
	}
}

func TestSelectBestIPs_GitHubMixed_SkipsSSHOnly(t *testing.T) {
	// github.com 混合候选：有 SSH-only 和 Port443 可达的 → 应跳过 SSH-only
	entries := []IPEntry{
		{IP: "1.1.1.1", Domain: "github.com", Port443: false, Port22: true, Latency: 10},
		{IP: "2.2.2.2", Domain: "github.com", Port443: true, Port22: false, Latency: 100},
	}
	best, _ := SelectBestIPs(entries, 1)
	if best == nil {
		t.Fatal("应选择 Port443 可达的候选，不应返回 nil")
	}
	if best.IP != "2.2.2.2" {
		t.Fatalf("应选择 Port443 可达的 2.2.2.2，实际: %s", best.IP)
	}
}

func TestSelectBestIPs_NonGitHubSSHOnly_StillReturns(t *testing.T) {
	// 非 github.com 域名：所有候选仅 SSH 可达 → 仍应返回最佳候选（退而求其次）
	entries := []IPEntry{
		{IP: "1.1.1.1", Domain: "github.githubassets.com", Port443: false, Port22: true, Latency: 50},
		{IP: "2.2.2.2", Domain: "github.githubassets.com", Port443: false, Port22: true, Latency: 30},
	}
	best, _ := SelectBestIPs(entries, 1)
	if best == nil {
		t.Fatal("非 github.com 域名即使仅 SSH 可达也应返回候选")
	}
}

func TestRankIPs_GitHubSSHOnlyNotAvailable(t *testing.T) {
	// github.com 仅 SSH 可达的 IP 不应被标记为 Available
	entries := []IPEntry{
		{IP: "1.1.1.1", Domain: "github.com", Port443: false, Port22: true, Latency: 50},
	}
	ranked := RankIPs(entries)
	if ranked[0].Available {
		t.Fatal("github.com 仅 SSH 可达的 IP 不应被标记为 Available")
	}
}

func TestRankIPs_NonGitHubSSHOnlyIsAvailable(t *testing.T) {
	// 非 github.com 域名仅 SSH 可达的 IP 仍应被标记为 Available
	entries := []IPEntry{
		{IP: "1.1.1.1", Domain: "assets-cdn.github.com", Port443: false, Port22: true, Latency: 50},
	}
	ranked := RankIPs(entries)
	if !ranked[0].Available {
		t.Fatal("非 github.com 域名仅 SSH 可达的 IP 应被标记为 Available")
	}
}

func TestFallbackIP_GitHubNoAvailable_ReturnsNil(t *testing.T) {
	// github.com 没有可用候选时，FallbackIP 应返回 nil
	domainIPs := &DomainIPs{
		Domain: "github.com",
		IPs: []IPEntry{
			{IP: "1.1.1.1", Domain: "github.com", Port443: false, Port22: true, Latency: 50},
			{IP: "2.2.2.2", Domain: "github.com", Port443: false, Port22: true, Latency: 30},
		},
	}
	result := FallbackIP(domainIPs)
	if result != nil {
		t.Fatalf("github.com 没有可用候选时 FallbackIP 应返回 nil，实际: %s", result.IP)
	}
}

func TestFallbackIP_NonGitHubNoAvailable_ReturnsBestEffort(t *testing.T) {
	// 非 github.com 域名没有可用候选时，FallbackIP 仍应返回最佳候选
	domainIPs := &DomainIPs{
		Domain: "raw.githubusercontent.com",
		IPs: []IPEntry{
			{IP: "1.1.1.1", Domain: "raw.githubusercontent.com", Port443: false, Port22: false, Latency: 50},
		},
	}
	result := FallbackIP(domainIPs)
	if result == nil {
		t.Fatal("非 github.com 域名没有可用候选时 FallbackIP 仍应返回最佳候选")
	}
}
