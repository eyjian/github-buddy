package storage

import (
	"os"
	"path/filepath"

	"github.com/eyjian/github-buddy/internal/platform"
)

// 数据目录下的子目录名
const (
	ConfigDir  = "config"
	CacheDir   = "cache"
	LogsDir    = "logs"
	BackupsDir = "backups"
)

// Manager 管理用户数据目录
type Manager struct {
	plat    *platform.Info
	dataDir string
}

// NewManager 创建存储管理器
func NewManager(plat *platform.Info) *Manager {
	return &Manager{
		plat:    plat,
		dataDir: plat.DataDir,
	}
}

// Init 初始化数据目录结构：~/.github-buddy/{config,cache,logs,backups}
func (m *Manager) Init() error {
	dirs := []string{
		m.dataDir,
		filepath.Join(m.dataDir, ConfigDir),
		filepath.Join(m.dataDir, CacheDir),
		filepath.Join(m.dataDir, LogsDir),
		filepath.Join(m.dataDir, BackupsDir),
	}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}

// IsInitialized 检查数据目录是否已初始化
func (m *Manager) IsInitialized() bool {
	_, err := os.Stat(m.dataDir)
	return err == nil
}

// DataDir 返回数据目录路径
func (m *Manager) DataDir() string {
	return m.dataDir
}

// ConfigPath 返回配置文件路径
func (m *Manager) ConfigPath() string {
	return filepath.Join(m.dataDir, "config.json")
}

// CachePath 返回缓存文件路径
func (m *Manager) CachePath() string {
	return filepath.Join(m.dataDir, CacheDir, "ip_cache.json")
}

// LogsPath 返回日志目录路径
func (m *Manager) LogsPath() string {
	return filepath.Join(m.dataDir, LogsDir)
}

// BackupsPath 返回备份目录路径
func (m *Manager) BackupsPath() string {
	return filepath.Join(m.dataDir, BackupsDir)
}
