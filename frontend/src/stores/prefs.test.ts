/**
 * 偏好设置单测：收藏、最近使用与持久化。
 */
import { createPinia, setActivePinia } from 'pinia'
import { nextTick } from 'vue'
import { beforeEach, describe, expect, it } from 'vitest'

import { DEFAULT_LOCALE, LOCALE_STORAGE_KEY } from '@/i18n'
import {
  ACCENT_STORAGE_KEY,
  FAVORITES_STORAGE_KEY,
  RECENT_LIMIT,
  RECENT_STORAGE_KEY,
  THEME_STORAGE_KEY,
  usePrefsStore,
} from '@/stores/prefs'

beforeEach(() => {
  localStorage.clear()
  setActivePinia(createPinia())
})

describe('收藏', () => {
  it('切换后写入 localStorage，且最近收藏在前', async () => {
    const prefs = usePrefsStore()

    prefs.toggleFavorite('base64')
    prefs.toggleFavorite('number-base')
    await nextTick()

    expect(prefs.isFavorite('base64')).toBe(true)
    expect(prefs.favorites).toEqual(['number-base', 'base64'])
    expect(JSON.parse(localStorage.getItem(FAVORITES_STORAGE_KEY) ?? '[]')).toEqual([
      'number-base',
      'base64',
    ])
  })

  it('再次切换会取消收藏', () => {
    const prefs = usePrefsStore()

    prefs.toggleFavorite('base64')
    prefs.toggleFavorite('base64')

    expect(prefs.isFavorite('base64')).toBe(false)
    expect(prefs.favorites).toEqual([])
  })
})

describe('最近使用', () => {
  it('去重并把最近一次排到最前', () => {
    const prefs = usePrefsStore()

    prefs.recordRecent('base64')
    prefs.recordRecent('number-base')
    prefs.recordRecent('base64')

    expect(prefs.recent).toEqual(['base64', 'number-base'])
  })

  it('超过上限时截断', async () => {
    const prefs = usePrefsStore()

    for (let i = 0; i < RECENT_LIMIT + 5; i += 1) {
      prefs.recordRecent(`tool-${i}`)
    }
    await nextTick()

    expect(prefs.recent).toHaveLength(RECENT_LIMIT)
    expect(prefs.recent[0]).toBe(`tool-${RECENT_LIMIT + 4}`)
    expect(JSON.parse(localStorage.getItem(RECENT_STORAGE_KEY) ?? '[]')).toHaveLength(RECENT_LIMIT)
  })

  it('clearRecent 清空记录', () => {
    const prefs = usePrefsStore()

    prefs.recordRecent('base64')
    prefs.clearRecent()

    expect(prefs.recent).toEqual([])
  })

  it('非法存储内容回退为空数组', () => {
    localStorage.setItem(RECENT_STORAGE_KEY, '{"not":"an array"}')
    setActivePinia(createPinia())

    expect(usePrefsStore().recent).toEqual([])
  })
})

describe('主色', () => {
  it('默认使用靛蓝', () => {
    expect(usePrefsStore().accent).toBe('indigo')
  })

  it('切换后写入 localStorage', async () => {
    const prefs = usePrefsStore()

    prefs.setAccent('teal')
    await nextTick()

    expect(prefs.accent).toBe('teal')
    expect(localStorage.getItem(ACCENT_STORAGE_KEY)).toBe('teal')
  })

  it('非法存储值回退到默认', () => {
    localStorage.setItem(ACCENT_STORAGE_KEY, 'not-a-color')
    setActivePinia(createPinia())

    expect(usePrefsStore().accent).toBe('indigo')
  })

  it('同步到 html 的 data-accent', async () => {
    const prefs = usePrefsStore()

    prefs.setAccent('rose')
    await nextTick()

    expect(document.documentElement.dataset.accent).toBe('rose')
  })
})

describe('主题', () => {
  it('默认跟随系统', () => {
    expect(usePrefsStore().theme).toBe('system')
  })

  it('切换后写入 localStorage 并同步 data-theme', async () => {
    const prefs = usePrefsStore()

    prefs.setTheme('dark')
    await nextTick()

    expect(prefs.resolvedTheme).toBe('dark')
    expect(localStorage.getItem(THEME_STORAGE_KEY)).toBe('dark')
    expect(document.documentElement.dataset.theme).toBe('dark')
  })

  it('非法存储值回退到 system', () => {
    localStorage.setItem(THEME_STORAGE_KEY, 'neon')
    setActivePinia(createPinia())

    expect(usePrefsStore().theme).toBe('system')
  })
})

describe('语言', () => {
  it('非法存储值回退到默认语言', () => {
    localStorage.setItem(LOCALE_STORAGE_KEY, 'fr-FR')
    setActivePinia(createPinia())

    expect(usePrefsStore().locale).toBe(DEFAULT_LOCALE)
  })

  it('切换后写入 localStorage', async () => {
    const prefs = usePrefsStore()

    prefs.setLocale('en-US')
    await nextTick()

    expect(localStorage.getItem(LOCALE_STORAGE_KEY)).toBe('en-US')
  })
})
