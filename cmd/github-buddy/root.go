package main

import (
	"github.com/spf13/cobra"
)

var (
	verbose bool
)

// rootCmd 是 github-buddy 的根命令
var rootCmd = &cobra.Command{
	Use:     "github-buddy",
	Short:   "GitHub 网络访问优化工具",
	Long:    "GitHub-Buddy 通过智能维护 GitHub 域名-IP 映射并修改系统 hosts 文件，\n解决开发者执行 go mod tidy 和 git clone 时的网络访问问题。",
	Version: version,
}

func init() {
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "输出详细日志到控制台")
}

// Execute 执行根命令
func Execute() error {
	return rootCmd.Execute()
}
