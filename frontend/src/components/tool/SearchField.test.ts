/**
 * SearchField 组件测试：清空按钮与进入页面即聚焦。
 *
 * 组件必须 `attachTo: document.body`，否则聚焦与点击行为不生效。
 */
import { enableAutoUnmount, mount } from '@vue/test-utils'
import { afterEach, describe, expect, it } from 'vitest'
import { nextTick } from 'vue'

import SearchField from '@/components/tool/SearchField.vue'

enableAutoUnmount(afterEach)

/** 挂载搜索框。 */
function mountSearchField(props: Record<string, unknown> = {}) {
  return mount(SearchField, {
    props: { modelValue: '', label: '搜索', ...props },
    attachTo: document.body,
  })
}

describe('searchField', () => {
  it('渲染内部输入框，标签通过隐式关联提供无障碍名称', () => {
    const wrapper = mountSearchField({ modelValue: 'png' })

    const input = wrapper.find('input')
    expect(input.exists()).toBe(true)
    expect(input.element.value).toBe('png')
    expect(wrapper.find('label').text()).toContain('搜索')

    wrapper.unmount()
  })

  it('聚焦后点清空按钮发出空串', async () => {
    const wrapper = mountSearchField({ modelValue: 'png' })

    await wrapper.find('input').trigger('focus')
    await nextTick()

    const clear = wrapper.find('.el-input__clear')
    expect(clear.exists()).toBe(true)

    await clear.trigger('click')

    const emitted = wrapper.emitted('update:modelValue')
    expect(emitted).toBeTruthy()
    expect(emitted?.[emitted.length - 1]).toEqual([''])

    wrapper.unmount()
  })

  it('autofocus 时进入页面焦点落在输入框', async () => {
    const wrapper = mountSearchField({ autofocus: true })
    await nextTick()

    expect(document.activeElement).toBe(wrapper.find('input').element)

    wrapper.unmount()
  })

  it('默认不自动聚焦', async () => {
    const wrapper = mountSearchField()
    await nextTick()

    expect(document.activeElement).not.toBe(wrapper.find('input').element)

    wrapper.unmount()
  })
})
