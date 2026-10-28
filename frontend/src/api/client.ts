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
 * 请求体为 JSON；二进制（multipart）工具用 `callToolBinary`。
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

/** 二进制工具的调用结果。 */
export interface BinaryCallResult {
  /** 响应体。 */
  blob: Blob
  /** 建议的文件名（来自 Content-Disposition）。 */
  filename: string
  /** 执行元信息。 */
  meta: Meta
}

/** 二进制请求的选项。 */
export interface BinaryRequestOptions extends RequestOptions {
  /** 上传进度回调。 */
  onUploadProgress?: (loaded: number, total: number) => void
}

/**
 * 执行一个返回二进制流的工具。
 *
 * multipart 工具的响应**不套**统一信封：
 *   - 成功 → `application/octet-stream`，警告走 `X-Warnings` 响应头；
 *   - 失败 → 仍是 JSON 信封（错误在响应头发出之前就能确定）。
 *
 * 因此这里按 Content-Type 分流：二进制取 blob，JSON 走信封拆包。
 *
 * ⚠️ 这里用 XMLHttpRequest 而不是 fetch：
 * fetch **不提供**上传进度事件，而本工具可能上传 100 MB，
 * 没有进度用户会以为卡死了。
 */
export function callToolBinary(
  slug: string,
  form: FormData,
  options: BinaryRequestOptions = {},
): Promise<BinaryCallResult> {
  const timeoutMs = options.timeoutMs ?? DEFAULT_TIMEOUT_MS
  const url = `${API_BASE}/api/v1/tools/${encodeURIComponent(slug)}`

  return new Promise<BinaryCallResult>((resolve, reject) => {
    const xhr = new XMLHttpRequest()
    xhr.open('POST', url)
    xhr.responseType = 'blob'
    xhr.timeout = timeoutMs
    xhr.setRequestHeader('Accept', 'application/octet-stream, application/json')

    const signal = options.signal
    const onAbort = (): void => {
      xhr.abort()
    }
    if (signal) {
      if (signal.aborted) {
        reject(toCancelError())
        return
      }
      signal.addEventListener('abort', onAbort, { once: true })
    }
    const cleanup = (): void => {
      signal?.removeEventListener('abort', onAbort)
    }

    xhr.upload.onprogress = (event: ProgressEvent): void => {
      if (event.lengthComputable) {
        options.onUploadProgress?.(event.loaded, event.total)
      }
    }

    xhr.onload = (): void => {
      cleanup()
      const contentType = xhr.getResponseHeader('Content-Type') ?? ''

      if (contentType.includes('application/json')) {
        // 错误路径：后端返回 JSON 信封。
        void readBlobText(xhr.response as Blob)
          .then((text: string) => {
            let envelope: Envelope<unknown>
            try {
              envelope = JSON.parse(text) as Envelope<unknown>
            } catch {
              reject(
                new ApiError(
                  xhr.status,
                  'INVALID_RESPONSE',
                  `unexpected response body (status ${xhr.status})`,
                ),
              )
              return
            }
            const payload = envelope.error
            reject(
              new ApiError(
                envelope.code || xhr.status,
                payload?.code ?? 'INTERNAL_ERROR',
                envelope.message || 'request failed',
                payload?.params ?? {},
                envelope.requestId,
              ),
            )
          })
          .catch(reject)
        return
      }

      if (xhr.status < 200 || xhr.status >= 300) {
        reject(new ApiError(xhr.status, 'INTERNAL_ERROR', `request failed (status ${xhr.status})`))
        return
      }

      resolve({
        blob: xhr.response as Blob,
        filename: filenameFromDisposition(xhr.getResponseHeader('Content-Disposition')),
        meta: metaFromHeaders(xhr),
      })
    }

    xhr.onerror = (): void => {
      cleanup()
      reject(new ApiError(0, 'NETWORK_ERROR', 'network error'))
    }
    xhr.ontimeout = (): void => {
      cleanup()
      reject(new ApiError(0, 'TIMEOUT', 'request timeout'))
    }
    xhr.onabort = (): void => {
      cleanup()
      reject(toCancelError())
    }

    // ⚠️ **不要**手动设置 Content-Type：
    // 浏览器需要自己补 multipart 的 boundary。
    xhr.send(form)
  })
}

/**
 * 读取 Blob 的文本内容。
 *
 * 用 `FileReader` 而不是 `blob.text()`：
 * 后者的可用性取决于宿主环境的 Blob 实现
 * （jsdom 就**没有**实现 `text()`），而 FileReader 是普遍可用的。
 */
function readBlobText(blob: Blob): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onload = () => {
      resolve(typeof reader.result === 'string' ? reader.result : '')
    }
    reader.onerror = () => {
      reject(reader.error ?? new Error('failed to read blob'))
    }
    reader.readAsText(blob)
  })
}

/** 从 Content-Disposition 响应头解析文件名。 */
function filenameFromDisposition(header: string | null): string {
  if (!header) {
    return ''
  }
  const utf8 = /filename\*=UTF-8''([^;]+)/i.exec(header)
  if (utf8?.[1]) {
    try {
      return decodeURIComponent(utf8[1].trim())
    } catch {
      // 编码损坏时退回 ASCII 文件名。
    }
  }
  const ascii = /filename="([^"]*)"/i.exec(header)
  return ascii?.[1] ?? ''
}

/** 从响应头构造元信息。 */
function metaFromHeaders(headers: XMLHttpRequest): Meta {
  const num = (name: string): number => {
    const raw = headers.getResponseHeader(name)
    const value = raw ? Number(raw) : Number.NaN
    return Number.isFinite(value) ? value : 0
  }
  return {
    durationMs: num('X-Duration-Ms'),
    inputBytes: num('X-Input-Bytes'),
    outputBytes: num('X-Output-Bytes'),
    warnings: parseWarningsHeader(headers.getResponseHeader('X-Warnings')),
  }
}

/** 解析 `X-Warnings` 响应头。 */
function parseWarningsHeader(raw: string | null): Warning[] {
  if (!raw) {
    return []
  }
  try {
    const parsed = JSON.parse(raw) as unknown
    if (Array.isArray(parsed)) {
      return parsed.filter((w): w is Warning => typeof w === 'object' && w !== null && 'code' in w)
    }
  } catch {
    // 响应头损坏时忽略：警告不影响结果可用性。
  }
  return []
}

/** 提取警告码列表。 */
export function warningCodes(warnings: Warning[] | undefined): string[] {
  if (!warnings) {
    return []
  }
  return warnings.map((w) => w.code)
}
