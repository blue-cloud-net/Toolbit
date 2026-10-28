/**
 * crontab-generator 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */
export type CronPreset =
  | 'every-minute'
  | 'every-5-minutes'
  | 'hourly'
  | 'daily'
  | 'weekly'
  | 'monthly'
  | 'yearly'

export interface CrontabGeneratorExample {
  /** 示例名称。 */
  label: string
  /** 预设名；空串表示自定义字段。 */
  preset: CronPreset | ''
  /** 分钟字段。 */
  minute: string
  /** 小时字段。 */
  hour: string
  /** 日字段。 */
  dayOfMonth: string
  /** 月字段。 */
  month: string
  /** 星期字段。 */
  dayOfWeek: string
  /** IANA 时区名。 */
  timezone: string
  /** 预览的触发次数。 */
  count: number
}

/** 默认示例：每天午夜。 */
export const DEFAULT_EXAMPLE: CrontabGeneratorExample = {
  label: '每天午夜执行',
  preset: 'daily',
  minute: '0',
  hour: '0',
  dayOfMonth: '*',
  month: '*',
  dayOfWeek: '*',
  timezone: 'UTC',
  count: 5,
}

/** 每五分钟示例。 */
export const EVERY_FIVE_EXAMPLE: CrontabGeneratorExample = {
  label: '每五分钟执行',
  preset: 'every-5-minutes',
  minute: '*/5',
  hour: '*',
  dayOfMonth: '*',
  month: '*',
  dayOfWeek: '*',
  timezone: 'Asia/Shanghai',
  count: 5,
}

/** 自定义字段示例。 */
export const CUSTOM_EXAMPLE: CrontabGeneratorExample = {
  label: '自定义字段',
  preset: '',
  minute: '30',
  hour: '2',
  dayOfMonth: '15',
  month: '*',
  dayOfWeek: '*',
  timezone: 'UTC',
  count: 3,
}

/** 全部示例。 */
export const EXAMPLES: readonly CrontabGeneratorExample[] = [
  DEFAULT_EXAMPLE,
  EVERY_FIVE_EXAMPLE,
  CUSTOM_EXAMPLE,
]
