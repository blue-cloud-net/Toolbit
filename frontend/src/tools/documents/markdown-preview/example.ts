/**
 * markdown-preview 工具的示例数据。
 *
 * 只用于演示，**不含**任何敏感材料。
 */

/** Markdown 方言。 */
export type MarkdownFlavor = 'gfm' | 'commonmark'

/** 一个 Markdown 预览示例。 */
export interface MarkdownPreviewExample {
  /** 示例名称。 */
  label: string
  /** 方言。 */
  flavor: MarkdownFlavor
  /** 源文。 */
  markdown: string
  /** 是否允许内嵌 HTML。 */
  allowHtml: boolean
  /** 是否把单换行视为换行。 */
  breaks: boolean
}

/** 默认示例：标题、列表与表格。 */
export const DEFAULT_EXAMPLE: MarkdownPreviewExample = {
  label: '标题与表格',
  flavor: 'gfm',
  markdown: [
    '# Toolbit 使用手册',
    '',
    '## 快速开始',
    '',
    '1. 打开首页',
    '2. 选择工具',
    '3. 粘贴内容并运行',
    '',
    '| 参数 | 说明 |',
    '| --- | --- |',
    '| slug | 工具标识 |',
  ].join('\n'),
  allowHtml: false,
  breaks: false,
}

/** CommonMark 示例。 */
export const COMMONMARK_EXAMPLE: MarkdownPreviewExample = {
  label: 'CommonMark 方言',
  flavor: 'commonmark',
  markdown: [
    '# CommonMark',
    '',
    '这段文字里的 **加粗** 与 `行内代码` 都按标准 CommonMark 渲染。',
  ].join('\n'),
  allowHtml: false,
  breaks: false,
}

/** 全部示例。 */
export const EXAMPLES: readonly MarkdownPreviewExample[] = [
  DEFAULT_EXAMPLE,
  COMMONMARK_EXAMPLE,
]
