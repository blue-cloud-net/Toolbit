/**
 * 与后端 DTO 对齐的传输层类型。
 *
 * 字段名必须与 `backend/internal/httpx/envelope.go` 逐字一致。
 */

/** 统一响应信封。 */
export interface Envelope<T = unknown> {
  /** 等于 HTTP 状态码；成功时为 0。 */
  code: number
  /** 英文兜底文案；成功时为 "ok"。 */
  message: string
  /** 业务结果；失败时为 null。 */
  data: T | null
  /** 错误详情；成功时为 null。 */
  error: ErrorPayload | null
  /** 执行元信息。 */
  meta: Meta
  /** 请求标识，用于日志关联。 */
  requestId: string
}

/** 信封中的错误部分。 */
export interface ErrorPayload {
  /** 机器可读错误码，见 docs/api/errors.md。 */
  code: string
  /** 可插值参数，可能缺省。 */
  params?: Record<string, unknown>
}

/** 执行元信息。 */
export interface Meta {
  /** 服务端处理耗时（毫秒）。 */
  durationMs: number
  /** 请求体字节数。 */
  inputBytes: number
  /** 响应体字节数。 */
  outputBytes: number
  /** 警告列表；无警告时为空数组。 */
  warnings: Warning[]
}

/** 警告实例。 */
export interface Warning {
  /** 警告码。 */
  code: string
  /** 可插值参数。 */
  params?: Record<string, unknown>
}

/** 服务元信息（GET /api/v1/meta）。 */
export interface MetaResponse {
  version: string
  commit: string
  buildTime: string
  locales: string[]
  features: string[]
}
