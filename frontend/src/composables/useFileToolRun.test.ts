/**
 * useFileToolRun 单测：四态迁移、上传进度、取消与下载。
 */
import { effectScope, nextTick } from 'vue'
import { afterEach, describe, expect, it, vi } from 'vitest'

import { useFileToolRun } from '@/composables/useFileToolRun'

/** 构造一个最小可用的 XMLHttpRequest 替身。 */
class FakeXHR {
  /** 上传进度事件。 */
  upload = { onprogress: null as ((event: ProgressEvent) => void) | null }
  /** load 回调。 */
  onload: (() => void) | null = null
  /** error 回调。 */
  onerror: (() => void) | null = null
  /** timeout 回调。 */
  ontimeout: (() => void) | null = null
  /** abort 回调。 */
  onabort: (() => void) | null = null
  /** 请求方法。 */
  method = ''
  /** 请求 URL。 */
  url = ''
  /** 响应状态码。 */
  status = 0
  /** 响应类型。 */
  responseType = ''
  /** 超时设置。 */
  timeout = 0
  /** 发送的请求体。 */
  body: FormData | null = null
  /** 响应头。 */
  headers: Record<string, string> = {}
  /** 响应体。 */
  response: Blob | string = ''
  /** 已中止。 */
  aborted = false
  /** 已发送。 */
  sent = false

  open(method: string, url: string): void {
    this.method = method
    this.url = url
  }

  setRequestHeader(name: string, value: string): void {
    this.headers[name] = value
  }

  getResponseHeader(name: string): string | null {
    return this.headers[name] ?? null
  }

  send(body: FormData): void {
    this.body = body
    this.sent = true
  }

  abort(): void {
    this.aborted = true
    this.onabort?.()
  }
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('useFileToolRun', () => {
  it('初始为 idle', () => {
    const scope = effectScope()
    const runner = scope.run(() => useFileToolRun({ slug: 'sm4-file-crypt' }))!

    expect(runner.state.value).toBe('idle')
    expect(runner.result.value).toBeNull()
    expect(runner.error.value).toBeNull()
    expect(runner.phase.value).toBe('idle')

    scope.stop()
  })

  it('成功后进入 success 并保留文件名与大小', async () => {
    const xhr = new FakeXHR()
    vi.stubGlobal('XMLHttpRequest', vi.fn(() => xhr))

    const scope = effectScope()
    const runner = scope.run(() => useFileToolRun({ slug: 'sm4-file-crypt' }))!

    const pending = runner.run(new FormData())

    // 模拟服务端返回二进制流。
    xhr.status = 200
    xhr.headers['Content-Type'] = 'application/octet-stream'
    xhr.headers['Content-Disposition'] = "attachment; filename=\"a.txt.sm4\""
    xhr.headers['X-Duration-Ms'] = '42'
    xhr.response = new Blob(['hello'], { type: 'application/octet-stream' })
    xhr.onload?.()

    await pending

    expect(runner.state.value).toBe('success')
    expect(runner.result.value?.filename).toBe('a.txt.sm4')
    expect(runner.result.value?.size).toBe(5)
    expect(runner.meta.value?.durationMs).toBe(42)
    expect(runner.phase.value).toBe('done')

    scope.stop()
  })

  it('上传进度会更新百分比', async () => {
    const xhr = new FakeXHR()
    vi.stubGlobal('XMLHttpRequest', vi.fn(() => xhr))

    const scope = effectScope()
    const runner = scope.run(() => useFileToolRun({ slug: 'sm4-file-crypt' }))!

    const pending = runner.run(new FormData())

    xhr.upload.onprogress?.({ lengthComputable: true, loaded: 25, total: 100 } as ProgressEvent)
    await nextTick()
    expect(runner.percent.value).toBe(25)
    expect(runner.phase.value).toBe('uploading')

    xhr.status = 200
    xhr.headers['Content-Type'] = 'application/octet-stream'
    xhr.response = new Blob(['x'])
    xhr.onload?.()

    await pending
    scope.stop()
  })

  it('JSON 错误信封被转成 ApiError', async () => {
    const xhr = new FakeXHR()
    vi.stubGlobal('XMLHttpRequest', vi.fn(() => xhr))

    const scope = effectScope()
    const runner = scope.run(() => useFileToolRun({ slug: 'sm4-file-crypt' }))!

    const pending = runner.run(new FormData())

    xhr.status = 400
    xhr.headers['Content-Type'] = 'application/json'
    // 真实 XHR 在 responseType='blob' 下返回 Blob，因此错误信封也是 Blob。
    xhr.response = new Blob([
      JSON.stringify({
        code: 400,
        message: 'invalid field value',
        data: null,
        error: { code: 'INVALID_FIELD', params: { field: 'key' } },
        meta: { durationMs: 1, inputBytes: 0, outputBytes: 0, warnings: [] },
        requestId: 'req-err',
      }),
    ])
    xhr.onload?.()

    await pending

    expect(runner.state.value).toBe('error')
    expect(runner.error.value?.code).toBe('INVALID_FIELD')

    scope.stop()
  })

  it('取消后不进入 error 态', async () => {
    const xhr = new FakeXHR()
    vi.stubGlobal('XMLHttpRequest', vi.fn(() => xhr))

    const scope = effectScope()
    const runner = scope.run(() => useFileToolRun({ slug: 'sm4-file-crypt' }))!

    const pending = runner.run(new FormData())
    runner.cancel()
    await pending

    expect(runner.state.value).not.toBe('error')

    scope.stop()
  })

  it('下载会创建并回收 object URL', async () => {
    const xhr = new FakeXHR()
    vi.stubGlobal('XMLHttpRequest', vi.fn(() => xhr))

    const created: string[] = []
    const revoked: string[] = []
    vi.stubGlobal('URL', {
      createObjectURL: vi.fn(() => {
        const url = `blob:test-${created.length}`
        created.push(url)
        return url
      }),
      revokeObjectURL: vi.fn((url: string) => {
        revoked.push(url)
      }),
    })

    const scope = effectScope()
    const runner = scope.run(() => useFileToolRun({ slug: 'sm4-file-crypt' }))!

    // jsdom 不实现导航：拦掉锚点点击。
    const clicked: string[] = []
    const clickSpy = vi
      .spyOn(HTMLAnchorElement.prototype, 'click')
      .mockImplementation(function (this: HTMLAnchorElement) {
        clicked.push(this.download)
      })

    const pending = runner.run(new FormData())
    xhr.status = 200
    xhr.headers['Content-Type'] = 'application/octet-stream'
    xhr.headers['Content-Disposition'] = 'attachment; filename="a.bin"'
    xhr.response = new Blob(['data'])
    xhr.onload?.()
    await pending

    runner.download()

    expect(clicked).toEqual(['a.bin'])
    clickSpy.mockRestore()

    expect(created).toHaveLength(1)
    expect(revoked).toEqual(created)

    scope.stop()
  })
})
