package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/eyjian/github-buddy/internal/backup"
	"github.com/eyjian/github-buddy/internal/logger"
	"github.com/spf13/cobra"
)

var rollbackCmd = &cobra.Command{
	Use:   "rollback",
	Short: "回滚 hosts 文件",
	Long:  "将 hosts 文件恢复为最近一次备份的版本。",
	RunE:  runRollback,
}

func init() {
	rootCmd.AddCommand(rollbackCmd)
}

func runRollback(cmd *cobra.Command, args []string) error {
	// 检查是否已初始化
	if !store.IsInitialized() {
		return fmt.Errorf("尚未初始化，请先执行 'github-buddy init'")
	}

	// 检查 hosts 文件权限
	if err := plat.CheckHostsWritable(); err != nil {
		return err
	}

	bakMgr := backup.NewManager(plat.HostPath, store.BackupsPath())

	// 检查备份是否存在
	if !bakMgr.HasBackup() {
		return fmt.Errorf("无可用备份，无法执行回滚")
	}

	// 获取备份时间
	bakTime, _ := bakMgr.LatestBackupTime()

	// 确认提示
	if !forceFlg {
		fmt.Printf("⚠️  即将将 hosts 恢复为 %s 的版本，是否继续？[Y/n] ", bakTime)
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))
		if input != "" && input != "y" && input != "yes" {
			fmt.Println("已取消回滚操作")
			return nil
		}
	}

	// 执行回滚
	fmt.Println("🔄 正在恢复 hosts 文件 ...")
	if err := bakMgr.Rollback(); err != nil {
		return fmt.Errorf("回滚失败: %w", err)
	}

	fmt.Println("✅ hosts 文件已恢复为备份版本")
	logger.Logger.Info().Str("backup_time", bakTime).Msg("hosts 文件回滚完成")
	return nil
}
