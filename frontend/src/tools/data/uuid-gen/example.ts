/**
 * uuid-gen 工具的示例数据。
 *
 * 只用于演示，**不含**任何敏感材料。
 */

/** UUID 版本。 */
export type UUIDVersion = 1 | 4 | 7

/** 一个 UUID 生成示例。 */
export interface UUIDExample {
  /** 示例名称。 */
  label: string
  /** UUID 版本。 */
  version: UUIDVersion
  /** 生成数量。 */
  count: number
  /** 是否输出大写。 */
  uppercase: boolean
  /** 是否保留连字符。 */
  hyphens: boolean
  /** 是否加花括号。 */
  braces: boolean
}

/** 默认示例：生成一个 v4。 */
export const DEFAULT_EXAMPLE: UUIDExample = {
  label: '生成一个 v4',
  version: 4,
  count: 1,
  uppercase: false,
  hyphens: true,
  braces: false,
}

/** 批量生成 v7 的示例。 */
export const BATCH_EXAMPLE: UUIDExample = {
  label: '批量生成 v7',
  version: 7,
  count: 5,
  uppercase: false,
  hyphens: true,
  braces: false,
}

/** 无连字符的大写示例。 */
export const COMPACT_EXAMPLE: UUIDExample = {
  label: '大写且无连字符',
  version: 4,
  count: 1,
  uppercase: true,
  hyphens: false,
  braces: true,
}

/** 全部示例。 */
export const EXAMPLES: readonly UUIDExample[] = [
  DEFAULT_EXAMPLE,
  BATCH_EXAMPLE,
  COMPACT_EXAMPLE,
]
