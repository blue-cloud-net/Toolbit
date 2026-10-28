/**
 * password-gen 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实密钥或敏感材料。
 */
export interface PasswordGenExample {
  /** 示例名称。 */
  label: string
  /** 长度。 */
  length: number
  /** 数量。 */
  count: number
  /** 字符集集合。 */
  charsets: readonly string[]
  /** 是否排除易混字符。 */
  excludeAmbiguous: boolean
  /** 是否要求每类字符至少一个。 */
  requireEachClass: boolean
}

/** 默认示例。 */
export const DEFAULT_EXAMPLE: PasswordGenExample = {
  label: '生成 16 位强密码',
  length: 16,
  count: 1,
  charsets: ['lower', 'upper', 'digit', 'symbol'],
  excludeAmbiguous: false,
  requireEachClass: true,
}

/** 批量示例。 */
export const BATCH_EXAMPLE: PasswordGenExample = {
  label: '批量生成 10 个 24 位密码',
  length: 24,
  count: 10,
  charsets: ['lower', 'upper', 'digit', 'symbol'],
  excludeAmbiguous: true,
  requireEachClass: true,
}

/** 全部示例。 */
export const EXAMPLES: readonly PasswordGenExample[] = [DEFAULT_EXAMPLE, BATCH_EXAMPLE]
