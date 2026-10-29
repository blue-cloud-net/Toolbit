package keys

import (
	"context"
	"encoding/asn1"
	"encoding/base64"
	"encoding/binary"
	"encoding/pem"
	"math/big"
	"strings"

	"github.com/blue-cloud-net/Toolbit/backend/internal/cryptosvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/pkisvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// SSH 密钥算法。
const (
	// SSHAlgorithmEd25519 是 Ed25519。
	SSHAlgorithmEd25519 = "Ed25519"
	// SSHAlgorithmRSA 是 RSA。
	SSHAlgorithmRSA = "RSA"
)

// SSHAlgorithm 返回受支持的 SSH 密钥算法（顺序稳定）。
func SSHAlgorithm() []string {
	return []string{SSHAlgorithmEd25519, SSHAlgorithmRSA}
}

// OpenSSH 私钥格式的魔数。
const sshPrivateMagic = "openssh-key-v1\x00"

// OpenSSH 私钥 PEM 头尾。
const (
	sshPrivatePEMHeader = "OPENSSH PRIVATE KEY"
	sshPrivatePEMFooter = "-----END OPENSSH PRIVATE KEY-----"
)

// SSHKeygenRequest 是 ssh-keygen 工具的请求。
type SSHKeygenRequest struct {
	// Algorithm 是密钥算法：Ed25519 / RSA；默认 Ed25519。
	Algorithm string `json:"algorithm"`
	// Bits 是 RSA 密钥位数；仅 RSA 有效，默认 3072。
	Bits int `json:"bits"`
	// Comment 是密钥注释；默认 user@toolbit。
	Comment string `json:"comment"`
}

// SSHKeygenResult 是 ssh-keygen 工具的响应。
type SSHKeygenResult struct {
	// Algorithm 是密钥算法。
	Algorithm string `json:"algorithm"`
	// PrivateKey 是 OpenSSH 私钥（PEM 文本）。
	PrivateKey string `json:"privateKey"`
	// PublicKey 是 authorized_keys 格式的公钥行。
	PublicKey string `json:"publicKey"`
	// Fingerprint 是 SHA256 指纹（base64，无填充）。
	Fingerprint string `json:"fingerprint"`
	// Bits 是 RSA 密钥位数；非 RSA 为 0。
	Bits int `json:"bits"`
	// Comment 是密钥注释。
	Comment string `json:"comment"`
}

// defaultSSHComment 是默认注释。
const defaultSSHComment = "user@toolbit"

// SSHKeygen 生成 OpenSSH 格式的密钥对。
//
// **只支持未加密私钥**（`openssh-key-v1` 的 `none` 加密变体）：
// 口令加密变体需要 bcrypt KDF 与密文格式转换，按 AGENTS §15
// 属于上游能力缺口，本版**跳过**，不做自研兜底。
//
// 密钥生成经 pkisvc；SHA-256 指纹经 cryptosvc 取摘要，
// 不自行实现任何密码学运算。
func SSHKeygen(ctx context.Context, req *SSHKeygenRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	algorithm := req.Algorithm
	if algorithm == "" {
		algorithm = SSHAlgorithmEd25519
	}
	comment := req.Comment
	if comment == "" {
		comment = defaultSSHComment
	}

	session := cryptosvc.FromContext(ctx)
	if session != nil {
		if err := session.Acquire(ctx); err != nil {
			return tools.Result{}, err
		}
		defer session.Release("ssh-keygen")
	}

	pair, err := pkisvc.Generate(algorithm, req.Bits, "")
	if err != nil {
		return tools.Result{}, err
	}
	defer func() { _ = pair.Close() }()

	var privatePEM []byte
	switch algorithm {
	case SSHAlgorithmRSA:
		privatePEM, err = pair.Private.MarshalPKCS1PEM()
	default:
		privatePEM, err = pair.Private.MarshalPKCS8PEM()
	}
	if err != nil {
		return tools.Result{}, err
	}

	publicPEM, err := pair.Public.MarshalPEM()
	if err != nil {
		return tools.Result{}, err
	}

	publicBlob, publicRaw, err := buildSSHPublicBlob(publicPEM, algorithm)
	if err != nil {
		return tools.Result{}, err
	}

	privateBlob, err := buildSSHPrivateBlob(privatePEM, publicBlob, publicRaw, algorithm, comment)
	if err != nil {
		return tools.Result{}, err
	}

	// 指纹：base64(SHA256(公钥 blob))，去掉尾部填充。
	digest, err := cryptosvc.Digest("SHA256", publicBlob)
	if err != nil {
		return tools.Result{}, err
	}
	fingerprint := base64.RawStdEncoding.EncodeToString(digest)

	authorizedKey := sshKeyTypeName(algorithm) + " " +
		base64.StdEncoding.EncodeToString(publicBlob) + " " + comment

	result := SSHKeygenResult{
		Algorithm:   algorithm,
		PrivateKey:  string(pem.EncodeToMemory(&pem.Block{Type: sshPrivatePEMHeader, Bytes: privateBlob})),
		PublicKey:   authorizedKey,
		Fingerprint: "SHA256:" + fingerprint,
		Comment:     comment,
	}
	if algorithm == SSHAlgorithmRSA {
		result.Bits = req.Bits
	}

	return tools.Result{Data: result}, nil
}

// sshKeyTypeName 返回 OpenSSH 的密钥类型标识。
func sshKeyTypeName(algorithm string) string {
	switch algorithm {
	case SSHAlgorithmRSA:
		return "ssh-rsa"
	default:
		return "ssh-ed25519"
	}
}

// buildSSHPublicBlob 构造 OpenSSH 公钥 blob。
//
// blob 结构是若干 SSH `string` 的拼接：
// 先是密钥类型名，再是算法相关的公钥分量。
//
// 第二个返回值是**原始公钥字节**（Ed25519 的 32 字节），
// 私钥块需要它拼出 `seed || pub`。
func buildSSHPublicBlob(publicPEM []byte, algorithm string) ([]byte, []byte, error) {
	block, _ := pem.Decode(publicPEM)
	if block == nil {
		return nil, nil, tools.InvalidField("algorithm", SSHAlgorithm(), "unsupported key format")
	}

	switch algorithm {
	case SSHAlgorithmRSA:
		key, err := parseRSAPublicKey(block.Bytes)
		if err != nil {
			return nil, nil, err
		}
		blob := sshString([]byte(sshKeyTypeName(algorithm)))
		blob = append(blob, sshMPInt(key.E)...)
		blob = append(blob, sshMPInt(key.N)...)
		return blob, nil, nil
	default:
		public, err := parseEd25519PublicKey(block.Bytes)
		if err != nil {
			return nil, nil, err
		}
		blob := sshString([]byte(sshKeyTypeName(algorithm)))
		blob = append(blob, sshString(public)...)
		return blob, public, nil
	}
}

// buildSSHPrivateBlob 构造 openssh-key-v1 私钥 blob。
//
// 结构：
//
//	"openssh-key-v1\0"
//	string  ciphername = "none"
//	string  kdfname    = "none"
//	string  kdfoptions = ""
//	uint32  密钥数量 = 1
//	string  公钥 blob
//	string  私钥块（未加密）
//
// 私钥块内含 checkint、密钥类型、算法分量与注释，
// 末尾按 8 字节对齐填充。
func buildSSHPrivateBlob(privatePEM, publicBlob, publicRaw []byte, algorithm, comment string) ([]byte, error) {
	block, _ := pem.Decode(privatePEM)
	if block == nil {
		return nil, tools.InvalidField("algorithm", SSHAlgorithm(), "unsupported key format")
	}

	privSection := []byte{}

	// checkint 是随机 32 位值，写两遍用于校验。
	checkInt, err := cryptosvc.RandomBytes(4)
	if err != nil {
		return nil, err
	}
	check := make([]byte, 4)
	copy(check, checkInt)
	privSection = append(privSection, check...)
	privSection = append(privSection, check...)

	switch algorithm {
	case SSHAlgorithmRSA:
		key, err := parseRSAPrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		privSection = append(privSection, sshString([]byte(sshKeyTypeName(algorithm)))...)
		privSection = append(privSection, sshMPInt(key.N)...)
		privSection = append(privSection, sshMPInt(key.E)...)
		privSection = append(privSection, sshMPInt(key.D)...)
		privSection = append(privSection, sshMPInt(key.Iqmp)...)
		privSection = append(privSection, sshMPInt(key.P)...)
		privSection = append(privSection, sshMPInt(key.Q)...)
	default:
		seed, err := parseEd25519PrivateKey(block.Bytes)
		if err != nil {
			return nil, err
		}
		privSection = append(privSection, sshString([]byte(sshKeyTypeName(algorithm)))...)
		privSection = append(privSection, sshString(publicRaw)...)
		// OpenSSH 的 Ed25519 私钥是 seed || pub 的 64 字节。
		privSection = append(privSection, sshString(append(append([]byte{}, seed...), publicRaw...))...)
	}

	privSection = append(privSection, sshString([]byte(comment))...)

	// 填充到 8 的倍数，填充字节依次为 1、2、3 …
	for pad := byte(1); len(privSection)%8 != 0; pad++ {
		privSection = append(privSection, pad)
	}

	blob := []byte(sshPrivateMagic)
	blob = append(blob, sshString([]byte("none"))...)
	blob = append(blob, sshString([]byte("none"))...)
	blob = append(blob, sshString(nil)...)
	blob = append(blob, sshUint32(1)...)
	blob = append(blob, sshString(publicBlob)...)
	blob = append(blob, sshString(privSection)...)

	return blob, nil
}

// sshString 编码 SSH 的 `string` 类型。
//
// 结构是 4 字节大端长度 + 内容。
func sshString(data []byte) []byte {
	out := make([]byte, 0, 4+len(data))
	out = append(out, sshUint32(uint32(len(data)))...)
	return append(out, data...)
}

// sshUint32 编码 4 字节大端无符号整数。
func sshUint32(value uint32) []byte {
	buffer := make([]byte, 4)
	binary.BigEndian.PutUint32(buffer, value)
	return buffer
}

// sshMPInt 编码 SSH 的 `mpint` 类型。
//
// 是大端二进制表示；最高位为 1 时前置 0x00 以免被读成负数。
func sshMPInt(value *big.Int) []byte {
	if value == nil || value.Sign() == 0 {
		return sshString(nil)
	}
	raw := value.Bytes()
	if raw[0]&0x80 != 0 {
		raw = append([]byte{0x00}, raw...)
	}
	return sshString(raw)
}

// Ed25519 的算法标识 OID。
var oidEd25519 = asn1.ObjectIdentifier{1, 3, 101, 112}

// rsaPublicKey 是 RSA 公钥的 ASN.1 结构。
type rsaPublicKey struct {
	N *big.Int
	E *big.Int
}

// rsaPrivateKey 是 PKCS#1 私钥的 ASN.1 结构。
//
// 字段必须与 PKCS#1 `RSAPrivateKey` **逐项对应**：
// 多一个字段会让 `asn1.Unmarshal` 报 “sequence truncated”。
// OpenSSH 需要的 `iqmp` 不在 PKCS#1 里，解析后单独算出。
type rsaPrivateKey struct {
	Version int
	N       *big.Int
	E       *big.Int
	D       *big.Int
	P       *big.Int
	Q       *big.Int
	DP      *big.Int
	DQ      *big.Int
	QInv    *big.Int
}

// pkcs8PrivateKey 是 PKCS#8 私钥的通用 ASN.1 结构。
type pkcs8PrivateKey struct {
	Version    int
	Algorithm  algorithmIdentifier
	PrivateKey []byte
}

// spkiPublicKey 是 SPKI 公钥的通用 ASN.1 结构。
type spkiPublicKey struct {
	Algorithm algorithmIdentifier
	PublicKey asn1.BitString
}

// algorithmIdentifier 是 X.509 的 AlgorithmIdentifier 结构。
//
// 这里用 `encoding/asn1` **结构解析**（AGENTS.md §4.15 允许的只读解析），
// 而不是 `crypto/x509/pkix`：后者属于禁止引入的标准库 crypto 包。
type algorithmIdentifier struct {
	Algorithm  asn1.ObjectIdentifier
	Parameters asn1.RawValue `asn1:"optional"`
}

// parseEd25519PublicKey 从 SPKI DER 提取 Ed25519 公钥字节。
func parseEd25519PublicKey(der []byte) ([]byte, error) {
	var key spkiPublicKey
	if _, err := asn1.Unmarshal(der, &key); err != nil {
		return nil, tools.InvalidField("algorithm", SSHAlgorithm(), "cannot parse public key")
	}
	if !key.Algorithm.Algorithm.Equal(oidEd25519) {
		return nil, tools.InvalidField("algorithm", SSHAlgorithm(), "not an Ed25519 key")
	}
	return key.PublicKey.Bytes, nil
}

// parseEd25519PrivateKey 从 PKCS#8 DER 提取 Ed25519 的私钥种子。
//
// 公钥**不**在这里从种子推导（那属于密钥运算）；
// 它由调用方从 SPKI 结构解析得到后传入。
func parseEd25519PrivateKey(der []byte) ([]byte, error) {
	var key pkcs8PrivateKey
	if _, err := asn1.Unmarshal(der, &key); err != nil {
		return nil, tools.InvalidField("algorithm", SSHAlgorithm(), "cannot parse private key")
	}
	if !key.Algorithm.Algorithm.Equal(oidEd25519) {
		return nil, tools.InvalidField("algorithm", SSHAlgorithm(), "not an Ed25519 key")
	}

	// PrivateKey 是内层 OCTET STRING 的 DER 编码。
	var inner []byte
	if _, err := asn1.Unmarshal(key.PrivateKey, &inner); err != nil {
		return nil, tools.InvalidField("algorithm", SSHAlgorithm(), "cannot parse private key")
	}
	if len(inner) != 32 {
		return nil, tools.InvalidField("algorithm", SSHAlgorithm(), "invalid private key length")
	}
	return inner, nil
}

// parseRSAPublicKey 从 SPKI DER 提取 RSA 公钥。
func parseRSAPublicKey(der []byte) (*rsaPublicKey, error) {
	var spki spkiPublicKey
	if _, err := asn1.Unmarshal(der, &spki); err != nil {
		return nil, tools.InvalidField("algorithm", SSHAlgorithm(), "cannot parse public key")
	}
	var key rsaPublicKey
	if _, err := asn1.Unmarshal(spki.PublicKey.Bytes, &key); err != nil {
		return nil, tools.InvalidField("algorithm", SSHAlgorithm(), "cannot parse public key")
	}
	return &key, nil
}

// parseRSAPrivateKey 从 PKCS#1 DER 提取 RSA 私钥分量。
//
// OpenSSH 的私钥块还需要 `iqmp`（q 模 p 的逆元），
// 它由 p 与 q 直接算出，属于**算术**而非密钥运算。
func parseRSAPrivateKey(der []byte) (*rsaPrivateKeyComponents, error) {
	var key rsaPrivateKey
	if _, err := asn1.Unmarshal(der, &key); err != nil {
		return nil, tools.InvalidField("algorithm", SSHAlgorithm(), "cannot parse private key")
	}
	if key.P == nil || key.Q == nil {
		return nil, tools.InvalidField("algorithm", SSHAlgorithm(), "invalid private key")
	}
	iqmp := new(big.Int).ModInverse(key.Q, key.P)
	if iqmp == nil {
		return nil, tools.InvalidField("algorithm", SSHAlgorithm(), "invalid private key")
	}
	return &rsaPrivateKeyComponents{
		N:    key.N,
		E:    key.E,
		D:    key.D,
		P:    key.P,
		Q:    key.Q,
		Iqmp: iqmp,
	}, nil
}

// rsaPrivateKeyComponents 是 OpenSSH 私钥块需要的 RSA 分量。
type rsaPrivateKeyComponents struct {
	N    *big.Int
	E    *big.Int
	D    *big.Int
	P    *big.Int
	Q    *big.Int
	Iqmp *big.Int
}

// Validate 做字段级校验。
func (r *SSHKeygenRequest) Validate() error {
	algorithm := r.Algorithm
	if algorithm == "" {
		algorithm = SSHAlgorithmEd25519
	}
	if !containsString(SSHAlgorithm(), algorithm) {
		return tools.InvalidField("algorithm", SSHAlgorithm(), algorithm)
	}
	if algorithm == SSHAlgorithmRSA {
		bits := r.Bits
		if bits == 0 {
			bits = 3072
		}
		if !containsInt(pkisvc.RSABits(), bits) {
			return tools.InvalidField("bits", pkisvc.RSABits(), bits)
		}
	}
	if len(r.Comment) > 128 {
		return tools.InvalidField("comment", []int{0, 128}, len(r.Comment))
	}
	if strings.ContainsAny(r.Comment, "\r\n") {
		return tools.InvalidField("comment", "single line comment", "invalid format")
	}
	return nil
}
