/**
 * CodeEditor 组件测试：行号窄轨与输入区的关系。
 *
 * 行号不再是独立成框的元素，而是编辑框内的一列 —— 两者共用同一个外框。
 */
import { enableAutoUnmount, mount } from '@vue/test-utils'
import { afterEach, describe, expect, it } from 'vitest'

import CodeEditor from '@/components/tool/CodeEditor.vue'
import { i18n } from '@/i18n'

enableAutoUnmount(afterEach)

/** 挂载编辑框。 */
function mountEditor(props: Record<string, unknown> = {}) {
  return mount(CodeEditor, {
    props: { modelValue: 'a\nb\nc', 'onUpdate:modelValue': () => {}, ...props },
    global: { plugins: [i18n] },
  })
}

describe('codeEditor 行号窄轨', () => {
  it('行号与输入区同处一个外框内', () => {
    const wrapper = mountEditor()
    const frame = wrapper.find('.code-editor__frame')

    expect(frame.exists()).toBe(true)
    expect(frame.find('.code-editor__gutter').exists()).toBe(true)
    expect(frame.find('.code-editor__area').exists()).toBe(true)
    // 窄轨是外框的子元素，不再是并列的独立盒子。
    expect(frame.find('.code-editor__gutter').element.parentElement).toBe(frame.element)
  })

  it('按行渲染行号', () => {
    const wrapper = mountEditor()

    expect(wrapper.find('.code-editor__gutter-numbers').text()).toBe('1\n2\n3')
  })

  it('不显示行号时不渲染窄轨', () => {
    const wrapper = mountEditor({ showLineNumbers: false })

    expect(wrapper.find('.code-editor__gutter').exists()).toBe(false)
  })

  it('软折行时隐藏行号，避免与内容错位', () => {
    const wrapper = mountEditor({ wrap: 'soft' })

    expect(wrapper.find('.code-editor__gutter').exists()).toBe(false)
    expect(wrapper.find('.code-editor__area').attributes('wrap')).toBe('soft')
  })

  it('默认不折行，保证行号与视觉行一一对应', () => {
    const wrapper = mountEditor()

    expect(wrapper.find('.code-editor__area').attributes('wrap')).toBe('off')
  })

  it('当前行号单独渲染，不会被高亮条盖住', () => {
    const wrapper = mountEditor()
    const current = wrapper.find('.code-editor__gutter-current')

    // 当前行号有自己的元素，且盖在整列行号之上（z-index 更高）。
    expect(current.exists()).toBe(true)
    expect(current.text().trim()).toBe('1')
    const style = (current.element as HTMLElement).style
    expect(Number(style.getPropertyValue('--tb-editor-cursor-line'))).toBe(1)
  })
})
