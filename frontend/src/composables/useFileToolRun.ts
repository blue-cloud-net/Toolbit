/**
 * 文件工具的运行状态机。
 *
 * 与 `useToolRun` 的差别：结果是**二进制文件**而不是文本，
 * 因此没有「复制」而只有「下载」，并额外提供上传进度。
 *
 * 状态四态与 `useToolRun` 一致：
 * `idle` / `loading` / `success` / `error`。
 */
import { onScopeDispose, ref, shallowRef } from 'vue'

import { ApiError, callToolBinary } from '@/api/client'
import { localizeError, type LocalizedError } from '@/i18n/errors'
import type { Meta } from '@/types/api'
import type { RunState } from '@/composables/useToolRun'

/** 上传进度的阶段。 */
export type UploadPhase = 'idle' | 'uploading' | 'processing' | 'done'

export interface UseFileToolRunOptions {
  /** 工具 slug。 */
  slug: string
  /** 超时（毫秒），默认取客户端默认值。 */
  timeoutMs?: number
}

/** 文件工具的运行结果。 */
export interface FileRunResult {
  /** 结果文件。 */
  blob: Blob
  /** 建议的文件名。 */
  filename: string
  /** 结果文件的字节数。 */
  size: number
}

export function useFileToolRun(options: UseFileToolRunOptions) {
  const state = ref<RunState>('idle')
  const result = shallowRef<FileRunResult | null>(null)
  const meta = shallowRef<Meta | null>(null)
  const error = shallowRef<LocalizedError | null>(null)
  const phase = ref<UploadPhase>('idle')
  const uploaded = ref(0)
  const uploadTotal = ref(0)

  let controller: AbortController | null = null

  /** 上传进度百分比；总量未知时为 null。 */
  const percent = ref<number | null>(null)

  /** 中止在途请求。 */
  function cancel(): void {
    if (controller) {
      controller.abort()
      controller = null
    }
  }

  /** 回到初始态。 */
  function reset(): void {
    cancel()
    state.value = 'idle'
    result.value = null
    meta.value = null
    error.value = null
    phase.value = 'idle'
    uploaded.value = 0
    uploadTotal.value = 0
    percent.value = null
  }

  /** 执行一次调用。 */
  async function run(form: FormData): Promise<void> {
    cancel()
    const current = new AbortController()
    controller = current

    state.value = 'loading'
    error.value = null
    phase.value = 'uploading'
    uploaded.value = 0
    uploadTotal.value = 0
    percent.value = null

    try {
      const response = await callToolBinary(options.slug, form, {
        signal: current.signal,
        ...(options.timeoutMs === undefined ? {} : { timeoutMs: options.timeoutMs }),
        onUploadProgress: (loaded, total) => {
          if (current.signal.aborted) {
            return
          }
          uploaded.value = loaded
          uploadTotal.value = total
          percent.value = total > 0 ? Math.round((loaded / total) * 100) : null
        },
      })
      if (current.signal.aborted) {
        return
      }
      phase.value = 'done'
      result.value = {
        blob: response.blob,
        filename: response.filename,
        size: response.blob.size,
      }
      meta.value = response.meta
      state.value = 'success'
    } catch (err) {
      if (current.signal.aborted) {
        return
      }
      if (err instanceof ApiError && err.isCanceled) {
        return
      }
      error.value = localizeError(err)
      meta.value = null
      state.value = 'error'
    } finally {
      if (controller === current) {
        controller = null
      }
    }
  }

  /**
   * 触发下载。
   *
   * 用临时 object URL，下载后**必须** revoke，
   * 否则 Blob 会一直驻留内存。
   */
  function download(): void {
    const current = result.value
    if (!current) {
      return
    }
    const url = URL.createObjectURL(current.blob)
    const anchor = document.createElement('a')
    anchor.href = url
    anchor.download = current.filename
    anchor.rel = 'noopener'
    document.body.appendChild(anchor)
    anchor.click()
    anchor.remove()
    URL.revokeObjectURL(url)
  }

  onScopeDispose(cancel)

  return { state, result, meta, error, phase, percent, uploaded, uploadTotal, run, reset, cancel, download }
}
