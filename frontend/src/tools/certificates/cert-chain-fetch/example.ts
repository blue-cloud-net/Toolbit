/**
 * cert-chain-fetch 工具的示例数据。
 */
export interface ChainFetchExample {
  /** 示例名称。 */
  label: string
  /** 目标主机。 */
  host: string
  /** 目标端口。 */
  port: number
  /** SNI。 */
  serverName: string
}

/** 默认示例：抓取公共站点证书链。 */
export const DEFAULT_EXAMPLE: ChainFetchExample = {
  label: '抓取公共站点证书链',
  host: 'example.com',
  port: 443,
  serverName: '',
}

/** 全部示例。 */
export const EXAMPLES: readonly ChainFetchExample[] = [DEFAULT_EXAMPLE]
