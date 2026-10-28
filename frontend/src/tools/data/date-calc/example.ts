/**
 * date-calc 工具的示例数据。
 *
 * 只用于演示，**不含**任何敏感材料。
 */

/** 计算操作。 */
export type DateOperation = 'diff' | 'add' | 'sub' | 'workdays'

/** 增减单位。 */
export type DateUnit = 'day' | 'week' | 'month' | 'year'

/** 一个日期计算示例。 */
export interface DateCalcExample {
  /** 示例名称。 */
  label: string
  /** 操作。 */
  operation: DateOperation
  /** 起始日期。 */
  dateA: string
  /** 结束日期；`diff` 与 `workdays` 时使用。 */
  dateB: string
  /** 增减数量；`add` 与 `sub` 时使用。 */
  days: number
  /** 增减单位。 */
  unit: DateUnit
  /** 节假日列表。 */
  holidays: string[]
}

/** 默认示例：计算相差天数。 */
export const DEFAULT_EXAMPLE: DateCalcExample = {
  label: '计算相差天数',
  operation: 'diff',
  dateA: '2024-01-01',
  dateB: '2024-01-31',
  days: 10,
  unit: 'day',
  holidays: [],
}

/** 加天数示例。 */
export const ADD_EXAMPLE: DateCalcExample = {
  label: '加十天',
  operation: 'add',
  dateA: '2024-01-01',
  dateB: '',
  days: 10,
  unit: 'day',
  holidays: [],
}

/** 减月份示例。 */
export const SUB_EXAMPLE: DateCalcExample = {
  label: '减两个月',
  operation: 'sub',
  dateA: '2024-03-15',
  dateB: '',
  days: 2,
  unit: 'month',
  holidays: [],
}

/** 工作日示例。 */
export const WORKDAYS_EXAMPLE: DateCalcExample = {
  label: '统计一周工作日',
  operation: 'workdays',
  dateA: '2024-01-01',
  dateB: '2024-01-07',
  days: 0,
  unit: 'day',
  holidays: ['2024-01-02'],
}

/** 全部示例。 */
export const EXAMPLES: readonly DateCalcExample[] = [
  DEFAULT_EXAMPLE,
  ADD_EXAMPLE,
  SUB_EXAMPLE,
  WORKDAYS_EXAMPLE,
]
