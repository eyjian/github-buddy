package main

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/eyjian/github-buddy/internal/backup"
	"github.com/eyjian/github-buddy/internal/cache"
	"github.com/eyjian/github-buddy/internal/config"
	"github.com/eyjian/github-buddy/internal/detector"
	"github.com/eyjian/github-buddy/internal/hosts"
	"github.com/eyjian/github-buddy/internal/logger"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "查看当前状态",
	Long:  "读取当前 hosts 中的 GitHub 域名映射，实时检测并以表格形式输出状态信息。",
	RunE:  runStatus,
}

func init() {
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	if !store.IsInitialized() {
		return fmt.Errorf("尚未初始化，请先执行 'github-buddy init'")
	}

	cfg := config.LoadOrDefault(store.ConfigPath())
	cacheMgr := cache.NewManager(store.CachePath())

	hf, err := hosts.ReadHostsFile(plat.HostPath)
	if err != nil {
		return fmt.Errorf("读取 hosts 文件失败: %w", err)
	}

	block := hosts.FindBlock(hf.Lines)
	if block == nil {
		fmt.Println("⚠️  hosts 文件中未找到 GitHub-Buddy 标记区块")
		fmt.Println("💡 提示: 运行 'github-buddy update' 写入 IP 映射")
		return nil
	}

	fmt.Print("📊 当前 GitHub 域名状态：\n\n")
	fmt.Printf("  %-34s %-17s %-8s %-6s %-6s %-6s %-6s %s\n",
		"域名", "IP", "延迟", "443", "22", "HTTPS", "浏览器", "状态")
	fmt.Printf("  %-34s %-17s %-8s %-6s %-6s %-6s %-6s %s\n",
		"----", "--", "----", "---", "--", "-----", "------", "----")

	tcpChecker := detector.NewTCPChecker(3 * time.Second)
	httpsChecker := detector.NewHTTPSChecker(5 * time.Second)
	pinger := detector.NewPinger(2, 3*time.Second)
	icmpOK := detector.IsPingAvailable()

	var githubWarnings []string

	for _, entry := range block.Entries {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)

		var wg sync.WaitGroup
		var port22 detector.TCPCheckResult
		var httpsRes detector.HTTPSCheckResult
		var pingRes detector.PingResult

		wg.Add(1)
		go func() {
			defer wg.Done()
			port22 = tcpChecker.CheckPort(ctx, entry.IP, 22)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			httpsRes = httpsChecker.Check(ctx, entry.IP, entry.Domain)
		}()

		if icmpOK {
			wg.Add(1)
			go func() {
				defer wg.Done()
				pingRes = pinger.Ping(ctx, entry.IP)
			}()
		}

		wg.Wait()
		cancel()

		latencyStr := "-"
		if pingRes.OK {
			latencyStr = fmt.Sprintf("%.0fms", pingRes.Latency)
		} else if httpsRes.BasicOK && httpsRes.Latency > 0 {
			latencyStr = fmt.Sprintf("%.0fms", httpsRes.Latency)
		} else if httpsRes.Port443OK && httpsRes.ConnectLatency > 0 {
			latencyStr = fmt.Sprintf("%.0fms", httpsRes.ConnectLatency)
		}

		fmt.Printf("  %-34s %-17s %-8s %-6s %-6s %-6s %-6s %s\n",
			entry.Domain,
			entry.IP,
			latencyStr,
			mark(httpsRes.Port443OK),
			mark(port22.OK),
			mark(httpsRes.BasicOK),
			browserMark(entry.Domain, httpsRes.BrowserOK),
			statusLabel(entry.Domain, httpsRes.Port443OK, port22.OK, httpsRes),
		)

		if entry.Domain == "github.com" && httpsRes.BasicOK && !httpsRes.BrowserOK {
			githubWarnings = append(githubWarnings, fmt.Sprintf("github.com 当前 IP %s 仅满足基础 HTTPS：首页校验未通过（%s）", entry.IP, browserReason(httpsRes.BrowserReason)))
		}
	}

	if len(githubWarnings) > 0 {
		fmt.Println()
		for _, warning := range githubWarnings {
			fmt.Printf("  ⚠ %s\n", warning)
		}
		fmt.Println("  💡 提示: 这种情况常见于‘hosts 已更新但浏览器仍打不开’，请优先重新执行 update，随后清理浏览器 DNS 缓存与 socket 连接池")
	}

	fmt.Println()
	if cachedData, err := cacheMgr.Load(); err == nil {
		age := time.Since(cachedData.UpdatedAt)
		expired := cacheMgr.IsExpired(cfg.UpdateInterval())
		expiredStr := "未过期"
		if expired {
			expiredStr = "已过期"
		}
		fmt.Printf("  📦 缓存状态: 上次更新 %s（%s前，%s）\n",
			cachedData.UpdatedAt.Format("2006-01-02 15:04:05"),
			formatDuration(age),
			expiredStr)
	} else {
		fmt.Println("  📦 缓存状态: 无缓存")
	}

	bakMgr := backup.NewManager(plat.HostPath, store.BackupsPath())
	if bakTime, err := bakMgr.LatestBackupTime(); err == nil {
		fmt.Printf("  💾 备份状态: 最近备份 %s\n", bakTime)
	} else {
		fmt.Println("  💾 备份状态: 无备份")
	}

	if cacheMgr.IsExpired(cfg.UpdateInterval()) {
		fmt.Println("\n💡 提示: 缓存已过期，建议运行 'github-buddy update' 更新 IP")
	}

	logger.Logger.Debug().Msg("status 命令执行完成")
	return nil
}

// formatDuration 格式化时间间隔为友好文本
func formatDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%d秒", int(d.Seconds()))
	}
	if d < time.Hour {
		return fmt.Sprintf("%d分钟", int(d.Minutes()))
	}
	if d < 24*time.Hour {
		return fmt.Sprintf("%.1f小时", d.Hours())
	}
	return fmt.Sprintf("%.1f天", d.Hours()/24)
}
