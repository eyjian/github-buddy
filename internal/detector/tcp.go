package detector

import (
	"context"
	"fmt"
	"net"
	"sync"
	"time"
)

// TCPCheckResult 表示一次 TCP 端口检测结果
type TCPCheckResult struct {
	IP      string
	Port    int
	OK      bool
	Latency float64 // 连接耗时（毫秒）
	Error   error
}

// TCPChecker TCP 端口检测器
type TCPChecker struct {
	timeout time.Duration
}

// NewTCPChecker 创建 TCP 端口检测器
func NewTCPChecker(timeout time.Duration) *TCPChecker {
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	return &TCPChecker{timeout: timeout}
}

// CheckPort 检测指定 IP 的 TCP 端口是否可连通
func (c *TCPChecker) CheckPort(ctx context.Context, ip string, port int) TCPCheckResult {
	result := TCPCheckResult{IP: ip, Port: port}
	addr := fmt.Sprintf("%s:%d", ip, port)

	start := time.Now()
	dialer := net.Dialer{Timeout: c.timeout}
	conn, err := dialer.DialContext(ctx, "tcp", addr)
	elapsed := time.Since(start)

	if err != nil {
		result.Error = fmt.Errorf("TCP 连接 %s 失败: %w", addr, err)
		return result
	}
	defer conn.Close()

	result.OK = true
	result.Latency = float64(elapsed.Milliseconds())
	return result
}

// CheckPorts 并发检测指定 IP 的多个 TCP 端口
func (c *TCPChecker) CheckPorts(ctx context.Context, ip string, ports []int) []TCPCheckResult {
	results := make([]TCPCheckResult, len(ports))
	var wg sync.WaitGroup

	for i, port := range ports {
		wg.Add(1)
		go func(idx int, p int) {
			defer wg.Done()
			results[idx] = c.CheckPort(ctx, ip, p)
		}(i, port)
	}

	wg.Wait()
	return results
}

// CheckIPPorts 对多个 IP 并发检测指定端口列表
func (c *TCPChecker) CheckIPPorts(ctx context.Context, ips []string, ports []int) map[string][]TCPCheckResult {
	result := make(map[string][]TCPCheckResult)
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, ip := range ips {
		wg.Add(1)
		go func(ipAddr string) {
			defer wg.Done()
			portResults := c.CheckPorts(ctx, ipAddr, ports)
			mu.Lock()
			result[ipAddr] = portResults
			mu.Unlock()
		}(ip)
	}

	wg.Wait()
	return result
}
