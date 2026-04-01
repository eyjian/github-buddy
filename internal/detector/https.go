package detector

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// HTTPSCheckResult 表示一次 HTTPS 验证结果
type HTTPSCheckResult struct {
	IP             string
	Domain         string
	OK             bool    // 最终 HTTPS 验证是否成功
	BasicOK        bool    // 基础 HTTPS 是否成功（TLS 握手 + HTTP 响应）
	BrowserOK      bool    // 浏览器导向验证是否成功（仅对 github.com 严格要求）
	BrowserReason  string  // 浏览器导向验证结果说明
	Port443OK      bool    // TCP 443 是否已建立连接（即便后续 TLS/HTTP 失败也会保留）
	ConnectLatency float64 // TCP 443 建连延迟（毫秒）
	Latency        float64 // HTTPS 完整延迟（毫秒），包含 TCP 建连 + TLS 握手 + HTTP 首包往返
	StatusCode     int
	BodyBytes      int64
	Error          error
}

// HTTPSChecker HTTPS 应用层验证器
// 通过真正的 HTTPS 请求（含 TLS 证书校验）验证候选 IP 的实际可用性
type HTTPSChecker struct {
	timeout time.Duration
	limiter chan struct{}
}

const (
	defaultHTTPSProbeConcurrency = 12
	githubValidationBodyLimit    = 12 * 1024
)

// NewHTTPSChecker 创建 HTTPS 验证器
func NewHTTPSChecker(timeout time.Duration) *HTTPSChecker {
	return NewHTTPSCheckerWithConcurrency(timeout, defaultHTTPSProbeConcurrency)
}

// NewHTTPSCheckerWithConcurrency 创建带并发限制的 HTTPS 验证器
func NewHTTPSCheckerWithConcurrency(timeout time.Duration, concurrency int) *HTTPSChecker {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	if concurrency <= 0 {
		concurrency = defaultHTTPSProbeConcurrency
	}
	return &HTTPSChecker{
		timeout: timeout,
		limiter: make(chan struct{}, concurrency),
	}
}

// browserUserAgent 模拟现代浏览器的 User-Agent，避免被服务端或中间设备区别对待
const browserUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36"

// Check 对指定 IP 执行 HTTPS 验证
// 流程：TCP 443 建连 → TLS 握手（验证证书匹配目标域名） → HTTP GET 请求 → 先看响应头，再按需读取少量正文校验页面特征
func (c *HTTPSChecker) Check(ctx context.Context, ip, domain string) HTTPSCheckResult {
	result := HTTPSCheckResult{IP: ip, Domain: domain}

	if c.limiter != nil {
		select {
		case c.limiter <- struct{}{}:
			defer func() { <-c.limiter }()
		case <-ctx.Done():
			result.Error = ctx.Err()
			return result
		}
	}

	start := time.Now()

	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: func(dialCtx context.Context, network, addr string) (net.Conn, error) {
			dialer := &net.Dialer{Timeout: c.timeout}
			connectStart := time.Now()
			conn, err := dialer.DialContext(dialCtx, "tcp", net.JoinHostPort(ip, "443"))
			if err != nil {
				return nil, fmt.Errorf("TCP 连接 %s:443 失败: %w", ip, err)
			}
			result.Port443OK = true
			result.ConnectLatency = float64(time.Since(connectStart).Milliseconds())
			return conn, nil
		},
		TLSClientConfig: &tls.Config{
			ServerName: domain,
			MinVersion: tls.VersionTLS12,
		},
		DisableKeepAlives:   true,
		ForceAttemptHTTP2:   true,
		MaxIdleConns:        0,
		MaxIdleConnsPerHost: 1,
	}
	defer transport.CloseIdleConnections()

	client := &http.Client{
		Transport: transport,
		Timeout:   c.timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 3 {
				return http.ErrUseLastResponse
			}
			if !sameRedirectHost(domain, req.URL.Hostname()) {
				return http.ErrUseLastResponse
			}
			return nil
		},
	}

	url := fmt.Sprintf("https://%s/", domain)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		result.Error = fmt.Errorf("创建 HTTPS 请求失败: %w", err)
		return result
	}
	req.Header.Set("User-Agent", browserUserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")

	resp, err := client.Do(req)
	result.Latency = float64(time.Since(start).Milliseconds())
	if err != nil {
		if !result.Port443OK {
			result.ConnectLatency = 0
		}
		result.Error = fmt.Errorf("HTTPS 验证失败 (IP=%s, Domain=%s): %w", ip, domain, err)
		return result
	}
	defer resp.Body.Close()

	result.StatusCode = resp.StatusCode
	if resp.StatusCode >= 500 {
		result.Error = fmt.Errorf("HTTPS 验证失败: 状态码 %d (IP=%s, Domain=%s)", resp.StatusCode, ip, domain)
		return result
	}

	result.BasicOK = true

	browserOK, browserReason, needBody := validateBrowserResponseHeaders(domain, resp)
	var body []byte
	if !browserOK && needBody {
		body, err = io.ReadAll(io.LimitReader(resp.Body, githubValidationBodyLimit))
		if err != nil {
			result.Error = fmt.Errorf("读取 HTTPS 响应失败 (IP=%s, Domain=%s): %w", ip, domain, err)
			return result
		}
		result.BodyBytes = int64(len(body))
		browserOK, browserReason = validateBrowserResponseBody(domain, body, browserReason)
		if resp.ContentLength > result.BodyBytes {
			result.BodyBytes = resp.ContentLength
		}
		result.Latency = float64(time.Since(start).Milliseconds())
		if domain != "github.com" {
			_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1024))
		}
	} else if resp.ContentLength > 0 {
		result.BodyBytes = resp.ContentLength
	}

	result.BrowserOK = browserOK
	result.BrowserReason = browserReason
	if domain != "github.com" {
		result.OK = result.BasicOK
	} else {
		result.OK = result.BasicOK && result.BrowserOK
	}

	if result.OK {
		return result
	}
	if domain == "github.com" && result.BasicOK && !result.BrowserOK {
		result.Error = fmt.Errorf("github.com 首页特征校验未通过: %s (IP=%s)", result.BrowserReason, ip)
		return result
	}
	result.Error = fmt.Errorf("HTTPS 验证失败 (IP=%s, Domain=%s)", ip, domain)
	return result
}

func sameRedirectHost(domain, host string) bool {
	domain = normalizeHost(domain)
	host = normalizeHost(host)
	if host == domain {
		return true
	}
	return domain == "github.com" && host == "www.github.com"
}

func normalizeHost(host string) string {
	return strings.TrimSuffix(strings.ToLower(strings.TrimSpace(host)), ".")
}

func validateBrowserResponse(domain string, resp *http.Response, body []byte) (bool, string) {
	ok, reason, needBody := validateBrowserResponseHeaders(domain, resp)
	if !ok && needBody {
		return validateBrowserResponseBody(domain, body, reason)
	}
	return ok, reason
}

func validateBrowserResponseHeaders(domain string, resp *http.Response) (bool, string, bool) {
	if domain != "github.com" {
		return true, "非 github.com 域名采用轻量 HTTPS 校验", false
	}
	if resp == nil {
		return false, "缺少响应对象", false
	}
	if resp.Request == nil || resp.Request.URL == nil {
		return false, "缺少最终请求信息", false
	}
	if !sameRedirectHost(domain, resp.Request.URL.Hostname()) {
		return false, fmt.Sprintf("最终跳转主机异常: %s", resp.Request.URL.Hostname()), false
	}

	contentType := strings.ToLower(resp.Header.Get("Content-Type"))
	if contentType != "" && !strings.Contains(contentType, "text/html") {
		return false, fmt.Sprintf("Content-Type 不是 HTML: %s", resp.Header.Get("Content-Type")), false
	}

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Sprintf("首页状态码异常: %d", resp.StatusCode), true
	}

	headerSignals := 0
	if strings.Contains(strings.ToLower(resp.Header.Get("Server")), "github.com") {
		headerSignals++
	}
	if resp.Header.Get("X-GitHub-Request-Id") != "" {
		headerSignals++
	}
	if strings.Contains(strings.ToLower(resp.Header.Get("Content-Security-Policy")), "github.githubassets.com") {
		headerSignals++
	}
	if hasGitHubSessionCookie(resp.Header.Values("Set-Cookie")) {
		headerSignals++
	}

	if headerSignals >= 2 {
		return true, "github.com 首页响应头特征校验通过", false
	}

	return false, "响应头缺少足够的 GitHub 首页特征", true
}

func validateBrowserResponseBody(domain string, body []byte, headerReason string) (bool, string) {
	if domain != "github.com" {
		return true, "非 github.com 域名采用轻量 HTTPS 校验"
	}

	if len(body) < 4096 {
		return false, joinBrowserReasons(headerReason, fmt.Sprintf("响应体过短: %dB", len(body)))
	}

	bodyLower := strings.ToLower(string(body))
	if looksLikeBlockPage(bodyLower) {
		return false, "疑似拦截页或错误中间页"
	}
	if !strings.Contains(bodyLower, "<!doctype html") || !strings.Contains(bodyLower, "<html") {
		return false, "缺少 HTML 首页结构"
	}
	if !strings.Contains(bodyLower, "github.githubassets.com") {
		return false, "缺少 GitHub 静态资源域名引用"
	}
	if !containsAny(bodyLower,
		"content=\"github\"",
		"<title>github",
		"octolytics-url",
		"app-argument=",
		"data-color-mode=",
		"copilotapioverrideurl",
		"mona sans",
		"monasans",
		"sign in to github",
		"github copilot",
	) {
		return false, "缺少 GitHub 首页关键身份特征"
	}

	return true, "github.com 首页特征校验通过"
}

func hasGitHubSessionCookie(values []string) bool {
	for _, value := range values {
		value = strings.ToLower(value)
		if strings.Contains(value, "_gh_sess=") || strings.Contains(value, "logged_in=") || strings.Contains(value, "_octo=") {
			return true
		}
	}
	return false
}

func joinBrowserReasons(parts ...string) string {
	filtered := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			filtered = append(filtered, part)
		}
	}
	if len(filtered) == 0 {
		return "原因未记录"
	}
	return strings.Join(filtered, "；")
}

func looksLikeBlockPage(body string) bool {
	return containsAny(body,
		"access denied",
		"request blocked",
		"web page blocked",
		"this site can't be reached",
		"security challenge",
		"captcha",
	) || (strings.Contains(body, "err_") && !strings.Contains(body, "github.githubassets.com"))
}

func containsAny(body string, markers ...string) bool {
	for _, marker := range markers {
		if strings.Contains(body, strings.ToLower(marker)) {
			return true
		}
	}
	return false
}

// CheckIPs 并发对多个 IP 执行 HTTPS 验证
func (c *HTTPSChecker) CheckIPs(ctx context.Context, ips []string, domain string) []HTTPSCheckResult {
	results := make([]HTTPSCheckResult, len(ips))
	var wg sync.WaitGroup

	for i, ip := range ips {
		wg.Add(1)
		go func(idx int, ipAddr string) {
			defer wg.Done()
			results[idx] = c.Check(ctx, ipAddr, domain)
		}(i, ip)
	}

	wg.Wait()
	return results
}
