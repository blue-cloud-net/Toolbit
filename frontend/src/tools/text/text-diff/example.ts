/**
 * text-diff 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */

/** 一个示例。 */
export interface TextDiffExample {
  /** 示例名称。 */
  label: string
  /** mode。 */
  mode: string
  /** left。 */
  left: string
  /** right。 */
  right: string
}

/** 默认示例。 */
export const DEFAULT_EXAMPLE: TextDiffExample = {
  label: 'text-diff 示例',
  mode: 'text',
  left: 'line A\nline B',
  right: 'line A\nline C',
}

/** 全部示例。 */
export const EXAMPLES: readonly TextDiffExample[] = [DEFAULT_EXAMPLE]
