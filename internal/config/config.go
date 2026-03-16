package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Config 应用配置
type Config struct {
	// 自动检查间隔（小时）
	UpdateIntervalHours int `json:"update_interval_hours"`
	// IP 数据源列表
	DataSources []DataSource `json:"data_sources"`
	// 需要维护的域名列表
	Domains []string `json:"domains"`
}

// DataSource IP 数据源配置
type DataSource struct {
	Name     string `json:"name"`
	URL      string `json:"url"`
	Priority int    `json:"priority"` // 优先级，数字越小优先级越高
	Enabled  bool   `json:"enabled"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		UpdateIntervalHours: 6,
		DataSources: []DataSource{
			{
				Name:     "GitHub520",
				URL:      "https://raw.hellogithub.com/hosts",
				Priority: 1,
				Enabled:  true,
			},
		},
		Domains: []string{
			"github.com",
			"ssh.github.com",
			"gist.github.com",
			"raw.githubusercontent.com",
			"api.github.com",
			"assets-cdn.github.com",
			"github.global.ssl.fastly.net",
			"collector.github.com",
			"avatars.githubusercontent.com",
			"codeload.github.com",
		},
	}
}

// UpdateInterval 返回更新间隔
func (c *Config) UpdateInterval() time.Duration {
	if c.UpdateIntervalHours <= 0 {
		c.UpdateIntervalHours = 6
	}
	return time.Duration(c.UpdateIntervalHours) * time.Hour
}

// Load 从文件加载配置
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	return &cfg, nil
}

// Save 将配置写入文件
func Save(path string, cfg *Config) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}

	return nil
}

// LoadOrDefault 加载配置，文件不存在时返回默认配置
func LoadOrDefault(path string) *Config {
	cfg, err := Load(path)
	if err != nil {
		return DefaultConfig()
	}
	return cfg
}
