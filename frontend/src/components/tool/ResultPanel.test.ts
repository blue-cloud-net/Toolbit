/**
 * ResultPanel 组件测试：四态渲染。
 */
import { mount } from '@vue/test-utils'
import { describe, expect, it } from 'vitest'

import ResultPanel from '@/components/tool/ResultPanel.vue'
import { i18n } from '@/i18n'

/** 统一的挂载选项。 */
const mountOptions = { global: { plugins: [i18n] } }

describe('ResultPanel', () => {
  it('idle 显示占位提示', () => {
    const wrapper = mount(ResultPanel, { ...mountOptions, props: { state: 'idle' } })

    expect(wrapper.text()).toContain('尚未运行')
    expect(wrapper.find('pre').exists()).toBe(false)
  })

  it('loading 显示运行中', () => {
    const wrapper = mount(ResultPanel, { ...mountOptions, props: { state: 'loading' } })

    expect(wrapper.text()).toContain('运行中')
  })

  it('success 渲染结果并提供复制按钮', () => {
    const wrapper = mount(ResultPanel, {
      ...mountOptions,
      props: {
        state: 'success',
        text: 'SGVsbG8=',
        meta: { durationMs: 4, inputBytes: 5, outputBytes: 8, warnings: [] },
      },
    })

    expect(wrapper.find('pre').text()).toBe('SGVsbG8=')
    expect(wrapper.text()).toContain('复制')
    expect(wrapper.text()).toContain('耗时 4 毫秒')
  })

  it('error 渲染错误码 chip 并派发重试事件', async () => {
    const wrapper = mount(ResultPanel, {
      ...mountOptions,
      props: {
        state: 'error',
        error: { code: 'INVALID_BASE64', text: 'Base64 串非法。', requestId: 'req-9' },
      },
    })

    expect(wrapper.text()).toContain('INVALID_BASE64')
    expect(wrapper.text()).toContain('Base64 串非法。')

    await wrapper.get('button').trigger('click')
    expect(wrapper.emitted('retry')).toHaveLength(1)
  })

  it('success 时展示警告词条', () => {
    const wrapper = mount(ResultPanel, {
      ...mountOptions,
      props: {
        state: 'success',
        text: 'x',
        meta: {
          durationMs: 1,
          inputBytes: 1,
          outputBytes: 1,
          warnings: [{ code: 'LARGE_INPUT' }],
        },
      },
    })

    expect(wrapper.text()).toContain('输入较大')
  })
})
