package platform

import (
	"os"
	"runtime"
	"testing"
)
func TestDetect(t *testing.T) {
	info, err := Detect()
	if err != nil {
		t.Fatalf("Detect() 失败: %v", err)
	}

	if info.OS != runtime.GOOS {
		t.Errorf("OS = %s, want %s", info.OS, runtime.GOOS)
	}

	if info.HomeDir == "" {
		t.Error("HomeDir 不应为空")
	}

	if info.HostPath == "" {
		t.Error("HostPath 不应为空")
	}

	if info.DataDir == "" {
		t.Error("DataDir 不应为空")
	}

	// 验证平台特定路径
	switch runtime.GOOS {
	case "windows":
		if info.HostsDir == "/etc" {
			t.Error("Windows 下 HostsDir 不应为 /etc")
		}
	default:
		if info.HostPath != "/etc/hosts" {
			t.Errorf("Linux/macOS 下 HostPath = %s, want /etc/hosts", info.HostPath)
		}
	}
}

func TestIsWindows(t *testing.T) {
	info := &Info{OS: "windows"}
	if !info.IsWindows() {
		t.Error("OS=windows 时 IsWindows() 应返回 true")
	}

	info = &Info{OS: "linux"}
	if info.IsWindows() {
		t.Error("OS=linux 时 IsWindows() 应返回 false")
	}
}

func TestIsRoot(t *testing.T) {
	// 仅验证函数可正常调用并返回布尔值
	result := IsRoot()
	t.Logf("IsRoot() = %v (UID=%d)", result, os.Getuid())
}

func TestPermissionError(t *testing.T) {
	isRoot := IsRoot()

	tests := []struct {
		os       string
		wantRoot string // root 用户时期望包含的关键词
		wantUser string // 非 root 用户时期望包含的关键词
	}{
		{"linux", "已是 root", "sudo"},
		{"darwin", "已是 root", "sudo"},
		{"windows", "管理员", "管理员"},
	}

	for _, tt := range tests {
		e := &PermissionError{Path: "/etc/hosts", OS: tt.os}
		msg := e.Error()
		if msg == "" {
			t.Errorf("OS=%s 时错误消息不应为空", tt.os)
		}

		var want string
		if isRoot {
			want = tt.wantRoot
		} else {
			want = tt.wantUser
		}
		if !contains(msg, want) {
			t.Errorf("OS=%s isRoot=%v: Error() = %q, 期望包含 %q", tt.os, isRoot, msg, want)
		}
	}
}

// contains 判断 s 中是否包含 substr
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		findSubstring(s, substr))
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
