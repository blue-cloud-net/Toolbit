package pkisvc

import (
	"encoding/hex"
	"math/big"

	"github.com/blue-cloud-net/tongsuo-go/asym"
	"github.com/blue-cloud-net/tongsuo-go/digest"
)

// KeyInfo 是密钥的公开元信息快照。
//
// 它是 pkisvc 对外的**唯一**元信息出口：本包的公开签名不出现
// tongsuo-go 的类型，因此上层拿不到底层库的 `*big.Int` 等结构。
type KeyInfo struct {
	// Algorithm 是算法标识（RSA / SM2 / EC / Ed25519 / Ed448 / X25519 / X448）。
	Algorithm string
	// KeyType 是密钥类型（private / public）。
	KeyType string
	// Format 是识别到的编码（pkcs1 / pkcs8 / spki）；仅 Inspect 填充。
	Format string
	// Encrypted 报告是否为加密私钥；仅 Inspect 填充。
	Encrypted bool
	// Bits 是 RSA 密钥位数；非 RSA 为 0。
	Bits int
	// Curve 是曲线名（P-256 / P-384 / P-521 / SM2）；非椭圆曲线为空。
	Curve string
	// FingerprintSHA256 是公钥 SPKI DER 的 SHA-256 指纹（小写十六进制）。
	FingerprintSHA256 string
	// RSA 是 RSA 私钥参数；仅私钥且算法为 RSA 时有值。
	RSA *RSAParams
}

// RSAParams 是 RSA 私钥的参数视图。
//
// 各值都是**小写十六进制**（不带 0x 前缀），便于与 `openssl` 的输出对照。
type RSAParams struct {
	// N 是模数。
	N string
	// E 是公钥指数。
	E string
	// D 是私钥指数。
	D string
	// P 是素因子 p。
	P string
	// Q 是素因子 q。
	Q string
	// Dp 是 CRT 系数 d mod (p-1)。
	Dp string
	// Dq 是 CRT 系数 d mod (q-1)。
	Dq string
	// Qi 是 CRT 系数 q⁻¹ mod p。
	Qi string
}

// describeKey 读取私钥的元信息。
func describeKey(priv *PrivateKey) (*KeyInfo, error) {
	if priv == nil || priv.key == nil {
		return nil, toolerrInvalidKey("key already closed")
	}
	params, err := asym.Params(priv.key)
	if err != nil {
		return nil, MapError(err)
	}

	info := &KeyInfo{
		Algorithm: priv.Algorithm(),
		KeyType:   KeyTypePrivate,
	}
	if params != nil {
		info.Bits = rsaBits(params)
		info.Curve = curveFromUpstream(params.Curve)
		if params.Type == "RSA" {
			info.RSA = &RSAParams{
				N:  hexInt(params.N),
				E:  hexInt(params.E),
				D:  hexInt(params.D),
				P:  hexInt(params.P),
				Q:  hexInt(params.Q),
				Dp: hexInt(params.Dmp1),
				Dq: hexInt(params.Dmq1),
				Qi: hexInt(params.Iqmp),
			}
		}
	}

	fp, err := publicFingerprint(priv)
	if err != nil {
		return nil, err
	}
	info.FingerprintSHA256 = fp

	return info, nil
}

// describePublicKey 读取公钥的元信息。
func describePublicKey(pub *PublicKey) (*KeyInfo, error) {
	if pub == nil || pub.key == nil {
		return nil, toolerrInvalidKey("key already closed")
	}
	params, err := asym.Params(pub.key)
	if err != nil {
		return nil, MapError(err)
	}

	info := &KeyInfo{
		Algorithm: pub.Algorithm(),
		KeyType:   KeyTypePublic,
	}
	if params != nil {
		info.Bits = rsaBits(params)
		info.Curve = curveFromUpstream(params.Curve)
	}

	fp, err := publicFingerprintFromPEM(mustPublicPEM(pub))
	if err != nil {
		return nil, err
	}
	info.FingerprintSHA256 = fp

	return info, nil
}

// publicFingerprint 计算私钥配对公钥的指纹。
func publicFingerprint(priv *PrivateKey) (string, error) {
	pemBytes, err := priv.MarshalPublicKeyPEM()
	if err != nil {
		return "", err
	}
	return publicFingerprintFromPEM(pemBytes)
}

// publicFingerprintFromPEM 计算 SPKI PEM 的 SHA-256 指纹。
//
// 指纹算在 **SPKI DER** 上：它只依赖公钥，因此私钥与它的公钥指纹相同，
// 可直接用来判断「是不是同一把钥匙」。
func publicFingerprintFromPEM(pemBytes []byte) (string, error) {
	der, err := pemToDER(pemBytes)
	if err != nil {
		return "", err
	}
	sum, err := digest.Sum("SHA256", der)
	if err != nil {
		return "", MapError(err)
	}
	return hex.EncodeToString(sum), nil
}

// mustPublicPEM 导出公钥 PEM；失败时返回空切片。
//
// 调用方（describePublicKey）已经确认公钥句柄有效，
// 因此这里的失败只可能是底层导出异常，交给上层按空值处理。
func mustPublicPEM(pub *PublicKey) []byte {
	out, err := pub.MarshalPEM()
	if err != nil {
		return nil
	}
	return out
}

// rsaBits 计算 RSA 密钥位数；非 RSA 返回 0。
func rsaBits(params *asym.KeyParams) int {
	if params == nil || params.Type != "RSA" || params.N == nil {
		return 0
	}
	return params.N.BitLen()
}

// hexInt 把大整数渲染为小写十六进制；nil 返回空串。
func hexInt(v *big.Int) string {
	if v == nil {
		return ""
	}
	return v.Text(16)
}
