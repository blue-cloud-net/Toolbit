/**
 * toml-format 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */
export interface TomlFormatExample {
  /** 示例名称。 */
  label: string
  /** 输入内容。 */
  input: string
  /** 操作：format / minify / validate。 */
  action: 'format' | 'minify' | 'validate'
  /** 缩进宽度。 */
  indent: number
}

/** 默认示例：美化一段 TOML。 */
export const DEFAULT_EXAMPLE: TomlFormatExample = {
  label: '美化 TOML',
  input: 'name = "toolbit"\nversion = 1\n\n[tags]\ncli = true\n',
  action: 'format',
  indent: 2,
}

/** 压缩示例：去掉空行。 */
export const MINIFY_EXAMPLE: TomlFormatExample = {
  label: '压缩 TOML',
  input: 'name = "toolbit"\n\nversion = 1\n',
  action: 'minify',
  indent: 2,
}

/** 全部示例。 */
export const EXAMPLES: readonly TomlFormatExample[] = [
  DEFAULT_EXAMPLE,
  MINIFY_EXAMPLE,
]
