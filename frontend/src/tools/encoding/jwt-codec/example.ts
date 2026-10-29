/**
 * jwt-codec 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据：密钥是随手编造的测试值。
 */
export interface JwtCodecExample {
  /** 示例名称。 */
  label: string
  /** 操作：decode / encode / verify。 */
  operation: 'decode' | 'encode' | 'verify'
  /** 令牌文本。 */
  token: string
  /** 头部 JSON 文本。 */
  header: string
  /** 载荷 JSON 文本。 */
  payload: string
  /** HMAC 密钥。 */
  secret: string
  /** 签名算法。 */
  algorithm: 'HS256' | 'HS384' | 'HS512'
}

/** 默认示例：解析一段 JWT。 */
export const DEFAULT_EXAMPLE: JwtCodecExample = {
  label: '解析 JWT',
  operation: 'decode',
  token:
    'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0b29sYml0IiwibmFtZSI6ImRlbW8iLCJpYXQiOjE3MzU2ODk2MDAsImV4cCI6MTczNTY4OTYwMH0.D9MwAs0iXS3XR6s3EQSUnyuDB7oTVqLuabXe20IBKR8',
  header: '',
  payload: '',
  secret: '',
  algorithm: 'HS256',
}

/** 生成示例：用测试密钥签出一段令牌。 */
export const ENCODE_EXAMPLE: JwtCodecExample = {
  label: '生成 JWT',
  operation: 'encode',
  token: '',
  header: '{"alg":"HS256"}',
  payload: '{"sub":"42","name":"demo"}',
  secret: 'toolbit-demo-secret',
  algorithm: 'HS256',
}

/** 校验示例：校验默认示例中的令牌。 */
export const VERIFY_EXAMPLE: JwtCodecExample = {
  label: '校验签名',
  operation: 'verify',
  token:
    'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0b29sYml0IiwibmFtZSI6ImRlbW8iLCJpYXQiOjE3MzU2ODk2MDAsImV4cCI6MTczNTY4OTYwMH0.D9MwAs0iXS3XR6s3EQSUnyuDB7oTVqLuabXe20IBKR8',
  header: '',
  payload: '',
  secret: 'toolbit-demo-secret',
  algorithm: 'HS256',
}

/** 全部示例。 */
export const EXAMPLES: readonly JwtCodecExample[] = [
  DEFAULT_EXAMPLE,
  ENCODE_EXAMPLE,
  VERIFY_EXAMPLE,
]
