package cache

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// IPCache IP 检测结果缓存
type IPCache struct {
	// 域名 -> 最优 IP
	BestIPs map[string]string `json:"best_ips"`
	// 域名 -> IP 列表及评分
	AllIPs map[string][]CachedIP `json:"all_ips"`
	// 上次更新时间
	UpdatedAt time.Time `json:"updated_at"`
	// 是否使用了 ICMP 检测
	ICMPUsed bool `json:"icmp_used"`
}

// CachedIP 缓存中的 IP 信息
type CachedIP struct {
	IP      string  `json:"ip"`
	Score   float64 `json:"score"`
	Latency float64 `json:"latency_ms"`
	Port443 bool    `json:"port_443"`
	Port22  bool    `json:"port_22"`
}

// Manager 缓存管理器
type Manager struct {
	cachePath string
}

// NewManager 创建缓存管理器
func NewManager(cachePath string) *Manager {
	return &Manager{cachePath: cachePath}
}

// Load 从文件加载缓存
func (m *Manager) Load() (*IPCache, error) {
	data, err := os.ReadFile(m.cachePath)
	if err != nil {
		return nil, fmt.Errorf("读取缓存文件失败: %w", err)
	}

	var cache IPCache
	if err := json.Unmarshal(data, &cache); err != nil {
		return nil, fmt.Errorf("解析缓存文件失败: %w", err)
	}

	return &cache, nil
}

// Save 将缓存写入文件
func (m *Manager) Save(cache *IPCache) error {
	data, err := json.MarshalIndent(cache, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化缓存失败: %w", err)
	}

	if err := os.WriteFile(m.cachePath, data, 0644); err != nil {
		return fmt.Errorf("写入缓存文件失败: %w", err)
	}

	return nil
}

// IsExpired 检查缓存是否过期
func (m *Manager) IsExpired(interval time.Duration) bool {
	cache, err := m.Load()
	if err != nil {
		return true // 缓存不存在或损坏，视为过期
	}

	return time.Since(cache.UpdatedAt) > interval
}

// Exists 检查缓存文件是否存在
func (m *Manager) Exists() bool {
	_, err := os.Stat(m.cachePath)
	return err == nil
}
