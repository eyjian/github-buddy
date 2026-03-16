package main

import (
	"github.com/spf13/cobra"

	"github.com/eyjian/github-buddy/internal/logger"
	"github.com/eyjian/github-buddy/internal/platform"
	"github.com/eyjian/github-buddy/internal/storage"
)

var (
	verbose  bool
	forceFlg bool
	plat     *platform.Info
	store    *storage.Manager
)

// rootCmd 是 github-buddy 的根命令
var rootCmd = &cobra.Command{
	Use:     "github-buddy",
	Short:   "GitHub 网络访问优化工具",
	Long:    "GitHub-Buddy 通过智能维护 GitHub 域名-IP 映射并修改系统 hosts 文件，\n解决开发者执行 go mod tidy 和 git clone 时的网络访问问题。",
	Version: version,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// 检测平台信息
		var err error
		plat, err = platform.Detect()
		if err != nil {
			logger.InitDefault(verbose)
			return
		}

		// 创建存储管理器
		store = storage.NewManager(plat)

		// 初始化日志系统
		if store.IsInitialized() {
			logger.Init(store.LogsPath(), verbose)
		} else {
			logger.InitDefault(verbose)
		}
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "输出详细日志到控制台")
	rootCmd.PersistentFlags().BoolVarP(&forceFlg, "force", "f", false, "跳过确认提示，强制执行")
}

// Execute 执行根命令
func Execute() error {
	return rootCmd.Execute()
}
