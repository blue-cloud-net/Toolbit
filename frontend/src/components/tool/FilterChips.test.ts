/**
 * FilterChips 组件测试：单选与多选两种模式的选中态。
 */
import { enableAutoUnmount, mount } from '@vue/test-utils'
import { afterEach, describe, expect, it } from 'vitest'

import FilterChips, { type FilterChipOption } from '@/components/tool/FilterChips.vue'

enableAutoUnmount(afterEach)

const OPTIONS: FilterChipOption[] = [
  { value: 'a', label: 'A', hint: '全球' },
  { value: 'b', label: 'B', hint: '中国' },
  { value: 'c', label: 'C' },
]

/** 挂载标签组。 */
function mountChips(props: Record<string, unknown> = {}) {
  return mount(FilterChips, {
    props: { options: OPTIONS, ...props },
    attachTo: document.body,
  })
}

/** 读取每个标签的选中态。 */
function pressedStates(wrapper: ReturnType<typeof mountChips>): boolean[] {
  return wrapper.findAll('button').map((button) => button.attributes('aria-pressed') === 'true')
}

/** 取第 n 个标签；不存在时直接失败，避免静默跳过断言。 */
function chipAt(wrapper: ReturnType<typeof mountChips>, index: number) {
  const button = wrapper.findAll('button')[index]
  if (!button) {
    throw new Error(`缺少第 ${index} 个标签`)
  }
  return button
}

describe('filterChips 单选模式', () => {
  it('高亮当前选中项', () => {
    const wrapper = mountChips({ modelValue: 'b' })

    expect(pressedStates(wrapper)).toEqual([false, true, false])
    expect(wrapper.findAll('.filter-chips__item--active')).toHaveLength(1)
  })

  it('点击其他项发出新的选中值', async () => {
    const wrapper = mountChips({ modelValue: 'a' })

    await chipAt(wrapper, 2).trigger('click')

    expect(wrapper.emitted('update:modelValue')).toEqual([['c']])
  })

  it('重复点击当前项则取消选择', async () => {
    const wrapper = mountChips({ modelValue: 'a' })

    await chipAt(wrapper, 0).trigger('click')

    expect(wrapper.emitted('update:modelValue')).toEqual([['']])
  })
})

describe('filterChips 多选模式', () => {
  it('高亮选中集合内的全部项', () => {
    const wrapper = mountChips({ selected: ['a', 'c'] })

    expect(pressedStates(wrapper)).toEqual([true, false, true])
  })

  it('点击未选中项追加到集合', async () => {
    const wrapper = mountChips({ selected: ['a'] })

    await chipAt(wrapper, 1).trigger('click')

    expect(wrapper.emitted('update:selected')?.[0]?.[0]).toEqual(['a', 'b'])
  })

  it('点击已选中项从集合移除', async () => {
    const wrapper = mountChips({ selected: ['a', 'b'] })

    await chipAt(wrapper, 0).trigger('click')

    expect(wrapper.emitted('update:selected')?.[0]?.[0]).toEqual(['b'])
  })

  it('禁用的项点击无效', async () => {
    const wrapper = mountChips({
      selected: ['a'],
      options: [{ value: 'a', label: 'A' }, { value: 'b', label: 'B', disabled: true }],
    })

    const button = chipAt(wrapper, 1)
    expect(button.attributes('disabled')).toBeDefined()

    await button.trigger('click')

    expect(wrapper.emitted('update:selected')).toBeUndefined()
  })
})
