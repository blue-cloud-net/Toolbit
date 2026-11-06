/**
 * Nginx 配置生成相关的类型。
 *
 * 与后端 `backend/internal/tools/system/nginx` 与
 * `backend/internal/api` 的响应结构保持一致。
 */

/** Nginx 模块元信息。 */
export interface NginxModuleInfo {
  /** 模块标识。 */
  name: string
  /** 模块标题。 */
  title: string
  /** 启用该模块的编译开关；为空表示默认编译。 */
  buildOption: string
  /** 是否默认编译。 */
  defaultBuild: boolean
  /** 是否属于商业订阅提供的模块。 */
  commercial: boolean
  /** 官方文档链接。 */
  docUrl: string
}

/** Nginx 指令元信息。 */
export interface NginxDirectiveInfo {
  /** 指令唯一标识：模块与指令名。 */
  id: string
  /** 指令名。 */
  name: string
  /** 所属模块。 */
  module: string
  /** 语法提示。 */
  syntax: string
  /** 默认值。 */
  default: string
  /** 可用上下文。 */
  contexts: string[]
  /** 是否是块指令。 */
  block: boolean
  /** 首次出现的版本。 */
  since: string
  /** 废弃的版本。 */
  deprecated: string
  /** 移除的版本。 */
  removed: string
  /** 是否含有仅商业订阅可用的参数。 */
  commercialParams: boolean
  /** 在目标版本中是否可用。 */
  available: boolean
  /** 在目标版本中是否已废弃。 */
  deprecatedNow: boolean
  /** 官方文档链接。 */
  docUrl: string
  /** 中文简述。 */
  summaryZh: string
  /** 英文简述。 */
  summaryEn: string
}

/** 指令目录端点的响应数据。 */
export interface NginxDirectivesData {
  /** 本次响应对应的目标版本。 */
  version: string
  /** 可选的目标版本列表。 */
  versions: string[]
  /** 快照对应的官方文档版本。 */
  docsVersion: string
  /** 官方文档地址。 */
  source: string
  /** 模块列表。 */
  modules: NginxModuleInfo[]
  /** 指令列表。 */
  directives: NginxDirectiveInfo[]
}

/** 配置树节点类型。 */
export type NginxNodeType = 'directive' | 'block' | 'raw'

/** 配置树节点。 */
export interface NginxNode {
  /** 节点标识。 */
  id: string
  /** 节点类型。 */
  type: NginxNodeType
  /** 所属上下文。 */
  context: string
  /** 指令名或块名。 */
  name: string
  /** 原样参数。 */
  args: string
  /** 原样片段内容。 */
  body: string
  /** 目标文件；留空表示跟随所在层级。 */
  file: string
  /** 注释。 */
  comment: string
  /** 子节点。 */
  children: NginxNode[]
}

/** 简单模式的一个站点。 */
export interface NginxSite {
  /** 站点标识。 */
  id: string
  /** 配置文件名；留空表示按域名自动生成。 */
  fileName: string
  /** 场景预设；留空表示沿用请求级预设。 */
  preset: string
  /** 站点域名。 */
  serverName: string
  /** 监听端口。 */
  listenPort: number
  /** 是否同时监听 IPv6。 */
  listenIPv6: boolean
  /** 站点根目录。 */
  root: string
  /** 索引文件列表。 */
  indexFiles: string[]
  /** 是否启用 HTTPS。 */
  enableSsl: boolean
  /** 证书路径。 */
  sslCertificate: string
  /** 私钥路径。 */
  sslPrivateKey: string
  /** TLS 档位。 */
  tlsProfile: string
  /** 是否启用反向代理。 */
  enableProxy: boolean
  /** 代理目标。 */
  proxyPass: string
  /** 代理路径。 */
  proxyPath: string
  /** 是否启用 WebSocket。 */
  enableWebSocket: boolean
  /** 是否启用负载均衡。 */
  enableLoadBalance: boolean
  /** 后端地址列表。 */
  upstreams: string[]
  /** 是否启用压缩。 */
  enableGzip: boolean
  /** 是否启用静态缓存。 */
  enableCache: boolean
  /** 是否输出安全响应头。 */
  enableSecurityHeaders: boolean
  /** 是否输出错误页。 */
  enableErrorPages: boolean
  /** 错误页目录。 */
  errorPageDir: string
  /** 是否输出 FastCGI 配置。 */
  enablePhp: boolean
  /** FastCGI 地址。 */
  phpSocket: string
}

/** 生成结果中的一个文件。 */
export interface NginxFile {
  /** 文件路径。 */
  path: string
  /** 文件内容。 */
  content: string
  /** 文件字节数。 */
  size: number
  /** 文件用途标识。 */
  description: string
}

/** 打包产物。 */
export interface NginxArchive {
  /** 建议的下载文件名。 */
  filename: string
  /** MIME 类型。 */
  contentType: string
  /** ZIP 内容的 Base64 编码。 */
  base64: string
  /** 压缩包内的条目数。 */
  entries: number
  /** 解压后的总字节数。 */
  uncompressedBytes: number
}

/** 生成诊断。 */
export interface NginxDiagnostic {
  /** 级别：info / warning / error。 */
  level: 'info' | 'warning' | 'error'
  /** 诊断码。 */
  code: string
  /** 相关上下文。 */
  context: string
  /** 相关节点标识。 */
  nodeId: string
  /** 说明文案。 */
  message: string
}

/** 结果摘要。 */
export interface NginxSummary {
  /** 站点数。 */
  siteCount: number
  /** 文件数。 */
  fileCount: number
  /** 配置树节点数。 */
  nodeCount: number
  /** 生成的指令条数。 */
  directiveCount: number
  /** 警告数。 */
  warningCount: number
  /** 错误数。 */
  errorCount: number
}

/** nginx-config 工具的响应数据。 */
export interface NginxConfigResult {
  /** 配置模式。 */
  mode: 'simple' | 'advanced'
  /** 目标版本。 */
  targetVersion: string
  /** 输出形态。 */
  output: 'bundle' | 'single'
  /** 生成的文件列表。 */
  files: NginxFile[]
  /** 打包产物。 */
  archive?: NginxArchive
  /** 生成诊断。 */
  diagnostics: NginxDiagnostic[]
  /** 结果摘要。 */
  summary: NginxSummary
}
