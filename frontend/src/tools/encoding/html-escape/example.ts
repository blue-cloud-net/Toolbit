/**
 * html-escape 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */

/** 一个示例。 */
export interface HTMLEscapeExample {
  /** 示例名称。 */
  label: string
  /** input。 */
  input: string
}

/** 默认示例。 */
export const DEFAULT_EXAMPLE: HTMLEscapeExample = {
  label: 'html-escape 示例',
  input: '<b>hello</b> & "world"',
}

/** 全部示例。 */
export const EXAMPLES: readonly HTMLEscapeExample[] = [DEFAULT_EXAMPLE]
