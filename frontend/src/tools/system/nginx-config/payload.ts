/**
 * Nginx 配置工具的请求体构造。
 *
 * 与页面分离，便于单测覆盖「模式切换」「输出切换」等行为。
 */
import type { NginxNode, NginxSite } from '@/types/nginx'

/** 配置模式。 */
export type NginxMode = 'simple' | 'advanced'

/** 输出形态。 */
export type NginxOutput = 'bundle' | 'single'

/** 响应形态。 */
export type NginxResponseShape = 'files' | 'bundle'

/** 生成请求体的入参。 */
export interface NginxPayloadInput {
  /** 配置模式。 */
  mode: NginxMode
  /** 输出形态。 */
  output: NginxOutput
  /** 目标版本；空字符串表示跟随文档版本。 */
  targetVersion: string
  /** 场景预设（简单模式）。 */
  preset: string
  /** 站点列表（简单模式）。 */
  sites: NginxSite[]
  /** 配置树（高级模式）。 */
  nodes: NginxNode[]
  /** 响应形态。 */
  responseShape: NginxResponseShape
}

/**
 * 构造工具请求体。
 *
 * 高级模式只提交配置树，简单模式只提交站点列表：
 * 两者由服务端展开成**同一棵配置树**，因此这里不做任何合并。
 */
export function buildNginxPayload(input: NginxPayloadInput): Record<string, unknown> {
  const payload: Record<string, unknown> = {
    mode: input.mode,
    output: input.output,
    responseShape: input.responseShape,
  }

  if (input.targetVersion !== '') {
    payload.targetVersion = input.targetVersion
  }

  if (input.mode === 'advanced') {
    payload.nodes = input.nodes
    return payload
  }

  payload.preset = input.preset
  payload.sites = input.sites
  return payload
}

/**
 * 解析后端地址列表。
 *
 * 逗号分隔，忽略空白项。
 */
export function parseUpstreams(raw: string): string[] {
  return raw
    .split(',')
    .map((item) => item.trim())
    .filter((item) => item !== '')
}

/** 生成站点标识。 */
function siteId(): string {
  return `s${Math.random().toString(36).slice(2, 8)}`
}

/** 创建一个空白站点。 */
export function createNginxSite(): NginxSite {
  return {
    id: siteId(),
    fileName: '',
    preset: '',
    serverName: 'example.com',
    listenPort: 80,
    listenIPv6: true,
    root: '/var/www/html',
    indexFiles: ['index.html'],
    enableSsl: false,
    sslCertificate: '/etc/ssl/certs/example.com.pem',
    sslPrivateKey: '/etc/ssl/private/example.com.key',
    tlsProfile: 'modern',
    enableProxy: false,
    proxyPass: 'http://127.0.0.1:8080',
    proxyPath: '/',
    enableWebSocket: false,
    enableLoadBalance: false,
    upstreams: [],
    enableGzip: false,
    enableCache: false,
    enableSecurityHeaders: false,
    enableErrorPages: false,
    errorPageDir: '/var/www/errors',
    enablePhp: false,
    phpSocket: 'unix:/run/php/php8.2-fpm.sock',
  }
}
