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
	"github.com/eyjian/github-buddy/internal/platform"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "初始化 github-buddy",
	Long:  "创建数据目录、生成默认配置、备份 hosts、执行首次 IP 检测并更新 hosts 文件。",
	RunE:  runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	fmt.Println("🚀 正在初始化 github-buddy ...")

	// 1. 检查是否已初始化
	if store.IsInitialized() {
		fmt.Println("⚠️  已初始化，可使用 'github-buddy update' 更新 IP")
		return nil
	}

	// 2. 创建目录结构
	fmt.Println("📁 创建数据目录 ...")
	if err := store.Init(); err != nil {
		return fmt.Errorf("创建数据目录失败: %w", err)
	}

	// 重新初始化日志（现在有了日志目录）
	logger.Init(store.LogsPath(), verbose)
	logger.Logger.Info().Msg("github-buddy 初始化开始")

	// 3. 生成默认配置
	fmt.Println("⚙️  生成默认配置 ...")
	cfg := config.DefaultConfig()
	if err := config.Save(store.ConfigPath(), cfg); err != nil {
		return fmt.Errorf("保存配置失败: %w", err)
	}

	// 4. 检查 hosts 文件权限
	if err := plat.CheckHostsWritable(); err != nil {
		return err
	}

	// 5. 备份当前 hosts
	fmt.Println("💾 备份当前 hosts 文件 ...")
	bakMgr := backup.NewManager(plat.HostPath, store.BackupsPath())
	if err := bakMgr.Backup(); err != nil {
		logger.Logger.Warn().Err(err).Msg("备份 hosts 文件失败")
		fmt.Printf("⚠️  备份失败: %v（继续执行）\n", err)
	}

	// 6. 执行首次 IP 检测
	fmt.Println("🔍 正在检测最优 GitHub IP（首次检测可能需要 10-30 秒）...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	det := detector.NewDetector(logger.Logger)
	result, err := det.DetectAll(ctx)
	if err != nil {
		return fmt.Errorf("IP 检测失败: %w", err)
	}

	// 7. 构建域名-IP 映射
	ipMap := buildIPMap(result)
	if len(ipMap) == 0 {
		return fmt.Errorf("未能检测到任何可用的 GitHub IP")
	}

	// 8. 更新 hosts 文件
	fmt.Println("📝 更新 hosts 文件 ...")
	if err := hosts.UpdateHostsFile(plat.HostPath, ipMap); err != nil {
		return fmt.Errorf("更新 hosts 文件失败: %w", err)
	}

	// 9. 保存缓存
	saveCache(store.CachePath(), result, ipMap)

	// 10. 刷新系统 DNS 缓存
	flushResult := platform.FlushDNSCache(context.Background(), plat.OS)
	if flushResult.Error != nil {
		logger.Logger.Warn().Err(flushResult.Error).Msg("自动刷新 DNS 缓存失败")
	}
	platform.PrintFlushResult(flushResult, plat.OS)

	// 11. 输出结果摘要
	fmt.Print("\n✅ 初始化完成！已更新以下域名映射：\n\n")
	printIPTable(ipMap, nil)

	logger.Logger.Info().Int("domains", len(ipMap)).Msg("github-buddy 初始化完成")
	fmt.Printf("\n💡 提示: 运行 'github-buddy status' 查看当前状态\n")
	return nil
}

// buildIPMap 从检测结果构建域名 -> 最优 IP 的映射
func buildIPMap(result *detector.DetectResult) map[string]string {
	ipMap := make(map[string]string)
	for domain, domainIPs := range result.Domains {
		if domainIPs.BestIP != nil {
			ipMap[domain] = domainIPs.BestIP.IP
		}
	}
	return ipMap
}

// saveCache 保存检测结果到缓存
func saveCache(cachePath string, result *detector.DetectResult, ipMap map[string]string) {
	cacheMgr := cache.NewManager(cachePath)
	c := &cache.IPCache{
		BestIPs:   ipMap,
		AllIPs:    make(map[string][]cache.CachedIP),
		UpdatedAt: time.Now(),
		ICMPUsed:  result.ICMPUsed,
	}

	for domain, domainIPs := range result.Domains {
		for _, ip := range domainIPs.IPs {
			c.AllIPs[domain] = append(c.AllIPs[domain], cache.CachedIP{
				IP:      ip.IP,
				Score:   ip.Score,
				Latency: ip.Latency,
				Port443: ip.Port443,
				Port22:  ip.Port22,
			})
		}
	}

	if err := cacheMgr.Save(c); err != nil {
		logger.Logger.Warn().Err(err).Msg("保存缓存失败")
	}
}

// printIPTable 输出域名-IP 对比表
func printIPTable(newIPs map[string]string, oldIPs map[string]string) {
	fmt.Printf("  %-40s %-18s", "域名", "IP")
	if oldIPs != nil {
		fmt.Printf(" %-18s", "旧 IP")
	}
	fmt.Println()
	fmt.Printf("  %-40s %-18s", "----", "--")
	if oldIPs != nil {
		fmt.Printf(" %-18s", "----")
	}
	fmt.Println()

	for _, domain := range sortedMapKeys(newIPs) {
		ip := newIPs[domain]
		fmt.Printf("  %-40s %-18s", domain, ip)
		if oldIPs != nil {
			oldIP, ok := oldIPs[domain]
			if ok && oldIP != ip {
				fmt.Printf(" %-18s", oldIP)
			} else if ok {
				fmt.Printf(" %-18s", "(无变化)")
			} else {
				fmt.Printf(" %-18s", "(新增)")
			}
		}
		fmt.Println()
	}
}

// sortedMapKeys 返回 map 的有序键列表
func sortedMapKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	for i := 0; i < len(keys)-1; i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	return keys
}
