/**
 * html-format 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */
export interface HtmlFormatExample {
  /** 示例名称。 */
  label: string
  /** 输入内容。 */
  input: string
  /** 操作：format / minify / validate。 */
  action: 'format' | 'minify' | 'validate'
  /** 缩进宽度。 */
  indent: number
}

/** 默认示例：美化一段 HTML 片段。 */
export const DEFAULT_EXAMPLE: HtmlFormatExample = {
  label: '美化 HTML',
  input: '<div class="wrap"><p>hello</p><img src="a.png" alt="a"></div>',
  action: 'format',
  indent: 2,
}

/** 压缩示例：去掉标签之间的空白。 */
export const MINIFY_EXAMPLE: HtmlFormatExample = {
  label: '压缩 HTML',
  input: '<div>\n  <p>hello</p>\n</div>',
  action: 'minify',
  indent: 2,
}

/** 全部示例。 */
export const EXAMPLES: readonly HtmlFormatExample[] = [
  DEFAULT_EXAMPLE,
  MINIFY_EXAMPLE,
]
