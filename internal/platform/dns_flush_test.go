package platform

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"testing"
	"time"
)

func TestFlushDNSCache_Linux(t *testing.T) {
	ctx := context.Background()
	result := FlushDNSCache(ctx, "linux")

	// 在测试环境中命令可能不存在，但不应 panic
	if result == nil {
		t.Fatal("FlushDNSCache 返回 nil")
	}

	// 无论成功与否，Message 不应为空
	if result.Message == "" {
		t.Error("FlushResult.Message 不应为空")
	}
}

func TestFlushDNSCache_Darwin(t *testing.T) {
	ctx := context.Background()
	result := FlushDNSCache(ctx, "darwin")

	if result == nil {
		t.Fatal("FlushDNSCache 返回 nil")
	}

	if result.Message == "" {
		t.Error("FlushResult.Message 不应为空")
	}
}

func TestFlushDNSCache_Windows(t *testing.T) {
	ctx := context.Background()
	result := FlushDNSCache(ctx, "windows")

	if result == nil {
		t.Fatal("FlushDNSCache 返回 nil")
	}

	if result.Message == "" {
		t.Error("FlushResult.Message 不应为空")
	}
}

func TestFlushDNSCache_UnsupportedOS(t *testing.T) {
	ctx := context.Background()
	result := FlushDNSCache(ctx, "freebsd")

	if result == nil {
		t.Fatal("FlushDNSCache 返回 nil")
	}

	if result.Success {
		t.Error("不支持的 OS 不应返回 Success=true")
	}

	if result.Error == nil {
		t.Error("不支持的 OS 应返回 Error")
	}
}

func TestFlushDNSCache_AutoDetectOS(t *testing.T) {
	ctx := context.Background()
	// 传空字符串应自动检测 OS
	result := FlushDNSCache(ctx, "")

	if result == nil {
		t.Fatal("FlushDNSCache 返回 nil")
	}

	// 自动检测不应导致 panic，Message 应有值
	if result.Message == "" {
		t.Error("FlushResult.Message 不应为空")
	}
}

func TestFlushDNSCache_Timeout(t *testing.T) {
	// 创建一个已过期的 context 来模拟超时
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Nanosecond)
	defer cancel()

	// 等待 context 过期
	time.Sleep(10 * time.Millisecond)

	result := FlushDNSCache(ctx, "linux")

	if result == nil {
		t.Fatal("FlushDNSCache 返回 nil")
	}

	// 超时的 context 应该导致失败
	if result.Success {
		t.Error("超时的 context 不应返回 Success=true")
	}
}

// captureStdout 捕获 stdout 输出用于验证
func captureStdout(fn func()) string {
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}

func TestPrintFlushResult_Success(t *testing.T) {
	result := &FlushResult{
		Success: true,
		Message: "已刷新",
	}

	output := captureStdout(func() {
		PrintFlushResult(result, "linux")
	})

	if !bytes.Contains([]byte(output), []byte("已自动刷新系统 DNS 缓存")) {
		t.Errorf("成功时应包含成功提示，实际输出: %s", output)
	}

	if !bytes.Contains([]byte(output), []byte("chrome://net-internals/#dns")) {
		t.Errorf("应始终包含 Chrome 浏览器缓存提示，实际输出: %s", output)
	}

	if !bytes.Contains([]byte(output), []byte("edge://net-internals/#dns")) {
		t.Errorf("应始终包含 Edge 浏览器缓存提示，实际输出: %s", output)
	}

	if !bytes.Contains([]byte(output), []byte("about:networking#dns")) {
		t.Errorf("应始终包含 Firefox 浏览器缓存提示，实际输出: %s", output)
	}
}

func TestPrintFlushResult_Failure_Linux(t *testing.T) {
	result := &FlushResult{
		Success: false,
		Message: "失败",
		Error:   fmt.Errorf("命令不存在"),
	}

	output := captureStdout(func() {
		PrintFlushResult(result, "linux")
	})

	if !bytes.Contains([]byte(output), []byte("自动刷新系统 DNS 缓存失败")) {
		t.Errorf("失败时应包含失败提示，实际输出: %s", output)
	}

	if !bytes.Contains([]byte(output), []byte("systemd-resolve")) {
		t.Errorf("Linux 失败时应包含 systemd-resolve 手动命令，实际输出: %s", output)
	}

	if !bytes.Contains([]byte(output), []byte("systemctl restart systemd-resolved")) {
		t.Errorf("Linux 失败时应包含 systemctl restart systemd-resolved 手动命令，实际输出: %s", output)
	}

	if !bytes.Contains([]byte(output), []byte("chrome://net-internals/#dns")) {
		t.Errorf("应始终包含浏览器缓存提示，实际输出: %s", output)
	}
}

func TestPrintFlushResult_Failure_Darwin(t *testing.T) {
	result := &FlushResult{
		Success: false,
		Message: "失败",
		Error:   fmt.Errorf("命令不存在"),
	}

	output := captureStdout(func() {
		PrintFlushResult(result, "darwin")
	})

	if !bytes.Contains([]byte(output), []byte("dscacheutil")) {
		t.Errorf("macOS 失败时应包含 dscacheutil 手动命令，实际输出: %s", output)
	}

	// macOS 应包含 Safari 缓存提示
	if !bytes.Contains([]byte(output), []byte("Safari")) {
		t.Errorf("macOS 应包含 Safari 浏览器缓存提示，实际输出: %s", output)
	}
}

func TestPrintFlushResult_Failure_Windows(t *testing.T) {
	result := &FlushResult{
		Success: false,
		Message: "失败",
		Error:   fmt.Errorf("命令不存在"),
	}

	output := captureStdout(func() {
		PrintFlushResult(result, "windows")
	})

	if !bytes.Contains([]byte(output), []byte("ipconfig /flushdns")) {
		t.Errorf("Windows 失败时应包含 ipconfig 手动命令，实际输出: %s", output)
	}
}

func TestPrintFlushResult_AlwaysShowBrowserHint(t *testing.T) {
	cases := []struct {
		name    string
		success bool
		osType  string
	}{
		{"success_linux", true, "linux"},
		{"failure_linux", false, "linux"},
		{"success_darwin", true, "darwin"},
		{"failure_darwin", false, "darwin"},
		{"success_windows", true, "windows"},
		{"failure_windows", false, "windows"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := &FlushResult{
				Success: tc.success,
				Message: "测试",
			}
			if !tc.success {
				result.Error = fmt.Errorf("测试错误")
			}

			output := captureStdout(func() {
				PrintFlushResult(result, tc.osType)
			})

			// 所有平台都应包含 Chrome、Edge、Firefox 提示
			if !bytes.Contains([]byte(output), []byte("chrome://net-internals/#dns")) {
				t.Errorf("[%s] 应始终包含 Chrome 浏览器缓存提示，实际输出: %s", tc.name, output)
			}
			if !bytes.Contains([]byte(output), []byte("edge://net-internals/#dns")) {
				t.Errorf("[%s] 应始终包含 Edge 浏览器缓存提示，实际输出: %s", tc.name, output)
			}
			if !bytes.Contains([]byte(output), []byte("about:networking#dns")) {
				t.Errorf("[%s] 应始终包含 Firefox 浏览器缓存提示，实际输出: %s", tc.name, output)
			}

			// 仅 macOS 应包含 Safari 提示
			hasSafari := bytes.Contains([]byte(output), []byte("Safari"))
			if tc.osType == "darwin" && !hasSafari {
				t.Errorf("[%s] macOS 应包含 Safari 浏览器缓存提示，实际输出: %s", tc.name, output)
			}
			if tc.osType != "darwin" && hasSafari {
				t.Errorf("[%s] 非 macOS 不应包含 Safari 浏览器缓存提示，实际输出: %s", tc.name, output)
			}
		})
	}
}
