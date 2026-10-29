/**
 * text-stats 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */

/** 一个示例。 */
export interface TextStatsExample {
  /** 示例名称。 */
  label: string
  /** text。 */
  text: string
}

/** 默认示例。 */
export const DEFAULT_EXAMPLE: TextStatsExample = {
  label: 'text-stats 示例',
  text: 'Hello 世界\n第二行',
}

/** 全部示例。 */
export const EXAMPLES: readonly TextStatsExample[] = [DEFAULT_EXAMPLE]
