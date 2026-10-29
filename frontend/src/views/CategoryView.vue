<script setup lang="ts">
/**
 * 分类页：列出某个分类下的全部工具。
 */
import { computed, watch } from 'vue'
import { useI18n } from 'vue-i18n'

import ToolCard from '@/components/discovery/ToolCard.vue'
import { setDocumentTitle } from '@/router'
import { useCatalogStore } from '@/stores/catalog'

const props = defineProps<{
  /** 分类 slug。 */
  category: string
}>()

const { t, locale } = useI18n()
const catalog = useCatalogStore()

const meta = computed(() => catalog.categoryMap.get(props.category))
const tools = computed(() => catalog.getToolsByCategory(props.category))

const title = computed(() => {
  const cat = meta.value
  if (!cat) {
    return props.category
  }
  return locale.value === 'en-US' ? cat.nameEn : cat.name
})

/** 目录数据就绪后同步页面标题。 */
watch(
  title,
  (value) => {
    if (meta.value) {
      setDocumentTitle(value)
    }
  },
  { immediate: true },
)
</script>

<template>
  <div class="category">
    <nav
      class="category__breadcrumb"
      :aria-label="t('common.nav.categories')"
    >
      <RouterLink :to="{ name: 'home' }">
        {{ t('common.nav.home') }}
      </RouterLink>
      <span aria-hidden="true">/</span>
      <span>{{ title }}</span>
    </nav>

    <h1 class="category__title">
      {{ title }}
    </h1>
    <p class="category__count">
      {{ t('home.toolCount', { count: tools.length }) }}
    </p>

    <p
      v-if="!meta && catalog.loaded"
      class="category__empty"
    >
      {{ t('common.states.notFound') }}
    </p>

    <div
      v-else
      class="category__grid"
    >
      <ToolCard
        v-for="tool in tools"
        :key="tool.slug"
        :tool="tool"
      />
    </div>
  </div>
</template>

<style scoped>
.category {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-3);
}

.category__breadcrumb {
  display: flex;
  gap: var(--tb-space-1);
  font-size: var(--tb-font-xs);
  color: var(--tb-text-tertiary);
}

.category__title {
  margin: 0;
  font-size: var(--tb-font-2xl);
  font-weight: 700;
}

.category__count {
  margin: 0;
  font-size: var(--tb-font-xs);
  color: var(--tb-text-tertiary);
}

.category__grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: var(--tb-space-3);
  margin-top: var(--tb-space-2);
}

.category__empty {
  color: var(--tb-text-tertiary);
}
</style>
