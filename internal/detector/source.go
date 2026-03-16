package detector

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// GitHub520 数据源 URL
const (
	GitHub520URL = "https://raw.hellogithub.com/hosts"
)

// Source 定义数据源接口
type Source interface {
	// FetchIPs 获取域名-IP 映射列表
	FetchIPs(ctx context.Context) (map[string][]string, error)
	// Name 返回数据源名称
	Name() string
}

// GitHub520Source 从 GitHub520 项目获取候选 IP
type GitHub520Source struct {
	url    string
	client *http.Client
}

// NewGitHub520Source 创建 GitHub520 数据源
func NewGitHub520Source() *GitHub520Source {
	return &GitHub520Source{
		url: GitHub520URL,
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

func (s *GitHub520Source) Name() string {
	return "GitHub520"
}

// FetchIPs 从 GitHub520 获取 hosts 格式的域名-IP 映射
// 格式示例：
//
//	140.82.114.4 github.com
//	185.199.108.133 raw.githubusercontent.com
func (s *GitHub520Source) FetchIPs(ctx context.Context) (map[string][]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("获取 GitHub520 数据失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub520 返回错误状态码: %d", resp.StatusCode)
	}

	result := make(map[string][]string)
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// 解析 "IP 域名" 格式
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			ip := fields[0]
			domain := fields[1]
			// 仅保留我们关注的 GitHub 域名
			if isTargetDomain(domain) {
				result[domain] = appendUnique(result[domain], ip)
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取数据流失败: %w", err)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("未从 GitHub520 获取到任何有效 IP 映射")
	}

	return result, nil
}

// MultiSource 多数据源聚合，支持 failover
type MultiSource struct {
	sources []Source
}

// NewMultiSource 创建多数据源聚合器
func NewMultiSource(sources ...Source) *MultiSource {
	return &MultiSource{sources: sources}
}

// FetchIPs 依次尝试各数据源，第一个成功的返回
func (m *MultiSource) FetchIPs(ctx context.Context) (map[string][]string, error) {
	var lastErr error
	for _, src := range m.sources {
		result, err := src.FetchIPs(ctx)
		if err == nil && len(result) > 0 {
			return result, nil
		}
		lastErr = err
	}
	return nil, fmt.Errorf("所有数据源均不可用, 最后一个错误: %w", lastErr)
}

// isTargetDomain 判断是否为我们关注的 GitHub 域名
func isTargetDomain(domain string) bool {
	for _, d := range TargetDomains {
		if d == domain {
			return true
		}
	}
	return false
}

// appendUnique 向切片追加不重复的元素
func appendUnique(slice []string, item string) []string {
	for _, s := range slice {
		if s == item {
			return slice
		}
	}
	return append(slice, item)
}

// TargetDomains 维护的目标 GitHub 域名清单
var TargetDomains = []string{
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
}
