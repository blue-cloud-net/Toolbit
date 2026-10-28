/**
 * hmac-gen 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实密钥或敏感材料。
 */
export interface HmacGenExample {
  /** 示例名称。 */
  label: string
  /** HMAC 算法。 */
  algorithm: string
  /** 密钥内容。 */
  key: string
  /** 密钥编码。 */
  keyEncoding: 'utf8' | 'hex'
  /** 消息内容。 */
  message: string
  /** 消息编码。 */
  messageEncoding: 'utf8' | 'hex'
}

/** 默认示例：HMAC-SM3。 */
export const DEFAULT_EXAMPLE: HmacGenExample = {
  label: '用 HMAC-SM3 计算认证码',
  algorithm: 'HMAC-SM3',
  key: '6b6579',
  keyEncoding: 'hex',
  message: 'The quick brown fox jumps over the lazy dog',
  messageEncoding: 'utf8',
}

/** HMAC-SHA256 示例。 */
export const SHA256_EXAMPLE: HmacGenExample = {
  label: '用 HMAC-SHA256 计算认证码',
  algorithm: 'HMAC-SHA256',
  key: '6b6579',
  keyEncoding: 'hex',
  message: 'The quick brown fox jumps over the lazy dog',
  messageEncoding: 'utf8',
}

/** 全部示例。 */
export const EXAMPLES: readonly HmacGenExample[] = [DEFAULT_EXAMPLE, SHA256_EXAMPLE]
