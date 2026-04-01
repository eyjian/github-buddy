package detector

// DefaultIPs 内置的兜底 IP 列表
// 当所有数据源均不可达时使用这份列表
// 注意：这些 IP 可能会过时，仅作为应急兜底
var DefaultIPs = map[string][]string{
	"github.com": {
		"20.205.243.166",
		"140.82.112.4",
		"140.82.114.4",
	},
	"alive.github.com": {
		"140.82.113.25",
	},
	"live.github.com": {
		"140.82.113.25",
	},
	"ssh.github.com": {
		"140.82.112.35",
		"140.82.113.35",
	},
	"gist.github.com": {
		"20.205.243.166",
	},
	"raw.githubusercontent.com": {
		"185.199.110.133",
		"185.199.108.133",
	},
	"api.github.com": {
		"20.205.243.168",
		"140.82.112.5",
	},
	"codeload.github.com": {
		"20.205.243.165",
		"140.82.112.10",
	},
	"avatars.githubusercontent.com": {
		"185.199.110.133",
	},
	"github.githubassets.com": {
		"185.199.111.154",
	},
	"objects.githubusercontent.com": {
		"185.199.110.133",
	},
	"media.githubusercontent.com": {
		"185.199.110.133",
	},
	"user-images.githubusercontent.com": {
		"185.199.110.133",
	},
	"github-cloud.s3.amazonaws.com": {
		"16.15.183.87",
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
