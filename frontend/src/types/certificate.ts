/**
 * 证书类工具的共享类型。
 *
 * 字段名必须与 `backend/internal/tools/pki/dto.go` 的 `json:` tag 逐字一致：
 * 名字对不上时页面不会报错，只会**一直空白**。
 *
 * ## 数组字段的可空性
 *
 * 后端对「证书没有该扩展」有两种编码，**语义不同**：
 *
 * - `null` —— 已查询、确认没有该项。适用于上游 tongsuo-go 在扩展缺失时
 *   返回 nil 的字段（`san` / `keyUsage` / `extendedKeyUsage` / `cpsUris` /
 *   `noticeNumbers`）。
 * - `[]` —— 结构性字段恒为空数组（`subject` / `issuer` / `extensions` /
 *   `policies` / `fingerprints` / `chain`），由 `make` 显式构造。
 *
 * 因此**可空字段的类型是 `T[] | null`**：TypeScript 会强制每个使用点显式
 * 处理 `null`（用 `toList()` 归一化），漏一处直接编译失败，
 * 而不是运行期抛 `null.length` 让整页白屏。
 */

/** 一个 RDN 条目（主体 / 签发者的单个字段）。 */
export interface NameEntry {
  /** 字段短名（如 CN、O）。 */
  field: string
  /** 字段值。 */
  value: string
}

/** 公钥元信息（摘要视图）。 */
export interface PublicKeyInfo {
  /** 算法标识（RSA / SM2 / EC …）。 */
  algorithm: string
  /** RSA 密钥位数；非 RSA 为 0。 */
  bits: number
  /** 曲线名；非椭圆曲线为空。 */
  curve: string
}

/** 公钥展开视图（含 SPKI、分量与指纹）。 */
export interface PublicKeyDetail {
  /** 算法标识（RSA / SM2 / EC …）。 */
  algorithm: string
  /** RSA 密钥位数；非 RSA 为 0。 */
  bits: number
  /** 曲线名；非椭圆曲线为空。 */
  curve: string
  /** 公钥的 SPKI PEM 文本。 */
  pem: string
  /** SPKI DER 的十六进制。 */
  spkiHex: string
  /** SPKI DER 的 SHA-256 指纹（小写十六进制）。 */
  fingerprint: string
  /** RSA 模数（大写十六进制）；非 RSA 为空。 */
  modulus: string
  /** RSA 公钥指数（大写十六进制）；非 RSA 为空。 */
  exponent: string
  /** 椭圆曲线公钥点 X 分量（大写十六进制）。 */
  x: string
  /** 椭圆曲线公钥点 Y 分量（大写十六进制）。 */
  y: string
}

/** 一个扩展密钥用途条目。 */
export interface ExtKeyUsageInfo {
  /**
   * OpenSSL 的短名（如 `serverAuth`）。
   *
   * OpenSSL 对象库里没有登记的 OID（如 RFC 9336 的 id-kp-documentSigning）
   * 没有名字，这里为空串 —— 靠 `oid` 定位。
   */
  name: string
  /** 点分 OID（如 `1.3.6.1.5.5.7.3.1`）。 */
  oid: string
}

/** 证书策略。 */
export interface CertificatePolicy {
  /** 策略标识符（点分形式）。 */
  oid: string
  /** 策略声明 URL 列表；无声明时为 `null`。 */
  cpsUris: string[] | null
  /** 通知机构名。 */
  organization: string
  /** 通知编号列表；无编号时为 `null`。 */
  noticeNumbers: number[] | null
  /** 通知正文。 */
  explicitText: string
}

/** 一个 X.509 扩展。 */
export interface CertificateExtension {
  /** OpenSSL NID。 */
  nid: number
  /** 点分 OID。 */
  oid: string
  /** 扩展短名；未知扩展为空。 */
  name: string
  /** 是否标记为 critical。 */
  critical: boolean
  /** 扩展值的十六进制。 */
  value: string
  /** 扩展的可读文本；未知扩展为空。 */
  text: string
}

/** 有效性判定。 */
export interface CertificateValidity {
  /** `valid` | `expired` | `not_yet_valid`。 */
  status: string
  /** 距离过期的剩余天数；已过期为负数。**小数**（保留 1 位）。 */
  daysRemaining: number
}

/** 一组指纹中的一项。 */
export interface CertificateFingerprint {
  /** 算法名（小写）。 */
  algorithm: string
  /** 指纹值（小写十六进制）。 */
  value: string
}

/** 证书元信息。 */
export interface CertificateInfo {
  /** 版本号（2 表示 v3）。 */
  version: number
  /** 主体的 RDN 条目。 */
  subject: NameEntry[]
  /** 主体的单行文本（OpenSSL 风格）。 */
  subjectText: string
  /** 签发者的 RDN 条目。 */
  issuer: NameEntry[]
  /** 签发者的单行文本（OpenSSL 风格）。 */
  issuerText: string
  /** 序列号（大写十六进制）。 */
  serial: string
  /** 生效时间（RFC3339）。 */
  notBefore: string
  /** 过期时间（RFC3339）。 */
  notAfter: string
  /** 公钥元信息。 */
  publicKey: PublicKeyInfo
  /** 签名算法短名。 */
  signatureAlgorithm: string
  /** 签名算法的点分 OID。 */
  signatureAlgorithmOid: string
  /** 是否为 CA 证书。 */
  isCa: boolean
  /** 路径长度约束；非 CA 或无该扩展为 -1。 */
  pathLen: number
  /** 主体备用名条目；无该扩展时为 `null`。 */
  san: string[] | null
  /** 密钥用途位名；无该扩展时为 `null`。 */
  keyUsage: string[] | null
  /** 扩展密钥用途；无该扩展时为 `null`。 */
  extendedKeyUsage: ExtKeyUsageInfo[] | null
  /** 主题密钥标识符（大写十六进制）。 */
  subjectKeyId: string
  /** 签发者密钥标识符（大写十六进制）。 */
  authorityKeyId: string
  /** 主体名的 X509_NAME_hash。 */
  subjectHash: string
  /** 签发者名的 X509_NAME_hash。 */
  issuerHash: string
  /** 全部扩展。 */
  extensions: CertificateExtension[]
  /** 是否自签名。 */
  selfSigned: boolean
  /** 签名字节的十六进制（小写）。 */
  signatureHex: string
  /** 证书策略；无该扩展时为空数组。 */
  policies: CertificatePolicy[]
}

/** 链上证书与上一级的连接诊断。 */
export interface ChainLink {
  /** 连接方式（`leaf` / `keyid` / `name` / `self`）。 */
  method: string
  /** 连接是否成立。 */
  ok: boolean
  /** 断链原因；正常时为空。 */
  reason: string
}

/** 链上的一张证书。 */
export interface ChainCertificate {
  /** 证书的 PEM 文本。 */
  pem: string
  /** 证书元信息。 */
  info: CertificateInfo
  /** 链中角色（`leaf` / `intermediate` / `root` / `unlinked`）。 */
  role: string
  /** 与上一级的连接诊断。 */
  link: ChainLink
  /** 有效性判定。 */
  validity: CertificateValidity
  /** 指纹列表（按请求的算法顺序）；无该扩展时为 `null`。 */
  fingerprints: CertificateFingerprint[]
  /** 公钥展开视图；读取失败时为 null。 */
  publicKey: PublicKeyDetail | null
}

/** cert-parse 的响应。 */
export interface CertParseResult {
  /** 证书元信息。 */
  certificate: CertificateInfo
  /** 有效性判定。 */
  validity: CertificateValidity
  /** 计算出的指纹。 */
  fingerprints: CertificateFingerprint[]
  /** 是否自签名。 */
  selfSigned: boolean
  /** 原始 DER（base64）。 */
  rawDerBase64: string
  /** 证书的 PEM 文本。 */
  pem: string
  /** 公钥展开视图。 */
  publicKey: PublicKeyDetail | null
  /** 文本视图（OpenSSL `x509 -text` 体例）。 */
  textView: string
}

/** cert-chain-parse 的响应。 */
export interface CertChainParseResult {
  /** 证书总数。 */
  count: number
  /** 排好序的证书（叶 → 根）。 */
  chain: ChainCertificate[]
  /** 输入是否已经有序。 */
  ordered: boolean
  /** 是否缺少根证书。 */
  missingRoot: boolean
  /** 链尾证书是否自签名。 */
  selfSigned: boolean
}
