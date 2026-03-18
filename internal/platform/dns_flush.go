package platform

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"time"
)

// FlushResult DNS 缓存刷新结果
type FlushResult struct {
	Success bool   // 是否刷新成功
	Message string // 结果描述信息
	Error   error  // 刷新失败时的错误（可为 nil）
}

// FlushDNSCache 根据当前操作系统类型自动刷新系统 DNS 缓存。
// osType 参数接受 "linux"、"darwin"、"windows"，传空字符串时自动检测。
// 所有外部命令执行均设置 5 秒超时。
func FlushDNSCache(ctx context.Context, osType string) *FlushResult {
	if osType == "" {
		osType = runtime.GOOS
	}

	// 为命令执行设置 5 秒超时
	cmdCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	switch osType {
	case "linux":
		return flushLinux(cmdCtx)
	case "darwin":
		return flushDarwin(cmdCtx)
	case "windows":
		return flushWindows(cmdCtx)
	default:
		return &FlushResult{
			Success: false,
			Message: fmt.Sprintf("不支持的操作系统: %s", osType),
			Error:   fmt.Errorf("unsupported OS: %s", osType),
		}
	}
}

// flushLinux 刷新 Linux 系统 DNS 缓存。
// 依次尝试: systemd-resolve --flush-caches → resolvectl flush-caches → systemctl restart systemd-resolved。
func flushLinux(ctx context.Context) *FlushResult {
	// 先尝试 systemd-resolve
	if err := exec.CommandContext(ctx, "systemd-resolve", "--flush-caches").Run(); err == nil {
		return &FlushResult{
			Success: true,
			Message: "已通过 systemd-resolve 刷新 DNS 缓存",
		}
	}

	// fallback 到 resolvectl
	if err := exec.CommandContext(ctx, "resolvectl", "flush-caches").Run(); err == nil {
		return &FlushResult{
			Success: true,
			Message: "已通过 resolvectl 刷新 DNS 缓存",
		}
	}

	// 最后 fallback 到 systemctl restart systemd-resolved
	if err := exec.CommandContext(ctx, "systemctl", "restart", "systemd-resolved").Run(); err == nil {
		return &FlushResult{
			Success: true,
			Message: "已通过 systemctl restart systemd-resolved 刷新 DNS 缓存",
		}
	}

	return &FlushResult{
		Success: false,
		Message: "systemd-resolve、resolvectl 和 systemctl 均不可用",
		Error:   fmt.Errorf("linux: systemd-resolve、resolvectl 和 systemctl restart systemd-resolved 均执行失败"),
	}
}

// flushDarwin 刷新 macOS 系统 DNS 缓存。
// 同时执行 dscacheutil -flushcache 和 killall -HUP mDNSResponder。
func flushDarwin(ctx context.Context) *FlushResult {
	// 执行 dscacheutil -flushcache
	err1 := exec.CommandContext(ctx, "dscacheutil", "-flushcache").Run()

	// 执行 killall -HUP mDNSResponder
	err2 := exec.CommandContext(ctx, "killall", "-HUP", "mDNSResponder").Run()

	if err1 != nil && err2 != nil {
		return &FlushResult{
			Success: false,
			Message: "dscacheutil 和 killall mDNSResponder 均执行失败",
			Error:   fmt.Errorf("darwin: dscacheutil 错误: %v, killall 错误: %v", err1, err2),
		}
	}

	return &FlushResult{
		Success: true,
		Message: "已刷新 macOS DNS 缓存",
	}
}

// flushWindows 刷新 Windows 系统 DNS 缓存。
// 执行 ipconfig /flushdns。
func flushWindows(ctx context.Context) *FlushResult {
	if err := exec.CommandContext(ctx, "ipconfig", "/flushdns").Run(); err != nil {
		return &FlushResult{
			Success: false,
			Message: "ipconfig /flushdns 执行失败",
			Error:   fmt.Errorf("windows: ipconfig /flushdns 错误: %v", err),
		}
	}

	return &FlushResult{
		Success: true,
		Message: "已刷新 Windows DNS 缓存",
	}
}

// PrintFlushResult 根据 DNS 缓存刷新结果输出友好的用户提示。
// 刷新成功输出 ✅ 提示，刷新失败输出 ⚠️ 警告及手动刷新指引。
// 无论成功或失败，始终输出浏览器缓存清除提示。
func PrintFlushResult(result *FlushResult, osType string) {
	if osType == "" {
		osType = runtime.GOOS
	}

	fmt.Println()
	if result.Success {
		fmt.Println("✅ 已自动刷新系统 DNS 缓存")
	} else {
		fmt.Println("⚠️  自动刷新系统 DNS 缓存失败，请手动执行:")
		// 如果已是 root 用户，则无需在提示中加 sudo 前缀
		prefix := "sudo "
		if IsRoot() {
			prefix = ""
		}
		switch osType {
		case "linux":
			fmt.Printf("   %ssystemd-resolve --flush-caches\n", prefix)
			fmt.Printf("   或: %sresolvectl flush-caches\n", prefix)
			fmt.Printf("   或: %ssystemctl restart systemd-resolved\n", prefix)
		case "darwin":
			fmt.Printf("   %sdscacheutil -flushcache && %skillall -HUP mDNSResponder\n", prefix, prefix)
		case "windows":
			fmt.Println("   ipconfig /flushdns")
		default:
			fmt.Println("   请根据您的操作系统查找对应的 DNS 缓存刷新命令")
		}
	}

	// 始终输出浏览器缓存清除提示
	fmt.Println("💡 提示: 如仍有问题，请手动清除浏览器 DNS 缓存")
	fmt.Println("   Chrome:  访问 chrome://net-internals/#dns 点击 \"Clear host cache\"")
	fmt.Println("   Edge:    访问 edge://net-internals/#dns 点击 \"Clear host cache\"")
	fmt.Println("   Firefox: 访问 about:networking#dns 点击 \"清除 DNS 缓存\"")
	if osType == "darwin" {
		fmt.Println("   Safari:  前往菜单 \"开发\" > \"清空缓存\"（需先在偏好设置中启用开发菜单）")
	}
}
