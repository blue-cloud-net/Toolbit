/**
 * number-base 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实密钥或敏感材料。
 */
export interface NumberBaseExample {
  /** 示例名称。 */
  label: string
  /** 待转换的值。 */
  value: string
  /** 源进制。 */
  from: number
  /** 目标进制。 */
  to: number
  /** 是否加进制前缀。 */
  prefix: boolean
}

/** 默认示例：十进制转十六进制。 */
export const DEFAULT_EXAMPLE: NumberBaseExample = {
  label: '十进制转十六进制',
  value: '255',
  from: 10,
  to: 16,
  prefix: false,
}

/** 全部示例。 */
export const EXAMPLES: readonly NumberBaseExample[] = [
  DEFAULT_EXAMPLE,
  {
    label: '十六进制转二进制',
    value: 'deadbeef',
    from: 16,
    to: 2,
    prefix: false,
  },
]
