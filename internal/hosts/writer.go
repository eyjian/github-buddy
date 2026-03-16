package hosts

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// WriteHostsFile 原子写入 hosts 文件
// 先写入临时文件，验证内容后再替换原文件
func WriteHostsFile(path string, lines []string) error {
	content := strings.Join(lines, "\n")
	// 确保文件以换行符结尾
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}

	// 写入临时文件
	dir := filepath.Dir(path)
	tmpFile, err := os.CreateTemp(dir, "github-buddy-hosts-*.tmp")
	if err != nil {
		// 如果临时文件创建失败（可能权限不够），尝试在系统临时目录创建
		tmpFile, err = os.CreateTemp("", "github-buddy-hosts-*.tmp")
		if err != nil {
			return fmt.Errorf("创建临时文件失败: %w", err)
		}
	}
	tmpPath := tmpFile.Name()

	// 写入内容
	if _, err := tmpFile.WriteString(content); err != nil {
		tmpFile.Close()
		os.Remove(tmpPath)
		return fmt.Errorf("写入临时文件失败: %w", err)
	}

	if err := tmpFile.Close(); err != nil {
		os.Remove(tmpPath)
		return fmt.Errorf("关闭临时文件失败: %w", err)
	}

	// 获取原文件的权限信息
	fileInfo, err := os.Stat(path)
	if err == nil {
		os.Chmod(tmpPath, fileInfo.Mode())
	}

	// 尝试原子替换（同一文件系统下 Rename 是原子操作）
	if err := os.Rename(tmpPath, path); err != nil {
		// Rename 失败（跨文件系统），回退到复制方式
		data, readErr := os.ReadFile(tmpPath)
		os.Remove(tmpPath)
		if readErr != nil {
			return fmt.Errorf("读取临时文件失败: %w", readErr)
		}
		perm := os.FileMode(0644)
		if fileInfo != nil {
			perm = fileInfo.Mode()
		}
		if err := os.WriteFile(path, data, perm); err != nil {
			return fmt.Errorf("写入 hosts 文件失败: %w", err)
		}
	}

	return nil
}

// UpdateHostsFile 更新 hosts 文件中的 GitHub 域名映射
// ipMap: 域名 -> 最优 IP 的映射
func UpdateHostsFile(path string, ipMap map[string]string) error {
	// 读取现有文件
	hf, err := ReadHostsFile(path)
	if err != nil {
		return fmt.Errorf("读取 hosts 文件失败: %w", err)
	}

	// 更新标记区块
	newLines := UpdateLines(hf.Lines, ipMap)

	// 写入文件
	return WriteHostsFile(path, newLines)
}
