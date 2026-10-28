// Package crypto 实现加密类工具。
//
// 本包**不**实现算法：只做「解码输入 → 校验字段 → 调用 cryptosvc → 编码 DTO」。
//
// 所有密码学运算都经 cryptosvc 转调 Tongsuo；本包**不得**直接 import tongsuo-go，
// 也**不得**用标准库 crypto/* 拼装密码学运算。
package crypto

import (
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// 密钥输入格式。
//
// 复用 `tools` 包中的共享定义（跨分类共享的输入类型），
// 本包只保留别名以便既有引用点不变。
const (
	// FormatAuto 自动识别：PEM → DER → base64 → hex。
	FormatAuto = tools.FormatAuto
	// FormatPEM 是 PEM 文本。
	FormatPEM = tools.FormatPEM
	// FormatDER 是 DER 二进制的 base64 表示。
	FormatDER = tools.FormatDER
	// FormatBase64 是 base64 文本。
	FormatBase64 = tools.FormatBase64
	// FormatHex 是十六进制文本。
	FormatHex = tools.FormatHex
)

// KeyFormats 返回受支持的密钥格式名（顺序稳定）。
func KeyFormats() []string {
	return tools.KeyFormats()
}

// ByteInput 是字节输入。
//
// 复用 `tools.Bytes`（跨分类共享的通用输入类型），
// 本包只保留别名以便既有引用点不变。
type ByteInput = tools.Bytes

// KeyInput 是密钥输入。
//
// 复用 `tools.KeyInput`（跨分类共享的密钥输入类型），
// 本包只保留别名以便既有引用点不变。
type KeyInput = tools.KeyInput
