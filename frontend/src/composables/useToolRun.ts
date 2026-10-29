/**
 * 工具运行状态机。
 *
 * 四态：`idle` / `loading` / `success` / `error`。
 * 组件卸载或重复运行时**必须**中止上一次请求。
 */
import { computed, onScopeDispose, ref, shallowRef } from 'vue'

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

  /** 输入回填器：工具页注册，用于「作为新输入」。 */
  const inputFiller = shallowRef<((text: string) => void) | null>(null)

  /**
   * 注册输入回填函数。
   *
   * 工具页把自己的主输入 ref 交给这里，结果区的「作为新输入」
   * 就能把结果文本回填到输入区。回填只改内存状态，**不写**任何浏览器存储。
   */
  function bindInput(filler: (text: string) => void): void {
    inputFiller.value = filler
  }

  /** 把文本回填到输入区；未注册回填器时返回 false。 */
  function useAsInput(text: string): boolean {
    if (!inputFiller.value) {
      return false
    }
    inputFiller.value(text)
    return true
  }

  /** 是否支持「作为新输入」。 */
  const canUseAsInput = computed(() => inputFiller.value !== null)

  return { state, data, meta, error, run, reset, cancel, bindInput, useAsInput, canUseAsInput }
}
