/**
 * time-format 工具的示例数据。
 */

/** 一个示例。 */
export interface TimeFormatExample {
  /** 示例名称。 */
  label: string
  /** 查询词：格式标识或关键字。 */
  query: string
}

/** 默认示例：查 unix-ms。 */
export const DEFAULT_EXAMPLE: TimeFormatExample = {
  label: 'Unix 毫秒时间戳',
  query: 'unix-ms',
}

/** 关键字示例。 */
export const KEYWORD_EXAMPLE: TimeFormatExample = {
  label: '按关键字查 RFC',
  query: 'rfc',
}

/** 全部示例。 */
export const EXAMPLES: readonly TimeFormatExample[] = [DEFAULT_EXAMPLE, KEYWORD_EXAMPLE]
