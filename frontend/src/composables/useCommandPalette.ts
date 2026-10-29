/**
 * 命令面板（Command Palette）。
 *
 * 全局搜索的入口状态与条目构建逻辑：
 *
 *   - `Ctrl/⌘ + K` 全局唤起 / 关闭；
 *   - 条目来自 catalog store 的工具与分类，加上 4 个内建页；
 *   - 匹配规则与首页搜索一致：中英文名任一命中即算匹配，忽略大小写；
 *   - 目录尚未加载时给出「加载中」状态，不静默显示空列表。
 *
 * 面板状态是**模块级单例**：顶栏的触发器与浮层是两个组件，
 * 必须共享同一份 `open` 才能互相唤起。
 *
 * 全局快捷键（`Ctrl/⌘ + K`）由 `CommandPalette` 组件独占注册 ——
 * 本 composable 会被多个组件调用，若在此注册监听器，
 * 一次按键会被 toggle 多次导致开关抵消。
 */
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import type { RouteLocationRaw } from 'vue-router'

import { useCatalogStore } from '@/stores/catalog'
import { isImplemented } from '@/tools'

/** 命令面板条目的类型。 */
export type PaletteItemKind = 'tool' | 'category' | 'page'

/** 命令面板条目。 */
export interface PaletteItem {
  kind: PaletteItemKind
  /** 中文标签。 */
  label: string
  /** 英文标签。 */
  labelEn: string
  /** 跳转目标。 */
  to: RouteLocationRaw
  /** 所属分类 slug（工具条目有）。 */
  category?: string
  /** 工具是否已在前端实现。 */
  implemented: boolean
}

// ── 模块级单例状态 ──────────────────────────────────────────────
const open = ref(false)
const keyword = ref('')
const activeIndex = ref(0)

/** 匹配模式：前缀优先，其次包含。 */
function score(item: PaletteItem, query: string): number {
  const label = item.label.toLowerCase()
  const labelEn = item.labelEn.toLowerCase()
  if (label.startsWith(query) || labelEn.startsWith(query)) {
    return 0
  }
  if (label.includes(query) || labelEn.includes(query)) {
    return 1
  }
  return -1
}

export function useCommandPalette() {
  const { t } = useI18n()
  const catalog = useCatalogStore()

  /** 全部可搜索条目。 */
  const items = computed<PaletteItem[]>(() => {
    const tools: PaletteItem[] = catalog.tools.map((tool) => ({
      kind: 'tool',
      label: tool.name,
      labelEn: tool.nameEn,
      to: { name: 'tool', params: { slug: tool.slug } },
      category: tool.category,
      implemented: isImplemented(tool.slug),
    }))

    const categories: PaletteItem[] = catalog.categories.map((cat) => ({
      kind: 'category',
      label: cat.name,
      labelEn: cat.nameEn,
      to: { name: 'category', params: { category: cat.slug } },
      implemented: true,
    }))

    const pages: PaletteItem[] = [
      {
        kind: 'page',
        label: t('common.nav.home'),
        labelEn: 'Home',
        to: { name: 'home' },
        implemented: true,
      },
      {
        kind: 'page',
        label: t('common.nav.recent'),
        labelEn: 'Recent',
        to: { name: 'recent' },
        implemented: true,
      },
      {
        kind: 'page',
        label: t('common.nav.favorites'),
        labelEn: 'Favorites',
        to: { name: 'favorites' },
        implemented: true,
      },
      {
        kind: 'page',
        label: t('common.nav.allTools'),
        labelEn: 'All Tools',
        to: { name: 'home', query: { view: 'all' } },
        implemented: true,
      },
    ]

    return [...tools, ...categories, ...pages]
  })

  /** 过滤后的条目：前缀命中排前。 */
  const results = computed<PaletteItem[]>(() => {
    const query = keyword.value.trim().toLowerCase()
    if (query === '') {
      return items.value
    }
    return items.value
      .map((item) => ({ item, rank: score(item, query) }))
      .filter((entry) => entry.rank >= 0)
      .sort((a, b) => a.rank - b.rank)
      .map((entry) => entry.item)
  })

  /** 目录是否仍在加载。 */
  const loading = computed(() => catalog.loading || (!catalog.loaded && catalog.tools.length === 0))

  /** 打开面板并重置状态。 */
  function show(): void {
    keyword.value = ''
    activeIndex.value = 0
    open.value = true
  }

  /** 关闭面板。 */
  function hide(): void {
    open.value = false
  }

  /** 切换面板显隐。 */
  function toggle(): void {
    if (open.value) {
      hide()
    } else {
      show()
    }
  }

  /** 移动选中项；`delta` 为 +1 下移、-1 上移，结果首尾循环。 */
  function move(delta: number): void {
    const total = results.value.length
    if (total === 0) {
      activeIndex.value = 0
      return
    }
    activeIndex.value = (activeIndex.value + delta + total) % total
  }

  /** 全局快捷键：`Ctrl/⌘ + K` 切换。 */
  function onKeydown(event: KeyboardEvent): void {
    if ((event.ctrlKey || event.metaKey) && event.key.toLowerCase() === 'k') {
      event.preventDefault()
      toggle()
    }
  }

  return {
    open,
    keyword,
    activeIndex,
    items,
    results,
    loading,
    show,
    hide,
    toggle,
    move,
    onKeydown,
  }
}

/** 重置面板状态（仅供测试使用）。 */
export function resetPalette(): void {
  open.value = false
  keyword.value = ''
  activeIndex.value = 0
}
