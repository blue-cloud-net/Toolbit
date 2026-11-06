package network

import (
	"regexp"
	"strings"
	"sync"

	"github.com/blue-cloud-net/Toolbit/backend/internal/tlssvc"
)

// cipherBitsPattern 匹配算法名后的密钥位数。
//
// 只允许在算法名后紧跟 `_` / `-` 或无分隔符的位数，避免把 SHA256 的
// 256 误当成分组长度。
var cipherBitsPattern = regexp.MustCompile(`(?:AES|ARIA|CAMELLIA|SM4)[_-]?(128|192|256)`)


// scanProtocols 逐个协议版本探测。
//
// 每个版本单独建连、单独握手，`MinVersion == MaxVersion` 强制锁定版本；
// 这样「服务端是否还接受 TLS 1.1」才是一个可以回答的问题。
func (s *scanner) scanProtocols() []ProtocolResult {
	versions := tlssvc.Versions()
	out := make([]ProtocolResult, 0, len(versions))
	for _, version := range versions {
		out = append(out, s.probeProtocol(version))
	}
	return out
}

// probeProtocol 探测单个协议版本。
//
// 结论靠**对照**得出：同一目标先用一个宽版本范围的基准握手确认「目标可达」，
// 再逐版本锁定。若基准握手本身就失败，则所有版本都标为「不可判定」而不是
// 「不支持」——这两者在页面上必须区分。
func (s *scanner) probeProtocol(version tlssvc.Version) ProtocolResult {
	result := ProtocolResult{Version: version.Name, VersionWire: version.Wire}
	state, err := s.probe(tlssvc.Options{
		ServerName:       s.sni,
		MinVersion:       version.Wire,
		MaxVersion:       version.Wire,
		NextProtos:       alpnProtos,
		HandshakeTimeout: probeTimeout,
	})
	if err != nil {
		result.Reason = reasonOf(err)
		return result
	}
	result.Supported = true
	result.NegotiatedCipher = state.CipherName
	return result
}

// scanCiphers 对每个受支持的协议版本逐套件探测。
//
// 只在**已确认支持**的版本上做，避免把「版本本身就不支持」误读成「所有套件都不支持」。
func (s *scanner) scanCiphers(protocols []ProtocolResult) []CipherResult {
	versions := tlssvc.Versions()
	out := make([]CipherResult, 0, maxCiphersPerVersion)
	for i, version := range versions {
		supported := false
		for _, item := range protocols {
			if item.VersionWire == version.Wire && item.Supported {
				supported = true
				break
			}
		}
		if !supported {
			continue
		}
		suites := selectSuites(version, i)
		results := make([]CipherResult, len(suites))
		runParallel(len(suites), func(index int) {
			suite := suites[index]
			results[index] = s.probeCipher(version, suite)
			s.reportItem(ScanPhaseCiphers, suite.Name)
		})
		out = append(out, results...)
	}
	return out
}

// probeCipher 探测单个套件。
//
// 该版本已在协议阶段用**默认套件**握手成功，因此这里唯一的变量就是套件：
// 单套件握手失败即意味着「服务端不接受这个套件」。
func (s *scanner) probeCipher(version tlssvc.Version, suite tlssvc.Suite) CipherResult {
	family, bits, kex := describeSuite(suite.Name)
	result := CipherResult{
		Version: version.Name,
		Name:    suite.Name,
		Id:      suite.ID,
		Family:  family,
		Bits:    bits,
		Kex:     kex,
		Verdict: SuiteRejected,
	}
	_, err := s.probe(tlssvc.Options{
		ServerName:       s.sni,
		MinVersion:       version.Wire,
		MaxVersion:       version.Wire,
		CipherSuites:     []string{suite.Name},
		HandshakeTimeout: probeTimeout,
	})
	if err == nil {
		result.Verdict = SuiteSupported
	}
	return result
}

// alpnProtos 是探测时提供的 ALPN 列表。
var alpnProtos = []string{"h2", "http/1.1"}

// selectSuites 从本机可用的套件里挑出值得探测的那些。
//
// 规则：
//
//   - 过滤掉 PSK / SRP / 匿名 / Kerberos / 空套件 / 信令套件 —— 它们不反映
//     站点对浏览器的兼容性；
//   - 每个版本最多保留 `maxCiphersPerVersion` 个，按强度由高到低截断；
//   - TLS 1.3 的套件数量少，通常全部保留。
func selectSuites(version tlssvc.Version, versionIndex int) []tlssvc.Suite {
	all := tlssvc.EnumerateSuites(version.Wire)
	kept := make([]tlssvc.Suite, 0, len(all))
	for _, suite := range all {
		if skipSuite(suite.Name) {
			continue
		}
		kept = append(kept, suite)
	}
	if len(kept) > maxCiphersPerVersion {
		kept = kept[:maxCiphersPerVersion]
	}
	_ = versionIndex
	return kept
}

// skipSuite 判断套件是否属于「与兼容性无关」的一类。
//
// 排除的是这些：PSK / SRP / 匿名（ADH / AECDH）/ DSS 签名 / Kerberos /
// 签名套件（SCSV）/ 空套件 / EXPORT / RC4 / NULL 加密 / SEED / IDEA。
// 它们要么与浏览器兼容性无关，要么本身就是弱套件 —— 弱套件的存在需要被
// **看见**，因此单独列在评级里，而不是从列表里抹掉。
func skipSuite(name string) bool {
	upper := strings.ToUpper(name)
	for _, fragment := range []string{
		"PSK", "SRP", "ANON", "ADH", "AECDH", "DSS", "KRB5", "SCSV",
		"RENEGOTIATION", "EMPTY", "EXPORT", "RC4", "NULL", "SEED", "IDEA",
	} {
		if strings.Contains(upper, fragment) {
			return true
		}
	}
	return false
}

// describeSuite 从套件名推断算法族、强度位数与密钥交换方式。
//
// 这里**不做密码学运算**，只是对名称做结构解析，用于分组展示与评分。
func describeSuite(name string) (family string, bits int, kex string) {
	upper := strings.ToUpper(name)

	switch {
	case strings.Contains(upper, "CHACHA20"):
		family = "AEAD"
		bits = 256
	case strings.Contains(upper, "GCM"), strings.Contains(upper, "CCM"):
		family = "AEAD"
	case strings.Contains(upper, "3DES"), strings.Contains(upper, "DES-CBC3"),
		strings.Contains(upper, "DES-EDE"):
		family = "3DES"
		bits = 112
	case strings.Contains(upper, "RC4"):
		family = "RC4"
		bits = 128
	case strings.Contains(upper, "AES"), strings.Contains(upper, "ARIA"),
		strings.Contains(upper, "CAMELLIA"), strings.Contains(upper, "SM4"):
		// TLS 里没有 GCM / CCM 的对称套件只剩 CBC。
		family = "CBC"
	default:
		family = "其它"
	}

	if bits == 0 {
		if match := cipherBitsPattern.FindStringSubmatch(upper); len(match) == 2 {
			switch match[1] {
			case "128":
				bits = 128
			case "192":
				bits = 192
			case "256":
				bits = 256
			}
		}
	}

	switch {
	case strings.HasPrefix(upper, "TLS_"):
		// TLS 1.3 只允许 (EC)DHE，套件名里不再标注密钥交换。
		kex = "ECDHE"
	case strings.HasPrefix(upper, "ECDHE-"):
		kex = "ECDHE"
	case strings.HasPrefix(upper, "DHE-"), strings.HasPrefix(upper, "EDH-"):
		kex = "DHE"
	case strings.HasPrefix(upper, "RSA-"):
		kex = "RSA"
	default:
		kex = "其它"
	}
	return family, bits, kex
}

// runParallel 以固定并发度执行 n 个任务，index 从 0 到 n-1。
//
// 并发上限固定为 `scanWorkers`：既缩短整体耗时，也不至于对一个站点
// 同时开过多连接。
func runParallel(n int, task func(index int)) {
	if n <= 0 {
		return
	}
	workers := scanWorkers
	if n < workers {
		workers = n
	}
	var wg sync.WaitGroup
	next := make(chan int)
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for index := range next {
				task(index)
			}
		}()
	}
	for i := 0; i < n; i++ {
		next <- i
	}
	close(next)
	wg.Wait()
}
