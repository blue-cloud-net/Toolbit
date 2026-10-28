// Package keys 实现密钥类工具。
//
// 本包**不**实现算法：只做「解码输入 → 校验字段 → 调用 pkisvc → 编码 DTO」。
//
// 所有密码学运算都经 pkisvc 转调 Tongsuo；本包**不得**直接 import tongsuo-go，
// 也**不得**用标准库 crypto/* 拼装密码学运算。
package keys

import (
	"encoding/base64"

	"github.com/blue-cloud-net/Toolbit/backend/internal/pkisvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// 密钥输入格式（复用 tools 包的共享定义）。
const (
	// FormatAuto 自动识别。
	FormatAuto = tools.FormatAuto
	// FormatPEM 是 PEM 文本。
	FormatPEM = tools.FormatPEM
	// FormatDER 是 DER 二进制的 base64 表示。
	FormatDER = tools.FormatDER
)

// 私钥输出格式。
const (
	// PrivateFormatPKCS8 是 PKCS#8 私钥（默认）。
	PrivateFormatPKCS8 = pkisvc.FormatPKCS8
	// PrivateFormatPKCS1 是 PKCS#1 私钥（仅 RSA）。
	PrivateFormatPKCS1 = pkisvc.FormatPKCS1
)

// 公钥输出格式。
const (
	// PublicFormatSPKI 是 SubjectPublicKeyInfo 公钥（默认）。
	PublicFormatSPKI = pkisvc.FormatSPKI
	// PublicFormatPKCS1 是 PKCS#1 公钥。
	//
	// ⚠️ 底层库**不提供** PKCS#1 公钥的导出入口，本项目不自研兜底，
	// 因此该取值恒返回 UNSUPPORTED_FORMAT
	// （见 docs/upstream/tongsuo-go-issues.md I-14）。
	PublicFormatPKCS1 = pkisvc.FormatPKCS1
)

// KeyInput 是密钥输入（复用 tools 包的共享类型）。
type KeyInput = tools.KeyInput

// PEMOutput 是 PEM 形式的输出。
type PEMOutput struct {
	// PEM 是 PEM 文本。
	PEM string `json:"pem"`
	// DERBase64 是 DER 二进制的 base64 表示；仅目标格式为 der 时有值。
	DERBase64 string `json:"derBase64,omitempty"`
}

// DetectedInfo 是识别到的密钥元信息。
type DetectedInfo struct {
	// Algorithm 是算法标识。
	Algorithm string `json:"algorithm"`
	// Format 是识别到的编码（pkcs1 / pkcs8 / spki）。
	Format string `json:"format"`
	// Encrypted 报告是否为加密私钥。
	Encrypted bool `json:"encrypted"`
	// Bits 是 RSA 位数；非 RSA 为 0。
	Bits int `json:"bits"`
	// Curve 是曲线名；非椭圆曲线为空。
	Curve string `json:"curve"`
}

// RSAParamsOutput 是 RSA 私钥参数的输出形态。
type RSAParamsOutput struct {
	// N 是模数（十六进制）。
	N string `json:"n"`
	// E 是公钥指数（十六进制）。
	E string `json:"e"`
	// D 是私钥指数（十六进制）。
	D string `json:"d"`
	// P 是素因子 p（十六进制）。
	P string `json:"p"`
	// Q 是素因子 q（十六进制）。
	Q string `json:"q"`
	// Dp 是 CRT 系数 d mod (p-1)（十六进制）。
	Dp string `json:"dp"`
	// Dq 是 CRT 系数 d mod (q-1)（十六进制）。
	Dq string `json:"dq"`
	// Qi 是 CRT 系数 q⁻¹ mod p（十六进制）。
	Qi string `json:"qi"`
}

// encodeDER 把 DER 字节编码为 base64 文本。
func encodeDER(der []byte) string {
	return base64.StdEncoding.EncodeToString(der)
}

// encodePEM 把 PEM 字节按 UTF-8 文本返回。
func encodePEM(pemBytes []byte) string {
	return string(pemBytes)
}

// decodeInput 把传输层的输入解码为字节，并报告它是 PEM 还是 DER。
func decodeInput(in KeyInput) (data []byte, isPEM bool, err error) {
	return in.Resolve()
}
