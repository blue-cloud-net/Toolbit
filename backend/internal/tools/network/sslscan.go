package network

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/blue-cloud-net/Toolbit/backend/internal/netguard"
	"github.com/blue-cloud-net/Toolbit/backend/internal/pkisvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tlssvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// 扫描阶段名（完整扫描用，顺序即执行顺序）。
const (
	// ScanPhaseResolve 表示域名解析。
	ScanPhaseResolve = "resolve"
	// ScanPhaseConnect 表示 TCP 建连。
	ScanPhaseConnect = "connect"
	// ScanPhaseHandshake 表示首次 TLS 握手。
	ScanPhaseHandshake = "handshake"
	// ScanPhaseCertificate 表示证书与链分析。
	ScanPhaseCertificate = "certificate"
	// ScanPhaseProtocols 表示协议版本矩阵。
	ScanPhaseProtocols = "protocols"
	// ScanPhaseCiphers 表示密码套件矩阵。
	ScanPhaseCiphers = "ciphers"
	// ScanPhaseFeatures 表示协议特性探测。
	ScanPhaseFeatures = "features"
	// ScanPhaseClients 表示客户端握手模拟。
	ScanPhaseClients = "clients"
	// ScanPhaseOffsite 表示出网附加检查（吊销 / CAA / HSTS）。
	ScanPhaseOffsite = "revocation"
	// ScanPhaseGrade 表示评级计算。
	ScanPhaseGrade = "grade"
)

// 检查项标识。
const (
	// CheckProtocols 是协议版本矩阵。
	CheckProtocols = "protocols"
	// CheckCiphers 是密码套件矩阵。
	CheckCiphers = "ciphers"
	// CheckFeatures 是协议特性。
	CheckFeatures = "features"
	// CheckClients 是客户端握手模拟。
	CheckClients = "clients"
	// CheckCert 是证书与链分析。
	CheckCert = "cert"
	// CheckRevocation 是吊销检查。
	CheckRevocation = "revocation"
	// CheckCAA 是 CAA 记录。
	CheckCAA = "caa"
	// CheckHSTS 是 HSTS 头。
	CheckHSTS = "hsts"
)

// 扫描模式。
const (
	// ScanModeQuick 只做协议矩阵与证书分析。
	ScanModeQuick = "quick"
	// ScanModeFull 做全部检查。
	ScanModeFull = "full"
)

// 单次扫描的预算约束（毫秒）。
const (
	// defaultScanBudgetMs 是默认总预算。
	defaultScanBudgetMs = 60000
	// maxScanBudgetMs 是允许的最大总预算。
	maxScanBudgetMs = 180000
	// probeTimeout 是单次探测握手超时。
	probeTimeout = 5 * time.Second
	// scanWorkers 是阶段内并发度。
	scanWorkers = 4
	// maxCiphersPerVersion 是每个版本最多探测的套件数。
	maxCiphersPerVersion = 40
)

// 特性状态。
const (
	// FeatureSupported 表示探测到支持。
	FeatureSupported = "supported"
	// FeatureNotSupported 表示探测到不支持。
	FeatureNotSupported = "not_supported"
	// FeatureUnavailable 表示当前引擎无法判定（**不是**不支持）。
	FeatureUnavailable = "unavailable"
)

// 套件判定结果。
const (
	// SuiteSupported 表示服务端接受该套件。
	SuiteSupported = "supported"
	// SuiteRejected 表示服务端拒绝该套件。
	SuiteRejected = "rejected"
	// SuiteUnavailable 表示本机未提供该套件，无法探测。
	SuiteUnavailable = "unavailable"
)

// ScanOptions 是一次完整扫描的参数。
type ScanOptions struct {
	// Host 是目标主机。
	Host string
	// Port 是目标端口，默认 443。
	Port int
	// SNI 是服务器名字指示；为空时用 Host。
	SNI string
	// Depth 是链深度上限。
	Depth int
	// Mode 是扫描模式：quick / full。
	Mode string
	// Checks 是要执行的检查项；为空表示按 Mode 推导。
	Checks []string
	// Clients 是要模拟的客户端画像 id；为空表示全部。
	Clients []string
	// BudgetMs 是总预算；0 表示默认值。
	BudgetMs int
}

// ScanEvent 是一次扫描进度通知。
type ScanEvent struct {
	// Phase 是阶段名。
	Phase string
	// Done 是已完成项数。
	Done int
	// Total 是总项数。
	Total int
	// Item 是当前正在处理的项（套件名或画像 id）。
	Item string
}

// ScanEndpoint 是目标端点信息。
type ScanEndpoint struct {
	// Host 是目标主机。
	Host string `json:"host"`
	// Port 是目标端口。
	Port int `json:"port"`
	// SNI 是实际使用的 SNI。
	SNI string `json:"sni"`
	// Addr 是实际连接的 IP。
	Addr string `json:"addr"`
	// ElapsedMs 是本次扫描耗时。
	ElapsedMs int64 `json:"elapsedMs"`
}

// ScanNegotiated 是首次握手的协商结果。
type ScanNegotiated struct {
	// Version 是协议版本短名。
	Version string `json:"version"`
	// VersionWire 是协议版本 wire 值。
	VersionWire uint16 `json:"versionWire"`
	// Cipher 是套件名。
	Cipher string `json:"cipher"`
	// CipherId 是套件 wire 值。
	CipherId uint16 `json:"cipherId"`
	// KeyExchangeGroup 是密钥交换组名。
	KeyExchangeGroup string `json:"keyExchangeGroup"`
	// PeerSignatureAlgorithm 是对端握手签名算法。
	PeerSignatureAlgorithm string `json:"peerSignatureAlgorithm"`
	// Alpn 是 ALPN 协商结果。
	Alpn string `json:"alpn"`
}

// ScanCertificate 是叶证书摘要与链验证结论。
type ScanCertificate struct {
	// Subject 是主体单行文本。
	Subject string `json:"subject"`
	// Issuer 是签发者单行文本。
	Issuer string `json:"issuer"`
	// Serial 是序列号（大写十六进制）。
	Serial string `json:"serial"`
	// NotBefore 是生效时间（RFC 3339）。
	NotBefore string `json:"notBefore"`
	// NotAfter 是失效时间（RFC 3339）。
	NotAfter string `json:"notAfter"`
	// DaysRemaining 是剩余天数（小数）。
	DaysRemaining float64 `json:"daysRemaining"`
	// KeyAlgorithm 是公钥算法。
	KeyAlgorithm string `json:"keyAlgorithm"`
	// KeyBits 是 RSA 位数；非 RSA 为 0。
	KeyBits int `json:"keyBits"`
	// Curve 是椭圆曲线名。
	Curve string `json:"curve"`
	// SignatureAlgorithm 是证书签名算法。
	SignatureAlgorithm string `json:"signatureAlgorithm"`
	// SanCount 是 SAN 条目数。
	SanCount int `json:"sanCount"`
	// SanMatchesHost 报告 SAN 是否覆盖目标主机。
	SanMatchesHost bool `json:"sanMatchesHost"`
	// IsCa 报告是否为 CA 证书。
	IsCa bool `json:"isCa"`
	// SelfSigned 报告是否自签名。
	SelfSigned bool `json:"selfSigned"`
	// SctCount 是证书内嵌的 SCT 数量。
	SctCount int `json:"sctCount"`
	// MustStaple 报告证书是否声明 must-staple。
	MustStaple bool `json:"mustStaple"`
	// ChainValid 报告链是否通过验证（含主机名匹配）。
	ChainValid bool `json:"chainValid"`
	// Verified 报告是否用系统信任根完成验证。
	Verified bool `json:"verified"`
	// ImplicitTrust 报告是否退化为链内自洽判定。
	ImplicitTrust bool `json:"implicitTrust"`
	// ChainErrors 是链验证错误码。
	ChainErrors []string `json:"chainErrors"`
}

// ProtocolResult 是某个协议版本的探测结论。
type ProtocolResult struct {
	// Version 是版本短名。
	Version string `json:"version"`
	// VersionWire 是版本 wire 值。
	VersionWire uint16 `json:"versionWire"`
	// Supported 报告该版本是否可协商。
	Supported bool `json:"supported"`
	// NegotiatedCipher 是该版本下单次握手协商出的套件。
	NegotiatedCipher string `json:"negotiatedCipher,omitempty"`
	// Reason 是失败原因分类；成功时为空。
	Reason string `json:"reason,omitempty"`
}

// CipherResult 是单个套件的探测结论。
type CipherResult struct {
	// Version 是该套件所属的协议版本短名。
	Version string `json:"version"`
	// Name 是套件名。
	Name string `json:"name"`
	// Id 是套件 wire 值。
	Id uint16 `json:"id"`
	// Family 是算法族（AEAD / CBC / RC4 / 3DES / 其他）。
	Family string `json:"family"`
	// Bits 是加密强度位数。
	Bits int `json:"bits"`
	// Kex 是密钥交换方式。
	Kex string `json:"kex"`
	// Verdict 是判定结果（supported / rejected / unavailable）。
	Verdict string `json:"verdict"`
}

// FeatureResult 是单个协议特性的探测结论。
type FeatureResult struct {
	// Id 是稳定标识。
	Id string `json:"id"`
	// Label 是中文标签（兜底展示用，**不**作为 i18n key）。
	Label string `json:"label"`
	// Status 是 supported / not_supported / unavailable。
	Status string `json:"status"`
	// Detail 是补充说明。
	Detail string `json:"detail,omitempty"`
}

// ClientResult 是一档客户端画像的握手模拟结论。
type ClientResult struct {
	// Id 是画像标识。
	Id string `json:"id"`
	// Vendor 是厂商。
	Vendor string `json:"vendor"`
	// Os 是操作系统。
	Os string `json:"os"`
	// Engine 是 TLS 引擎。
	Engine string `json:"engine"`
	// ProfileVersion 是画像版本。
	ProfileVersion string `json:"profileVersion"`
	// Supported 报告该画像能否完成握手。
	Supported bool `json:"supported"`
	// Reason 是失败归因；成功时为空。
	Reason string `json:"reason,omitempty"`
	// NegotiatedVersion 是模拟握手协商出的版本。
	NegotiatedVersion string `json:"negotiatedVersion,omitempty"`
	// Cipher 是模拟握手协商出的套件。
	Cipher string `json:"cipher,omitempty"`
	// Alpn 是模拟握手的 ALPN 结果。
	Alpn string `json:"alpn,omitempty"`
}

// CAAResult 是 CAA 记录查询结论。
type CAAResult struct {
	// Queried 报告是否执行了查询。
	Queried bool `json:"queried"`
	// Records 是读到的 CAA 记录。
	Records []string `json:"records"`
	// Permitted 报告签发者是否被 CAA 允许；无记录时为 true。
	Permitted bool `json:"permitted"`
	// Violations 是不被允许的标签值。
	Violations []string `json:"violations"`
	// Error 是查询失败原因。
	Error string `json:"error,omitempty"`
}

// RevocationEntry 是一条吊销检查记录。
type RevocationEntry struct {
	// Url 是检查地址。
	Url string `json:"url"`
	// Status 是 good / revoked / unknown。
	Status string `json:"status"`
	// ThisUpdate 是本次更新时刻。
	ThisUpdate string `json:"thisUpdate,omitempty"`
	// NextUpdate 是下次更新时刻。
	NextUpdate string `json:"nextUpdate,omitempty"`
	// Error 是失败原因。
	Error string `json:"error,omitempty"`
}

// RevocationResult 是吊销检查汇总。
type RevocationResult struct {
	// Ocsp 是 OCSP 查询记录。
	Ocsp []RevocationEntry `json:"ocsp"`
	// Crl 是 CRL 查询记录。
	Crl []RevocationEntry `json:"crl"`
}

// HSTSResult 是 HSTS 头检查结论。
type HSTSResult struct {
	// Present 报告是否返回了 Strict-Transport-Security。
	Present bool `json:"present"`
	// MaxAge 是 max-age（秒）。
	MaxAge int64 `json:"maxAge"`
	// IncludeSubDomains 报告是否含 includeSubDomains。
	IncludeSubDomains bool `json:"includeSubDomains"`
	// Preload 报告是否含 preload。
	Preload bool `json:"preload"`
	// Error 是请求失败原因。
	Error string `json:"error,omitempty"`
}

// GradeCategories 是四个象限的得分。
type GradeCategories struct {
	// Certificate 是证书象限得分。
	Certificate int `json:"certificate"`
	// Protocols 是协议象限得分。
	Protocols int `json:"protocols"`
	// KeyExchange 是密钥交换象限得分。
	KeyExchange int `json:"keyExchange"`
	// CipherStrength 是加密强度象限得分。
	CipherStrength int `json:"cipherStrength"`
}

// GradeIssue 是一条评级说明。
type GradeIssue struct {
	// Id 是稳定标识。
	Id string `json:"id"`
	// Impact 是影响程度：critical / high / medium / low / unknown。
	Impact string `json:"impact"`
	// Detail 是补充说明。
	Detail string `json:"detail,omitempty"`
}

// GradeResult 是综合评级。
type GradeResult struct {
	// Letter 是等级字母。
	Letter string `json:"letter"`
	// Score 是四象限平均分。
	Score int `json:"score"`
	// Categories 是四象限得分。
	Categories GradeCategories `json:"categories"`
	// Caps 是命中的档位上限。
	Caps []string `json:"caps"`
	// Issues 是评级说明。
	Issues []GradeIssue `json:"issues"`
}

// ScanResult 是一次完整扫描的结果。
type ScanResult struct {
	// Endpoint 是目标端点。
	Endpoint ScanEndpoint `json:"endpoint"`
	// Negotiated 是首次握手的协商结果。
	Negotiated ScanNegotiated `json:"negotiated"`
	// Certificate 是叶证书摘要。
	Certificate ScanCertificate `json:"certificate"`
	// Chain 是整条证书链（与 cert-chain-parse 的响应同构）。
	Chain *pkisvc.ChainInfo `json:"chain"`
	// Protocols 是协议版本矩阵。
	Protocols []ProtocolResult `json:"protocols"`
	// Ciphers 是密码套件矩阵。
	Ciphers []CipherResult `json:"ciphers"`
	// Features 是协议特性。
	Features []FeatureResult `json:"features"`
	// Clients 是客户端握手模拟。
	Clients []ClientResult `json:"clients"`
	// Caa 是 CAA 查询结论。
	Caa CAAResult `json:"caa"`
	// Revocation 是吊销检查结论。
	Revocation RevocationResult `json:"revocation"`
	// Hsts 是 HSTS 头结论。
	Hsts HSTSResult `json:"hsts"`
	// Grade 是综合评级。
	Grade GradeResult `json:"grade"`
}

// ValidateScanOptions 校验扫描参数。
func ValidateScanOptions(opts *ScanOptions) error {
	if _, err := normalizeHost(opts.Host); err != nil {
		return err
	}
	if _, err := normalizePort(opts.Port, 443); err != nil {
		return err
	}
	if opts.Depth < 0 || opts.Depth > maxChainDepth {
		return tools.InvalidField("depth", []string{"0..10"}, opts.Depth)
	}
	switch opts.Mode {
	case "", ScanModeQuick, ScanModeFull:
	default:
		return tools.InvalidField("mode", []string{ScanModeQuick, ScanModeFull}, opts.Mode)
	}
	known := map[string]bool{}
	for _, id := range AllChecks() {
		known[id] = true
	}
	for _, id := range opts.Checks {
		if !known[id] {
			return tools.InvalidField("checks", AllChecks(), id)
		}
	}
	profiles := clientProfileIndex()
	for _, id := range opts.Clients {
		if _, ok := profiles[id]; !ok {
			return tools.InvalidField("clients", ClientProfileIDs(), id)
		}
	}
	if opts.BudgetMs < 0 || opts.BudgetMs > maxScanBudgetMs {
		return tools.InvalidField("budgetMs", []string{"0..180000"}, opts.BudgetMs)
	}
	return nil
}

// AllChecks 返回全部检查项标识（顺序稳定）。
func AllChecks() []string {
	return []string{
		CheckProtocols, CheckCiphers, CheckFeatures, CheckClients,
		CheckCert, CheckRevocation, CheckCAA, CheckHSTS,
	}
}

// enabledChecks 按 Mode 与 Checks 推导实际要跑的检查项。
func enabledChecks(opts ScanOptions) map[string]bool {
	out := make(map[string]bool, len(AllChecks()))
	if len(opts.Checks) > 0 {
		for _, id := range opts.Checks {
			out[id] = true
		}
		return out
	}
	if opts.Mode == ScanModeQuick {
		out[CheckProtocols] = true
		out[CheckCert] = true
		return out
	}
	for _, id := range AllChecks() {
		out[id] = true
	}
	return out
}

// scanner 承载一次扫描的共享状态。
type scanner struct {
	// guard 是出网防护。
	guard *netguard.Guard
	// ctx 是扫描上下文。
	ctx context.Context
	// host 是目标主机。
	host string
	// port 是目标端口。
	port int
	// sni 是 SNI。
	sni string
	// depth 是链深度上限。
	depth int
	// onEvent 是进度回调，可为 nil。
	onEvent func(ScanEvent)
	// chainDER 是首次握手拿到的证书链 DER（叶 → 根）。
	chainDER [][]byte
	// done 是已完成项数。
	done int
	// total 是总项数。
	total int
	// mu 保护 done 与进度回调。
	mu sync.Mutex
}

// report 通知一次阶段级进度（推进 done）。
func (s *scanner) report(phase, item string) {
	s.mu.Lock()
	s.done++
	done := s.done
	s.mu.Unlock()
	if s.onEvent == nil {
		return
	}
	s.onEvent(ScanEvent{Phase: phase, Done: done, Total: s.total, Item: item})
}

// reportItem 通知一次阶段内子项进度（**不**推进 done）。
//
// 套件与客户端画像的数量随目标不同而变，若每项都 +1，done 会超过 total。
func (s *scanner) reportItem(phase, item string) {
	if s.onEvent == nil {
		return
	}
	s.mu.Lock()
	done := s.done
	s.mu.Unlock()
	s.onEvent(ScanEvent{Phase: phase, Done: done, Total: s.total, Item: item})
}

// probe 在已校验的地址上做一次握手。
//
// 每次探测都重新建连（HTTP/1.1 的连接复用语义不适用于「换参数再握手」），
// 但地址只解析一次 —— DNS 由 `DialTCP` 内部复用同一解析结果。
func (s *scanner) probe(opts tlssvc.Options) (*tlssvc.State, error) {
	raw, _, err := s.guard.DialTCP(s.ctx, s.host, s.port)
	if err != nil {
		return nil, err
	}
	session, hsErr := tlssvc.Handshake(s.ctx, raw, opts)
	if hsErr != nil {
		return nil, hsErr
	}
	defer func() { _ = session.Close() }()
	state := session.State()
	return &state, nil
}

// ScanTLS 执行一次完整扫描。
func ScanTLS(
	ctx context.Context, g *netguard.Guard, opts ScanOptions,
	onEvent func(ScanEvent),
) (*ScanResult, []toolerr.Warning, error) {
	if err := requireNetwork(g); err != nil {
		return nil, nil, err
	}
	if err := ValidateScanOptions(&opts); err != nil {
		return nil, nil, err
	}

	host, _ := normalizeHost(opts.Host)
	port, _ := normalizePort(opts.Port, 443)
	sni := trimOr(opts.SNI, host)
	budget := opts.BudgetMs
	if budget <= 0 {
		budget = defaultScanBudgetMs
	}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(budget)*time.Millisecond)
	defer cancel()

	checks := enabledChecks(opts)
	s := &scanner{
		guard: g, ctx: ctx, host: host, port: port, sni: sni,
		depth: opts.Depth, onEvent: onEvent,
		total: countSteps(checks),
	}

	started := time.Now()
	result := &ScanResult{
		Endpoint:   ScanEndpoint{Host: host, Port: port, SNI: sni},
		Protocols:  []ProtocolResult{},
		Ciphers:    []CipherResult{},
		Features:   []FeatureResult{},
		Clients:    []ClientResult{},
		Caa:        CAAResult{Records: []string{}, Violations: []string{}},
		Revocation: RevocationResult{Ocsp: []RevocationEntry{}, Crl: []RevocationEntry{}},
	}

	// 阶段 1：基础握手 + 证书链。
	s.report(ScanPhaseResolve, "")
	main, err := TLSProbe(ctx, g, TLSProbeOptions{
		Host: host, Port: port, SNI: sni, Depth: opts.Depth,
		TimeoutMs: budget,
	}, func(p TLSProgress) {
		// 底层探测自带的 4 阶段进度已经包含在建连与握手内，这里不再重复上报。
	})
	if err != nil {
		return nil, nil, err
	}
	s.report(ScanPhaseConnect, "")
	s.report(ScanPhaseHandshake, "")
	result.Endpoint.Addr = main.Addr
	result.Negotiated = ScanNegotiated{
		Version:                main.NegotiatedVersion,
		VersionWire:            main.NegotiatedVersionWire,
		Cipher:                 main.CipherSuite,
		CipherId:               main.CipherSuiteID,
		KeyExchangeGroup:       main.KeyExchangeGroup,
		PeerSignatureAlgorithm: main.PeerSignatureAlgorithm,
	}

	certInfo, chainInfo, leafURLs := analyzeCertificate(main, host)
	result.Certificate = certInfo
	result.Chain = chainInfo
	s.chainDER = main.ChainDER
	s.report(ScanPhaseCertificate, "")

	if checks[CheckProtocols] {
		result.Protocols = s.scanProtocols()
		s.report(ScanPhaseProtocols, "")
	}
	if checks[CheckCiphers] {
		result.Ciphers = s.scanCiphers(result.Protocols)
		s.report(ScanPhaseCiphers, "")
	}
	if checks[CheckFeatures] {
		result.Features = s.scanFeatures(result.Negotiated, certInfo)
		s.report(ScanPhaseFeatures, "")
	}
	if checks[CheckClients] {
		result.Clients = s.scanClients(opts.Clients)
		s.report(ScanPhaseClients, "")
	}
	if checks[CheckRevocation] || checks[CheckCAA] || checks[CheckHSTS] {
		result.Revocation, result.Caa, result.Hsts = s.scanOffsite(leafURLs, host, port, checks)
		s.report(ScanPhaseOffsite, "")
	}

	result.Grade = gradeScan(result, checks)
	s.report(ScanPhaseGrade, "")

	result.Endpoint.ElapsedMs = time.Since(started).Milliseconds()
	warnings := scanWarnings(result)
	return result, warnings, nil
}

// countSteps 统计本次扫描会上报的进度步数。
//
// 固定步骤 5 步（解析 / 建连 / 握手 / 证书 / 评级）；出网附加检查
// （吊销 / CAA / HSTS）合并在同一个阶段内，只算一步。
func countSteps(checks map[string]bool) int {
	steps := 5
	for _, id := range []string{CheckProtocols, CheckCiphers, CheckFeatures, CheckClients} {
		if checks[id] {
			steps++
		}
	}
	if checks[CheckRevocation] || checks[CheckCAA] || checks[CheckHSTS] {
		steps++
	}
	return steps
}

// trimOr 返回去掉空白的 value；为空时回落到 fallback。
func trimOr(value, fallback string) string {
	if trimmed := strings.TrimSpace(value); trimmed != "" {
		return trimmed
	}
	return fallback
}
