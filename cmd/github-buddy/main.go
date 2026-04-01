package main

import (
	"fmt"
	"os"
)

// 版本信息，通过编译时 -ldflags 注入
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if err := Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "错误: %v\n", err)
		os.Exit(1)
	}
}
