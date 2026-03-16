package detector

import (
	"testing"
)

func TestScoreIP_FullScore(t *testing.T) {
	entry := &IPEntry{
		Latency:  10,
		LossRate: 0,
		Port443:  true,
		Port22:   true,
	}
	score := ScoreIP(entry)
	if score < 90 {
		t.Errorf("满分场景评分应≥90, 实际: %.2f", score)
	}
}

func TestScoreIP_HighLatency(t *testing.T) {
	entry := &IPEntry{
		Latency:  400,
		LossRate: 0,
		Port443:  true,
		Port22:   true,
	}
	score := ScoreIP(entry)
	if score > 70 {
		t.Errorf("高延迟场景评分应＜70, 实际: %.2f", score)
	}
}

func TestScoreIP_PortsDown(t *testing.T) {
	entry := &IPEntry{
		Latency:  10,
		LossRate: 0,
		Port443:  false,
		Port22:   false,
	}
	score := ScoreIP(entry)
	if score > 80 {
		t.Errorf("端口不通场景评分应＜80, 实际: %.2f", score)
	}
}

func TestScoreIP_PacketLoss(t *testing.T) {
	entry := &IPEntry{
		Latency:  10,
		LossRate: 0.5,
		Port443:  true,
		Port22:   true,
	}
	score := ScoreIP(entry)
	full := ScoreIP(&IPEntry{Latency: 10, LossRate: 0, Port443: true, Port22: true})
	if score >= full {
		t.Errorf("有丢包的评分(%.2f)应低于无丢包(%.2f)", score, full)
	}
}

func TestRankIPs(t *testing.T) {
	entries := []IPEntry{
		{IP: "1.1.1.1", Latency: 100, LossRate: 0, Port443: true, Port22: true},
		{IP: "2.2.2.2", Latency: 10, LossRate: 0, Port443: true, Port22: true},
		{IP: "3.3.3.3", Latency: 50, LossRate: 0.5, Port443: true, Port22: false},
	}
	ranked := RankIPs(entries)
	if ranked[0].IP != "2.2.2.2" {
		t.Errorf("最优 IP 应为 2.2.2.2, 实际: %s", ranked[0].IP)
	}
}

func TestSelectBestIPs(t *testing.T) {
	entries := []IPEntry{
		{IP: "1.1.1.1", Latency: 100, LossRate: 0, Port443: true, Port22: true},
		{IP: "2.2.2.2", Latency: 10, LossRate: 0, Port443: true, Port22: true},
		{IP: "3.3.3.3", Latency: 50, LossRate: 0, Port443: true, Port22: true},
	}
	best, backups := SelectBestIPs(entries, 2)
	if best == nil {
		t.Fatal("最优 IP 不应为 nil")
	}
	if best.IP != "2.2.2.2" {
		t.Errorf("最优 IP 应为 2.2.2.2, 实际: %s", best.IP)
	}
	if len(backups) != 2 {
		t.Errorf("应有 2 个备选 IP, 实际: %d", len(backups))
	}
}

func TestSelectBestIPs_Empty(t *testing.T) {
	best, backups := SelectBestIPs(nil, 2)
	if best != nil {
		t.Error("空列表应返回 nil")
	}
	if backups != nil {
		t.Error("空列表备选应为 nil")
	}
}

func TestParsePingLatency(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   float64
	}{
		{
			name:   "Linux格式",
			output: "rtt min/avg/max/mdev = 1.234/5.678/9.012/0.123 ms",
			want:   5.678,
		},
		{
			name:   "Windows格式",
			output: "Average = 12ms",
			want:   12,
		},
		{
			name:   "无匹配",
			output: "no output",
			want:   0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parsePingLatency(tt.output)
			if got != tt.want {
				t.Errorf("parsePingLatency() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParsePingLoss(t *testing.T) {
	tests := []struct {
		name   string
		output string
		want   float64
	}{
		{
			name:   "0%丢包",
			output: "3 packets transmitted, 3 received, 0% packet loss",
			want:   0,
		},
		{
			name:   "50%丢包",
			output: "2 packets transmitted, 1 received, 50% packet loss",
			want:   0.5,
		},
		{
			name:   "100%丢包",
			output: "3 packets transmitted, 0 received, 100% packet loss",
			want:   1.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parsePingLoss(tt.output)
			if got != tt.want {
				t.Errorf("parsePingLoss() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsTargetDomain(t *testing.T) {
	if !isTargetDomain("github.com") {
		t.Error("github.com 应为目标域名")
	}
	if isTargetDomain("google.com") {
		t.Error("google.com 不应为目标域名")
	}
}

func TestGetDefaultIPs(t *testing.T) {
	ips := GetDefaultIPs()
	if len(ips) == 0 {
		t.Error("默认 IP 列表不应为空")
	}
	if _, ok := ips["github.com"]; !ok {
		t.Error("默认 IP 列表应包含 github.com")
	}
}
