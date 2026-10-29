/**
 * nginx-config 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */

/** 一个示例。 */
export interface NginxConfigExample {
  /** 示例名称。 */
  label: string
  /** serverName。 */
  serverName: string
  /** root。 */
  root: string
}

/** 默认示例。 */
export const DEFAULT_EXAMPLE: NginxConfigExample = {
  label: 'nginx-config 示例',
  serverName: 'example.com',
  root: '/var/www/html',
}

/** 全部示例。 */
export const EXAMPLES: readonly NginxConfigExample[] = [DEFAULT_EXAMPLE]
