/**
 * dns-tester 工具的示例数据。
 */
export interface DNSTesterExample {
  /** 示例名称。 */
  label: string
  /** 检测项：whois / dnssec / ptr。 */
  action: 'whois' | 'dnssec' | 'ptr'
  /** 域名。 */
  domain: string
  /** IP 地址。 */
  ip: string
}

/** 默认示例：查询 DNSSEC 状态。 */
export const DEFAULT_EXAMPLE: DNSTesterExample = {
  label: '查询 DNSSEC 状态',
  action: 'dnssec',
  domain: 'example.com',
  ip: '',
}

/** 全部示例。 */
export const EXAMPLES: readonly DNSTesterExample[] = [DEFAULT_EXAMPLE]
