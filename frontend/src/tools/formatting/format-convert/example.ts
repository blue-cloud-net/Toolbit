/**
 * format-convert 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */
export interface FormatConvertExample {
  /** 示例名称。 */
  label: string
  /** 输入内容。 */
  input: string
  /** 源格式。 */
  from: 'json' | 'yaml' | 'toml' | 'xml'
  /** 目标格式。 */
  to: 'json' | 'yaml' | 'toml' | 'xml'
  /** 缩进宽度。 */
  indent: number
  /** 是否美化输出。 */
  pretty: boolean
}

/** 默认示例：JSON 转 YAML。 */
export const DEFAULT_EXAMPLE: FormatConvertExample = {
  label: 'JSON 转 YAML',
  input: '{"name":"toolbit","count":3}',
  from: 'json',
  to: 'yaml',
  indent: 2,
  pretty: true,
}

/** YAML 转 JSON 示例。 */
export const YAML_TO_JSON_EXAMPLE: FormatConvertExample = {
  label: 'YAML 转 JSON',
  input: 'name: toolbit\ncount: 3\n',
  from: 'yaml',
  to: 'json',
  indent: 2,
  pretty: true,
}

/** 全部示例。 */
export const EXAMPLES: readonly FormatConvertExample[] = [
  DEFAULT_EXAMPLE,
  YAML_TO_JSON_EXAMPLE,
]
