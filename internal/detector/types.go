package detector

// IPEntry 表示一个 IP 检测结果
type IPEntry struct {
	IP           string  `json:"ip"`
	Domain       string  `json:"domain"`
	Latency      float64 `json:"latency_ms"`       // 延迟（毫秒）
	LossRate     float64 `json:"loss_rate"`        // 丢包率（0.0-1.0）
	Port443      bool    `json:"port_443"`         // TCP 443 端口是否连通
	Port22       bool    `json:"port_22"`          // TCP 22 端口是否连通
	HTTPS        bool    `json:"https"`            // HTTPS 验证是否通过（TLS 握手 + HTTP 响应）
	HTTPSLatency float64 `json:"https_latency_ms"` // HTTPS 验证延迟（毫秒）
	Score        float64 `json:"score"`            // 质量评分（0-100）
	Available    bool    `json:"available"`        // 是否可用
}

// DomainIPs 表示一个域名对应的多个候选 IP
type DomainIPs struct {
	Domain string    `json:"domain"`
	IPs    []IPEntry `json:"ips"`
	BestIP *IPEntry  `json:"best_ip,omitempty"` // 最优 IP
}

// DetectResult 表示一次完整检测的结果
type DetectResult struct {
	Domains   map[string]*DomainIPs `json:"domains"`
	Timestamp int64                 `json:"timestamp"`
	ICMPUsed  bool                  `json:"icmp_used"`  // 是否使用了 ICMP 检测
	HTTPSUsed bool                  `json:"https_used"` // 是否使用了 HTTPS 验证
}
