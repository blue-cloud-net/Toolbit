/**
 * 运行栏组件测试：`Ctrl / ⌘ + Enter` 运行快捷键及其守卫。
 *
 * 快捷键监听挂在 `window` 上，因此用例必须用 `attachTo` 把组件挂到文档里，
 * 事件才能冒泡 / 捕获到 `window`。
 */
import { enableAutoUnmount, mount } from '@vue/test-utils'
import { afterEach, describe, expect, it, vi } from 'vitest'

import RunBar from '@/components/tool/RunBar.vue'
import { i18n } from '@/i18n'

enableAutoUnmount(afterEach)

/** 挂载运行栏，并接到文档上。 */
function mountRunBar(props: Record<string, unknown> = {}) {
  return mount(RunBar, {
    props,
    attachTo: document.body,
    global: { plugins: [i18n] },
  })
}

/** 派发一次 `Ctrl / ⌘ + Enter`。 */
function pressEnter(target: EventTarget, options: { meta?: boolean } = {}): KeyboardEvent {
  const event = new KeyboardEvent('keydown', {
    key: 'Enter',
    ctrlKey: options.meta !== true,
    metaKey: options.meta === true,
    bubbles: true,
    cancelable: true,
  })
  target.dispatchEvent(event)
  return event
}

describe('runBar 快捷键', () => {
  it('按 Ctrl + Enter 触发一次 run', () => {
    const wrapper = mountRunBar()
    pressEnter(wrapper.element)

    expect(wrapper.emitted('run')).toHaveLength(1)
  })

  it('按 ⌘ + Enter 同样触发 run', () => {
    const wrapper = mountRunBar()
    pressEnter(wrapper.element, { meta: true })

    expect(wrapper.emitted('run')).toHaveLength(1)
  })

  it('单独按 Enter 不触发', () => {
    const wrapper = mountRunBar()
    wrapper.element.dispatchEvent(
      new KeyboardEvent('keydown', { key: 'Enter', bubbles: true, cancelable: true }),
    )

    expect(wrapper.emitted('run')).toBeUndefined()
  })

  it('运行中或不可运行时不触发', () => {
    const running = mountRunBar({ running: true })
    pressEnter(running.element)
    expect(running.emitted('run')).toBeUndefined()

    const disabled = mountRunBar({ disabled: true })
    pressEnter(disabled.element)
    expect(disabled.emitted('run')).toBeUndefined()
  })

  it('弹窗内的按键不触发', () => {
    const wrapper = mountRunBar()
    const dialog = document.createElement('div')
    dialog.setAttribute('role', 'dialog')
    document.body.append(dialog)
    pressEnter(dialog)
    dialog.remove()

    expect(wrapper.emitted('run')).toBeUndefined()
  })

  it('快捷键事件被本组件接管，页面内遗留实现不会重复运行', () => {
    const wrapper = mountRunBar()
    const inner = document.createElement('input')
    wrapper.element.append(inner)

    /** 模拟工具页里遗留的 `@keydown` 处理。 */
    const legacy = vi.fn()
    inner.addEventListener('keydown', legacy)
    const event = pressEnter(inner)

    expect(legacy).not.toHaveBeenCalled()
    expect(event.defaultPrevented).toBe(true)
    expect(wrapper.emitted('run')).toHaveLength(1)
  })

  it('卸载后不再响应快捷键', () => {
    const wrapper = mountRunBar()
    wrapper.unmount()
    pressEnter(document.body)

    expect(wrapper.emitted('run')).toBeUndefined()
  })
})
