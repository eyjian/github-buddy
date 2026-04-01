package detector

import (
	"bufio"
	"context"
	"fmt"
	"net/http"
	"net/netip"
	"strings"
	"time"
)

// 数据源 URL
const (
	GitHub520URL = "https://raw.hellogithub.com/hosts"
	Ineo6URL     = "https://gitlab.com/ineo6/hosts/-/raw/master/hosts"
)

// Source 定义数据源接口
type Source interface {
	// FetchIPs 获取域名-IP 映射列表
	FetchIPs(ctx context.Context) (map[string][]string, error)
	// Name 返回数据源名称
	Name() string
}

// HTTPHostsSource 从 hosts 文本源获取候选 IP
type HTTPHostsSource struct {
	name   string
	url    string
	client *http.Client
}

// NewHTTPHostsSource 创建通用 hosts 数据源
func NewHTTPHostsSource(name, url string) *HTTPHostsSource {
	return &HTTPHostsSource{
		name: name,
		url:  url,
		client: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// NewGitHub520Source 创建 GitHub520 数据源
func NewGitHub520Source() *HTTPHostsSource {
	return NewHTTPHostsSource("GitHub520", GitHub520URL)
}

// NewIneo6Source 创建 ineo6/hosts 数据源
func NewIneo6Source() *HTTPHostsSource {
	return NewHTTPHostsSource("Ineo6", Ineo6URL)
}

func (s *HTTPHostsSource) Name() string {
	return s.name
}

// FetchIPs 获取 hosts 格式的域名-IP 映射
func (s *HTTPHostsSource) FetchIPs(ctx context.Context) (map[string][]string, error) {
	return fetchHostsFromURL(ctx, s.client, s.url, s.name)
}

// fetchHostsFromURL 通用的 hosts 格式数据获取与解析函数
func fetchHostsFromURL(ctx context.Context, client *http.Client, url, sourceName string) (map[string][]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("获取 %s 数据失败: %w", sourceName, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%s 返回错误状态码: %d", sourceName, resp.StatusCode)
	}

	result := make(map[string][]string)
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		ip := fields[0]
		if !isUsableCandidateIP(ip) {
			continue
		}

		for _, domain := range fields[1:] {
			if !isTargetDomain(domain) {
				continue
			}
			result[domain] = appendUnique(result[domain], ip)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取数据流失败: %w", err)
	}

	if len(result) == 0 {
		return nil, fmt.Errorf("未从 %s 获取到任何有效 IP 映射", sourceName)
	}

	return result, nil
}

// MultiSource 多数据源聚合，支持汇总多个成功结果
type MultiSource struct {
	sources []Source
}

// NewMultiSource 创建多数据源聚合器
func NewMultiSource(sources ...Source) *MultiSource {
	return &MultiSource{sources: sources}
}

func (m *MultiSource) Name() string {
	return "multi-source"
}

// FetchIPs 聚合所有成功数据源的结果，全部失败时返回错误
func (m *MultiSource) FetchIPs(ctx context.Context) (map[string][]string, error) {
	merged := make(map[string][]string)
	var lastErr error
	var successCount int

	for _, src := range m.sources {
		result, err := src.FetchIPs(ctx)
		if err != nil {
			lastErr = err
			continue
		}

		successCount++
		for domain, ips := range result {
			for _, ip := range ips {
				if isUsableCandidateIP(ip) {
					merged[domain] = appendUnique(merged[domain], ip)
				}
			}
		}
	}

	if successCount == 0 || len(merged) == 0 {
		if lastErr == nil {
			lastErr = fmt.Errorf("所有数据源均未返回有效候选 IP")
		}
		return nil, fmt.Errorf("所有数据源均不可用, 最后一个错误: %w", lastErr)
	}

	return merged, nil
}

// filterCandidateMap 过滤候选 IP 映射中的无效地址
func filterCandidateMap(candidates map[string][]string) map[string][]string {
	filtered := make(map[string][]string, len(candidates))
	for domain, ips := range candidates {
		if !isTargetDomain(domain) {
			continue
		}
		for _, ip := range ips {
			if isUsableCandidateIP(ip) {
				filtered[domain] = appendUnique(filtered[domain], ip)
			}
		}
	}
	return filtered
}

// isTargetDomain 判断是否为我们关注的 GitHub 域名
func isTargetDomain(domain string) bool {
	_, ok := targetDomainSet[domain]
	return ok
}

var targetDomainSet = func() map[string]struct{} {
	set := make(map[string]struct{}, len(TargetDomains))
	for _, domain := range TargetDomains {
		set[domain] = struct{}{}
	}
	return set
}()

// RFC 6890 等定义的非公网与保留地址通过显式网段判断过滤，
// 避免依赖运行环境对 IsPrivate/特殊地址判断的差异。
func isUsableCandidateIP(raw string) bool {
	addr, err := netip.ParseAddr(strings.TrimSpace(raw))
	if err != nil {
		return false
	}

	addr = addr.Unmap()
	if !addr.IsValid() || addr.Zone() != "" || !addr.IsGlobalUnicast() {
		return false
	}

	if addr.Is4() {
		return !isSpecialUseIPv4(addr.As4())
	}
	if addr.Is6() {
		return !isSpecialUseIPv6(addr.As16())
	}
	return false
}

func isSpecialUseIPv4(ip [4]byte) bool {
	switch {
	case ip[0] == 0:
		return true
	case ip[0] == 10:
		return true
	case ip[0] == 100 && ip[1] >= 64 && ip[1] <= 127:
		return true
	case ip[0] == 127:
		return true
	case ip[0] == 169 && ip[1] == 254:
		return true
	case ip[0] == 172 && ip[1] >= 16 && ip[1] <= 31:
		return true
	case ip[0] == 192 && ip[1] == 0 && ip[2] == 0:
		return true
	case ip[0] == 192 && ip[1] == 0 && ip[2] == 2:
		return true
	case ip[0] == 192 && ip[1] == 168:
		return true
	case ip[0] == 198 && (ip[1] == 18 || ip[1] == 19):
		return true
	case ip[0] == 198 && ip[1] == 51 && ip[2] == 100:
		return true
	case ip[0] == 203 && ip[1] == 0 && ip[2] == 113:
		return true
	case ip[0] >= 224:
		return true
	default:
		return false
	}
}

func isSpecialUseIPv6(ip [16]byte) bool {
	switch {
	case ip == [16]byte{}:
		return true
	case ip == [16]byte{15: 1}:
		return true
	case ip[0] == 0xfc || ip[0] == 0xfd:
		return true
	case ip[0] == 0xfe && ip[1]&0xc0 == 0x80:
		return true
	case ip[0] == 0xff:
		return true
	case ip[0] == 0x20 && ip[1] == 0x01 && ip[2] == 0x0d && ip[3] == 0xb8:
		return true
	case ip[0] == 0x20 && ip[1] == 0x01 && ip[2] == 0x00 && ip[3] == 0x02 && ip[4] == 0x00 && ip[5] == 0x00:
		return true
	case ip[0] == 0x20 && ip[1] == 0x01 && ip[2] == 0x00 && ip[3] == 0x10:
		return true
	case ip[0] == 0x20 && ip[1] == 0x01 && ip[2] == 0x00 && ip[3] == 0x20:
		return true
	default:
		return false
	}
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
	"alive.github.com",
	"live.github.com",
	"ssh.github.com",
	"gist.github.com",
	"raw.githubusercontent.com",
	"api.github.com",
	"codeload.github.com",
	"avatars.githubusercontent.com",
	"github.githubassets.com",
	"objects.githubusercontent.com",
	"media.githubusercontent.com",
	"user-images.githubusercontent.com",
	"github-cloud.s3.amazonaws.com",
	"collector.github.com",
}
