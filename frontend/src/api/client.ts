/**
 * API 客户端。
 *
 * 后端通信的**唯一**入口：组件内禁止直接 `fetch`。
 * 统一拆信封，并把 `error.code` 映射为本地化文案。
 */
import type { Envelope, Meta, Warning } from '@/types/api'
import type { CatalogResponse } from '@/types/tools'

/** API 基地址；默认与前端同源。 */
const API_BASE = import.meta.env.VITE_API_BASE ?? ''

/** 默认超时（毫秒）。 */
export const DEFAULT_TIMEOUT_MS = 30_000

/**
 * API 错误。
 *
 * 携带机器可读错误码与可插值参数，由调用方本地化。
 */
export class ApiError extends Error {
  /** HTTP 状态码。 */
  readonly status: number
  /** 机器可读错误码。 */
  readonly code: string
  /** 可插值参数。 */
  readonly params: Record<string, unknown>
  /** 请求标识。 */
  readonly requestId: string

  constructor(
    status: number,
    code: string,
    message: string,
    params: Record<string, unknown> = {},
    requestId = '',
  ) {
    super(message)
    this.name = 'ApiError'
    this.status = status
    this.code = code
    this.params = params
    this.requestId = requestId
  }

  /** 是否为「用户主动取消」。 */
  get isCanceled(): boolean {
    return this.code === 'CANCELED'
  }
}

/** 请求选项。 */
export interface RequestOptions {
  /** 外部取消信号。 */
  signal?: AbortSignal
  /** 超时（毫秒）。 */
  timeoutMs?: number
  /** 附加请求头。 */
  headers?: Record<string, string>
}

/** 工具执行结果。 */
export interface ToolCallResult<T = unknown> {
  /** 业务数据。 */
  data: T
  /** 执行元信息。 */
  meta: Meta
}

/** 把 AbortError 归一化成 CANCELED 错误码。 */
function toCancelError(): ApiError {
  return new ApiError(0, 'CANCELED', 'request canceled')
}

/** 把任意异常归一化成 ApiError。 */
function toApiError(err: unknown): ApiError {
  if (err instanceof ApiError) {
    return err
  }
  if (err instanceof DOMException && err.name === 'AbortError') {
    return toCancelError()
  }
  return new ApiError(0, 'NETWORK_ERROR', err instanceof Error ? err.message : 'network error')
}

/**
 * 合并取消信号与超时。
 *
 * 超时用独立的 AbortController，与外部信号合并：
 * 外部取消时不再触发超时定时器。
 */
function withTimeout(
  signal: AbortSignal | undefined,
  timeoutMs: number,
): { signal: AbortSignal; cleanup: () => void } {
  const controller = new AbortController()
  const timer = setTimeout(() => controller.abort(new DOMException('timeout', 'TimeoutError')), timeoutMs)

  const onAbort = () => controller.abort(signal?.reason)
  if (signal) {
    if (signal.aborted) {
      controller.abort(signal.reason)
    } else {
      signal.addEventListener('abort', onAbort, { once: true })
    }
  }

  return {
    signal: controller.signal,
    cleanup: () => {
      clearTimeout(timer)
      signal?.removeEventListener('abort', onAbort)
    },
  }
}

/**
 * 执行一次 API 调用并拆信封。
 *
 * 后端约定：HTTP 状态码等于信封的 `code` 字段。
 */
async function request<T>(
  path: string,
  init: RequestInit,
  options: RequestOptions = {},
): Promise<ToolCallResult<T>> {
  const timeoutMs = options.timeoutMs ?? DEFAULT_TIMEOUT_MS
  const { signal, cleanup } = withTimeout(options.signal, timeoutMs)

  let response: Response
  try {
    response = await fetch(`${API_BASE}${path}`, {
      ...init,
      signal,
      headers: {
        Accept: 'application/json',
        ...options.headers,
        ...init.headers,
      },
    })
  } catch (err) {
    cleanup()
    if (signal.aborted) {
      if (options.signal?.aborted) {
        throw toCancelError()
      }
      throw new ApiError(0, 'TIMEOUT', 'request timeout')
    }
    throw toApiError(err)
  } finally {
    cleanup()
  }

  let envelope: Envelope<T>
  try {
    envelope = (await response.json()) as Envelope<T>
  } catch {
    throw new ApiError(
      response.status,
      'INVALID_RESPONSE',
      `unexpected response body (status ${response.status})`,
    )
  }

  if (!response.ok || envelope.code !== 0 || envelope.error !== null) {
    const payload = envelope.error
    throw new ApiError(
      envelope.code || response.status,
      payload?.code ?? 'INTERNAL_ERROR',
      envelope.message || 'request failed',
      payload?.params ?? {},
      envelope.requestId,
    )
  }

  const meta: Meta = envelope.meta ?? {
    durationMs: 0,
    inputBytes: 0,
    outputBytes: 0,
    warnings: [],
  }

  return { data: envelope.data as T, meta }
}

/** 拉取工具目录。 */
export function fetchCatalog(options: RequestOptions = {}): Promise<ToolCallResult<CatalogResponse>> {
  return request<CatalogResponse>('/api/v1/tools', { method: 'GET' }, options)
}

/**
 * 执行一个工具。
 *
 * 请求体为 JSON；multipart 与二进制工具由后续版本扩展。
 */
export function callTool<T = unknown>(
  slug: string,
  payload: unknown,
  options: RequestOptions = {},
): Promise<ToolCallResult<T>> {
  return request<T>(
    `/api/v1/tools/${encodeURIComponent(slug)}`,
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    },
    options,
  )
}

/** 提取警告码列表。 */
export function warningCodes(warnings: Warning[] | undefined): string[] {
  if (!warnings) {
    return []
  }
  return warnings.map((w) => w.code)
}
