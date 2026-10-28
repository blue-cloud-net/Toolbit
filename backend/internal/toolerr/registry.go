package toolerr

import "net/http"

// Spec 描述一个错误码的传输层与语义层属性。
type Spec struct {
	// HTTP 是该错误码对应的 HTTP 状态码。
	HTTP int
	// Message 是英文兜底文案，**不得**包含用户输入。
	Message string
	// I18nKey 是前端本地化词条的 key，形式为 errors.<CODE>。
	I18nKey string
}

// Registry 是错误码的唯一事实来源。
//
// 键一律写成带引号的字符串字面量：scripts/check-docs.mjs 从本文件解析错误码，
// 若改用「常量名 + ErrorCode 类型」的形式，常量名会被误当成错误码。
var Registry = map[Code]Spec{
	// 请求与格式。
	"INVALID_REQUEST":  {HTTP: http.StatusBadRequest, Message: "invalid request", I18nKey: "errors.INVALID_REQUEST"},
	"MISSING_FIELD":    {HTTP: http.StatusBadRequest, Message: "missing required field", I18nKey: "errors.MISSING_FIELD"},
	"INVALID_FIELD":    {HTTP: http.StatusBadRequest, Message: "invalid field value", I18nKey: "errors.INVALID_FIELD"},
	"INVALID_ENCODING": {HTTP: http.StatusBadRequest, Message: "unsupported encoding", I18nKey: "errors.INVALID_ENCODING"},
	"INVALID_HEX":      {HTTP: http.StatusBadRequest, Message: "invalid hex string", I18nKey: "errors.INVALID_HEX"},
	"INVALID_BASE64":   {HTTP: http.StatusBadRequest, Message: "invalid base64 string", I18nKey: "errors.INVALID_BASE64"},
	"INVALID_PEM":      {HTTP: http.StatusBadRequest, Message: "invalid PEM data", I18nKey: "errors.INVALID_PEM"},
	"INVALID_DER":      {HTTP: http.StatusBadRequest, Message: "invalid DER data", I18nKey: "errors.INVALID_DER"},
	"INVALID_KEY":      {HTTP: http.StatusBadRequest, Message: "invalid key", I18nKey: "errors.INVALID_KEY"},
	"INVALID_CERT":     {HTTP: http.StatusBadRequest, Message: "invalid certificate", I18nKey: "errors.INVALID_CERT"},
	"INVALID_CSR":      {HTTP: http.StatusBadRequest, Message: "invalid certificate request", I18nKey: "errors.INVALID_CSR"},
	"INVALID_CRL":      {HTTP: http.StatusBadRequest, Message: "invalid CRL", I18nKey: "errors.INVALID_CRL"},
	"INVALID_PKCS7":    {HTTP: http.StatusBadRequest, Message: "invalid PKCS#7 data", I18nKey: "errors.INVALID_PKCS7"},
	"INVALID_PKCS12":   {HTTP: http.StatusBadRequest, Message: "invalid PKCS#12 data", I18nKey: "errors.INVALID_PKCS12"},
	"INVALID_JWK":      {HTTP: http.StatusBadRequest, Message: "invalid JWK", I18nKey: "errors.INVALID_JWK"},
	"INVALID_JSON":     {HTTP: http.StatusBadRequest, Message: "invalid JSON", I18nKey: "errors.INVALID_JSON"},
	"INVALID_YAML":     {HTTP: http.StatusBadRequest, Message: "invalid YAML", I18nKey: "errors.INVALID_YAML"},
	"INVALID_TOML":     {HTTP: http.StatusBadRequest, Message: "invalid TOML", I18nKey: "errors.INVALID_TOML"},
	"INVALID_XML":      {HTTP: http.StatusBadRequest, Message: "invalid XML", I18nKey: "errors.INVALID_XML"},
	"INVALID_JWT":      {HTTP: http.StatusBadRequest, Message: "invalid JWT", I18nKey: "errors.INVALID_JWT"},
	"INVALID_CIDR":     {HTTP: http.StatusBadRequest, Message: "invalid CIDR", I18nKey: "errors.INVALID_CIDR"},
	"INVALID_REGEX":    {HTTP: http.StatusBadRequest, Message: "invalid regular expression", I18nKey: "errors.INVALID_REGEX"},
	"INVALID_CRON":     {HTTP: http.StatusBadRequest, Message: "invalid crontab expression", I18nKey: "errors.INVALID_CRON"},
	"INVALID_CHMOD":    {HTTP: http.StatusBadRequest, Message: "invalid permission value", I18nKey: "errors.INVALID_CHMOD"},
	"INVALID_TIME":     {HTTP: http.StatusBadRequest, Message: "invalid time value", I18nKey: "errors.INVALID_TIME"},
	"INVALID_DOMAIN":   {HTTP: http.StatusBadRequest, Message: "invalid domain", I18nKey: "errors.INVALID_DOMAIN"},

	// 算法与参数。
	"UNSUPPORTED_ALGORITHM": {HTTP: http.StatusBadRequest, Message: "unsupported algorithm", I18nKey: "errors.UNSUPPORTED_ALGORITHM"},
	"UNSUPPORTED_MODE":      {HTTP: http.StatusBadRequest, Message: "unsupported mode", I18nKey: "errors.UNSUPPORTED_MODE"},
	"UNSUPPORTED_PADDING":   {HTTP: http.StatusBadRequest, Message: "unsupported padding", I18nKey: "errors.UNSUPPORTED_PADDING"},
	"UNSUPPORTED_FORMAT":    {HTTP: http.StatusBadRequest, Message: "unsupported format", I18nKey: "errors.UNSUPPORTED_FORMAT"},
	"UNSUPPORTED_VERSION":   {HTTP: http.StatusBadRequest, Message: "unsupported version", I18nKey: "errors.UNSUPPORTED_VERSION"},
	"KEY_SIZE_MISMATCH":     {HTTP: http.StatusBadRequest, Message: "key size mismatch", I18nKey: "errors.KEY_SIZE_MISMATCH"},
	"IV_SIZE_MISMATCH":      {HTTP: http.StatusBadRequest, Message: "iv size mismatch", I18nKey: "errors.IV_SIZE_MISMATCH"},
	"PASSPHRASE_REQUIRED":   {HTTP: http.StatusBadRequest, Message: "passphrase required", I18nKey: "errors.PASSPHRASE_REQUIRED"},
	"PASSPHRASE_INCORRECT":  {HTTP: http.StatusBadRequest, Message: "passphrase incorrect", I18nKey: "errors.PASSPHRASE_INCORRECT"},
	"TAG_REQUIRED":          {HTTP: http.StatusBadRequest, Message: "authentication tag required", I18nKey: "errors.TAG_REQUIRED"},
	"TAG_INVALID":           {HTTP: http.StatusBadRequest, Message: "invalid authentication tag", I18nKey: "errors.TAG_INVALID"},
	"DECRYPT_FAILED":        {HTTP: http.StatusBadRequest, Message: "decryption failed", I18nKey: "errors.DECRYPT_FAILED"},
	"SIGNATURE_INVALID":     {HTTP: http.StatusBadRequest, Message: "invalid signature", I18nKey: "errors.SIGNATURE_INVALID"},
	"PLAINTEXT_TOO_LONG":    {HTTP: http.StatusBadRequest, Message: "plaintext too long", I18nKey: "errors.PLAINTEXT_TOO_LONG"},
	"KEY_MISMATCH":          {HTTP: http.StatusBadRequest, Message: "key mismatch", I18nKey: "errors.KEY_MISMATCH"},
	"SAN_INVALID":           {HTTP: http.StatusBadRequest, Message: "invalid subject alternative name", I18nKey: "errors.SAN_INVALID"},
	"SERIAL_INVALID":        {HTTP: http.StatusBadRequest, Message: "invalid serial number", I18nKey: "errors.SERIAL_INVALID"},
	"NOT_AFTER_BEFORE_NOW":  {HTTP: http.StatusBadRequest, Message: "notAfter is before notBefore", I18nKey: "errors.NOT_AFTER_BEFORE_NOW"},

	// 限制与传输。
	"TOO_LARGE":              {HTTP: http.StatusRequestEntityTooLarge, Message: "request body too large", I18nKey: "errors.TOO_LARGE"},
	"TOO_MANY_ITEMS":         {HTTP: http.StatusBadRequest, Message: "too many items", I18nKey: "errors.TOO_MANY_ITEMS"},
	"TIMEOUT":                {HTTP: http.StatusGatewayTimeout, Message: "processing timeout", I18nKey: "errors.TIMEOUT"},
	"RATE_LIMITED":           {HTTP: http.StatusTooManyRequests, Message: "rate limit exceeded", I18nKey: "errors.RATE_LIMITED"},
	"TOOL_NOT_FOUND":         {HTTP: http.StatusNotFound, Message: "tool not found", I18nKey: "errors.TOOL_NOT_FOUND"},
	"TOOL_DISABLED":          {HTTP: http.StatusNotFound, Message: "tool disabled", I18nKey: "errors.TOOL_DISABLED"},
	"METHOD_NOT_ALLOWED":     {HTTP: http.StatusMethodNotAllowed, Message: "method not allowed", I18nKey: "errors.METHOD_NOT_ALLOWED"},
	"UNSUPPORTED_MEDIA_TYPE": {HTTP: http.StatusUnsupportedMediaType, Message: "unsupported media type", I18nKey: "errors.UNSUPPORTED_MEDIA_TYPE"},
	"CROSS_ORIGIN_BLOCKED":   {HTTP: http.StatusForbidden, Message: "cross-origin request blocked", I18nKey: "errors.CROSS_ORIGIN_BLOCKED"},
	"NETWORK_DISABLED":       {HTTP: http.StatusForbidden, Message: "network access disabled", I18nKey: "errors.NETWORK_DISABLED"},
	"SSRF_BLOCKED":           {HTTP: http.StatusBadRequest, Message: "target blocked by policy", I18nKey: "errors.SSRF_BLOCKED"},
	"DNS_ERROR":              {HTTP: http.StatusBadGateway, Message: "dns resolution failed", I18nKey: "errors.DNS_ERROR"},
	"TLS_ERROR":              {HTTP: http.StatusBadGateway, Message: "tls handshake failed", I18nKey: "errors.TLS_ERROR"},
	"HOST_UNREACHABLE":       {HTTP: http.StatusBadGateway, Message: "host unreachable", I18nKey: "errors.HOST_UNREACHABLE"},
	"UPSTREAM_TIMEOUT":       {HTTP: http.StatusGatewayTimeout, Message: "upstream timeout", I18nKey: "errors.UPSTREAM_TIMEOUT"},
	"UPSTREAM_ERROR":         {HTTP: http.StatusBadGateway, Message: "upstream error", I18nKey: "errors.UPSTREAM_ERROR"},
	"REDIRECT_LIMIT":         {HTTP: http.StatusBadGateway, Message: "too many redirects", I18nKey: "errors.REDIRECT_LIMIT"},

	// 服务端。
	"INTERNAL_ERROR":          {HTTP: http.StatusInternalServerError, Message: "internal error", I18nKey: "errors.INTERNAL_ERROR"},
	"CRYPTO_OPERATION_FAILED": {HTTP: http.StatusInternalServerError, Message: "crypto operation failed", I18nKey: "errors.CRYPTO_OPERATION_FAILED"},
	"CRYPTO_INIT_FAILED":      {HTTP: http.StatusInternalServerError, Message: "crypto initialization failed", I18nKey: "errors.CRYPTO_INIT_FAILED"},
	"STORAGE_ERROR":           {HTTP: http.StatusInternalServerError, Message: "storage error", I18nKey: "errors.STORAGE_ERROR"},
	"NOT_IMPLEMENTED":         {HTTP: http.StatusNotImplemented, Message: "not implemented", I18nKey: "errors.NOT_IMPLEMENTED"},
	"STATS_UNAUTHORIZED":      {HTTP: http.StatusUnauthorized, Message: "stats token missing or invalid", I18nKey: "errors.STATS_UNAUTHORIZED"},
}
