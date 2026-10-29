/**
 * text-convert 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */
export type TextStyle =
  | 'upper'
  | 'lower'
  | 'camel'
  | 'pascal'
  | 'snake'
  | 'kebab'
  | 'constant'
  | 'title'
  | 'sentence'
  | 'dot'
  | 'path'

export interface TextConvertExample {
  /** 示例名称。 */
  label: string
  /** 待转换文本。 */
  input: string
  /** 目标命名风格。 */
  style: TextStyle
}

/** 默认示例：连字符转驼峰。 */
export const DEFAULT_EXAMPLE: TextConvertExample = {
  label: '连字符转驼峰',
  input: 'http-server',
  style: 'camel',
}

/** 转下划线示例。 */
export const SNAKE_EXAMPLE: TextConvertExample = {
  label: '转下划线风格',
  input: 'HTTPServer',
  style: 'snake',
}

/** 转常量示例。 */
export const CONSTANT_EXAMPLE: TextConvertExample = {
  label: '转常量风格',
  input: 'user profile setting',
  style: 'constant',
}

/** 全部示例。 */
export const EXAMPLES: readonly TextConvertExample[] = [
  DEFAULT_EXAMPLE,
  SNAKE_EXAMPLE,
  CONSTANT_EXAMPLE,
]
