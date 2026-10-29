/**
 * chmod-calc 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */
export type PermBit = 'read' | 'write' | 'execute'
export type PermClass = 'user' | 'group' | 'other'

export interface ChmodCalcExample {
  /** 示例名称。 */
  label: string
  /** 权限：符号形式或八进制形式。 */
  mode: string
  /** 属主权限位；非空时覆盖 `mode` 中的属主取值。 */
  user: PermBit[]
  /** 属组权限位。 */
  group: PermBit[]
  /** 其他人权限位。 */
  other: PermBit[]
}

/** 默认示例：八进制 755。 */
export const DEFAULT_EXAMPLE: ChmodCalcExample = {
  label: '常见八进制权限',
  mode: '755',
  user: [],
  group: [],
  other: [],
}

/** 符号形式示例。 */
export const SYMBOLIC_EXAMPLE: ChmodCalcExample = {
  label: '符号形式权限',
  mode: 'rwxr-x---',
  user: [],
  group: [],
  other: [],
}

/** 权限位覆盖示例。 */
export const OVERRIDE_EXAMPLE: ChmodCalcExample = {
  label: '覆盖其他人权限',
  mode: '755',
  user: [],
  group: [],
  other: ['read'],
}

/** 全部示例。 */
export const EXAMPLES: readonly ChmodCalcExample[] = [
  DEFAULT_EXAMPLE,
  SYMBOLIC_EXAMPLE,
  OVERRIDE_EXAMPLE,
]
