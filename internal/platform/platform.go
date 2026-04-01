package platform

import (
	"os"
	"path/filepath"
	"runtime"
)

// IsRoot 判断当前用户是否为 root（UID 为 0）。
// Windows 下始终返回 false（Windows 使用管理员提权而非 root）。
func IsRoot() bool {
	return os.Getuid() == 0
}

// Info 包含当前平台的关键信息
type Info struct {
	OS       string // "windows", "darwin", "linux"
	HostsDir string // hosts 文件所在目录
	HostPath string // hosts 文件完整路径
	HomeDir  string // 用户主目录
	DataDir  string // ~/.github-buddy/ 数据目录
}

// Detect 检测当前平台信息
func Detect() (*Info, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	info := &Info{
		OS:      runtime.GOOS,
		HomeDir: homeDir,
		DataDir: filepath.Join(homeDir, ".github-buddy"),
	}

	switch runtime.GOOS {
	case "windows":
		info.HostsDir = filepath.Join(os.Getenv("SystemRoot"), "System32", "drivers", "etc")
		info.HostPath = filepath.Join(info.HostsDir, "hosts")
	default: // linux, darwin
		info.HostsDir = "/etc"
		info.HostPath = "/etc/hosts"
	}

	return info, nil
}

// IsWindows 判断当前系统是否为 Windows
func (p *Info) IsWindows() bool {
	return p.OS == "windows"
}

// CheckHostsWritable 检测 hosts 文件是否可写
func (p *Info) CheckHostsWritable() error {
	// 尝试以追加模式打开 hosts 文件来检测写权限
	f, err := os.OpenFile(p.HostPath, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return &PermissionError{
			Path:    p.HostPath,
			OS:      p.OS,
			RawErr:  err,
		}
	}
	f.Close()
	return nil
}

// PermissionError 权限不足错误，包含平台特定的提权指引
type PermissionError struct {
	Path   string
	OS     string
	RawErr error
}

func (e *PermissionError) Error() string {
	switch e.OS {
	case "windows":
		return "权限不足: 请以管理员身份运行命令提示符后重试"
	default:
		if IsRoot() {
			return "权限不足: 当前已是 root 用户，请检查 hosts 文件权限"
		}
		return "权限不足: 请使用 sudo github-buddy 运行"
	}
}

func (e *PermissionError) Unwrap() error {
	return e.RawErr
}
