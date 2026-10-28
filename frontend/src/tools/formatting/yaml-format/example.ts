/**
 * yaml-format 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */
export interface YamlFormatExample {
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
}

/** 默认示例：美化一段 YAML。 */
export const DEFAULT_EXAMPLE: YamlFormatExample = {
  label: '美化 YAML',
  input: '# 注释\nname: toolbit\nversion: 1\ntags:\n  - cli\n  - web\n',
  action: 'format',
  indent: 2,
  sortKeys: false,
}

/** 压缩示例：去掉空行与文档分隔符。 */
export const MINIFY_EXAMPLE: YamlFormatExample = {
  label: '压缩 YAML',
  input: 'name: toolbit\n\nversion: 1\n',
  action: 'minify',
  indent: 2,
  sortKeys: false,
}

/** 全部示例。 */
export const EXAMPLES: readonly YamlFormatExample[] = [
  DEFAULT_EXAMPLE,
  MINIFY_EXAMPLE,
]
