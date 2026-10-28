/**
 * keygen 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实密钥或敏感材料。
 */
export interface KeyGenExample {
  /** 示例名称。 */
  label: string
  /** 密钥算法。 */
  algorithm: string
  /** RSA 位数；非 RSA 为 0。 */
  bits: number
  /** 曲线名；非椭圆曲线为空。 */
  curve: string
  /** 是否用口令加密输出。 */
  encrypt: boolean
}

/** 默认示例：RSA 2048。 */
export const DEFAULT_EXAMPLE: KeyGenExample = {
  label: '生成 RSA 2048 密钥对',
  algorithm: 'RSA',
  bits: 2048,
  curve: '',
  encrypt: false,
}

/** 国密示例：SM2。 */
export const SM2_EXAMPLE: KeyGenExample = {
  label: '生成 SM2 密钥对',
  algorithm: 'SM2',
  bits: 0,
  curve: 'SM2',
  encrypt: false,
}

/** 全部示例。 */
export const EXAMPLES: readonly KeyGenExample[] = [DEFAULT_EXAMPLE, SM2_EXAMPLE]
