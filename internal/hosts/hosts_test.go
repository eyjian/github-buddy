package hosts

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParseLine(t *testing.T) {
	tests := []struct {
		name   string
		line   string
		wantIP string
		wantDm string
		isNil  bool
	}{
		{"正常行", "1.2.3.4 example.com", "1.2.3.4", "example.com", false},
		{"带注释", "1.2.3.4 example.com # comment", "1.2.3.4", "example.com", false},
		{"注释行", "# this is a comment", "", "", true},
		{"空行", "", "", "", true},
		{"只有 IP", "1.2.3.4", "", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entry := parseLine(tt.line)
			if tt.isNil {
				if entry != nil {
					t.Error("期望返回 nil")
				}
				return
			}
			if entry == nil {
				t.Fatal("不期望返回 nil")
			}
			if entry.IP != tt.wantIP {
				t.Errorf("IP = %s, want %s", entry.IP, tt.wantIP)
			}
			if entry.Domain != tt.wantDm {
				t.Errorf("Domain = %s, want %s", entry.Domain, tt.wantDm)
			}
		})
	}
}

func TestFindBlock(t *testing.T) {
	lines := []string{
		"# some comment",
		"1.2.3.4 existing.com",
		"# GitHub-Buddy Auto-Generated Start",
		"5.6.7.8 github.com",
		"9.10.11.12 api.github.com",
		"# GitHub-Buddy Auto-Generated End",
		"13.14.15.16 other.com",
	}

	block := FindBlock(lines)
	if block == nil {
		t.Fatal("应找到标记区块")
	}
	if block.StartLine != 2 {
		t.Errorf("StartLine = %d, want 2", block.StartLine)
	}
	if block.EndLine != 5 {
		t.Errorf("EndLine = %d, want 5", block.EndLine)
	}
	if len(block.Entries) != 2 {
		t.Errorf("Entries = %d, want 2", len(block.Entries))
	}
}

func TestFindBlock_NotFound(t *testing.T) {
	lines := []string{"1.2.3.4 example.com", "# just a comment"}
	block := FindBlock(lines)
	if block != nil {
		t.Error("不应找到标记区块")
	}
}

func TestUpdateLines_NewBlock(t *testing.T) {
	lines := []string{"# existing content", "1.2.3.4 example.com"}
	ipMap := map[string]string{"github.com": "5.6.7.8"}
	result := UpdateLines(lines, ipMap)

	found := false
	for _, line := range result {
		if strings.Contains(line, "GitHub-Buddy Auto-Generated Start") {
			found = true
			break
		}
	}
	if !found {
		t.Error("应追加标记区块")
	}
}

func TestUpdateLines_ReplaceBlock(t *testing.T) {
	lines := []string{
		"# existing",
		"# GitHub-Buddy Auto-Generated Start",
		"1.1.1.1 github.com",
		"# GitHub-Buddy Auto-Generated End",
		"# other",
	}
	ipMap := map[string]string{"github.com": "2.2.2.2"}
	result := UpdateLines(lines, ipMap)

	// 验证旧 IP 已被替换
	for _, line := range result {
		if strings.Contains(line, "1.1.1.1") {
			t.Error("旧 IP 应被替换")
		}
	}

	// 验证新 IP 存在
	found := false
	for _, line := range result {
		if strings.Contains(line, "2.2.2.2") {
			found = true
			break
		}
	}
	if !found {
		t.Error("新 IP 应存在")
	}

	// 验证区块外内容未被修改
	if result[0] != "# existing" || result[len(result)-1] != "# other" {
		t.Error("区块外内容不应被修改")
	}
}

func TestRemoveBlock(t *testing.T) {
	lines := []string{
		"# before",
		"# GitHub-Buddy Auto-Generated Start",
		"1.1.1.1 github.com",
		"# GitHub-Buddy Auto-Generated End",
		"# after",
	}
	result := RemoveBlock(lines)
	if len(result) != 2 {
		t.Errorf("删除区块后应剩 2 行, 实际: %d", len(result))
	}
	if result[0] != "# before" || result[1] != "# after" {
		t.Error("区块外内容不应被修改")
	}
}

func TestDetectConflicts(t *testing.T) {
	lines := []string{
		"1.1.1.1 github.com",
		"# GitHub-Buddy Auto-Generated Start",
		"2.2.2.2 github.com",
		"# GitHub-Buddy Auto-Generated End",
	}
	ipMap := map[string]string{"github.com": "2.2.2.2"}
	conflicts := DetectConflicts(lines, []string{"github.com"}, ipMap)
	if len(conflicts) != 1 {
		t.Errorf("应检测到 1 个冲突, 实际: %d", len(conflicts))
	}
	if len(conflicts) > 0 && conflicts[0].ExistIP != "1.1.1.1" {
		t.Errorf("冲突 IP = %s, want 1.1.1.1", conflicts[0].ExistIP)
	}
}

func TestReadWriteHostsFile(t *testing.T) {
	// 创建临时文件
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "hosts")
	content := "# test hosts\n1.2.3.4 example.com\n"
	os.WriteFile(tmpFile, []byte(content), 0644)

	// 读取
	hf, err := ReadHostsFile(tmpFile)
	if err != nil {
		t.Fatalf("读取失败: %v", err)
	}
	if len(hf.Entries) != 1 {
		t.Errorf("Entries = %d, want 1", len(hf.Entries))
	}

	// 写入
	newLines := UpdateLines(hf.Lines, map[string]string{"github.com": "5.6.7.8"})
	err = WriteHostsFile(tmpFile, newLines)
	if err != nil {
		t.Fatalf("写入失败: %v", err)
	}

	// 验证写入结果
	data, _ := os.ReadFile(tmpFile)
	if !strings.Contains(string(data), "5.6.7.8") {
		t.Error("写入内容应包含新 IP")
	}
	if !strings.Contains(string(data), "example.com") {
		t.Error("原有内容应保留")
	}
}
