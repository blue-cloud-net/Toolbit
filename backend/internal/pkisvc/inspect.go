package pkisvc

// Inspect 解析密钥内容，返回其元信息与**规范化**后的 PEM。
//
// source 是 PEM 文本或 DER 二进制，由 isPEM 报告；
// passphrase 是加密私钥的口令（源不是加密私钥时可为空）。
//
// 规范化规则：
//
//   - 私钥 → PKCS#8 PEM；源为加密私钥时**保持加密**，
//     不会因为「解析」而把私钥解密暴露出去；
//   - 公钥 → SPKI PEM。
//
// 证书不是密钥：本函数只接受私钥与公钥，
// 证书请用 cert-parse（见 docs/roadmap.md §8）。
func Inspect(source []byte, isPEM bool, passphrase string) (*KeyInfo, []byte, error) {
	src, err := detectSource(ConvertOptions{Source: source, SourceIsPEM: isPEM})
	if err != nil {
		return nil, nil, err
	}

	switch src.Kind {
	case sourcePrivate:
		priv, err := LoadPrivateKey(src.PEM, passphrase)
		if err != nil {
			return nil, nil, err
		}
		defer func() { _ = priv.Close() }()

		info, err := describeKey(priv)
		if err != nil {
			return nil, nil, err
		}
		info.Format = src.Encoding
		info.Encrypted = src.Encrypted

		normalized, err := marshalCanonicalPrivate(priv, src.Encrypted, passphrase)
		if err != nil {
			return nil, nil, err
		}
		return info, normalized, nil
	case sourcePublic:
		pub, err := LoadPublicKey(src.PEM)
		if err != nil {
			return nil, nil, err
		}
		defer func() { _ = pub.Close() }()

		info, err := describePublicKey(pub)
		if err != nil {
			return nil, nil, err
		}
		info.Format = src.Encoding

		normalized, err := pub.MarshalPEM()
		if err != nil {
			return nil, nil, err
		}
		return info, normalized, nil
	default:
		return nil, nil, toolerrInvalidKey("unsupported key material")
	}
}
