/**
 * ecdh 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实密钥或敏感材料。
 */
export interface ECDHExample {
  /** 示例名称。 */
  label: string
  /** 曲线。 */
  curve: string
  /** 是否启用 KDF 派生。 */
  useKDF: boolean
  /** KDF 摘要。 */
  hash: string
  /** 派生长度。 */
  length: number
}

/** 默认示例：P-256 协商。 */
export const DEFAULT_EXAMPLE: ECDHExample = {
  label: 'P-256 协商共享密钥',
  curve: 'P-256',
  useKDF: false,
  hash: 'SHA256',
  length: 32,
}

/** 带 KDF 的示例。 */
export const KDF_EXAMPLE: ECDHExample = {
  label: 'P-256 协商并用 HKDF 派生',
  curve: 'P-256',
  useKDF: true,
  hash: 'SHA256',
  length: 32,
}

/** 全部示例。 */
export const EXAMPLES: readonly ECDHExample[] = [DEFAULT_EXAMPLE, KDF_EXAMPLE]
