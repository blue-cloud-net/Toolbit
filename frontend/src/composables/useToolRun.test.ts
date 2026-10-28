/**
 * useToolRun 单测：四态迁移、取消与卸载清理。
 */
import { effectScope } from 'vue'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { useToolRun } from '@/composables/useToolRun'

/** 构造成功信封。 */
function okEnvelope(data: unknown): Response {
  return new Response(
    JSON.stringify({
      code: 0,
      message: 'ok',
      data,
      error: null,
      meta: { durationMs: 5, inputBytes: 1, outputBytes: 2, warnings: [] },
      requestId: 'req-ok',
    }),
    { status: 200, headers: { 'Content-Type': 'application/json' } },
  )
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('useToolRun', () => {
  it('初始为 idle', () => {
    const scope = effectScope()
    const runner = scope.run(() => useToolRun({ slug: 'base64' }))!

    expect(runner.state.value).toBe('idle')
    expect(runner.data.value).toBeNull()
    expect(runner.error.value).toBeNull()

    scope.stop()
  })

  it('成功后进入 success 并保留 meta', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(okEnvelope({ output: 'SGVsbG8=' })))

    const scope = effectScope()
    const runner = scope.run(() => useToolRun<{ output: string }>({ slug: 'base64' }))!

    await runner.run({ operation: 'encode' })

    expect(runner.state.value).toBe('success')
    expect(runner.data.value?.output).toBe('SGVsbG8=')
    expect(runner.meta.value?.durationMs).toBe(5)

    scope.stop()
  })

  it('失败后进入 error 并本地化错误码', async () => {
    vi.stubGlobal(
      'fetch',
      vi.fn().mockResolvedValue(
        new Response(
          JSON.stringify({
            code: 400,
            message: 'invalid field value',
            data: null,
            error: { code: 'INVALID_FIELD', params: { field: 'variant' } },
            meta: { durationMs: 1, inputBytes: 0, outputBytes: 0, warnings: [] },
            requestId: 'req-err',
          }),
          { status: 400, headers: { 'Content-Type': 'application/json' } },
        ),
      ),
    )

    const scope = effectScope()
    const runner = scope.run(() => useToolRun({ slug: 'base64' }))!

    await runner.run({})

    expect(runner.state.value).toBe('error')
    expect(runner.error.value?.code).toBe('INVALID_FIELD')
    expect(runner.error.value?.text).toContain('variant')
    expect(runner.error.value?.requestId).toBe('req-err')

    scope.stop()
  })

  it('reset 回到 idle', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(okEnvelope({ output: 'x' })))

    const scope = effectScope()
    const runner = scope.run(() => useToolRun<{ output: string }>({ slug: 'base64' }))!

    await runner.run({})
    runner.reset()

    expect(runner.state.value).toBe('idle')
    expect(runner.data.value).toBeNull()

    scope.stop()
  })

  it('作用域销毁会中止在途请求', () => {
    const abortSpy = vi.fn()
    vi.stubGlobal(
      'fetch',
      vi.fn().mockImplementation((_url: string, init: RequestInit) => {
        const signal = init.signal as AbortSignal
        signal.addEventListener('abort', abortSpy)
        return new Promise(() => {
          /* 永不 resolve，模拟在途请求 */
        })
      }),
    )

    const scope = effectScope()
    const runner = scope.run(() => useToolRun({ slug: 'base64' }))!

    void runner.run({})
    scope.stop()

    expect(abortSpy).toHaveBeenCalled()
  })
})
