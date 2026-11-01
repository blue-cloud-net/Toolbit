/**
 * dns-servers 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */

/** 一个示例。 */
export interface DNSServersExample {
  /** 示例名称。 */
  label: string
  /** 搜索词：服务商名（中英文）或地址。 */
  query: string
}

/** 默认示例：按中文服务商名搜索。 */
export const DEFAULT_EXAMPLE: DNSServersExample = {
  label: '阿里公共 DNS',
  query: '阿里',
}

/** 按英文品牌名搜索。 */
export const ENGLISH_EXAMPLE: DNSServersExample = {
  label: 'Cloudflare',
  query: 'Cloudflare',
}

/** 按地址搜索。 */
export const ADDRESS_EXAMPLE: DNSServersExample = {
  label: '按地址查',
  query: '1.1.1.1',
}

/** 全部示例。 */
export const EXAMPLES: readonly DNSServersExample[] = [
  DEFAULT_EXAMPLE,
  ENGLISH_EXAMPLE,
  ADDRESS_EXAMPLE,
]
