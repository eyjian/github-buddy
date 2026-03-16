package detector

// DefaultIPs 内置的兜底 IP 列表
// 当所有数据源均不可达时使用这份列表
// 注意：这些 IP 可能会过时，仅作为应急兜底
var DefaultIPs = map[string][]string{
	"github.com": {
		"20.205.243.166",
		"20.27.177.113",
		"140.82.121.3",
	},
	"ssh.github.com": {
		"20.205.243.160",
	},
	"gist.github.com": {
		"140.82.112.4",
	},
	"raw.githubusercontent.com": {
		"185.199.108.133",
		"185.199.109.133",
		"185.199.110.133",
		"185.199.111.133",
	},
	"api.github.com": {
		"20.205.243.168",
		"140.82.112.5",
	},
	"assets-cdn.github.com": {
		"185.199.108.153",
		"185.199.109.153",
		"185.199.110.153",
		"185.199.111.153",
	},
	"github.global.ssl.fastly.net": {
		"146.75.77.194",
	},
	"collector.github.com": {
		"140.82.113.21",
	},
	"avatars.githubusercontent.com": {
		"185.199.108.133",
		"185.199.109.133",
	},
	"codeload.github.com": {
		"140.82.121.9",
		"20.205.243.165",
	},
}

// GetDefaultIPs 返回内置默认 IP 列表的副本
func GetDefaultIPs() map[string][]string {
	result := make(map[string][]string, len(DefaultIPs))
	for domain, ips := range DefaultIPs {
		copied := make([]string, len(ips))
		copy(copied, ips)
		result[domain] = copied
	}
	return result
}
