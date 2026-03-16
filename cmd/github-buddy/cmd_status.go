package main

import (
	"context"
	"fmt"
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
	// 检查是否已初始化
	if !store.IsInitialized() {
		return fmt.Errorf("尚未初始化，请先执行 'github-buddy init'")
	}

	cfg := config.LoadOrDefault(store.ConfigPath())
	cacheMgr := cache.NewManager(store.CachePath())

	// 读取当前 hosts 中的 GitHub 域名映射
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

	// 表头
	fmt.Printf("  %-40s %-18s %-10s %-8s %-8s %s\n",
		"域名", "IP", "延迟", "443端口", "22端口", "状态")
	fmt.Printf("  %-40s %-18s %-10s %-8s %-8s %s\n",
		"----", "--", "----", "------", "-----", "----")

	// 对当前配置的 IP 执行实时检测
	tcpChecker := detector.NewTCPChecker(3 * time.Second)
	pinger := detector.NewPinger(2, 3*time.Second)
	icmpOK := detector.IsPingAvailable()

	for _, entry := range block.Entries {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)

		// TCP 端口检测
		port443 := tcpChecker.CheckPort(ctx, entry.IP, 443)
		port22 := tcpChecker.CheckPort(ctx, entry.IP, 22)

		latencyStr := "-"
		if icmpOK {
			pingResult := pinger.Ping(ctx, entry.IP)
			if pingResult.OK {
				latencyStr = fmt.Sprintf("%.0fms", pingResult.Latency)
			}
		} else if port443.OK {
			latencyStr = fmt.Sprintf("%.0fms", port443.Latency)
		}

		p443 := "✗"
		if port443.OK {
			p443 = "✓"
		}
		p22 := "✗"
		if port22.OK {
			p22 = "✓"
		}

		status := "✗ 不可用"
		if port443.OK && port22.OK {
			status = "✓ 正常"
		} else if port443.OK {
			status = "⚠ 部分"
		}

		fmt.Printf("  %-40s %-18s %-10s %-8s %-8s %s\n",
			entry.Domain, entry.IP, latencyStr, p443, p22, status)

		cancel()
	}

	// 缓存状态
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

	// 备份状态
	bakMgr := backup.NewManager(plat.HostPath, store.BackupsPath())
	if bakTime, err := bakMgr.LatestBackupTime(); err == nil {
		fmt.Printf("  💾 备份状态: 最近备份 %s\n", bakTime)
	} else {
		fmt.Println("  💾 备份状态: 无备份")
	}

	// 检查缓存过期时自动提示
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
