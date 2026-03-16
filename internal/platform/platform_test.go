package platform

import (
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

func TestPermissionError(t *testing.T) {
	tests := []struct {
		os   string
		want string
	}{
		{"linux", "sudo"},
		{"darwin", "sudo"},
		{"windows", "管理员"},
	}

	for _, tt := range tests {
		e := &PermissionError{Path: "/etc/hosts", OS: tt.os}
		msg := e.Error()
		if msg == "" {
			t.Errorf("OS=%s 时错误消息不应为空", tt.os)
		}
	}
}
