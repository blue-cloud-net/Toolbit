package toolerr

// WarningCode 是警告码。
//
// 警告**不改变** HTTP 状态，也**不**影响 data 的可用性；
// 它们出现在响应的 meta.warnings 中，用于提示用户注意风险。
type WarningCode string

// 全部警告码。取值与 docs/api/errors.md §6 一一对应。
const (
	// WarnInsecureECB 使用了 ECB 模式（无扩散）。
	WarnInsecureECB WarningCode = "INSECURE_ECB"
	// WarnNoAuthentication 对称加密未提供认证标签。
	WarnNoAuthentication WarningCode = "NO_AUTHENTICATION"
	// WarnSmallRSAKey RSA 密钥长度小于 2048。
	WarnSmallRSAKey WarningCode = "SMALL_RSA_KEY"
	// WarnWeakHash 使用了 MD5 / SHA-1。
	WarnWeakHash WarningCode = "WEAK_HASH"
	// WarnExpiredCert 证书已过期。
	WarnExpiredCert WarningCode = "EXPIRED_CERT"
	// WarnNotYetValid 证书尚未生效。
	WarnNotYetValid WarningCode = "NOT_YET_VALID"
	// WarnSelfSignedCert 自签名证书。
	WarnSelfSignedCert WarningCode = "SELF_SIGNED_CERT"
	// WarnUntrustedChain 证书链无法追溯到受信任根。
	WarnUntrustedChain WarningCode = "UNTRUSTED_CHAIN"
	// WarnPaddingOracleRisk 使用了 PKCS#1 v1.5 填充。
	WarnPaddingOracleRisk WarningCode = "PADDING_ORACLE_RISK"
	// WarnLargeInput 输入超过常见阈值，响应可能很大。
	WarnLargeInput WarningCode = "LARGE_INPUT"
	// WarnLossyFormat 格式转换有损（信息丢失）。
	WarnLossyFormat WarningCode = "LOSSY_FORMAT"
	// WarnRegexDialectDifference 正则方言与目标语言存在差异。
	WarnRegexDialectDifference WarningCode = "REGEX_DIALECT_DIFFERENCE"
	// WarnTimezoneAssumed 时间转换假定了某个时区。
	WarnTimezoneAssumed WarningCode = "TIMEZONE_ASSUMED"
	// WarnCommentsLost 注释被丢弃。
	WarnCommentsLost WarningCode = "COMMENTS_LOST"
)

// WarningSpec 描述一个警告码的语义与本地化 key。
type WarningSpec struct {
	// Message 是英文兜底文案，**不得**包含用户输入。
	Message string
	// I18nKey 是前端本地化词条的 key，形式为 errors.<CODE>。
	I18nKey string
}

// Warnings 是警告码的唯一事实来源。
//
// 与 Registry 分开：警告没有 HTTP 状态，混在一起会让 HTTP 映射失去意义。
var Warnings = map[WarningCode]WarningSpec{
	"INSECURE_ECB":             {Message: "ECB mode provides no diffusion", I18nKey: "errors.INSECURE_ECB"},
	"NO_AUTHENTICATION":        {Message: "no authentication tag provided", I18nKey: "errors.NO_AUTHENTICATION"},
	"SMALL_RSA_KEY":            {Message: "RSA key size below 2048 bits", I18nKey: "errors.SMALL_RSA_KEY"},
	"WEAK_HASH":                {Message: "weak hash algorithm", I18nKey: "errors.WEAK_HASH"},
	"EXPIRED_CERT":             {Message: "certificate has expired", I18nKey: "errors.EXPIRED_CERT"},
	"NOT_YET_VALID":            {Message: "certificate is not yet valid", I18nKey: "errors.NOT_YET_VALID"},
	"SELF_SIGNED_CERT":         {Message: "self-signed certificate", I18nKey: "errors.SELF_SIGNED_CERT"},
	"UNTRUSTED_CHAIN":          {Message: "certificate chain is not trusted", I18nKey: "errors.UNTRUSTED_CHAIN"},
	"PADDING_ORACLE_RISK":      {Message: "PKCS#1 v1.5 padding is oracle-prone", I18nKey: "errors.PADDING_ORACLE_RISK"},
	"LARGE_INPUT":              {Message: "input is large, response may be big", I18nKey: "errors.LARGE_INPUT"},
	"LOSSY_FORMAT":             {Message: "conversion is lossy", I18nKey: "errors.LOSSY_FORMAT"},
	"REGEX_DIALECT_DIFFERENCE": {Message: "regex dialect differs from target language", I18nKey: "errors.REGEX_DIALECT_DIFFERENCE"},
	"TIMEZONE_ASSUMED":         {Message: "a timezone was assumed", I18nKey: "errors.TIMEZONE_ASSUMED"},
	"COMMENTS_LOST":            {Message: "comments were dropped", I18nKey: "errors.COMMENTS_LOST"},
}

// Warning 是一个具体的警告实例。
type Warning struct {
	// Code 是警告码。
	Code WarningCode `json:"code"`
	// Params 是可插值参数，只放结构化元信息。
	Params map[string]any `json:"params,omitempty"`
}

// NewWarning 构造一个警告。
func NewWarning(code WarningCode, params map[string]any) Warning {
	return Warning{Code: code, Params: params}
}

// WithParam 追加一个可插值参数并返回自身。
func (w Warning) WithParam(key string, value any) Warning {
	if w.Params == nil {
		w.Params = make(map[string]any, 1)
	}
	w.Params[key] = value
	return w
}

// Message 返回警告的英文兜底文案。
func (w Warning) Message() string {
	spec, ok := Warnings[w.Code]
	if !ok {
		return "unknown warning"
	}
	return spec.Message
}

// I18nKey 返回警告的本地化 key。
func (w Warning) I18nKey() string {
	spec, ok := Warnings[w.Code]
	if !ok {
		return "errors.INTERNAL_ERROR"
	}
	return spec.I18nKey
}
