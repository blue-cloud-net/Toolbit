/**
 * 工具运行状态机。
 *
 * 四态：`idle` / `loading` / `success` / `error`。
 * 组件卸载或重复运行时**必须**中止上一次请求。
 */
import { onScopeDispose, ref, shallowRef } from 'vue'

import { ApiError, callTool } from '@/api/client'
import { localizeError, type LocalizedError } from '@/i18n/errors'
import type { Meta } from '@/types/api'

/** 运行状态。 */
export type RunState = 'idle' | 'loading' | 'success' | 'error'

export interface UseToolRunOptions {
  /** 工具 slug。 */
  slug: string
  /** 超时（毫秒），默认取客户端默认值。 */
  timeoutMs?: number
}

export function useToolRun<T = unknown>(options: UseToolRunOptions) {
  const state = ref<RunState>('idle')
  const data = shallowRef<T | null>(null)
  const meta = shallowRef<Meta | null>(null)
  const error = shallowRef<LocalizedError | null>(null)

  let controller: AbortController | null = null

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
    data.value = null
    meta.value = null
    error.value = null
  }

  /** 执行一次调用。 */
  async function run(payload: unknown): Promise<void> {
    cancel()
    const current = new AbortController()
    controller = current

    state.value = 'loading'
    error.value = null

    try {
      const result = await callTool<T>(options.slug, payload, {
        signal: current.signal,
        ...(options.timeoutMs === undefined ? {} : { timeoutMs: options.timeoutMs }),
      })
      if (current.signal.aborted) {
        return
      }
      data.value = result.data
      meta.value = result.meta
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

  onScopeDispose(cancel)

  return { state, data, meta, error, run, reset, cancel }
}
