// Package toolerr 定义 Toolbit 的错误码注册表与统一错误类型。
//
// 错误码的**唯一事实来源**是本包的 Registry；docs/api/errors.md 是它的镜像文档，
// 由 scripts/check-docs.mjs 校验覆盖率。
//
// 设计要点：
//
//   - 错误码全大写、下划线分隔，全局唯一；
//   - HTTP 状态表达「请求能否被处理」，错误码表达「具体原因」；
//   - 错误文案是英文兜底文案，**不得**包含用户输入；
//   - 可插值参数只放结构化元信息（字段名、期望长度、索引、上限）。
package toolerr

// Code 是机器可读的错误码。
//
// 常量名统一使用 Code 前缀，值使用全大写下划线形式。
type Code string

// 请求与格式类错误（Invalid *）。
const (
	CodeInvalidRequest  Code = "INVALID_REQUEST"
	CodeMissingField    Code = "MISSING_FIELD"
	CodeInvalidField    Code = "INVALID_FIELD"
	CodeInvalidEncoding Code = "INVALID_ENCODING"
	CodeInvalidHex      Code = "INVALID_HEX"
	CodeInvalidBase64   Code = "INVALID_BASE64"
	CodeInvalidPEM      Code = "INVALID_PEM"
	CodeInvalidDER      Code = "INVALID_DER"
	CodeInvalidKey      Code = "INVALID_KEY"
	CodeInvalidCert     Code = "INVALID_CERT"
	CodeInvalidCSR      Code = "INVALID_CSR"
	CodeInvalidCRL      Code = "INVALID_CRL"
	CodeInvalidPKCS7    Code = "INVALID_PKCS7"
	CodeInvalidPKCS12   Code = "INVALID_PKCS12"
	CodeInvalidJWK      Code = "INVALID_JWK"
	CodeInvalidJSON     Code = "INVALID_JSON"
	CodeInvalidYAML     Code = "INVALID_YAML"
	CodeInvalidTOML     Code = "INVALID_TOML"
	CodeInvalidXML      Code = "INVALID_XML"
	CodeInvalidJWT      Code = "INVALID_JWT"
	CodeInvalidCIDR     Code = "INVALID_CIDR"
	CodeInvalidRegex    Code = "INVALID_REGEX"
	CodeInvalidCron     Code = "INVALID_CRON"
	CodeInvalidChmod    Code = "INVALID_CHMOD"
	CodeInvalidTime     Code = "INVALID_TIME"
	CodeInvalidDomain   Code = "INVALID_DOMAIN"
)

// 算法与参数类错误（Unsupported * / Key* / Passphrase* 等）。
const (
	CodeUnsupportedAlgorithm Code = "UNSUPPORTED_ALGORITHM"
	CodeUnsupportedMode      Code = "UNSUPPORTED_MODE"
	CodeUnsupportedPadding   Code = "UNSUPPORTED_PADDING"
	CodeUnsupportedFormat    Code = "UNSUPPORTED_FORMAT"
	CodeUnsupportedVersion   Code = "UNSUPPORTED_VERSION"
	CodeKeySizeMismatch      Code = "KEY_SIZE_MISMATCH"
	CodeIVSizeMismatch       Code = "IV_SIZE_MISMATCH"
	CodePassphraseRequired   Code = "PASSPHRASE_REQUIRED"
	CodePassphraseIncorrect  Code = "PASSPHRASE_INCORRECT"
	CodeTagRequired          Code = "TAG_REQUIRED"
	CodeTagInvalid           Code = "TAG_INVALID"
	CodeDecryptFailed        Code = "DECRYPT_FAILED"
	CodeSignatureInvalid     Code = "SIGNATURE_INVALID"
	CodePlaintextTooLong     Code = "PLAINTEXT_TOO_LONG"
	CodeKeyMismatch          Code = "KEY_MISMATCH"
	CodeSANInvalid           Code = "SAN_INVALID"
	CodeSerialInvalid        Code = "SERIAL_INVALID"
	CodeNotAfterBeforeNow    Code = "NOT_AFTER_BEFORE_NOW"
)

// 限制与传输类错误。
const (
	CodeTooLarge             Code = "TOO_LARGE"
	CodeTooManyItems         Code = "TOO_MANY_ITEMS"
	CodeTimeout              Code = "TIMEOUT"
	CodeRateLimited          Code = "RATE_LIMITED"
	CodeToolNotFound         Code = "TOOL_NOT_FOUND"
	CodeToolDisabled         Code = "TOOL_DISABLED"
	CodeMethodNotAllowed     Code = "METHOD_NOT_ALLOWED"
	CodeUnsupportedMediaType Code = "UNSUPPORTED_MEDIA_TYPE"
	CodeCrossOriginBlocked   Code = "CROSS_ORIGIN_BLOCKED"
	CodeNetworkDisabled      Code = "NETWORK_DISABLED"
	CodeSSRFBlocked          Code = "SSRF_BLOCKED"
	CodeDNSError             Code = "DNS_ERROR"
	CodeTLSError             Code = "TLS_ERROR"
	CodeHostUnreachable      Code = "HOST_UNREACHABLE"
	CodeUpstreamTimeout      Code = "UPSTREAM_TIMEOUT"
	CodeUpstreamError        Code = "UPSTREAM_ERROR"
	CodeRedirectLimit        Code = "REDIRECT_LIMIT"
)

// 服务端类错误。
const (
	CodeInternalError         Code = "INTERNAL_ERROR"
	CodeCryptoOperationFailed Code = "CRYPTO_OPERATION_FAILED"
	CodeCryptoInitFailed      Code = "CRYPTO_INIT_FAILED"
	CodeStorageError          Code = "STORAGE_ERROR"
	CodeNotImplemented        Code = "NOT_IMPLEMENTED"
	CodeStatsUnauthorized     Code = "STATS_UNAUTHORIZED"
)
