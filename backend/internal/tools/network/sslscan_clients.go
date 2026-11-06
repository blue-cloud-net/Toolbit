package network

import (
	"errors"
	"sort"

	"github.com/blue-cloud-net/Toolbit/backend/internal/netguard"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tlssvc"
)

// TLS 1.3 套件名（三件套）。
var tls13Suites = []string{
	"TLS_AES_128_GCM_SHA256",
	"TLS_AES_256_GCM_SHA384",
	"TLS_CHACHA20_POLY1305_SHA256",
}

// 现代客户端常见的 TLS 1.2 套件集。
var modernSuites = []string{
	"ECDHE-ECDSA-AES128-GCM-SHA256",
	"ECDHE-RSA-AES128-GCM-SHA256",
	"ECDHE-ECDSA-AES256-GCM-SHA384",
	"ECDHE-RSA-AES256-GCM-SHA384",
	"ECDHE-ECDSA-CHACHA20-POLY1305",
	"ECDHE-RSA-CHACHA20-POLY1305",
	"AES128-GCM-SHA256",
	"AES256-GCM-SHA384",
}

// 遗留客户端（Java 8 时代）的套件集：含 CBC，不含 TLS 1.3。
var legacySuites = []string{
	"ECDHE-RSA-AES128-GCM-SHA256",
	"ECDHE-ECDSA-AES128-GCM-SHA256",
	"ECDHE-RSA-AES256-GCM-SHA384",
	"ECDHE-ECDSA-AES256-GCM-SHA384",
	"ECDHE-RSA-AES128-SHA",
	"ECDHE-RSA-AES256-SHA",
	"ECDHE-ECDSA-AES128-SHA",
	"AES128-SHA",
	"AES256-SHA",
}

// 现代客户端的密钥交换组。
var modernGroups = []string{"X25519", "P-256", "P-384"}

// 遗留客户端的密钥交换组（不含 X25519）。
var legacyGroups = []string{"P-256", "P-384", "P-521"}

// ClientProfile 是一档客户端握手画像。
//
// 它描述的是「以这些参数发起握手」，**不是**真实 ClientHello 的字节回放：
// 扩展顺序、GREASE 与 TLS 指纹都不一致。结论措辞必须遵循这一点。
type ClientProfile struct {
	// Id 是稳定标识。
	Id string
	// Vendor 是厂商。
	Vendor string
	// Os 是操作系统。
	Os string
	// Engine 是 TLS 引擎。
	Engine string
	// Version 是画像版本。
	Version string
	// MinVersion / MaxVersion 是协议版本范围。
	MinVersion uint16
	MaxVersion uint16
	// CipherSuites 是套件白名单；为空表示用库默认。
	CipherSuites []string
	// Curves 是密钥交换组白名单；为空表示用库默认。
	Curves []string
	// NextProtos 是 ALPN 列表。
	NextProtos []string
}

// clientProfiles 是内置画像表（顺序稳定）。
//
// 套件与组名只写**本机支持**的名字：`knownSuites` / `knownGroups` 在每次
// 组装握手参数时过滤，名字不被识别时该维度降级为「不裁剪」，
// 避免整档画像直接失效。
func clientProfiles() []ClientProfile {
	return []ClientProfile{
		{
			Id: "chrome", Vendor: "Google", Os: "Windows", Engine: "BoringSSL", Version: "120+",
			MinVersion: tlssvc.VersionTLS12, MaxVersion: tlssvc.VersionTLS13,
			CipherSuites: append(append([]string{}, tls13Suites...), modernSuites...),
			Curves:       modernGroups,
			NextProtos:   alpnProtos,
		},
		{
			Id: "firefox", Vendor: "Mozilla", Os: "Windows", Engine: "NSS", Version: "121+",
			MinVersion: tlssvc.VersionTLS12, MaxVersion: tlssvc.VersionTLS13,
			CipherSuites: append(append([]string{}, tls13Suites...), modernSuites...),
			Curves:       modernGroups,
			NextProtos:   alpnProtos,
		},
		{
			Id: "safari", Vendor: "Apple", Os: "macOS", Engine: "SecureTransport", Version: "17+",
			MinVersion: tlssvc.VersionTLS12, MaxVersion: tlssvc.VersionTLS13,
			CipherSuites: append(append([]string{}, tls13Suites...), modernSuites...),
			Curves:       modernGroups,
			NextProtos:   alpnProtos,
		},
		{
			Id: "edge", Vendor: "Microsoft", Os: "Windows", Engine: "BoringSSL", Version: "120+",
			MinVersion: tlssvc.VersionTLS12, MaxVersion: tlssvc.VersionTLS13,
			CipherSuites: append(append([]string{}, tls13Suites...), modernSuites...),
			Curves:       modernGroups,
			NextProtos:   alpnProtos,
		},
		{
			Id: "android", Vendor: "Google", Os: "Android 10+", Engine: "BoringSSL", Version: "10+",
			MinVersion: tlssvc.VersionTLS12, MaxVersion: tlssvc.VersionTLS13,
			CipherSuites: append(append([]string{}, tls13Suites...), modernSuites...),
			Curves:       modernGroups,
			NextProtos:   alpnProtos,
		},
		{
			Id: "ios", Vendor: "Apple", Os: "iOS 16+", Engine: "Network.framework", Version: "16+",
			MinVersion: tlssvc.VersionTLS12, MaxVersion: tlssvc.VersionTLS13,
			CipherSuites: append(append([]string{}, tls13Suites...), modernSuites...),
			Curves:       modernGroups,
			NextProtos:   alpnProtos,
		},
		{
			Id: "java11", Vendor: "Oracle", Os: "跨平台", Engine: "JSSE", Version: "11",
			MinVersion: tlssvc.VersionTLS12, MaxVersion: tlssvc.VersionTLS13,
			CipherSuites: append(append([]string{}, tls13Suites...), modernSuites...),
			Curves:       legacyGroups,
		},
		{
			Id: "java8", Vendor: "Oracle", Os: "跨平台", Engine: "JSSE", Version: "8",
			MinVersion: tlssvc.VersionTLS12, MaxVersion: tlssvc.VersionTLS12,
			CipherSuites: legacySuites,
			Curves:       legacyGroups,
		},
		{
			Id: "openssl111", Vendor: "OpenSSL", Os: "跨平台", Engine: "OpenSSL", Version: "1.1.1",
			MinVersion: tlssvc.VersionTLS10, MaxVersion: tlssvc.VersionTLS13,
			CipherSuites: append(append([]string{}, tls13Suites...), legacySuites...),
			Curves:       legacyGroups,
		},
		{
			Id: "openssl3", Vendor: "OpenSSL", Os: "跨平台", Engine: "OpenSSL", Version: "3.x",
			MinVersion: tlssvc.VersionTLS12, MaxVersion: tlssvc.VersionTLS13,
			CipherSuites: append(append([]string{}, tls13Suites...), modernSuites...),
			Curves:       modernGroups,
		},
		{
			Id: "schannel", Vendor: "Microsoft", Os: "Windows", Engine: "Schannel", Version: "TLS 1.2",
			MinVersion: tlssvc.VersionTLS12, MaxVersion: tlssvc.VersionTLS12,
			CipherSuites: legacySuites,
			Curves:       legacyGroups,
		},
		{
			Id: "go", Vendor: "Go", Os: "跨平台", Engine: "crypto/tls", Version: "1.22+",
			MinVersion: tlssvc.VersionTLS12, MaxVersion: tlssvc.VersionTLS13,
		},
	}
}

// clientProfileIndex 按 id 索引画像。
func clientProfileIndex() map[string]ClientProfile {
	profiles := clientProfiles()
	out := make(map[string]ClientProfile, len(profiles))
	for _, profile := range profiles {
		out[profile.Id] = profile
	}
	return out
}

// ClientProfileIDs 返回全部画像 id（字典序，用于错误提示）。
func ClientProfileIDs() []string {
	profiles := clientProfiles()
	out := make([]string, 0, len(profiles))
	for _, profile := range profiles {
		out = append(out, profile.Id)
	}
	sort.Strings(out)
	return out
}

// scanClients 逐档模拟客户端握手。
func (s *scanner) scanClients(wanted []string) []ClientResult {
	profiles := clientProfiles()
	if len(wanted) > 0 {
		index := clientProfileIndex()
		selected := make([]ClientProfile, 0, len(wanted))
		for _, id := range wanted {
			if profile, ok := index[id]; ok {
				selected = append(selected, profile)
			}
		}
		profiles = selected
	}

	results := make([]ClientResult, len(profiles))
	runParallel(len(profiles), func(index int) {
		profile := profiles[index]
		results[index] = s.probeClient(profile)
		s.reportItem(ScanPhaseClients, profile.Id)
	})
	return results
}

// probeClient 模拟一档客户端。
func (s *scanner) probeClient(profile ClientProfile) ClientResult {
	result := ClientResult{
		Id:             profile.Id,
		Vendor:         profile.Vendor,
		Os:             profile.Os,
		Engine:         profile.Engine,
		ProfileVersion: profile.Version,
	}

	state, err := s.probe(s.clientOptions(profile, true, true, true))
	if err == nil {
		result.Supported = true
		result.NegotiatedVersion = state.VersionName
		result.Cipher = state.CipherName
		result.Alpn = state.ALPN
		return result
	}
	result.Reason = s.attributeClientFailure(profile)
	return result
}

// clientOptions 按画像组装握手参数。
//
// 三个开关分别控制是否裁剪套件、组与 ALPN：
// 归因过程要逐项放宽，才能定位到底是哪一维不匹配。
func (s *scanner) clientOptions(profile ClientProfile, ciphers, curves, alpn bool) tlssvc.Options {
	opts := tlssvc.Options{
		ServerName:       s.sni,
		MinVersion:       profile.MinVersion,
		MaxVersion:       profile.MaxVersion,
		HandshakeTimeout: probeTimeout,
	}
	if ciphers {
		opts.CipherSuites = knownSuites(profile.CipherSuites)
	}
	if curves {
		opts.Curves = knownGroups(profile.Curves)
	}
	if alpn {
		opts.NextProtos = profile.NextProtos
	}
	return opts
}

// attributeClientFailure 定位一档客户端握手失败的原因。
//
// 归因靠**逐步放宽**的对照握手，而不是匹配错误文本：
//
//  1. 只保留版本范围 → 失败说明版本范围本身不可用；
//  2. 版本范围 + 画像套件 → 失败说明套件无交集；
//  3. 版本范围 + 画像组 → 失败说明组无交集；
//  4. 以上都通过 → 归为握手被拒绝。
func (s *scanner) attributeClientFailure(profile ClientProfile) string {
	if _, err := s.probe(s.clientOptions(profile, false, false, false)); err != nil {
		return "noCommonVersion"
	}
	if _, err := s.probe(s.clientOptions(profile, true, false, false)); err != nil {
		return "noCommonCipher"
	}
	if _, err := s.probe(s.clientOptions(profile, false, true, false)); err != nil {
		return "noCommonGroup"
	}
	return "handshakeFailure"
}

// knownSuites 过滤掉本机不认识的套件名。
func knownSuites(names []string) []string {
	out := make([]string, 0, len(names))
	for _, name := range names {
		if tlssvc.HasSuite(name) {
			out = append(out, name)
		}
	}
	return out
}

// knownGroups 过滤掉本机不支持的组名。
func knownGroups(names []string) []string {
	out := make([]string, 0, len(names))
	for _, name := range names {
		if tlssvc.HasGroup(name) {
			out = append(out, name)
		}
	}
	return out
}

// reasonOf 把探测失败归类为稳定的归因标识。
//
// 只有上游提供了明确的错误种类时才细分；其余一律归到 `handshakeFailure`，
// 具体原因放在详情里，**不**用文本匹配做分支判断。
func reasonOf(err error) string {
	if err == nil {
		return ""
	}

	var netErr *netguard.Error
	if errors.As(err, &netErr) {
		return "network"
	}

	var hsErr *tlssvc.HandshakeError
	if errors.As(err, &hsErr) {
		switch hsErr.Reason {
		case tlssvc.ReasonNoCommonVersion:
			return "noCommonVersion"
		case tlssvc.ReasonNoCommonCipher:
			return "noCommonCipher"
		case tlssvc.ReasonNetwork:
			return "network"
		case tlssvc.ReasonTimeout:
			return "timeout"
		default:
			return "handshakeFailure"
		}
	}
	return "handshakeFailure"
}
