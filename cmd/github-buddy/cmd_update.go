package main

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"

	"github.com/eyjian/github-buddy/internal/backup"
	"github.com/eyjian/github-buddy/internal/cache"
	"github.com/eyjian/github-buddy/internal/config"
	"github.com/eyjian/github-buddy/internal/detector"
	"github.com/eyjian/github-buddy/internal/hosts"
	"github.com/eyjian/github-buddy/internal/logger"
	"github.com/eyjian/github-buddy/internal/platform"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "强制更新 GitHub IP",
	Long:  "强制执行完整 IP 检测（忽略缓存），备份 hosts 并更新为最优 IP。",
	RunE:  runUpdate,
}

func init() {
	rootCmd.AddCommand(updateCmd)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	if !store.IsInitialized() {
		return fmt.Errorf("尚未初始化，请先执行 'github-buddy init'")
	}

	fmt.Println("🔄 正在更新 GitHub IP ...")

	if err := plat.CheckHostsWritable(); err != nil {
		return err
	}

	cfg := config.LoadOrDefault(store.ConfigPath())
	cacheMgr := cache.NewManager(store.CachePath())
	var oldIPs map[string]string
	if oldCache, err := cacheMgr.Load(); err == nil {
		oldIPs = oldCache.BestIPs
	}

	fmt.Println("💾 备份当前 hosts 文件 ...")
	bakMgr := backup.NewManager(plat.HostPath, store.BackupsPath())
	if err := bakMgr.Backup(); err != nil {
		logger.Logger.Warn().Err(err).Msg("备份 hosts 文件失败")
	}

	fmt.Println("🔍 正在检测最优 GitHub IP ...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	det := detector.NewDetector(logger.Logger)
	result, err := det.DetectAll(ctx)
	if err != nil {
		return fmt.Errorf("IP 检测失败: %w", err)
	}

	ipMap := buildIPMap(result)
	if len(ipMap) == 0 {
		return fmt.Errorf("未能检测到任何可用的 GitHub IP")
	}

	// github.com 保护策略：如果本次检测未找到合格的 github.com IP，保留旧 IP 不覆盖
	if _, hasGitHub := ipMap["github.com"]; !hasGitHub {
		if oldIPs != nil {
			if oldGitHubIP, ok := oldIPs["github.com"]; ok && oldGitHubIP != "" {
				fmt.Printf("⚠️  本次未找到 github.com 的合格 IP（至少需要 443 端口可达），保留旧 IP: %s\n", oldGitHubIP)
				ipMap["github.com"] = oldGitHubIP
			} else {
				fmt.Println("⚠️  本次未找到 github.com 的合格 IP（至少需要 443 端口可达），且无旧 IP 可保留")
			}
		} else {
			fmt.Println("⚠️  本次未找到 github.com 的合格 IP（至少需要 443 端口可达）")
		}
	}

	hf, err := hosts.ReadHostsFile(plat.HostPath)
	if err == nil {
		conflicts := hosts.DetectConflicts(hf.Lines, cfg.Domains, ipMap)
		if len(conflicts) > 0 {
			fmt.Println("\n" + hosts.FormatConflicts(conflicts))
			if !forceFlg {
				fmt.Println("  使用 --force 参数自动采用工具推荐的最优 IP")
			}
			if forceFlg {
				newLines := hosts.RemoveConflictLines(hf.Lines, conflicts)
				hosts.WriteHostsFile(plat.HostPath, newLines)
			}
		}
	}

	fmt.Println("📝 更新 hosts 文件 ...")
	if err := hosts.UpdateHostsFile(plat.HostPath, ipMap); err != nil {
		return fmt.Errorf("更新 hosts 文件失败: %w", err)
	}

	saveCache(store.CachePath(), result, ipMap)

	flushResult := platform.FlushDNSCache(context.Background(), plat.OS)
	if flushResult.Error != nil {
		logger.Logger.Warn().Err(flushResult.Error).Msg("自动刷新 DNS 缓存失败")
	}
	platform.PrintFlushResult(flushResult, plat.OS)

	fmt.Print("\n✅ 更新完成！域名映射变化：\n\n")
	printIPTable(ipMap, oldIPs)
	printGitHubBrowserSummary(result)

	logger.Logger.Info().Int("domains", len(ipMap)).Msg("IP 更新完成")
	return nil
}
