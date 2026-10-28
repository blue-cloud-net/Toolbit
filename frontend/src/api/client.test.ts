/**
 * API 客户端单测。
 *
 * 全部用 `vi.stubGlobal('fetch')` 拦截网络，**禁止**依赖真实后端。
 */
import { afterEach, describe, expect, it, vi } from 'vitest'

import { ApiError, callTool, fetchCatalog, warningCodes } from '@/api/client'

/** 构造一个成功信封响应。 */
function okEnvelope(data: unknown): Response {
  return new Response(
    JSON.stringify({
      code: 0,
      message: 'ok',
      data,
      error: null,
      meta: { durationMs: 3, inputBytes: 10, outputBytes: 20, warnings: [] },
      requestId: 'req-1',
    }),
    { status: 200, headers: { 'Content-Type': 'application/json' } },
  )
}

/** 构造一个失败信封响应。 */
function errEnvelope(status: number, code: string, params: Record<string, unknown>): Response {
  return new Response(
    JSON.stringify({
      code: status,
      message: 'invalid field value',
      data: null,
      error: { code, params },
      meta: { durationMs: 1, inputBytes: 0, outputBytes: 0, warnings: [] },
      requestId: 'req-2',
    }),
    { status, headers: { 'Content-Type': 'application/json' } },
  )
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('fetchCatalog', () => {
  it('拆出 data 与 meta', async () => {
    const payload = {
      version: 1,
      categories: [{ slug: 'encoding', name: '编码', nameEn: 'Encoding', order: 4, icon: 'code' }],
      tools: [],
    }
    const fetchMock = vi.fn().mockResolvedValue(okEnvelope(payload))
    vi.stubGlobal('fetch', fetchMock)

    const result = await fetchCatalog()

    expect(result.data.version).toBe(1)
    expect(result.meta.durationMs).toBe(3)
    expect(fetchMock).toHaveBeenCalledOnce()
  })

  it('把 HTTP 失败转成 ApiError', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(errEnvelope(400, 'INVALID_FIELD', { field: 'value' })))

    await expect(fetchCatalog()).rejects.toBeInstanceOf(ApiError)
  })
})

describe('callTool', () => {
  it('带上 JSON 请求体', async () => {
    const fetchMock = vi.fn().mockResolvedValue(okEnvelope({ output: 'SGVsbG8=' }))
    vi.stubGlobal('fetch', fetchMock)

    const result = await callTool('base64', { operation: 'encode' })

    expect(result.data).toEqual({ output: 'SGVsbG8=' })
    const init = fetchMock.mock.calls[0]?.[1] as RequestInit
    expect(init.method).toBe('POST')
    expect(init.body).toBe(JSON.stringify({ operation: 'encode' }))
  })

  it('保留错误码与插值参数', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(errEnvelope(400, 'INVALID_FIELD', { field: 'value', expected: '2..36' })),
    )

    const error = await callTool('number-base', {}).catch((e: unknown) => e)

    expect(error).toBeInstanceOf(ApiError)
    const apiError = error as ApiError
    expect(apiError.code).toBe('INVALID_FIELD')
    expect(apiError.status).toBe(400)
    expect(apiError.params.field).toBe('value')
    expect(apiError.requestId).toBe('req-2')
  })

  it('外部取消归一化为 CANCELED', async () => {
    const controller = new AbortController()
    vi.stubGlobal(
      'fetch',
      vi.fn().mockImplementation((_url: string, init: RequestInit) => {
        const signal = init.signal as AbortSignal
        return new Promise((_resolve, reject) => {
          signal.addEventListener('abort', () => {
            reject(new DOMException('aborted', 'AbortError'))
          })
        })
      }),
    )

    const pending = callTool('base64', {}, { signal: controller.signal })
    controller.abort()

    await expect(pending).rejects.toMatchObject({ code: 'CANCELED' })
  })

  it('非 JSON 响应体归为 INVALID_RESPONSE', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(new Response('<html>oops</html>', { status: 502 })),
    )

    await expect(callTool('base64', {})).rejects.toMatchObject({ code: 'INVALID_RESPONSE' })
  })
})

describe('warningCodes', () => {
  it('提取警告码，空值返回空数组', () => {
    expect(warningCodes(undefined)).toEqual([])
    expect(warningCodes([{ code: 'LARGE_INPUT' }, { code: 'WEAK_HASH' }])).toEqual([
      'LARGE_INPUT',
      'WEAK_HASH',
    ])
  })
})
