/**
 * regex-tester 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */
export interface RegexTesterExample {
  /** 示例名称。 */
  label: string
  /** 正则表达式。 */
  pattern: string
  /** 正则标志。 */
  flags: string[]
  /** 待匹配文本。 */
  input: string
  /** 匹配条数上限；0 表示使用服务端默认值。 */
  maxMatches: number
  /** 捕获组名称，按顺序对应第 1、2 … 组。 */
  groupNames: string[]
}

/** 默认示例：匹配邮箱地址。 */
export const DEFAULT_EXAMPLE: RegexTesterExample = {
  label: '匹配邮箱地址',
  pattern: '(\\w+)@(\\w+)\\.com',
  flags: [],
  input: 'mail: alice@example.com\nmail: bob@example.com',
  maxMatches: 0,
  groupNames: ['user', 'domain'],
}

/** 忽略大小写示例。 */
export const IGNORE_CASE_EXAMPLE: RegexTesterExample = {
  label: '忽略大小写',
  pattern: 'error',
  flags: ['i'],
  input: 'ERROR\nError\nerror',
  maxMatches: 0,
  groupNames: [],
}

/** 提取日志字段示例。 */
export const LOG_FIELDS_EXAMPLE: RegexTesterExample = {
  label: '提取日志字段',
  pattern: '^(\\S+)\\s+(INFO|WARN|ERROR)\\s+(.*)$',
  flags: ['m'],
  input: '2026-10-29 INFO service started\n2026-10-29 ERROR upstream timeout',
  maxMatches: 0,
  groupNames: ['time', 'level', 'message'],
}

/** 全部示例。 */
export const EXAMPLES: readonly RegexTesterExample[] = [
  DEFAULT_EXAMPLE,
  IGNORE_CASE_EXAMPLE,
  LOG_FIELDS_EXAMPLE,
]
