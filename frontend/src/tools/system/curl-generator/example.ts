/**
 * curl-generator 工具的示例数据。
 *
 * 凭据一律使用**明显的假值**，不含任何真实密钥或口令。
 */
export type HttpMethod = 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE' | 'HEAD' | 'OPTIONS'
export type BodyFormat = 'json' | 'form' | 'raw'
export type AuthType = 'none' | 'basic' | 'bearer'
export type Shell = 'bash' | 'sh' | 'powershell' | 'cmd'

/** 一个请求头。 */
export interface CurlHeaderExample {
  /** 头名称。 */
  name: string
  /** 头值。 */
  value: string
  /** 是否以占位符呈现。 */
  sensitive: boolean
}

export interface CurlGeneratorExample {
  /** 示例名称。 */
  label: string
  /** HTTP 方法。 */
  method: HttpMethod
  /** 请求地址。 */
  url: string
  /** 请求头。 */
  headers: CurlHeaderExample[]
  /** 请求体。 */
  body: string
  /** 请求体类型。 */
  bodyFormat: BodyFormat
  /** 认证方式。 */
  auth: AuthType
  /** Basic 认证用户名。 */
  username: string
  /** Basic 认证口令。 */
  password: string
  /** Bearer Token。 */
  token: string
  /** 是否跳过 TLS 证书校验。 */
  insecure: boolean
  /** 是否跟随重定向。 */
  followRedirects: boolean
  /** 超时秒数；0 表示不设置。 */
  timeout: number
  /** 目标 Shell。 */
  shell: Shell
}

/** 默认示例：一个普通的 GET 请求。 */
export const DEFAULT_EXAMPLE: CurlGeneratorExample = {
  label: '普通 GET 请求',
  method: 'GET',
  url: 'https://api.example.com/v1/status',
  headers: [{ name: 'Accept', value: 'application/json', sensitive: false }],
  body: '',
  bodyFormat: 'json',
  auth: 'none',
  username: '',
  password: '',
  token: '',
  insecure: false,
  followRedirects: false,
  timeout: 0,
  shell: 'bash',
}

/** POST JSON 示例：含一个以占位符呈现的请求头。 */
export const JSON_POST_EXAMPLE: CurlGeneratorExample = {
  label: 'POST JSON 请求',
  method: 'POST',
  url: 'https://api.example.com/v1/items',
  headers: [
    { name: 'Content-Type', value: 'application/json', sensitive: false },
    { name: 'X-Api-Key', value: 'demo-api-key', sensitive: true },
  ],
  body: '{\n  "name": "demo"\n}',
  bodyFormat: 'json',
  auth: 'none',
  username: '',
  password: '',
  token: '',
  insecure: false,
  followRedirects: true,
  timeout: 30,
  shell: 'bash',
}

/** Basic 认证示例：口令是明显的假值。 */
export const BASIC_AUTH_EXAMPLE: CurlGeneratorExample = {
  label: 'Basic 认证请求',
  method: 'GET',
  url: 'https://api.example.com/v1/me',
  headers: [],
  body: '',
  bodyFormat: 'json',
  auth: 'basic',
  username: 'demo-user',
  password: 'demo-password',
  token: '',
  insecure: false,
  followRedirects: false,
  timeout: 0,
  shell: 'bash',
}

/** Bearer Token 示例：Token 是明显的假值。 */
export const BEARER_EXAMPLE: CurlGeneratorExample = {
  label: 'Bearer Token 请求',
  method: 'GET',
  url: 'https://api.example.com/v1/me',
  headers: [],
  body: '',
  bodyFormat: 'json',
  auth: 'bearer',
  username: '',
  password: '',
  token: 'demo-token',
  insecure: false,
  followRedirects: false,
  timeout: 0,
  shell: 'powershell',
}

/** 全部示例。 */
export const EXAMPLES: readonly CurlGeneratorExample[] = [
  DEFAULT_EXAMPLE,
  JSON_POST_EXAMPLE,
  BASIC_AUTH_EXAMPLE,
  BEARER_EXAMPLE,
]
