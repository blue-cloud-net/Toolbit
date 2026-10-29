<script setup lang="ts">
/**
 * 首页：搜索 + 分类 + 全部工具。
 */
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import CategoryCard from '@/components/discovery/CategoryCard.vue'
import ToolCard from '@/components/discovery/ToolCard.vue'
import { useGlobalHotkeys } from '@/composables/useGlobalHotkeys'
import { useCatalogStore } from '@/stores/catalog'
import { isImplemented } from '@/tools'

const { t } = useI18n()
const catalog = useCatalogStore()

const keyword = ref('')
const searchInput = ref<HTMLInputElement | null>(null)

/** `/` 聚焦首页搜索框。 */
useGlobalHotkeys({
  focusSearch: () => searchInput.value,
})

/** 是否只显示已上线的工具。 */
const onlyImplemented = ref(false)

const sortedCategories = computed(() => [...catalog.categories].sort((a, b) => a.order - b.order))

/** 已实现的工具数。 */
const implementedCount = computed(
  () => catalog.tools.filter((tool) => isImplemented(tool.slug)).length,
)

/** 按中英文名过滤工具。 */
const filteredTools = computed(() => {
  const q = keyword.value.trim().toLowerCase()
  let list = [...catalog.tools].sort((a, b) => {
    if (a.category !== b.category) {
      return a.category.localeCompare(b.category)
    }
    return a.order - b.order
  })
  if (onlyImplemented.value) {
    list = list.filter((tool) => isImplemented(tool.slug))
  }
  if (q === '') {
    return list
  }
  return list.filter(
    (tool) => tool.name.toLowerCase().includes(q) || tool.nameEn.toLowerCase().includes(q),
  )
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
        <input
          ref="searchInput"
          v-model="keyword"
          type="search"
          class="home__search-input"
          :placeholder="t('home.searchPlaceholder')"
        >
      </label>
    </section>

    <section
      v-if="!isSearching"
      class="home__section"
    >
      <h2 class="home__section-title">
        {{ t('home.categoriesTitle') }}
      </h2>
      <div class="home__categories">
        <CategoryCard
          v-for="cat in sortedCategories"
          :key="cat.slug"
          :category="cat"
        />
      </div>
    </section>

    <section class="home__section">
      <h2 class="home__section-title">
        {{ isSearching ? t('home.toolsTitle') : t('home.toolsTitle') }}
        <span class="home__count">{{ t('home.toolCount', { count: filteredTools.length }) }}</span>
        <span class="home__progress">
          {{ t('home.enabledProgress', { done: implementedCount, total: catalog.tools.length }) }}
        </span>
        <label class="home__filter">
          <input
            v-model="onlyImplemented"
            type="checkbox"
          >
          <span>{{ t('home.enabledOnly') }}</span>
        </label>
      </h2>

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
      <p
        v-else-if="filteredTools.length === 0"
        class="home__empty"
      >
        {{ t('home.searchEmpty') }}
      </p>

      <div
        v-else
        class="home__grid"
      >
        <ToolCard
          v-for="tool in filteredTools"
          :key="tool.slug"
          :tool="tool"
        />
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
  padding: var(--tb-space-2) var(--tb-space-3);
  font-family: inherit;
  font-size: var(--tb-font-base);
  color: var(--tb-text);
  background: var(--tb-bg-elevated);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-control);
}

.home__search-input:focus-visible {
  border-color: var(--tb-primary);
}

.home__section {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-3);
}

.home__section-title {
  position: sticky;
  top: var(--tb-header-h);
  z-index: 5;
  display: flex;
  align-items: baseline;
  gap: var(--tb-space-2);
  margin: 0;
  padding: var(--tb-space-2) 0;
  font-size: var(--tb-font-lg);
  font-weight: 600;
  background: var(--tb-bg);
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
  display: flex;
  align-items: center;
  gap: var(--tb-space-1);
  margin-left: auto;
  font-size: var(--tb-font-sm);
  font-weight: 400;
  color: var(--tb-text-secondary);
  cursor: pointer;
}

.home__filter input {
  width: 16px;
  height: 16px;
  accent-color: var(--tb-primary);
  cursor: pointer;
}

.home__categories {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(150px, 1fr));
  gap: var(--tb-space-2);
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
