package backup

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const (
	maxBackups     = 10                     // 最多保留历史备份数
	bakFileName    = "hosts.bak"            // hosts 同目录备份文件名
	checksumFile   = "backup_checksums.json" // 校验和文件
)

// Manager 备份管理器
type Manager struct {
	hostsPath  string // hosts 文件路径
	hostsDir   string // hosts 文件所在目录
	backupsDir string // ~/.github-buddy/backups/ 目录
}

// ChecksumRecord 校验和记录
type ChecksumRecord struct {
	FilePath  string `json:"file_path"`
	SHA256    string `json:"sha256"`
	CreatedAt string `json:"created_at"`
}

// NewManager 创建备份管理器
func NewManager(hostsPath, backupsDir string) *Manager {
	return &Manager{
		hostsPath:  hostsPath,
		hostsDir:   filepath.Dir(hostsPath),
		backupsDir: backupsDir,
	}
}

// Backup 执行 hosts 文件备份
// 1. 备份到 hosts 同目录的 hosts.bak
// 2. 备份到 ~/.github-buddy/backups/ 带时间戳
// 3. 记录 SHA256 校验和
func (m *Manager) Backup() error {
	// 读取原始 hosts 文件
	data, err := os.ReadFile(m.hostsPath)
	if err != nil {
		return fmt.Errorf("读取 hosts 文件失败: %w", err)
	}

	// 计算 SHA256 校验和
	hash := sha256.Sum256(data)
	checksum := hex.EncodeToString(hash[:])
	timestamp := time.Now().Format("20060102-150405")

	// 1. 备份到 hosts 同目录
	bakPath := filepath.Join(m.hostsDir, bakFileName)
	if err := os.WriteFile(bakPath, data, 0644); err != nil {
		// hosts 同目录可能没有写权限，仅记录警告，继续
		fmt.Fprintf(os.Stderr, "警告: 无法备份到 %s: %v\n", bakPath, err)
	}

	// 2. 备份到 backups 目录（带时间戳）
	if err := os.MkdirAll(m.backupsDir, 0755); err != nil {
		return fmt.Errorf("创建备份目录失败: %w", err)
	}

	histBakName := fmt.Sprintf("hosts.bak.%s", timestamp)
	histBakPath := filepath.Join(m.backupsDir, histBakName)
	if err := os.WriteFile(histBakPath, data, 0644); err != nil {
		return fmt.Errorf("写入历史备份失败: %w", err)
	}

	// 3. 记录校验和
	record := ChecksumRecord{
		FilePath:  histBakPath,
		SHA256:    checksum,
		CreatedAt: timestamp,
	}
	if err := m.saveChecksum(record); err != nil {
		fmt.Fprintf(os.Stderr, "警告: 保存校验和失败: %v\n", err)
	}

	// 4. 清理过旧的备份
	m.pruneBackups()

	return nil
}

// Rollback 从最近的备份恢复 hosts 文件
func (m *Manager) Rollback() error {
	// 优先从 backups 目录恢复最新备份
	bakPath, err := m.latestBackup()
	if err != nil {
		return fmt.Errorf("查找备份失败: %w", err)
	}

	data, err := os.ReadFile(bakPath)
	if err != nil {
		return fmt.Errorf("读取备份文件失败: %w", err)
	}

	// 验证校验和
	if err := m.verifyChecksum(bakPath, data); err != nil {
		fmt.Fprintf(os.Stderr, "警告: %v\n", err)
	}

	// 恢复 hosts 文件
	fileInfo, _ := os.Stat(m.hostsPath)
	perm := os.FileMode(0644)
	if fileInfo != nil {
		perm = fileInfo.Mode()
	}

	if err := os.WriteFile(m.hostsPath, data, perm); err != nil {
		return fmt.Errorf("恢复 hosts 文件失败: %w", err)
	}

	return nil
}

// LatestBackupTime 返回最近备份的时间
func (m *Manager) LatestBackupTime() (string, error) {
	bakPath, err := m.latestBackup()
	if err != nil {
		return "", err
	}
	info, err := os.Stat(bakPath)
	if err != nil {
		return "", err
	}
	return info.ModTime().Format("2006-01-02 15:04:05"), nil
}

// HasBackup 检查是否有可用的备份
func (m *Manager) HasBackup() bool {
	_, err := m.latestBackup()
	return err == nil
}

// latestBackup 查找最新的备份文件路径
func (m *Manager) latestBackup() (string, error) {
	// 先查 backups 目录
	entries, err := os.ReadDir(m.backupsDir)
	if err == nil {
		var bakFiles []string
		for _, e := range entries {
			if strings.HasPrefix(e.Name(), "hosts.bak.") {
				bakFiles = append(bakFiles, filepath.Join(m.backupsDir, e.Name()))
			}
		}
		if len(bakFiles) > 0 {
			sort.Strings(bakFiles)
			return bakFiles[len(bakFiles)-1], nil // 最新的
		}
	}

	// 回退到 hosts 同目录的 hosts.bak
	bakPath := filepath.Join(m.hostsDir, bakFileName)
	if _, err := os.Stat(bakPath); err == nil {
		return bakPath, nil
	}

	return "", fmt.Errorf("无可用备份，无法执行回滚")
}

// pruneBackups 清理过旧的备份，保留最近 maxBackups 份
func (m *Manager) pruneBackups() {
	entries, err := os.ReadDir(m.backupsDir)
	if err != nil {
		return
	}

	var bakFiles []string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "hosts.bak.") {
			bakFiles = append(bakFiles, e.Name())
		}
	}

	if len(bakFiles) <= maxBackups {
		return
	}

	sort.Strings(bakFiles)
	// 删除最旧的
	for _, name := range bakFiles[:len(bakFiles)-maxBackups] {
		os.Remove(filepath.Join(m.backupsDir, name))
	}
}

// saveChecksum 保存备份文件的校验和
func (m *Manager) saveChecksum(record ChecksumRecord) error {
	checksumPath := filepath.Join(m.backupsDir, checksumFile)

	var records []ChecksumRecord
	data, err := os.ReadFile(checksumPath)
	if err == nil {
		json.Unmarshal(data, &records)
	}

	records = append(records, record)

	// 只保留最近 maxBackups 条记录
	if len(records) > maxBackups {
		records = records[len(records)-maxBackups:]
	}

	jsonData, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(checksumPath, jsonData, 0644)
}

// verifyChecksum 验证备份文件的校验和
func (m *Manager) verifyChecksum(bakPath string, data []byte) error {
	checksumPath := filepath.Join(m.backupsDir, checksumFile)

	var records []ChecksumRecord
	jsonData, err := os.ReadFile(checksumPath)
	if err != nil {
		return fmt.Errorf("校验和文件不存在，跳过验证")
	}

	if err := json.Unmarshal(jsonData, &records); err != nil {
		return fmt.Errorf("校验和文件格式错误")
	}

	// 查找匹配的记录
	hash := sha256.Sum256(data)
	actualChecksum := hex.EncodeToString(hash[:])

	for _, r := range records {
		if r.FilePath == bakPath {
			if r.SHA256 != actualChecksum {
				return fmt.Errorf("备份文件校验和不匹配（可能已损坏），期望: %s, 实际: %s", r.SHA256[:16], actualChecksum[:16])
			}
			return nil // 校验通过
		}
	}

	return nil // 没有记录，跳过验证
}

// FileChecksum 计算文件的 SHA256 校验和
func FileChecksum(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}
