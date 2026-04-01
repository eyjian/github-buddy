package hosts

import (
	"fmt"
	"strings"
)

// ConflictEntry 表示一个域名冲突
type ConflictEntry struct {
	Domain    string // 冲突域名
	ExistIP   string // hosts 中已存在的 IP
	NewIP     string // 工具推荐的 IP
	LineNum   int    // 在 hosts 文件中的行号（1-based）
}

// DetectConflicts 检测标记区块外是否存在 GitHub 域名冲突
// targetDomains: 需要检查的域名列表
// ipMap: 工具推荐的域名-IP 映射
func DetectConflicts(lines []string, targetDomains []string, ipMap map[string]string) []ConflictEntry {
	block := FindBlock(lines)
	var conflicts []ConflictEntry

	targetSet := make(map[string]bool)
	for _, d := range targetDomains {
		targetSet[d] = true
	}

	for i, line := range lines {
		// 跳过标记区块内的行
		if block != nil && i >= block.StartLine && i <= block.EndLine {
			continue
		}

		entry := parseLine(line)
		if entry == nil {
			continue
		}

		// 检查是否为目标域名
		if targetSet[entry.Domain] {
			newIP, ok := ipMap[entry.Domain]
			if ok && newIP != entry.IP {
				conflicts = append(conflicts, ConflictEntry{
					Domain:  entry.Domain,
					ExistIP: entry.IP,
					NewIP:   newIP,
					LineNum: i + 1,
				})
			}
		}
	}

	return conflicts
}

// FormatConflicts 格式化冲突信息为用户可读字符串
func FormatConflicts(conflicts []ConflictEntry) string {
	if len(conflicts) == 0 {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("发现 %d 个域名冲突（hosts 文件中已存在以下 GitHub 域名配置）：\n\n", len(conflicts)))

	for _, c := range conflicts {
		sb.WriteString(fmt.Sprintf("  行 %d: %s -> %s（工具推荐: %s）\n", c.LineNum, c.Domain, c.ExistIP, c.NewIP))
	}

	return sb.String()
}

// RemoveConflictLines 从 hosts 行列表中移除冲突行
func RemoveConflictLines(lines []string, conflicts []ConflictEntry) []string {
	removeSet := make(map[int]bool)
	for _, c := range conflicts {
		removeSet[c.LineNum-1] = true // LineNum 是 1-based
	}

	var result []string
	for i, line := range lines {
		if !removeSet[i] {
			result = append(result, line)
		}
	}
	return result
}
