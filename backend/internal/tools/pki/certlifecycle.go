package pki

import (
	"context"
	"math"
	"strconv"
	"time"

	"github.com/blue-cloud-net/Toolbit/backend/internal/pkisvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// CertLifecycleRequest 是 cert-lifecycle 工具的请求。
type CertLifecycleRequest struct {
	// Certificate 是证书内容（PEM 或 DER）。
	Certificate string `json:"certificate"`
	// IsPEM 是否是 PEM 文本；默认 true。
	IsPEM bool `json:"isPem"`
	// WarningDays 是进入警告期的剩余天数阈值；默认 30。
	WarningDays int `json:"warningDays"`
}

// CertLifecycleResult 是 cert-lifecycle 工具的响应。
type CertLifecycleResult struct {
	// Subject 是主体单行文本。
	Subject string `json:"subject"`
	// Issuer 是签发者单行文本。
	Issuer string `json:"issuer"`
	// NotBefore 是生效时间。
	NotBefore string `json:"notBefore"`
	// NotAfter 是过期时间。
	NotAfter string `json:"notAfter"`
	// ValidDays 是有效天数。
	ValidDays float64 `json:"validDays"`
	// DaysRemaining 是距离过期的剩余天数；已过期为负数。
	//
	// 它是**小数**（保留 1 位）：临期证书需要区分「还剩 9.6 天」与「还剩 9 天」。
	DaysRemaining float64 `json:"daysRemaining"`
	// HoursRemaining 是距离过期的剩余小时数；已过期为负数。
	HoursRemaining int `json:"hoursRemaining"`
	// Status 是状态：valid / expiring / expired / notYetValid。
	Status string `json:"status"`
	// IsSelfSigned 是否是自签名证书。
	IsSelfSigned bool `json:"isSelfSigned"`
	// IsCA 是否是 CA 证书。
	IsCA bool `json:"isCa"`
	// KeyAlgorithm 是公钥算法。
	KeyAlgorithm string `json:"keyAlgorithm"`
	// KeyBits 是 RSA 密钥位数；非 RSA 为 0。
	KeyBits int `json:"keyBits"`
	// Curve 是椭圆曲线名；非椭圆曲线为空。
	Curve string `json:"curve"`
	// SignatureAlgorithm 是签名算法。
	SignatureAlgorithm string `json:"signatureAlgorithm"`
	// SAN 是主体备用名列表。
	SAN []string `json:"san"`
	// Recommendations 是续期建议。
	Recommendations []string `json:"recommendations"`
}

// 证书生命周期状态。
const (
	// CertStatusValid 处于有效期内。
	CertStatusValid = "valid"
	// CertStatusExpiring 即将过期。
	CertStatusExpiring = "expiring"
	// CertStatusExpired 已过期。
	CertStatusExpired = "expired"
	// CertStatusNotYetValid 尚未生效。
	CertStatusNotYetValid = "notYetValid"
)

// defaultCertWarningDays 是默认的警告期天数。
const defaultCertWarningDays = 30

// CertLifecycle 计算证书的有效期区间、到期倒计时与续期建议。
//
// 证书解析走 pkisvc（**不**直接用标准库 `crypto/x509`），
// 保证与仓库其他证书工具的口径一致。
func CertLifecycle(ctx context.Context, req *CertLifecycleRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	source := []byte(req.Certificate)
	cert, err := pkisvc.LoadCertificate(source, req.IsPEM)
	if err != nil {
		return tools.Result{}, err
	}
	defer func() {
		_ = cert.Close()
	}()

	now := time.Now()
	info, err := pkisvc.Describe(cert, now, nil)
	if err != nil {
		return tools.Result{}, err
	}

	validity := pkisvc.CheckValidity(cert, now)

	result := CertLifecycleResult{
		Subject:            info.SubjectText,
		Issuer:             info.IssuerText,
		NotBefore:          info.NotBefore.UTC().Format(time.RFC3339),
		NotAfter:           info.NotAfter.UTC().Format(time.RFC3339),
		ValidDays:          pkisvc.RoundDays(info.NotAfter.Sub(info.NotBefore)),
		DaysRemaining:      validity.DaysRemaining,
		HoursRemaining:     int(info.NotAfter.Sub(now).Hours()),
		Status:             CertStatusValid,
		IsSelfSigned:       info.SelfSigned,
		IsCA:               info.IsCA,
		KeyAlgorithm:       info.PublicKey.Algorithm,
		KeyBits:            info.PublicKey.Bits,
		Curve:              info.PublicKey.Curve,
		SignatureAlgorithm: info.SignatureAlgorithm,
		SAN:                info.SAN,
		Recommendations:    []string{},
	}

	switch {
	case validity.Status == pkisvc.ValidityNotYetValid:
		result.Status = CertStatusNotYetValid
		result.Recommendations = append(result.Recommendations,
			"证书尚未生效，请检查系统时间或证书的生效时间设置")
	case validity.Status == pkisvc.ValidityExpired:
		result.Status = CertStatusExpired
		result.Recommendations = append(result.Recommendations,
			"证书已过期，需要立即续期并替换")
	case validity.DaysRemaining <= float64(req.WarningDays):
		result.Status = CertStatusExpiring
		result.Recommendations = append(result.Recommendations,
			"证书将在 "+formatDays(validity.DaysRemaining)+" 天内过期，请尽快安排续期")
	default:
		result.Recommendations = append(result.Recommendations,
			"证书在有效期内，剩余 "+formatDays(validity.DaysRemaining)+" 天")
	}

	if result.IsSelfSigned {
		result.Recommendations = append(result.Recommendations,
			"自签名证书需要客户端单独信任该 CA，不适合公开环境")
	}
	if len(result.SAN) == 0 && !result.IsCA {
		result.Recommendations = append(result.Recommendations,
			"证书没有主体备用名（SAN），现代客户端会要求 SAN 中包含访问域名")
	}
	if result.ValidDays > 0 && result.ValidDays > 365*3.0 {
		result.Recommendations = append(result.Recommendations,
			"有效期超过三年，建议确认是否符合机构的证书策略要求")
	}

	return tools.Result{Data: result}, nil
}

// formatDays 把天数渲染成字符串。
//
// 口径与前端 `certDetail.ts` 的 `formatDays` **必须一致**：
// 不足 10 天时保留 1 位小数（「9.6 天」），否则**截断**（「820.7」→「820」）。
//
// 用截断而非四舍五入：820.7 天说成「821 天」会显得比实际更紧迫，
// 而前端 `Math.trunc` 也会截断 —— 两边必须一样。
func formatDays(value float64) string {
	if math.Abs(value) < 10 {
		return strconv.FormatFloat(value, 'f', 1, 64)
	}
	return strconv.FormatFloat(math.Trunc(value), 'f', 0, 64)
}

// Validate 做字段级校验。
func (r *CertLifecycleRequest) Validate() error {
	if r.Certificate == "" {
		return tools.MissingField("certificate")
	}
	if r.WarningDays < 0 || r.WarningDays > 3650 {
		return tools.InvalidField("warningDays", []int{0, 3650}, r.WarningDays)
	}
	if r.WarningDays == 0 {
		r.WarningDays = defaultCertWarningDays
	}
	return nil
}
