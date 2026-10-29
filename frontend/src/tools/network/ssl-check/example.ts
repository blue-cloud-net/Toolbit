/**
 * ssl-check 工具的示例数据。
 */
export interface SSLExample {
  /** 示例名称。 */
  label: string
  /** 目标主机。 */
  host: string
  /** 目标端口。 */
  port: number
  /** SNI。 */
  sni: string
}

/** 默认示例：检测公共站点。 */
export const DEFAULT_EXAMPLE: SSLExample = {
  label: '检测公共站点',
  host: 'example.com',
  port: 443,
  sni: '',
}

/** 全部示例。 */
export const EXAMPLES: readonly SSLExample[] = [DEFAULT_EXAMPLE]
