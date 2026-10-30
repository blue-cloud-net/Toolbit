/**
 * 自动运行（`autoRun`）。
 *
 * `backend/catalog/tools.yaml` 的 `autoRun` 字段声明某个工具是否
 * 「在参数变更后自动触发」。此前该字段无人消费，本模块把它接通：
 *
 *   - 只有 catalog 标记 `autoRun: true` 的工具才会自动运行；
 *   - 输入变化后**防抖**再触发，避免逐字输入时打满限流；
 *   - 敏感工具、需出网的工具、超时较长的工具一律不开启（catalog 侧把关）。
 *
 * 「一次取回全量、之后纯本地筛选」的工具**不用**本模块，
 * 改用 [`useLiveFilter`](./useLiveFilter.ts)。
 *
 * 用法（在工具页中）：
 *
 * ```ts
 * const { state, run } = useToolRun<Result>({ slug: 'hash-calc' })
 *
 * useAutoRun({
 *   slug: 'hash-calc',
 *   run: () => run({ ... }),
 *   watch: [input, algorithms, inputEncoding],
 * })
 * ```
 */
import { onScopeDispose, watch, type WatchSource } from 'vue'

import { useCatalogStore } from '@/stores/catalog'

/** 自动运行的防抖间隔（毫秒）。 */
export const AUTO_RUN_DEBOUNCE_MS = 400

export interface UseAutoRunOptions {
  /** 工具 slug；用于查 catalog 的 `autoRun` 标记。 */
  slug: string
  /** 触发运行的函数。 */
  run: () => void
  /** 监听的输入源；任一变化即触发防抖运行。 */
  watch: WatchSource[]
  /** 防抖间隔（毫秒），默认 `AUTO_RUN_DEBOUNCE_MS`。 */
  debounceMs?: number
  /** 额外的启用条件；返回 false 时不自动运行。 */
  enabled?: () => boolean
}

export function useAutoRun(options: UseAutoRunOptions): void {
  const catalog = useCatalogStore()
  const debounceMs = options.debounceMs ?? AUTO_RUN_DEBOUNCE_MS

  let timer: ReturnType<typeof setTimeout> | null = null

  function clearTimer(): void {
    if (timer !== null) {
      clearTimeout(timer)
      timer = null
    }
  }

  const stop = watch(
    options.watch,
    () => {
      // catalog 未加载或该工具未开启 autoRun 时不自动运行。
      const meta = catalog.getTool(options.slug)
      if (!meta?.autoRun) {
        return
      }
      if (options.enabled && !options.enabled()) {
        return
      }
      clearTimer()
      timer = setTimeout(() => {
        timer = null
        options.run()
      }, debounceMs)
    },
    { deep: true },
  )

  onScopeDispose(() => {
    clearTimer()
    stop()
  })
}
