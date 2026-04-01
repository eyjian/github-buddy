package backup

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBackupAndRollback(t *testing.T) {
	tmpDir := t.TempDir()
	hostsPath := filepath.Join(tmpDir, "hosts")
	backupsDir := filepath.Join(tmpDir, "backups")

	// 创建测试 hosts 文件
	originalContent := "# original hosts\n127.0.0.1 localhost\n"
	os.WriteFile(hostsPath, []byte(originalContent), 0644)

	mgr := NewManager(hostsPath, backupsDir)

	// 测试备份
	err := mgr.Backup()
	if err != nil {
		t.Fatalf("备份失败: %v", err)
	}

	// 验证 backups 目录有备份文件
	entries, _ := os.ReadDir(backupsDir)
	bakCount := 0
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "hosts.bak.") {
			bakCount++
		}
	}
	if bakCount == 0 {
		t.Error("backups 目录应有备份文件")
	}

	// 修改 hosts 文件
	modifiedContent := "# modified hosts\n192.168.1.1 myhost\n"
	os.WriteFile(hostsPath, []byte(modifiedContent), 0644)

	// 验证 hosts 已被修改
	data, _ := os.ReadFile(hostsPath)
	if !strings.Contains(string(data), "modified") {
		t.Error("hosts 应已被修改")
	}

	// 测试回滚
	err = mgr.Rollback()
	if err != nil {
		t.Fatalf("回滚失败: %v", err)
	}

	// 验证回滚结果
	data, _ = os.ReadFile(hostsPath)
	if !strings.Contains(string(data), "original") {
		t.Error("回滚后应恢复原始内容")
	}
}

func TestHasBackup_NoBackup(t *testing.T) {
	tmpDir := t.TempDir()
	mgr := NewManager(filepath.Join(tmpDir, "hosts"), filepath.Join(tmpDir, "backups"))
	if mgr.HasBackup() {
		t.Error("不应有备份")
	}
}

func TestPruneBackups(t *testing.T) {
	tmpDir := t.TempDir()
	backupsDir := filepath.Join(tmpDir, "backups")
	os.MkdirAll(backupsDir, 0755)

	// 创建 12 个备份文件
	for i := 0; i < 12; i++ {
		name := filepath.Join(backupsDir, "hosts.bak.20260316-10000"+string(rune('0'+i)))
		os.WriteFile(name, []byte("test"), 0644)
	}

	mgr := NewManager(filepath.Join(tmpDir, "hosts"), backupsDir)
	mgr.pruneBackups()

	entries, _ := os.ReadDir(backupsDir)
	bakCount := 0
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "hosts.bak.") {
			bakCount++
		}
	}
	if bakCount > maxBackups {
		t.Errorf("备份数量应≤%d, 实际: %d", maxBackups, bakCount)
	}
}

func TestFileChecksum(t *testing.T) {
	tmpDir := t.TempDir()
	file := filepath.Join(tmpDir, "test.txt")
	os.WriteFile(file, []byte("hello world"), 0644)

	checksum1, err := FileChecksum(file)
	if err != nil {
		t.Fatalf("计算校验和失败: %v", err)
	}
	if checksum1 == "" {
		t.Error("校验和不应为空")
	}

	// 相同内容应有相同校验和
	checksum2, _ := FileChecksum(file)
	if checksum1 != checksum2 {
		t.Error("相同内容校验和应一致")
	}

	// 修改内容后校验和应不同
	os.WriteFile(file, []byte("hello world 2"), 0644)
	checksum3, _ := FileChecksum(file)
	if checksum1 == checksum3 {
		t.Error("不同内容校验和应不同")
	}
}
