/**
 * signature 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实密钥或敏感材料。
 */
export interface SignatureExample {
  /** 示例名称。 */
  label: string
  /** 算法。 */
  algorithm: 'SM2' | 'RSA' | 'ECDSA' | 'ED25519' | 'ED448'
  /** 操作。 */
  operation: 'sign' | 'verify'
  /** 消息。 */
  message: string
  /** 消息编码。 */
  messageEncoding: 'utf8' | 'hex'
}

/** 默认示例：SM2 签名。 */
export const DEFAULT_EXAMPLE: SignatureExample = {
  label: '用 SM2 签名一段文本',
  algorithm: 'SM2',
  operation: 'sign',
  message: '待签名内容',
  messageEncoding: 'utf8',
}

/** Ed25519 示例。 */
export const ED25519_EXAMPLE: SignatureExample = {
  label: '用 Ed25519 签名一段文本',
  algorithm: 'ED25519',
  operation: 'sign',
  message: '待签名内容',
  messageEncoding: 'utf8',
}

/** 全部示例。 */
export const EXAMPLES: readonly SignatureExample[] = [DEFAULT_EXAMPLE, ED25519_EXAMPLE]
