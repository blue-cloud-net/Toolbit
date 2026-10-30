/**
 * 键入即筛选（live filter）。
 *
 * 面向「一次性取回全量数据集、之后纯本地检索」的工具
 * （如 MIME 类型查询、HTTP 状态码查询）：
 *
 *   - 进入页面拉一次数据，之后**不再发请求**；
 *   - 用户键入的查询词与所选分类在**前端本地**完成过滤，零延迟、无 loading 闪烁；
 *   - 页面**不提供**「运行」按钮与运行快捷键 —— 没有需要触发的动作。
 *
 * 与 `useAutoRun` 的分工：`useAutoRun` 是「每次输入都打后端」（带防抖与限流），
 * 本模块是「只取一次、之后本地筛」。两者不可混用。
 *
 * 过滤逻辑抽成纯函数 `filterEntries`，便于脱离网络单测。
 */
import { computed, onMounted, type ComputedRef, type Ref, type ShallowRef } from 'vue'

import { useToolRun, type RunState } from '@/composables/useToolRun'
import type { Meta } from '@/types/api'
import type { LocalizedError } from '@/i18n/errors'

/** 默认查询词归一化：去首尾空白并转小写。 */
function defaultNormalize(raw: string): string {
  return raw.trim().toLowerCase()
}

/** 过滤选项。 */
export interface FilterOptions<T> {
  /** 查询词归一化；缺省去首尾空白并转小写。 */
  normalize?: (raw: string) => string
  /** 条目是否命中查询词。 */
  match: (entry: T, term: string) => boolean
  /** 命中档位；数字越小越靠前，缺省保持原顺序。 */
  rank?: (entry: T, term: string) => number
  /** 条目的分类键；提供后才会按分类过滤与计数。 */
  category?: (entry: T) => string
  /** 当前分类；空串表示「全部」。 */
  categoryRef?: Ref<string>
}

/**
 * 纯函数过滤：按查询词与分类过滤条目，并按命中档位排序。
 *
 * 同档位时保持原始顺序（稳定排序），因此「精确命中排在最前」可由 `rank` 表达。
 */
export function filterEntries<T>(
  entries: readonly T[],
  rawTerm: string,
  options: FilterOptions<T>,
): T[] {
  const normalize = options.normalize ?? defaultNormalize
  const term = normalize(rawTerm)
  const category = options.categoryRef?.value ?? ''

  const hits = entries.filter((entry) => {
    if (category !== '' && options.category !== undefined && options.category(entry) !== category) {
      return false
    }
    if (term === '') {
      return true
    }
    return options.match(entry, term)
  })

  const rank = options.rank
  if (rank === undefined) {
    return hits
  }
  return hits
    .map((entry, index) => ({ entry, rank: rank(entry, term), index }))
    .sort((a, b) => (a.rank === b.rank ? a.index - b.index : a.rank - b.rank))
    .map((item) => item.entry)
}

/** `useLiveFilter` 的入参。 */
export interface UseLiveFilterOptions<T, R> extends FilterOptions<T> {
  /** 工具 slug。 */
  slug: string
  /** 从响应数据中取出待筛选条目。 */
  select: (data: R) => T[]
  /** 查询词（外部持有的 ref）。 */
  query: Ref<string>
  /** 请求体；缺省为空对象。 */
  payload?: () => unknown
}

/** `useLiveFilter` 的返回值。 */
export interface UseLiveFilterReturn<T, R> {
  /** 运行状态。 */
  state: Ref<RunState>
  /** 响应数据。 */
  data: ShallowRef<R | null>
  /** 执行元信息。 */
  meta: ShallowRef<Meta | null>
  /** 错误。 */
  error: ShallowRef<LocalizedError | null>
  /** 是否正在运行。 */
  running: ComputedRef<boolean>
  /** 全量条目。 */
  entries: ComputedRef<T[]>
  /** 过滤后的条目。 */
  filtered: ComputedRef<T[]>
  /** 各分类条目数（基于全量数据集）。 */
  counts: ComputedRef<Record<string, number>>
  /** 归一化后的查询词。 */
  term: ComputedRef<string>
  /** 是否处于筛选态。 */
  hasFilter: ComputedRef<boolean>
  /** 重新拉取全量数据集。 */
  reload: () => void
  /** 清空查询词（不重新请求）。 */
  clear: () => void
}

/**
 * 键入即筛选。
 *
 * 进入页面自动拉一次全量数据集，之后所有筛选都在本地完成。
 * 首屏加载**不依赖** catalog 的 `autoRun`：页面没有「运行」按钮兜底，
 * 必须自行取数，且不要求 catalog 先就绪。
 */
export function useLiveFilter<T, R = unknown>(
  options: UseLiveFilterOptions<T, R>,
): UseLiveFilterReturn<T, R> {
  const { state, data, meta, error, run } = useToolRun<R>({ slug: options.slug })

  const running = computed(() => state.value === 'loading')

  /** 重新拉取全量数据集。 */
  function reload(): void {
    void run(options.payload ? options.payload() : {})
  }

  /** 全量条目。 */
  const entries = computed<T[]>(() => (data.value === null ? [] : options.select(data.value)))

  /** 归一化后的查询词。 */
  const term = computed(() => (options.normalize ?? defaultNormalize)(options.query.value))

  /** 过滤结果。 */
  const filtered = computed<T[]>(() => filterEntries(entries.value, options.query.value, options))

  /** 各分类条目数（基于全量数据集，不受筛选影响）。 */
  const counts = computed<Record<string, number>>(() => {
    const result: Record<string, number> = {}
    if (options.category === undefined) {
      return result
    }
    for (const entry of entries.value) {
      const key = options.category(entry)
      result[key] = (result[key] ?? 0) + 1
    }
    return result
  })

  /** 是否处于筛选态。 */
  const hasFilter = computed(
    () => term.value !== '' || (options.categoryRef?.value ?? '') !== '',
  )

  /** 清空查询词（不重新请求）。 */
  function clear(): void {
    options.query.value = ''
  }

  onMounted(reload)

  return {
    state,
    data,
    meta,
    error,
    running,
    entries,
    filtered,
    counts,
    term,
    hasFilter,
    reload,
    clear,
  }
}
