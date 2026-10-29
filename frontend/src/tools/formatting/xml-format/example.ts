/**
 * xml-format 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */
export interface XmlFormatExample {
  /** 示例名称。 */
  label: string
  /** 输入内容。 */
  input: string
  /** 操作：format / minify / validate。 */
  action: 'format' | 'minify' | 'validate'
  /** 缩进宽度。 */
  indent: number
  /** 是否保留注释。 */
  keepComments: boolean
}

/** 默认示例：美化一段 XML。 */
export const DEFAULT_EXAMPLE: XmlFormatExample = {
  label: '美化 XML',
  input: '<root><name>toolbit</name><item id="1">first</item></root>',
  action: 'format',
  indent: 2,
  keepComments: false,
}

/** 压缩示例：去掉标签之间的空白。 */
export const MINIFY_EXAMPLE: XmlFormatExample = {
  label: '压缩 XML',
  input: '<root>\n  <name>toolbit</name>\n</root>',
  action: 'minify',
  indent: 2,
  keepComments: false,
}

/** 全部示例。 */
export const EXAMPLES: readonly XmlFormatExample[] = [
  DEFAULT_EXAMPLE,
  MINIFY_EXAMPLE,
]
