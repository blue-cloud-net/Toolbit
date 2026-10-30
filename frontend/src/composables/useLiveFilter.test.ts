/**
 * useLiveFilter 单测：纯函数过滤与组合式行为。
 *
 * 组合式必须在**真实组件**里跑 —— `onMounted` 只在组件挂载时触发，
 * 因此用例用一个内联组件作为宿主。
 */
import { flushPromises, mount } from '@vue/test-utils'
import { afterEach, describe, expect, it, vi } from 'vitest'
import { defineComponent, h, nextTick, ref } from 'vue'

import { filterEntries, useLiveFilter, type UseLiveFilterReturn } from '@/composables/useLiveFilter'

/** 一个测试条目。 */
interface Item {
  /** 标识。 */
  id: string
  /** 名称。 */
  name: string
  /** 分组。 */
  group: string
}

/** 响应体。 */
interface ItemResponse {
  /** 条目列表。 */
  items: Item[]
}

/** 测试数据集。 */
const ITEMS: Item[] = [
  { id: 'a', name: 'alpha', group: 'x' },
  { id: 'b', name: 'beta', group: 'y' },
  { id: 'c', name: 'alphabet', group: 'x' },
]

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

describe('filterEntries', () => {
  it('空查询返回全部条目', () => {
    const result = filterEntries(ITEMS, '   ', { match: () => true })

    expect(result.map((item) => item.id)).toEqual(['a', 'b', 'c'])
  })

  it('按匹配函数过滤', () => {
    const result = filterEntries(ITEMS, 'alpha', {
      match: (entry, term) => entry.name.includes(term),
    })

    expect(result.map((item) => item.id)).toEqual(['a', 'c'])
  })

  it('按命中档位排序，同档保持原顺序', () => {
    const result = filterEntries(ITEMS, 'a', {
      match: (entry, term) => entry.name.includes(term),
      rank: (entry) => (entry.group === 'x' ? 0 : 1),
    })

    expect(result.map((item) => item.id)).toEqual(['a', 'c', 'b'])
  })

  it('分类筛选与查询词叠加', () => {
    const category = ref('x')
    const result = filterEntries(ITEMS, 'a', {
      match: (entry, term) => entry.name.includes(term),
      category: (entry) => entry.group,
      categoryRef: category,
    })

    expect(result.map((item) => item.id)).toEqual(['a', 'c'])
  })

  it('未提供分类函数时忽略分类', () => {
    const category = ref('y')
    const result = filterEntries(ITEMS, 'beta', {
      match: (entry, term) => entry.name.includes(term),
      categoryRef: category,
    })

    expect(result.map((item) => item.id)).toEqual(['b'])
  })
})

describe('useLiveFilter', () => {
  /** 挂载宿主组件并返回组合式句柄。 */
  function mountLive() {
    const query = ref('')
    const category = ref('')
    let live!: UseLiveFilterReturn<Item, ItemResponse>

    const host = defineComponent({
      setup() {
        live = useLiveFilter<Item, ItemResponse>({
          slug: 'demo',
          query,
          select: (result) => result.items,
          match: (entry, term) => entry.name.includes(term),
          category: (entry) => entry.group,
          categoryRef: category,
        })
        return () => h('div')
      },
    })

    const wrapper = mount(host, { attachTo: document.body })

    return { wrapper, query, category, live: () => live }
  }

  it('进入页面自动取数', async () => {
    const fetchMock = vi.fn().mockResolvedValue(okEnvelope({ items: ITEMS }))
    vi.stubGlobal('fetch', fetchMock)

    const { wrapper, live } = mountLive()
    await flushPromises()

    expect(fetchMock).toHaveBeenCalledTimes(1)
    expect(live().state.value).toBe('success')
    expect(live().entries.value).toHaveLength(3)
    expect(live().filtered.value).toHaveLength(3)

    wrapper.unmount()
  })

  it('键入只在本地过滤，不再发请求', async () => {
    const fetchMock = vi.fn().mockResolvedValue(okEnvelope({ items: ITEMS }))
    vi.stubGlobal('fetch', fetchMock)

    const { wrapper, query, live } = mountLive()
    await flushPromises()

    query.value = 'beta'
    await nextTick()

    expect(live().filtered.value.map((item) => item.id)).toEqual(['b'])
    expect(fetchMock).toHaveBeenCalledTimes(1)

    wrapper.unmount()
  })

  it('clear 只清查询词，不重新请求', async () => {
    const fetchMock = vi.fn().mockResolvedValue(okEnvelope({ items: ITEMS }))
    vi.stubGlobal('fetch', fetchMock)

    const { wrapper, query, live } = mountLive()
    await flushPromises()

    query.value = 'beta'
    await nextTick()
    expect(live().hasFilter.value).toBe(true)

    live().clear()
    await nextTick()

    expect(query.value).toBe('')
    expect(live().hasFilter.value).toBe(false)
    expect(live().filtered.value).toHaveLength(3)
    expect(fetchMock).toHaveBeenCalledTimes(1)

    wrapper.unmount()
  })

  it('counts 统计各分类条目数，不受筛选影响', async () => {
    const fetchMock = vi.fn().mockResolvedValue(okEnvelope({ items: ITEMS }))
    vi.stubGlobal('fetch', fetchMock)

    const { wrapper, query, live } = mountLive()
    await flushPromises()

    expect(live().counts.value).toEqual({ x: 2, y: 1 })

    query.value = 'beta'
    await nextTick()

    expect(live().counts.value).toEqual({ x: 2, y: 1 })

    wrapper.unmount()
  })

  it('取数失败进入 error 态', async () => {
    const fetchMock = vi.fn().mockResolvedValue(
      new Response(
        JSON.stringify({
          code: 400,
          message: 'bad request',
          data: null,
          error: { code: 'INVALID_INPUT', message: 'bad request', params: {} },
          meta: null,
          requestId: 'req-bad',
        }),
        { status: 400, headers: { 'Content-Type': 'application/json' } },
      ),
    )
    vi.stubGlobal('fetch', fetchMock)

    const { wrapper, live } = mountLive()
    await flushPromises()

    expect(live().state.value).toBe('error')
    expect(live().entries.value).toHaveLength(0)

    wrapper.unmount()
  })
})
