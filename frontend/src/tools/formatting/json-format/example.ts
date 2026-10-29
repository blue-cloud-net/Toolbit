/**
 * json-format 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */
export interface JsonFormatExample {
  /** 示例名称。 */
  label: string
  /** 输入内容。 */
  input: string
  /** 操作：format / minify / validate。 */
  action: 'format' | 'minify' | 'validate'
  /** 缩进宽度。 */
  indent: number
  /** 是否按键排序。 */
  sortKeys: boolean
  /** 是否转义非 ASCII 字符。 */
  escapeUnicode: boolean
}

/** 默认示例：美化一段 JSON。 */
export const DEFAULT_EXAMPLE: JsonFormatExample = {
  label: '美化 JSON',
  input: '{"name":"toolbit","version":1,"tags":["cli","web"]}',
  action: 'format',
  indent: 2,
  sortKeys: false,
  escapeUnicode: false,
}

/** 压缩示例：去掉多余空白。 */
export const MINIFY_EXAMPLE: JsonFormatExample = {
  label: '压缩 JSON',
  input: '{\n  "a" : 1\n}',
  action: 'minify',
  indent: 2,
  sortKeys: false,
  escapeUnicode: false,
}

/** 排序示例：按键名排序。 */
export const SORT_EXAMPLE: JsonFormatExample = {
  label: '按键排序',
  input: '{"b":1,"a":2,"c":3}',
  action: 'format',
  indent: 4,
  sortKeys: true,
  escapeUnicode: false,
}

/** 全部示例。 */
export const EXAMPLES: readonly JsonFormatExample[] = [
  DEFAULT_EXAMPLE,
  MINIFY_EXAMPLE,
  SORT_EXAMPLE,
]
