/**
 * exit-code 工具的示例数据。
 */

/** 一个示例。 */
export interface ExitCodeExample {
  /** 示例名称。 */
  label: string
  /** 查询词：退出码或关键字。 */
  query: string
}

/** 默认示例：查 127。 */
export const DEFAULT_EXAMPLE: ExitCodeExample = {
  label: '127 命令未找到',
  query: '127',
}

/** 关键字示例。 */
export const KEYWORD_EXAMPLE: ExitCodeExample = {
  label: '按关键字查段错误',
  query: '段错误',
}

/** 全部示例。 */
export const EXAMPLES: readonly ExitCodeExample[] = [DEFAULT_EXAMPLE, KEYWORD_EXAMPLE]
