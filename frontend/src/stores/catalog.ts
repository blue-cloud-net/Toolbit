/**
 * 工具目录 store。
 *
 * 目录数据来自 `GET /api/v1/tools`，只含展示元信息；
 * 「是否已实现」由 `src/tools/index.ts` 的静态登记表决定，
 * 目录响应本身**不含**该字段。
 */
import { defineStore } from 'pinia'
import { computed, ref } from 'vue'

import { ApiError, fetchCatalog } from '@/api/client'
import { isImplemented } from '@/tools'
import type { CatalogResponse, CategoryMeta, ToolMeta } from '@/types/tools'

export const useCatalogStore = defineStore('catalog', () => {
  const categories = ref<CategoryMeta[]>([])
  const tools = ref<ToolMeta[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)
  const loaded = ref(false)

  /** 分类 slug → 分类元信息。 */
  const categoryMap = computed(() => {
    const map = new Map<string, CategoryMeta>()
    for (const cat of categories.value) {
      map.set(cat.slug, cat)
    }
    return map
  })

  /** slug → 工具元信息。 */
  const toolMap = computed(() => {
    const map = new Map<string, ToolMeta>()
    for (const tool of tools.value) {
      map.set(tool.slug, tool)
    }
    return map
  })

  /** 已实现的工具（前端已登记组件的 slug）。 */
  const implementedTools = computed(() => tools.value.filter((tool) => isImplemented(tool.slug)))

  /** 按分类分组后的工具，已实现的排在前面。 */
  const groupedTools = computed(() => {
    const groups = new Map<string, ToolMeta[]>()
    for (const tool of tools.value) {
      const list = groups.get(tool.category) ?? []
      list.push(tool)
      groups.set(tool.category, list)
    }
    for (const list of groups.values()) {
      list.sort((a, b) => {
        const aImpl = isImplemented(a.slug)
        const bImpl = isImplemented(b.slug)
        if (aImpl !== bImpl) {
          return aImpl ? -1 : 1
        }
        return a.order - b.order
      })
    }
    return groups
  })

  /** 拉取目录；已加载时默认跳过。 */
  async function load(force = false): Promise<void> {
    if (loading.value) {
      return
    }
    if (loaded.value && !force) {
      return
    }
    loading.value = true
    error.value = null
    try {
      const resp = await fetchCatalog()
      applyCatalog(resp.data)
      loaded.value = true
    } catch (err) {
      error.value = err instanceof ApiError ? err.message : String(err)
    } finally {
      loading.value = false
    }
  }

  /** 写入目录数据。 */
  function applyCatalog(data: CatalogResponse): void {
    categories.value = data.categories ?? []
    tools.value = data.tools ?? []
  }

  /** 按 slug 取工具元信息。 */
  function getTool(slug: string): ToolMeta | undefined {
    return toolMap.value.get(slug)
  }

  /** 按分类 slug 取工具列表。 */
  function getToolsByCategory(category: string): ToolMeta[] {
    return groupedTools.value.get(category) ?? []
  }

  return {
    categories,
    tools,
    loading,
    error,
    loaded,
    categoryMap,
    toolMap,
    implementedTools,
    groupedTools,
    load,
    applyCatalog,
    getTool,
    getToolsByCategory,
  }
})
