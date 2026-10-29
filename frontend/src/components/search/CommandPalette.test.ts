/**
 * 命令面板组件测试：唤起、过滤、键盘导航、关闭与焦点归还。
 */
import { enableAutoUnmount, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { afterEach, beforeEach, describe, expect, it } from 'vitest'
import { createRouter, createWebHistory, type Router } from 'vue-router'

import CommandPalette from '@/components/search/CommandPalette.vue'
import { resetPalette } from '@/composables/useCommandPalette'
import { i18n } from '@/i18n'
import { useCatalogStore } from '@/stores/catalog'

/** 目录里的示例数据。 */
const CATEGORIES = [
  { slug: 'encoding', name: '编码转换', nameEn: 'Encoding', order: 1, icon: 'code' },
]

const TOOLS = [
  {
    slug: 'base64',
    category: 'encoding',
    order: 1,
    name: 'Base64 编解码',
    nameEn: 'Base64 Encode/Decode',
    enabled: true,
    transport: 'json' as const,
    rateClass: 'light' as const,
    maxBodyBytes: 1024,
    timeoutMs: 5000,
    sensitive: false,
    requiresNetwork: false,
    autoRun: true,
    wide: false,
    i18nKey: 'tools.base64',
  },
  {
    slug: 'url-codec',
    category: 'encoding',
    order: 2,
    name: 'URL 编解码',
    nameEn: 'URL Encode/Decode',
    enabled: true,
    transport: 'json' as const,
    rateClass: 'light' as const,
    maxBodyBytes: 1024,
    timeoutMs: 5000,
    sensitive: false,
    requiresNetwork: false,
    autoRun: true,
    wide: false,
    i18nKey: 'tools.url-codec',
  },
]

/** 构造带最小路由表的测试用 router。 */
function makeRouter(): Router {
  return createRouter({
    history: createWebHistory(),
    routes: [
      { path: '/', name: 'home', component: { template: '<div />' } },
      { path: '/recent', name: 'recent', component: { template: '<div />' } },
      { path: '/favorites', name: 'favorites', component: { template: '<div />' } },
      { path: '/c/:category', name: 'category', component: { template: '<div />' } },
      { path: '/tools/:slug', name: 'tool', component: { template: '<div />' } },
    ],
  })
}

/** 挂载命令面板。 */
async function mountPalette(): Promise<ReturnType<typeof mount>> {
  const router = makeRouter()
  await router.push('/')
  await router.isReady()

  const pinia = createPinia()
  setActivePinia(pinia)
  const catalog = useCatalogStore()
  catalog.applyCatalog({ version: 1, categories: CATEGORIES, tools: TOOLS })

  const wrapper = mount(CommandPalette, {
    global: { plugins: [router, pinia, i18n] },
    attachTo: document.body,
  })
  await wrapper.vm.$nextTick()
  return wrapper
}

// 包装器自动卸载：否则全局 keydown 监听器会在用例间累积。
enableAutoUnmount(afterEach)

beforeEach(() => {
  document.body.innerHTML = ''
  resetPalette()
})

describe('CommandPalette', () => {
  it('默认不渲染面板', async () => {
    const wrapper = await mountPalette()

    expect(wrapper.find('[role="dialog"]').exists()).toBe(false)
  })

  it('Ctrl+K 唤起面板，再按一次关闭', async () => {
    const wrapper = await mountPalette()

    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', ctrlKey: true }))
    await wrapper.vm.$nextTick()
    expect(wrapper.find('[role="dialog"]').exists()).toBe(true)

    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', ctrlKey: true }))
    await wrapper.vm.$nextTick()
    expect(wrapper.find('[role="dialog"]').exists()).toBe(false)
  })

  it('⌘+K 同样可以唤起', async () => {
    const wrapper = await mountPalette()

    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', metaKey: true }))
    await wrapper.vm.$nextTick()

    expect(wrapper.find('[role="dialog"]').exists()).toBe(true)
  })

  it('输入关键字后过滤中文名', async () => {
    const wrapper = await mountPalette()

    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', ctrlKey: true }))
    await wrapper.vm.$nextTick()

    const input = wrapper.get('input')
    await input.setValue('URL')
    await wrapper.vm.$nextTick()

    const labels = wrapper.findAll('.palette__item-label').map((n) => n.text())
    expect(labels).toContain('URL 编解码')
    expect(labels).not.toContain('Base64 编解码')
  })

  it('输入关键字后过滤英文名', async () => {
    const wrapper = await mountPalette()

    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', ctrlKey: true }))
    await wrapper.vm.$nextTick()

    const input = wrapper.get('input')
    await input.setValue('base64')
    await wrapper.vm.$nextTick()

    const labels = wrapper.findAll('.palette__item-label').map((n) => n.text())
    expect(labels).toContain('Base64 编解码')
    expect(labels).not.toContain('URL 编解码')
  })

  it('↓ 移动选中项，到边界后循环', async () => {
    const wrapper = await mountPalette()

    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', ctrlKey: true }))
    await wrapper.vm.$nextTick()

    const panel = wrapper.get('[role="dialog"]')
    const items = wrapper.findAll('.palette__item')
    expect(items.length).toBeGreaterThan(2)
    expect(items[0]!.classes()).toContain('palette__item--active')

    await panel.trigger('keydown', { key: 'ArrowDown' })
    await wrapper.vm.$nextTick()
    expect(wrapper.findAll('.palette__item')[1]!.classes()).toContain('palette__item--active')

    await panel.trigger('keydown', { key: 'ArrowUp' })
    await wrapper.vm.$nextTick()
    expect(wrapper.findAll('.palette__item')[0]!.classes()).toContain('palette__item--active')

    // 从首项继续上移 → 回到末项（循环）。
    await panel.trigger('keydown', { key: 'ArrowUp' })
    await wrapper.vm.$nextTick()
    const all = wrapper.findAll('.palette__item')
    expect(all[all.length - 1]!.classes()).toContain('palette__item--active')
  })

  it('Enter 打开选中项并跳转', async () => {
    const wrapper = await mountPalette()

    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', ctrlKey: true }))
    await wrapper.vm.$nextTick()

    const panel = wrapper.get('[role="dialog"]')
    await panel.trigger('keydown', { key: 'ArrowDown' })
    await panel.trigger('keydown', { key: 'Enter' })
    await wrapper.vm.$nextTick()

    expect(wrapper.find('[role="dialog"]').exists()).toBe(false)
  })

  it('Esc 关闭面板', async () => {
    const wrapper = await mountPalette()

    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', ctrlKey: true }))
    await wrapper.vm.$nextTick()
    expect(wrapper.find('[role="dialog"]').exists()).toBe(true)

    await wrapper.get('[role="dialog"]').trigger('keydown', { key: 'Escape' })
    await wrapper.vm.$nextTick()
    expect(wrapper.find('[role="dialog"]').exists()).toBe(false)
  })

  it('Tab 被焦点陷阱拦截，焦点留在输入框', async () => {
    const wrapper = await mountPalette()

    window.dispatchEvent(new KeyboardEvent('keydown', { key: 'k', ctrlKey: true }))
    await wrapper.vm.$nextTick()

    const input = wrapper.get('input')
    await input.trigger('focus')

    const event = new KeyboardEvent('keydown', { key: 'Tab', cancelable: true })
    wrapper.get('[role="dialog"]').element.dispatchEvent(event)
    await wrapper.vm.$nextTick()

    expect(event.defaultPrevented).toBe(true)
    expect(document.activeElement).toBe(wrapper.get('input').element)
  })
})
