package detector

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

// Detector 是统一的 IP 检测器，协调数据源获取、多维度检测、评分筛选
type Detector struct {
	source       Source
	pinger       *Pinger
	tcpChecker   *TCPChecker
	httpsChecker *HTTPSChecker
	logger       zerolog.Logger
	icmpOK       bool // ICMP 是否可用
}

// NewDetector 创建检测器
func NewDetector(logger zerolog.Logger) *Detector {
	d := &Detector{
		source:       NewMultiSource(NewGitHub520Source(), NewIneo6Source()),
		pinger:       NewPinger(3, 5*time.Second),
		tcpChecker:   NewTCPChecker(3 * time.Second),
		httpsChecker: NewHTTPSChecker(5 * time.Second),
		logger:       logger,
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
// 2. 对每个 IP 执行多维度检测（ICMP + TCP 443 + TCP 22 + HTTPS）
// 3. 评分排序，选择最优和备选 IP
func (d *Detector) DetectAll(ctx context.Context) (*DetectResult, error) {
	ips, err := d.source.FetchIPs(ctx)
	if err != nil {
		d.logger.Warn().Err(err).Msg("从数据源获取 IP 失败，使用内置默认 IP 列表")
		ips = GetDefaultIPs()
	}
	ips = filterCandidateMap(ips)

	d.logger.Info().Int("domains", len(ips)).Msg("获取候选 IP 列表完成")

	result := &DetectResult{
		Domains:   make(map[string]*DomainIPs),
		Timestamp: time.Now().Unix(),
		ICMPUsed:  d.icmpOK,
		HTTPSUsed: true,
	}

	var wg sync.WaitGroup
	var mu sync.Mutex

	for domain, ipList := range ips {
		if len(ipList) == 0 {
			continue
		}
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

	best, _ := SelectBestIPs(domainResult.IPs, 2)
	domainResult.BestIP = best

	if best != nil {
		d.logger.Debug().
			Str("domain", domain).
			Str("best_ip", best.IP).
			Float64("score", best.Score).
			Float64("latency_ms", best.Latency).
			Bool("https", best.HTTPS).
			Bool("basic_https", best.BasicHTTPS).
			Bool("browser_ok", best.BrowserOK).
			Msg("域名最优 IP")
	}

	return domainResult
}

// detectIP 对单个 IP 执行多维度检测（HTTPS + TCP 22 + ICMP 并发）
func (d *Detector) detectIP(ctx context.Context, domain, ip string) IPEntry {
	entry := IPEntry{
		IP:     ip,
		Domain: domain,
	}

	var wg sync.WaitGroup
	var httpsRes HTTPSCheckResult
	var port22Res TCPCheckResult
	var pingRes PingResult

	wg.Add(1)
	go func() {
		defer wg.Done()
		httpsRes = d.httpsChecker.Check(ctx, ip, domain)
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		port22Res = d.tcpChecker.CheckPort(ctx, ip, 22)
	}()

	if d.icmpOK {
		wg.Add(1)
		go func() {
			defer wg.Done()
			pingRes = d.pinger.Ping(ctx, ip)
		}()
	}

	wg.Wait()

	entry.Port443 = httpsRes.Port443OK
	entry.Port22 = port22Res.OK
	entry.BasicHTTPS = httpsRes.BasicOK
	entry.HTTPS = httpsRes.OK
	entry.HTTPSStatusCode = httpsRes.StatusCode
	entry.HTTPSLatency = httpsRes.Latency
	entry.BrowserOK = httpsRes.BrowserOK
	entry.BrowserReason = httpsRes.BrowserReason

	if pingRes.OK {
		entry.Latency = pingRes.Latency
		entry.LossRate = pingRes.LossRate
	} else if httpsRes.BasicOK && httpsRes.Latency > 0 {
		entry.Latency = httpsRes.Latency
	} else if httpsRes.Port443OK && httpsRes.ConnectLatency > 0 {
		entry.Latency = httpsRes.ConnectLatency
	}

	entry.Score = ScoreIP(&entry)
	entry.Available = entry.HTTPS || entry.BasicHTTPS || entry.Port443 || entry.Port22

	if httpsRes.Error != nil {
		d.logger.Debug().
			Str("ip", ip).
			Str("domain", domain).
			Bool("basic_https", httpsRes.BasicOK).
			Bool("browser_ok", httpsRes.BrowserOK).
			Str("browser_reason", httpsRes.BrowserReason).
			Err(httpsRes.Error).
			Msg("HTTPS 验证失败")
	}

	return entry
}

// FallbackIP 在主 IP 失效时切换到备选 IP
// 传入当前域名的检测结果，返回可用的备选 IP
// 对 github.com 域名：没有可用候选时返回 nil，避免写入不可用 IP 导致退化
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

	// github.com 域名：没有可用候选时不退而求其次，返回 nil
	if domainIPs.Domain == "github.com" {
		return nil
	}

	if len(ranked) > 0 {
		return &ranked[0]
	}
	return nil
}
