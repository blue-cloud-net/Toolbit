/**
 * markdown-convert 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */

/** 一个示例。 */
export interface MarkdownConvertExample {
  /** 示例名称。 */
  label: string
  /** input。 */
  input: string
}

/** 默认示例。 */
export const DEFAULT_EXAMPLE: MarkdownConvertExample = {
  label: 'markdown-convert 示例',
  input: '# Title\n\nSome **bold** text.',
}

/** 全部示例。 */
export const EXAMPLES: readonly MarkdownConvertExample[] = [DEFAULT_EXAMPLE]
