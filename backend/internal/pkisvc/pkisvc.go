// Package pkisvc 是 PKI 与密钥格式服务层。
//
// 本包与 cryptosvc 并列，是**仅有的两个**允许 import tongsuo-go 的层
// （见 AGENTS.md §4.1）：
//
//   - cryptosvc —— 密码学**原语**：摘要 / MAC / 对称加解密 / 非对称运算 / 签名；
//   - pkisvc    —— 密钥与 **PKI 领域**：密钥生成与格式、X.509、JWK、PKCS 容器。
//
// 分工理由：X.509 / JWK / PKCS#12 的句柄模型与 cryptosvc 的原语模型差异很大，
// 混在一个包里会让两边都臃肿；拆开后 cryptosvc 保持「纯运算」，pkisvc 保持
// 「纯格式与结构」。
//
// # 依赖方向
//
//	api → tools → {cryptosvc, pkisvc} → tongsuo-go
//
// pkisvc **不**依赖 tools（避免成环）；pkisvc 可依赖 cryptosvc（复用摘要白名单
// 与错误映射），但**不**被 cryptosvc 依赖。
//
// # 句柄纪律
//
//   - 本包返回的所有原生句柄都**必须**显式 Close；
//   - Close **幂等**：重复调用只返回首次的结果；
//   - 句柄由调用方（tools 层）登记到 cryptosvc.Session，请求结束时按 LIFO 释放；
//   - 本包的**公开签名不出现** tongsuo-go 的类型（asym.* / x509.*），
//     以免 tools 层被迫 import tongsuo-go 而破坏「只有两个密码学层」的边界。
//
// # 与 cryptosvc 的关系
//
// 两个包的句柄类型**互不通用**：cryptosvc 的 PrivateKey 内部字段不可导出，
// 且 cryptosvc 的公开签名不出现 asym.*，因此 pkisvc 自带一套句柄包装。
package pkisvc

import "time"

// 密钥算法标识。
const (
	// AlgRSA 是 RSA 算法。
	AlgRSA = "RSA"
	// AlgSM2 是国密 SM2 算法。
	AlgSM2 = "SM2"
	// AlgEC 是椭圆曲线 ECDSA 算法。
	AlgEC = "EC"
	// AlgEd25519 是 Ed25519 算法。
	AlgEd25519 = "Ed25519"
	// AlgEd448 是 Ed448 算法。
	AlgEd448 = "Ed448"
	// AlgX25519 是 X25519 算法。
	AlgX25519 = "X25519"
	// AlgX448 是 X448 算法。
	AlgX448 = "X448"
)

// 曲线标识（对外友好名）。
const (
	// CurveP256 是 NIST P-256 曲线。
	CurveP256 = "P-256"
	// CurveP384 是 NIST P-384 曲线。
	CurveP384 = "P-384"
	// CurveP521 是 NIST P-521 曲线。
	CurveP521 = "P-521"
	// CurveSM2 是国密 SM2 推荐曲线。
	CurveSM2 = "SM2"
)

// zeroTime 是「未指定时刻」的零值。
//
// 它用于区分「按当前时间判定」与「回拨到某个历史时刻」两种语义：
// 零值表示不调用 SetTime，非零值表示按该时刻验证。
var zeroTime time.Time

// KeyAlgorithms 返回受支持的密钥算法名（顺序稳定）。
func KeyAlgorithms() []string {
	return []string{
		AlgRSA,
		AlgSM2,
		AlgEC,
		AlgEd25519,
		AlgEd448,
		AlgX25519,
		AlgX448,
	}
}

// ECCurves 返回受支持的椭圆曲线名（顺序稳定）。
//
// 该列表同时适用于 EC 与 SM2：SM2 使用其中的 CurveSM2。
func ECCurves() []string {
	return []string{CurveP256, CurveP384, CurveP521, CurveSM2}
}

// RSABits 返回受支持的 RSA 密钥位数（顺序稳定）。
func RSABits() []int {
	return []int{1024, 2048, 3072, 4096}
}

// curveToUpstream 把对外曲线名映射为底层库的曲线名。
//
// 底层库用 OpenSSL 的曲线名（prime256v1 等），本项目对外用 P-256 这样的
// 友好名。SM2 **不**在映射内：SM2 走独立的生成入口，不复用 EC。
func curveToUpstream(curve string) (string, bool) {
	switch curve {
	case CurveP256:
		return "prime256v1", true
	case CurveP384:
		return "secp384r1", true
	case CurveP521:
		return "secp521r1", true
	default:
		return "", false
	}
}

// curveFromUpstream 把底层库的曲线名映射为对外友好名。
//
// 未登记的曲线名原样返回（不报错）：底层库可能新增曲线，
// 本项目不应因为不认识就拒绝对外展示。
func curveFromUpstream(curve string) string {
	switch curve {
	case "prime256v1":
		return CurveP256
	case "secp384r1":
		return CurveP384
	case "secp521r1":
		return CurveP521
	case "sm2", "sm2p256v1":
		return CurveSM2
	default:
		return curve
	}
}
