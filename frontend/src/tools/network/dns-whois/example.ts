/**
 * dns-whois 工具的示例数据。
 */

/** 一个示例。 */
export interface DNSWhoisExample {
  /** 示例名称。 */
  label: string
  /** 域名；可为空。 */
  domain: string
  /** IP 地址；可为空。 */
  ip: string
}

/** 默认示例：查询域名注册信息。 */
export const DEFAULT_EXAMPLE: DNSWhoisExample = {
  label: '查询 example.com',
  domain: 'example.com',
  ip: '',
}

/** 全部示例。 */
export const EXAMPLES: readonly DNSWhoisExample[] = [DEFAULT_EXAMPLE]
