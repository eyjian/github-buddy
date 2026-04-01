package logger

import (
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger 全局日志实例
var Logger zerolog.Logger

// Init 初始化日志系统
// logsDir: 日志文件目录
// verbose: 是否输出 DEBUG 级别到控制台
func Init(logsDir string, verbose bool) {
	// 控制台输出（彩色美化）
	consoleWriter := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.DateTime,
	}

	// 日志级别
	if verbose {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
	}

	var writers []io.Writer
	writers = append(writers, consoleWriter)

	// 文件输出（带轮转）
	if logsDir != "" {
		os.MkdirAll(logsDir, 0755)
		logFile := &lumberjack.Logger{
			Filename:   filepath.Join(logsDir, "github-buddy.log"),
			MaxSize:    10, // 单文件最大 10MB
			MaxBackups: 5,
			MaxAge:     30, // 保留 30 天
			Compress:   false,
		}
		writers = append(writers, logFile)
	}

	multi := zerolog.MultiLevelWriter(writers...)
	Logger = zerolog.New(multi).With().Timestamp().Logger()
}

// InitDefault 使用默认配置初始化日志（仅控制台输出）
func InitDefault(verbose bool) {
	Init("", verbose)
}
