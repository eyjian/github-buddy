package detector

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// Detector 是统一的 IP 检测器，协调数据源获取、多维度检测、评分筛选
type Detector struct {
	source     *MultiSource
	pinger     *Pinger
	tcpChecker *TCPChecker
	logger     zerolog.Logger
	icmpOK     bool // ICMP 是否可用
}

// NewDetector 创建检测器
func NewDetector(logger zerolog.Logger) *Detector {
	d := &Detector{
		source:     NewMultiSource(NewGitHub520Source()),
		pinger:     NewPinger(3, 5*time.Second),
		tcpChecker: NewTCPChecker(3 * time.Second),
		logger:     logger,
	}

	// 检测 ICMP 是否可用
	d.icmpOK = IsPingAvailable()
	if !d.icmpOK {
		d.logger.Warn().Msg("ICMP ping 不可用，降级为 TCP-only 检测模式")
	}

	return d
}

// DetectAll 执行完整的 IP 检测流程
// 1. 从数据源获取候选 IP（失败则使用默认列表）
// 2. 对每个 IP 执行多维度检测（ICMP + TCP 443 + TCP 22）
// 3. 评分排序，选择最优和备选 IP
func (d *Detector) DetectAll(ctx context.Context) (*DetectResult, error) {
	// 1. 获取候选 IP
	ips, err := d.source.FetchIPs(ctx)
	if err != nil {
		d.logger.Warn().Err(err).Msg("从数据源获取 IP 失败，使用内置默认 IP 列表")
		ips = GetDefaultIPs()
	}

	d.logger.Info().Int("domains", len(ips)).Msg("获取候选 IP 列表完成")

	// 2. 并发检测所有域名的所有 IP
	result := &DetectResult{
		Domains:   make(map[string]*DomainIPs),
		Timestamp: time.Now().Unix(),
		ICMPUsed:  d.icmpOK,
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	for domain, ipList := range ips {
		wg.Add(1)
		go func(dom string, candidates []string) {
			defer wg.Done()
			domainResult := d.detectDomain(ctx, dom, candidates)
			mu.Lock()
			result.Domains[dom] = domainResult
			mu.Unlock()
		}(domain, ipList)
	}

	wg.Wait()

	d.logger.Info().Int("domains", len(result.Domains)).Bool("icmp_used", d.icmpOK).Msg("IP 检测完成")
	return result, nil
}

// detectDomain 对单个域名的所有候选 IP 进行检测和排序
func (d *Detector) detectDomain(ctx context.Context, domain string, ips []string) *DomainIPs {
	domainResult := &DomainIPs{
		Domain: domain,
		IPs:    make([]IPEntry, 0, len(ips)),
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, ip := range ips {
		wg.Add(1)
		go func(ipAddr string) {
			defer wg.Done()
			entry := d.detectIP(ctx, domain, ipAddr)
			mu.Lock()
			domainResult.IPs = append(domainResult.IPs, entry)
			mu.Unlock()
		}(ip)
	}

	wg.Wait()

	// 评分排序并选择最优 IP
	best, _ := SelectBestIPs(domainResult.IPs, 2)
	domainResult.BestIP = best

	if best != nil {
		d.logger.Debug().
			Str("domain", domain).
			Str("best_ip", best.IP).
			Float64("score", best.Score).
			Float64("latency_ms", best.Latency).
			Msg("域名最优 IP")
	}

	return domainResult
}

// detectIP 对单个 IP 执行多维度检测
func (d *Detector) detectIP(ctx context.Context, domain, ip string) IPEntry {
	entry := IPEntry{
		IP:     ip,
		Domain: domain,
	}

	var wg sync.WaitGroup

	// TCP 443 端口检测
	wg.Add(1)
	go func() {
		defer wg.Done()
		r := d.tcpChecker.CheckPort(ctx, ip, 443)
		entry.Port443 = r.OK
		if r.OK && entry.Latency <= 0 {
			entry.Latency = r.Latency
		}
	}()

	// TCP 22 端口检测
	wg.Add(1)
	go func() {
		defer wg.Done()
		r := d.tcpChecker.CheckPort(ctx, ip, 22)
		entry.Port22 = r.OK
	}()

	// ICMP ping 检测（如果可用）
	if d.icmpOK {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r := d.pinger.Ping(ctx, ip)
			if r.OK {
				entry.Latency = r.Latency
				entry.LossRate = r.LossRate
			}
		}()
	}

	wg.Wait()

	entry.Score = ScoreIP(&entry)
	entry.Available = entry.Port443 || entry.Port22
	return entry
}

// FallbackIP 在主 IP 失效时切换到备选 IP
// 传入当前域名的检测结果，返回可用的备选 IP
func FallbackIP(domainIPs *DomainIPs) *IPEntry {
	if domainIPs == nil || len(domainIPs.IPs) == 0 {
		return nil
	}

	ranked := RankIPs(domainIPs.IPs)
	for i := range ranked {
		if ranked[i].Available {
			return &ranked[i]
		}
	}

	// 没有可用的，返回评分最高的
	if len(ranked) > 0 {
		return &ranked[0]
	}
	return nil
}
