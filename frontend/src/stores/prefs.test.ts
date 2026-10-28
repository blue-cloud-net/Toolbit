/**
 * 偏好设置单测：收藏、最近使用与持久化。
 */
import { createPinia, setActivePinia } from 'pinia'
import { nextTick } from 'vue'
import { beforeEach, describe, expect, it } from 'vitest'

import {
  FAVORITES_STORAGE_KEY,
  RECENT_LIMIT,
  RECENT_STORAGE_KEY,
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
