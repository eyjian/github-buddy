package storage

import (
	"os"
	"path/filepath"

	"github.com/eyjian/github-buddy/internal/hosts"
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

// IsInitialized 检查是否已完成初始化
// 同时检查数据目录是否存在和 hosts 文件中是否包含 github-buddy 标记区块，
// 避免目录已创建但 hosts 未修改时误判为已初始化
func (m *Manager) IsInitialized() bool {
	// 检查数据目录是否存在
	if _, err := os.Stat(m.dataDir); err != nil {
		return false
	}
	// 检查 hosts 文件中是否存在 github-buddy 标记区块
	return hosts.HasBlock(m.plat.HostPath)
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
