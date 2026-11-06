/**
 * nginx-config 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */

/** 一个示例。 */
export interface NginxConfigExample {
  /** 示例名称。 */
  label: string
  /** 场景预设。 */
  preset: string
  /** 站点域名。 */
  serverName: string
  /** 站点根目录。 */
  root: string
}

/** 默认示例。 */
export const DEFAULT_EXAMPLE: NginxConfigExample = {
  label: 'nginx-config 示例',
  preset: 'static',
  serverName: 'example.com',
  root: '/var/www/html',
}

/** 全部示例。 */
export const EXAMPLES: readonly NginxConfigExample[] = [
  DEFAULT_EXAMPLE,
  {
    label: '单页应用示例',
    preset: 'spa',
    serverName: 'app.example.com',
    root: '/var/www/app',
  },
  {
    label: 'HTTPS 示例',
    preset: 'tls',
    serverName: 'secure.example.com',
    root: '/var/www/secure',
  },
  {
    label: '反向代理示例',
    preset: 'proxy',
    serverName: 'api.example.com',
    root: '/var/www/api',
  },
]
