/**
 * dns-lookup 工具的示例数据。
 */
export interface DNSLookupExample {
  /** 示例名称。 */
  label: string
  /** 域名。 */
  host: string
  /** 记录类型。 */
  type: string
}

/** 默认示例：查询 A 记录。 */
export const DEFAULT_EXAMPLE: DNSLookupExample = {
  label: '查询 A 记录',
  host: 'example.com',
  type: 'A',
}

/** 全部示例。 */
export const EXAMPLES: readonly DNSLookupExample[] = [DEFAULT_EXAMPLE]
