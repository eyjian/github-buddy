package detector

import (
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// PingResult 表示一次 ping 检测结果
type PingResult struct {
	IP       string
	Latency  float64 // 平均延迟（毫秒）
	LossRate float64 // 丢包率（0.0-1.0）
	OK       bool    // 是否成功
	Error    error
}

// Pinger 通过系统 ping 命令检测 IP 可达性
type Pinger struct {
	count   int           // ping 次数
	timeout time.Duration // 超时时间
}

// NewPinger 创建 Pinger 实例
func NewPinger(count int, timeout time.Duration) *Pinger {
	if count <= 0 {
		count = 3
	}
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &Pinger{count: count, timeout: timeout}
}

// Ping 对指定 IP 执行 ping 检测
func (p *Pinger) Ping(ctx context.Context, ip string) PingResult {
	result := PingResult{IP: ip}

	args := p.buildArgs(ip)
	cmd := exec.CommandContext(ctx, "ping", args...)

	output, err := cmd.CombinedOutput()
	if err != nil {
		// ping 命令执行失败，可能是权限问题或目标不可达
		result.Error = fmt.Errorf("ping %s 失败: %w", ip, err)
		result.LossRate = 1.0
		return result
	}

	// 解析 ping 输出
	outputStr := string(output)
	result.Latency = parsePingLatency(outputStr)
	result.LossRate = parsePingLoss(outputStr)
	result.OK = result.LossRate < 1.0 && result.Latency > 0

	return result
}

// IsPingAvailable 检测系统 ping 命令是否可用
func IsPingAvailable() bool {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("ping", "-n", "1", "-w", "1000", "127.0.0.1")
	default:
		cmd = exec.Command("ping", "-c", "1", "-W", "1", "127.0.0.1")
	}
	return cmd.Run() == nil
}

// buildArgs 根据操作系统构建 ping 命令参数
func (p *Pinger) buildArgs(ip string) []string {
	timeoutSec := int(p.timeout.Seconds())
	if timeoutSec <= 0 {
		timeoutSec = 5
	}

	switch runtime.GOOS {
	case "windows":
		return []string{
			"-n", strconv.Itoa(p.count),
			"-w", strconv.Itoa(timeoutSec * 1000), // Windows 用毫秒
			ip,
		}
	default: // linux, darwin
		return []string{
			"-c", strconv.Itoa(p.count),
			"-W", strconv.Itoa(timeoutSec),
			ip,
		}
	}
}

// parsePingLatency 从 ping 输出中解析平均延迟（毫秒）
func parsePingLatency(output string) float64 {
	// Linux/macOS 格式: rtt min/avg/max/mdev = 1.234/5.678/9.012/0.123 ms
	re := regexp.MustCompile(`=\s*[\d.]+/([\d.]+)/[\d.]+`)
	matches := re.FindStringSubmatch(output)
	if len(matches) >= 2 {
		if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
			return val
		}
	}

	// Windows 格式: Average = 5ms 或 平均 = 5ms
	re2 := regexp.MustCompile(`(?:Average|平均)\s*=\s*(\d+)\s*ms`)
	matches2 := re2.FindStringSubmatch(output)
	if len(matches2) >= 2 {
		if val, err := strconv.ParseFloat(matches2[1], 64); err == nil {
			return val
		}
	}

	return 0
}

// parsePingLoss 从 ping 输出中解析丢包率
func parsePingLoss(output string) float64 {
	// Linux/macOS 格式: 3 packets transmitted, 3 received, 0% packet loss
	// Windows 格式: (0% 丢失) 或 (0% loss)
	re := regexp.MustCompile(`(\d+)%\s*(?:packet\s+loss|丢失|loss)`)
	matches := re.FindStringSubmatch(output)
	if len(matches) >= 2 {
		if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
			return val / 100.0
		}
	}

	// 如果无法解析，检查是否有 "0 received" 表示全部丢失
	if strings.Contains(output, "0 received") || strings.Contains(output, "0 packets received") {
		return 1.0
	}

	return 0
}
