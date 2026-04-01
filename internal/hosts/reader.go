package hosts

import (
	"bufio"
	"os"
	"strings"
)

// Entry 表示 hosts 文件中的一条记录
type Entry struct {
	IP      string // IP 地址
	Domain  string // 域名
	Comment string // 行尾注释
	Raw     string // 原始行内容
}

// HostsFile 表示解析后的 hosts 文件
type HostsFile struct {
	Entries  []Entry  // 所有条目
	Lines    []string // 所有原始行（保留格式）
	FilePath string   // 文件路径
}

// ReadHostsFile 读取并解析 hosts 文件
func ReadHostsFile(path string) (*HostsFile, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	hf := &HostsFile{
		FilePath: path,
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		hf.Lines = append(hf.Lines, line)

		entry := parseLine(line)
		if entry != nil {
			hf.Entries = append(hf.Entries, *entry)
		}
	}

	return hf, scanner.Err()
}

// GetEntriesByDomain 获取指定域名的所有 IP 映射
func (hf *HostsFile) GetEntriesByDomain(domain string) []Entry {
	var result []Entry
	for _, e := range hf.Entries {
		if e.Domain == domain {
			result = append(result, e)
		}
	}
	return result
}

// GetAllDomains 获取所有域名列表
func (hf *HostsFile) GetAllDomains() []string {
	seen := make(map[string]bool)
	var domains []string
	for _, e := range hf.Entries {
		if !seen[e.Domain] {
			seen[e.Domain] = true
			domains = append(domains, e.Domain)
		}
	}
	return domains
}

// parseLine 解析 hosts 文件中的一行
func parseLine(line string) *Entry {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "#") {
		return nil
	}

	// 处理行内注释
	comment := ""
	if idx := strings.Index(trimmed, "#"); idx > 0 {
		comment = strings.TrimSpace(trimmed[idx:])
		trimmed = strings.TrimSpace(trimmed[:idx])
	}

	fields := strings.Fields(trimmed)
	if len(fields) < 2 {
		return nil
	}

	return &Entry{
		IP:      fields[0],
		Domain:  fields[1],
		Comment: comment,
		Raw:     line,
	}
}
