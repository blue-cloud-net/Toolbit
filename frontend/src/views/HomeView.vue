<script setup lang="ts">
/**
 * 首页：搜索 + 按分类分组的全部工具。
 *
 * 工具**按分类分组**展示，每个分组的标题可点击，进入该分类的工具列表页。
 */
import { ElCheckbox } from 'element-plus/es/components/checkbox/index'
import 'element-plus/es/components/checkbox/style/css'
import { ElInput, type InputInstance } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import CategoryHeader from '@/components/discovery/CategoryHeader.vue'
import ToolCard from '@/components/discovery/ToolCard.vue'
import { useGlobalHotkeys } from '@/composables/useGlobalHotkeys'
import { useCatalogStore } from '@/stores/catalog'
import { isImplemented } from '@/tools'
import type { ToolMeta } from '@/types/tools'

const { t } = useI18n()
const catalog = useCatalogStore()

const keyword = ref('')
const searchInput = ref<InputInstance | null>(null)

/** `/` 聚焦首页搜索框。 */
useGlobalHotkeys({
  focusSearch: () => searchInput.value?.input ?? null,
})

/** 是否只显示已上线的工具。 */
const onlyImplemented = ref(false)

const sortedCategories = computed(() => [...catalog.categories].sort((a, b) => a.order - b.order))

/** 已实现的工具数。 */
const implementedCount = computed(
  () => catalog.tools.filter((tool) => isImplemented(tool.slug)).length,
)

/** 按中英文名过滤工具。 */
function matches(tool: ToolMeta): boolean {
  const q = keyword.value.trim().toLowerCase()
  if (q === '') {
    return true
  }
  return tool.name.toLowerCase().includes(q) || tool.nameEn.toLowerCase().includes(q)
}

/** 命中过滤条件的工具总数（用于标题上的计数）。 */
const filteredCount = computed(
  () =>
    catalog.tools.filter(
      (tool) => (!onlyImplemented.value || isImplemented(tool.slug)) && matches(tool),
    ).length,
)

/**
 * 按分类分组的工具。
 *
 * 分类顺序跟随 catalog 的 `order`；组内按工具自身的 `order` 排列。
 * 未登记在 catalog 中的工具归入末尾的「其他」分组，避免漏掉。
 */
const groupedTools = computed(() => {
  const groups = sortedCategories.value
    .map((cat) => ({
      category: cat,
      tools: catalog
        .getToolsByCategory(cat.slug)
        .filter((tool) => (!onlyImplemented.value || isImplemented(tool.slug)) && matches(tool))
        .sort((a, b) => a.order - b.order),
    }))
    .filter((group) => group.tools.length > 0)

  const known = new Set(sortedCategories.value.map((cat) => cat.slug))
  const others = catalog.tools
    .filter((tool) => !known.has(tool.category))
    .filter((tool) => (!onlyImplemented.value || isImplemented(tool.slug)) && matches(tool))
    .sort((a, b) => a.order - b.order)
  if (others.length > 0) {
    groups.push({
      category: {
        slug: '__other__',
        name: '其他',
        nameEn: 'Others',
        order: Number.MAX_SAFE_INTEGER,
        icon: 'document',
      },
      tools: others,
    })
  }

  return groups
})

const isSearching = computed(() => keyword.value.trim() !== '')
</script>

<template>
  <div class="home">
    <section class="home__hero">
      <h1 class="home__title">
        {{ t('home.title') }}
      </h1>
      <p class="home__subtitle">
        {{ t('home.subtitle') }}
      </p>

      <label class="home__search">
        <span class="tb-sr-only">{{ t('common.actions.search') }}</span>
        <ElInput
          ref="searchInput"
          v-model="keyword"
          class="home__search-input"
          :placeholder="t('home.searchPlaceholder')"
          :aria-label="t('common.actions.search')"
          clearable
        />
      </label>
    </section>

    <p
      v-if="catalog.error"
      class="home__error"
    >
      {{ t('common.states.failed') }}：{{ catalog.error }}
    </p>
    <p
      v-else-if="catalog.loading"
      class="home__loading"
    >
      {{ t('common.states.loading') }}…
    </p>

    <!-- 分类分组：标题可点进分类页，组内是工具卡片 -->
    <section
      v-else
      class="home__section"
    >
      <div class="home__section-head">
        <h2 class="home__section-title">
          {{ isSearching ? t('home.searchResultTitle') : t('home.toolsTitle') }}
        </h2>
        <span class="home__count">{{ t('home.toolCount', { count: filteredCount }) }}</span>
        <span class="home__progress">
          {{ t('home.enabledProgress', { done: implementedCount, total: catalog.tools.length }) }}
        </span>
        <ElCheckbox
          v-model="onlyImplemented"
          class="home__filter"
        >
          {{ t('home.enabledOnly') }}
        </ElCheckbox>
      </div>

      <p
        v-if="groupedTools.length === 0"
        class="home__empty"
      >
        {{ t('home.searchEmpty') }}
      </p>

      <div
        v-else
        class="home__groups"
      >
        <section
          v-for="group in groupedTools"
          :key="group.category.slug"
          class="home__group"
        >
          <CategoryHeader
            :category="group.category"
            :count="group.tools.length"
          />
          <div class="home__grid">
            <ToolCard
              v-for="tool in group.tools"
              :key="tool.slug"
              :tool="tool"
            />
          </div>
        </section>
      </div>
    </section>
  </div>
</template>

<style scoped>
.home {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-8);
}

.home__hero {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-2);
}

.home__title {
  margin: 0;
  font-size: var(--tb-font-3xl);
  font-weight: 700;
  letter-spacing: -0.01em;
}

.home__subtitle {
  margin: 0;
  color: var(--tb-text-secondary);
}

.home__search {
  margin-top: var(--tb-space-2);
  max-width: 420px;
}

.home__search-input {
  width: 100%;
}

.home__section {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-3);
}

/* 分组标题行：粘在顶栏下方，滚动时始终可见 */
.home__section-head {
  position: sticky;
  top: var(--tb-header-h);
  z-index: 5;
  display: flex;
  align-items: baseline;
  gap: var(--tb-space-2);
  margin: 0;
  padding: var(--tb-space-2) 0;
  background: var(--tb-bg);
}

.home__section-title {
  margin: 0;
  font-size: var(--tb-font-lg);
  font-weight: 600;
}

.home__count {
  font-size: var(--tb-font-xs);
  font-weight: 400;
  color: var(--tb-text-tertiary);
}

.home__progress {
  font-size: var(--tb-font-xs);
  font-weight: 400;
  color: var(--tb-text-tertiary);
}

.home__filter {
  margin-left: auto;
  font-size: var(--tb-font-sm);
  font-weight: 400;
  color: var(--tb-text-secondary);
}

/* 分类分组：每个分类一节，节内是工具网格 */
.home__groups {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-6);
}

.home__group {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-3);
}

.home__grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: var(--tb-space-3);
}

.home__loading,
.home__empty,
.home__error {
  margin: 0;
  font-size: var(--tb-font-sm);
  color: var(--tb-text-tertiary);
}

.home__error {
  color: var(--tb-danger);
}
</style>
