package detector

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"
)

// HTTPSCheckResult 表示一次 HTTPS 验证结果
type HTTPSCheckResult struct {
	IP      string
	Domain  string
	OK      bool    // TLS 握手 + HTTP 响应是否成功
	Latency float64 // HTTPS 完整延迟（毫秒），包含 TLS 握手 + HTTP 往返
	Error   error
}

// HTTPSChecker HTTPS 应用层验证器
// 通过真正的 HTTPS 请求（含 TLS 证书校验）验证候选 IP 的实际可用性
type HTTPSChecker struct {
	timeout time.Duration
}

// NewHTTPSChecker 创建 HTTPS 验证器
func NewHTTPSChecker(timeout time.Duration) *HTTPSChecker {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &HTTPSChecker{timeout: timeout}
}

// Check 对指定 IP 执行 HTTPS 验证
// 流程：TLS 握手（验证证书匹配目标域名） → HTTP HEAD 请求 → 验证状态码
func (c *HTTPSChecker) Check(ctx context.Context, ip, domain string) HTTPSCheckResult {
	result := HTTPSCheckResult{IP: ip, Domain: domain}

	start := time.Now()

	// 创建自定义 Transport：强制连接指定 IP，但 TLS 验证目标域名
	transport := &http.Transport{
		DialTLSContext: func(dialCtx context.Context, network, addr string) (net.Conn, error) {
			// 直接连接候选 IP 的 443 端口
			dialer := &net.Dialer{Timeout: c.timeout}
			rawConn, err := dialer.DialContext(dialCtx, "tcp", net.JoinHostPort(ip, "443"))
			if err != nil {
				return nil, fmt.Errorf("TCP 连接 %s:443 失败: %w", ip, err)
			}

			// 在 TCP 连接上进行 TLS 握手，ServerName 设为目标域名以验证证书
			tlsConn := tls.Client(rawConn, &tls.Config{
				ServerName: domain,
				MinVersion: tls.VersionTLS12,
			})

			if err := tlsConn.HandshakeContext(dialCtx); err != nil {
				rawConn.Close()
				return nil, fmt.Errorf("TLS 握手失败 (IP=%s, Domain=%s): %w", ip, domain, err)
			}

			return tlsConn, nil
		},
		DisableKeepAlives: true,
	}

	client := &http.Client{
		Transport: transport,
		Timeout:   c.timeout,
		// 不跟随重定向，只检查第一个响应
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	// 构造 HTTP HEAD 请求，Host 设为目标域名
	url := fmt.Sprintf("https://%s/", domain)
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		result.Error = fmt.Errorf("创建 HTTPS 请求失败: %w", err)
		return result
	}

	resp, err := client.Do(req)
	elapsed := time.Since(start)

	if err != nil {
		result.Error = fmt.Errorf("HTTPS 验证失败 (IP=%s, Domain=%s): %w", ip, domain, err)
		return result
	}
	defer resp.Body.Close()

	// 状态码 < 500 视为可用（允许 3xx 重定向、4xx 认证等正常响应）
	if resp.StatusCode < 500 {
		result.OK = true
		result.Latency = float64(elapsed.Milliseconds())
	} else {
		result.Error = fmt.Errorf("HTTPS 验证失败: 状态码 %d (IP=%s, Domain=%s)", resp.StatusCode, ip, domain)
	}

	return result
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
