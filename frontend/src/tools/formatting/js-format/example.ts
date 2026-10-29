/**
 * js-format 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */
export interface JsFormatExample {
  /** 示例名称。 */
  label: string
  /** 输入内容。 */
  input: string
  /** 操作：format / minify / validate。 */
  action: 'format' | 'minify' | 'validate'
  /** 缩进宽度。 */
  indent: number
  /** 引号风格。 */
  quotes: 'double' | 'single'
  /** 是否保留分号。 */
  semi: boolean
}

/** 默认示例：美化一段 JavaScript。 */
export const DEFAULT_EXAMPLE: JsFormatExample = {
  label: '美化 JavaScript',
  input: 'function add(a, b) {\n  return a + b;\n}\n',
  action: 'format',
  indent: 2,
  quotes: 'double',
  semi: true,
}

/** 压缩示例：去掉注释与空白。 */
export const MINIFY_EXAMPLE: JsFormatExample = {
  label: '压缩 JavaScript',
  input: 'function add(a, b) {\n  // 求和\n  return a + b;\n}\n',
  action: 'minify',
  indent: 2,
  quotes: 'double',
  semi: true,
}

/** 全部示例。 */
export const EXAMPLES: readonly JsFormatExample[] = [
  DEFAULT_EXAMPLE,
  MINIFY_EXAMPLE,
]
