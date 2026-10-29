/**
 * cn-regions 工具的示例数据。
 *
 * 只用于演示，**不含**任何敏感材料。
 */

/** 一个行政区划查询示例。 */
export interface CNRegionsExample {
  /** 示例名称。 */
  label: string
  /** 名称或代码关键词。 */
  query: string
  /** 省级行政区名称或代码。 */
  province: string
}

/** 默认示例：按名称查询城市。 */
export const DEFAULT_EXAMPLE: CNRegionsExample = {
  label: '按名称查询城市',
  query: '杭州',
  province: '',
}

/** 按省份限定范围的示例。 */
export const PROVINCE_EXAMPLE: CNRegionsExample = {
  label: '按省份限定范围',
  query: '',
  province: '浙江省',
}

/** 按省份代码限定范围的示例。 */
export const PROVINCE_CODE_EXAMPLE: CNRegionsExample = {
  label: '按省份代码限定范围',
  query: '',
  province: '110000',
}

/** 全部示例。 */
export const EXAMPLES: readonly CNRegionsExample[] = [
  DEFAULT_EXAMPLE,
  PROVINCE_EXAMPLE,
  PROVINCE_CODE_EXAMPLE,
]
