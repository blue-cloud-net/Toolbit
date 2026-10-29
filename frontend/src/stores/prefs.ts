/**
 * 偏好设置。
 *
 * 只存「语言 / 主题 / 主色 / 侧栏折叠」这类非敏感偏好；
 * **禁止**把工具输入写进任何浏览器存储。
 */
import { defineStore } from 'pinia'
import { computed, nextTick, ref, watch } from 'vue'

import { DEFAULT_LOCALE, LOCALE_STORAGE_KEY, i18n, isLocale, type Locale } from '@/i18n'
import { ACCENT_PRESETS, type AccentKey } from '@/config/accents'

/** 主题模式。 */
export type ThemeMode = 'system' | 'light' | 'dark'

/** 全部可选主色（顺序即选择器展示顺序）。 */
export const ACCENT_KEYS: readonly AccentKey[] = ACCENT_PRESETS.map((item) => item.key)

/** 默认主色。 */
export const DEFAULT_ACCENT: AccentKey = 'indigo'

/** 主题在 localStorage 中的键。 */
export const THEME_STORAGE_KEY = 'toolbit.theme'

/** 主色在 localStorage 中的键。 */
export const ACCENT_STORAGE_KEY = 'toolbit.accent'

/** 侧栏折叠状态在 localStorage 中的键。 */
export const SIDEBAR_STORAGE_KEY = 'toolbit.sidebar'

/** 收藏工具 slug 列表在 localStorage 中的键。 */
export const FAVORITES_STORAGE_KEY = 'toolbit.favorites'

/** 最近使用工具 slug 列表在 localStorage 中的键。 */
export const RECENT_STORAGE_KEY = 'toolbit.recent'

/** 最近使用最多保留的条数。 */
export const RECENT_LIMIT = 12

/** 本应用写入浏览器存储的键前缀；清理缓存只动这个前缀下的键。 */
export const STORAGE_PREFIX = 'toolbit.'

/** 读取存储的偏好；无值或非法值回退到默认值。 */
function readStored(key: string): string | null {
  if (typeof localStorage === 'undefined') {
    return null
  }
  return localStorage.getItem(key)
}

/** 读取存储的字符串数组；非法内容回退为空数组。 */
function readStoredList(key: string): string[] {
  const raw = readStored(key)
  if (!raw) {
    return []
  }
  try {
    const parsed: unknown = JSON.parse(raw)
    if (!Array.isArray(parsed)) {
      return []
    }
    return parsed.filter((item): item is string => typeof item === 'string')
  } catch {
    return []
  }
}

/** 删除某个存储中所有以 `toolbit.` 开头的键，返回被删除的键名。 */
function removeStoredKeys(storage: Storage): string[] {
  const keys: string[] = []
  try {
    for (let i = 0; i < storage.length; i += 1) {
      const key = storage.key(i)
      if (key && key.startsWith(STORAGE_PREFIX)) {
        keys.push(key)
      }
    }
    for (const key of keys) {
      storage.removeItem(key)
    }
  } catch {
    // 存储被禁用（隐私模式 / 配额满）时忽略：清理失败不应阻断调用方。
  }
  return keys
}

/**
 * 删除本应用写入的全部本地偏好。
 *
 * 只按 `toolbit.` 前缀删除，不碰同域其他应用或浏览器自身的缓存；
 * 返回被删除的键名，便于调用方判断是否真的清理过。
 */
export function clearStoredData(): string[] {
  const removed: string[] = []
  if (typeof localStorage !== 'undefined') {
    removed.push(...removeStoredKeys(localStorage))
  }
  if (typeof sessionStorage !== 'undefined') {
    removed.push(...removeStoredKeys(sessionStorage))
  }
  return removed
}

/**
 * 清空本源（origin）的 Cache Storage。
 *
 * 当前没有注册 Service Worker，正常情况下为空；
 * 这里做兜底，避免旧版本遗留的离线缓存继续生效。
 */
export async function clearBrowserCaches(): Promise<number> {
  if (typeof caches === 'undefined' || typeof caches.keys !== 'function') {
    return 0
  }
  try {
    const names = await caches.keys()
    await Promise.all(names.map((name) => caches.delete(name)))
    return names.length
  } catch {
    return 0
  }
}

/** 判断是否为受支持的主题模式。 */
export function isThemeMode(value: unknown): value is ThemeMode {
  return value === 'system' || value === 'light' || value === 'dark'
}

/** 判断是否为主色标识。 */
export function isAccent(value: unknown): value is AccentKey {
  return typeof value === 'string' && (ACCENT_KEYS as readonly string[]).includes(value)
}

export const usePrefsStore = defineStore('prefs', () => {
  const storedLocale = readStored(LOCALE_STORAGE_KEY)
  const locale = ref<Locale>(isLocale(storedLocale) ? storedLocale : DEFAULT_LOCALE)

  const storedTheme = readStored(THEME_STORAGE_KEY)
  const theme = ref<ThemeMode>(isThemeMode(storedTheme) ? storedTheme : 'system')

  const storedAccent = readStored(ACCENT_STORAGE_KEY)
  const accent = ref<AccentKey>(isAccent(storedAccent) ? storedAccent : DEFAULT_ACCENT)

  const sidebarCollapsed = ref(readStored(SIDEBAR_STORAGE_KEY) === '1')

  /** 移动端导航抽屉是否打开（仅窄屏有效，不持久化）。 */
  const mobileNavOpen = ref(false)

  /** 收藏的工具 slug（有序，最近收藏在前）。 */
  const favorites = ref<string[]>(readStoredList(FAVORITES_STORAGE_KEY))

  /** 最近使用的工具 slug（有序，最近使用在前）。 */
  const recent = ref<string[]>(readStoredList(RECENT_STORAGE_KEY))

  /** 系统偏好是否为深色。 */
  const systemPrefersDark = ref(false)

  /** 实际生效的主题（把 system 解析为 light / dark）。 */
  const resolvedTheme = computed<'light' | 'dark'>(() => {
    if (theme.value === 'system') {
      return systemPrefersDark.value ? 'dark' : 'light'
    }
    return theme.value
  })

  /** 把主题写到 `<html data-theme>`，并同步 Element Plus 主色。 */
  function applyTheme(): void {
    if (typeof document === 'undefined') {
      return
    }
    document.documentElement.dataset.theme = resolvedTheme.value
    document.documentElement.dataset.accent = accent.value
  }

  /** 同步 `<html lang>`。 */
  function applyLocale(): void {
    if (typeof document === 'undefined') {
      return
    }
    document.documentElement.lang = locale.value
  }

  /** 切换语言。 */
  function setLocale(next: Locale): void {
    locale.value = next
  }

  /** 切换主题模式。 */
  function setTheme(next: ThemeMode): void {
    theme.value = next
  }

  /** 切换主色。 */
  function setAccent(next: AccentKey): void {
    accent.value = next
  }

  /** 折叠或展开侧栏。 */
  function toggleSidebar(): void {
    sidebarCollapsed.value = !sidebarCollapsed.value
  }

  /** 打开 / 关闭移动端导航抽屉。 */
  function toggleMobileNav(): void {
    mobileNavOpen.value = !mobileNavOpen.value
  }

  /** 关闭移动端导航抽屉。 */
  function closeMobileNav(): void {
    mobileNavOpen.value = false
  }

  /** 判断工具是否已收藏。 */
  function isFavorite(slug: string): boolean {
    return favorites.value.includes(slug)
  }

  /** 切换收藏状态。 */
  function toggleFavorite(slug: string): void {
    favorites.value = isFavorite(slug)
      ? favorites.value.filter((item) => item !== slug)
      : [slug, ...favorites.value]
  }

  /** 记录一次工具使用（去重并截断到上限）。 */
  function recordRecent(slug: string): void {
    recent.value = [slug, ...recent.value.filter((item) => item !== slug)].slice(0, RECENT_LIMIT)
  }

  /** 清空最近使用。 */
  function clearRecent(): void {
    recent.value = []
  }

  /** 把全部偏好恢复为默认值（只改内存态，持久化由 watch 接管）。 */
  function resetToDefaults(): void {
    locale.value = DEFAULT_LOCALE
    theme.value = 'system'
    accent.value = DEFAULT_ACCENT
    sidebarCollapsed.value = false
    favorites.value = []
    recent.value = []
  }

  /**
   * 清理本地缓存并恢复默认配置。
   *
   * 顺序：先删存储 → 重置内存态 → 等 watch 把默认值写回 → 再删一次，
   * 这样清理后浏览器存储回到「空」状态，下次访问等同首次访问。
   * 完成后调用方应重新加载页面，让分栏比例等页面内状态一并复位。
   */
  async function clearLocalData(): Promise<string[]> {
    const removed = clearStoredData()
    resetToDefaults()
    await nextTick()
    clearStoredData()
    await clearBrowserCaches()
    return removed
  }

  /** 监听系统主题变化。 */
  function watchSystemTheme(): void {
    if (typeof window === 'undefined' || !window.matchMedia) {
      return
    }
    const query = window.matchMedia('(prefers-color-scheme: dark)')
    systemPrefersDark.value = query.matches
    const onChange = (event: MediaQueryListEvent) => {
      systemPrefersDark.value = event.matches
    }
    query.addEventListener('change', onChange)
  }

  watch(locale, (next) => {
    ;(i18n.global.locale as unknown as { value: string }).value = next
    applyLocale()
    if (typeof localStorage !== 'undefined') {
      localStorage.setItem(LOCALE_STORAGE_KEY, next)
    }
  })

  watch(theme, (next) => {
    applyTheme()
    if (typeof localStorage !== 'undefined') {
      localStorage.setItem(THEME_STORAGE_KEY, next)
    }
  })

  watch(accent, (next) => {
    applyTheme()
    if (typeof localStorage !== 'undefined') {
      localStorage.setItem(ACCENT_STORAGE_KEY, next)
    }
  })

  watch(sidebarCollapsed, (next) => {
    if (typeof localStorage !== 'undefined') {
      localStorage.setItem(SIDEBAR_STORAGE_KEY, next ? '1' : '0')
    }
  })

  watch(favorites, (next) => {
    if (typeof localStorage !== 'undefined') {
      localStorage.setItem(FAVORITES_STORAGE_KEY, JSON.stringify(next))
    }
  })

  watch(recent, (next) => {
    if (typeof localStorage !== 'undefined') {
      localStorage.setItem(RECENT_STORAGE_KEY, JSON.stringify(next))
    }
  })

  watch(resolvedTheme, applyTheme)

  /** 初始化：应用持久化偏好并监听系统主题。 */
  function init(): void {
    applyLocale()
    applyTheme()
    watchSystemTheme()
  }

  return {
    locale,
    theme,
    resolvedTheme,
    accent,
    sidebarCollapsed,
    mobileNavOpen,
    systemPrefersDark,
    favorites,
    recent,
    setLocale,
    setTheme,
    setAccent,
    toggleSidebar,
    toggleMobileNav,
    closeMobileNav,
    isFavorite,
    toggleFavorite,
    recordRecent,
    clearRecent,
    resetToDefaults,
    clearLocalData,
    init,
  }
})
