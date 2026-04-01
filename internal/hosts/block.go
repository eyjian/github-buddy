package hosts

import (
	"fmt"
	"os"
	"strings"
)

// 标记区块的起止标记
const (
	BlockStart = "# GitHub-Buddy Auto-Generated Start"
	BlockEnd   = "# GitHub-Buddy Auto-Generated End"
)

// Block 表示工具管理的标记区块
type Block struct {
	StartLine int    // 起始行号（0-based）
	EndLine   int    // 结束行号（0-based）
	Entries   []Entry // 区块内的条目
}

// HasBlock 检查指定 hosts 文件中是否存在 github-buddy 标记区块
// 用于判断 hosts 文件是否已被 github-buddy 修改过
func HasBlock(hostsPath string) bool {
	data, err := os.ReadFile(hostsPath)
	if err != nil {
		return false
	}
	content := string(data)
	return strings.Contains(content, BlockStart) && strings.Contains(content, BlockEnd)
}

// FindBlock 在 hosts 文件的行列表中查找标记区块
func FindBlock(lines []string) *Block {
	block := &Block{StartLine: -1, EndLine: -1}

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == BlockStart {
			block.StartLine = i
		} else if trimmed == BlockEnd {
			block.EndLine = i
			break
		}
	}

	if block.StartLine < 0 || block.EndLine < 0 || block.EndLine <= block.StartLine {
		return nil
	}

	// 解析区块内的条目
	for i := block.StartLine + 1; i < block.EndLine; i++ {
		entry := parseLine(lines[i])
		if entry != nil {
			block.Entries = append(block.Entries, *entry)
		}
	}

	return block
}

// BuildBlockContent 根据域名-IP 映射构建标记区块内容
func BuildBlockContent(ipMap map[string]string) []string {
	var lines []string
	lines = append(lines, BlockStart)

	// 按域名排序输出，确保输出稳定
	for _, domain := range sortedKeys(ipMap) {
		ip := ipMap[domain]
		lines = append(lines, fmt.Sprintf("%-20s %s", ip, domain))
	}

	lines = append(lines, BlockEnd)
	return lines
}

// UpdateLines 在 hosts 文件行列表中更新或追加标记区块
// 返回更新后的行列表
func UpdateLines(lines []string, ipMap map[string]string) []string {
	block := FindBlock(lines)
	newBlock := BuildBlockContent(ipMap)

	if block != nil {
		// 替换现有标记区块
		var result []string
		result = append(result, lines[:block.StartLine]...)
		result = append(result, newBlock...)
		result = append(result, lines[block.EndLine+1:]...)
		return result
	}

	// 标记区块不存在，追加到文件末尾
	result := make([]string, len(lines))
	copy(result, lines)

	// 确保前面有空行分隔
	if len(result) > 0 && strings.TrimSpace(result[len(result)-1]) != "" {
		result = append(result, "")
	}
	result = append(result, newBlock...)

	return result
}

// RemoveBlock 从 hosts 文件行列表中删除标记区块
func RemoveBlock(lines []string) []string {
	block := FindBlock(lines)
	if block == nil {
		return lines
	}

	var result []string
	result = append(result, lines[:block.StartLine]...)
	result = append(result, lines[block.EndLine+1:]...)
	return result
}

// sortedKeys 返回 map 的有序键列表
func sortedKeys(m map[string]string) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	// 简单冒泡排序，域名数量很少
	for i := 0; i < len(keys)-1; i++ {
		for j := i + 1; j < len(keys); j++ {
			if keys[i] > keys[j] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	return keys
}
