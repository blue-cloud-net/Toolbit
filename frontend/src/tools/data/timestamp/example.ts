/**
 * timestamp 工具的示例数据。
 *
 * 只用于演示，**不含**任何敏感材料。
 */

/** 转换方向。 */
export type TimestampOperation = 'toDate' | 'toTimestamp'

/** 时间戳单位。 */
export type TimestampUnit = 's' | 'ms' | 'us' | 'ns'

/** 一个时间戳示例。 */
export interface TimestampExample {
  /** 示例名称。 */
  label: string
  /** 转换方向。 */
  operation: TimestampOperation
  /** 输入值：时间戳或日期文本。 */
  value: string
  /** 时间戳单位。 */
  unit: TimestampUnit
  /** IANA 时区名；空串表示交由服务端按 UTC 处理。 */
  timezone: string
}

/** 默认示例：秒级时间戳转日期。 */
export const DEFAULT_EXAMPLE: TimestampExample = {
  label: '秒时间戳转日期',
  operation: 'toDate',
  value: '1700000000',
  unit: 's',
  timezone: 'Asia/Shanghai',
}

/** 毫秒级时间戳示例。 */
export const MILLISECOND_EXAMPLE: TimestampExample = {
  label: '毫秒时间戳转日期',
  operation: 'toDate',
  value: '1700000000123',
  unit: 'ms',
  timezone: 'Asia/Shanghai',
}

/** 日期转时间戳示例。 */
export const TO_TIMESTAMP_EXAMPLE: TimestampExample = {
  label: '日期转时间戳',
  operation: 'toTimestamp',
  value: '2023-11-14',
  unit: 's',
  timezone: 'UTC',
}

/** 全部示例。 */
export const EXAMPLES: readonly TimestampExample[] = [
  DEFAULT_EXAMPLE,
  MILLISECOND_EXAMPLE,
  TO_TIMESTAMP_EXAMPLE,
]
