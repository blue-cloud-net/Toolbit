<script setup lang="ts">
/**
 * 分类卡片：显示分类名与该分类下的工具数。
 */
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

import { useCatalogStore } from '@/stores/catalog'
import type { CategoryMeta } from '@/types/tools'

const props = defineProps<{
  /** 分类元信息。 */
  category: CategoryMeta
}>()

const { t, locale } = useI18n()
const catalog = useCatalogStore()

const displayName = computed(() =>
  locale.value === 'en-US' ? props.category.nameEn : props.category.name,
)

const count = computed(() => catalog.getToolsByCategory(props.category.slug).length)
</script>

<template>
  <RouterLink
    class="category-card"
    :to="{ name: 'category', params: { category: category.slug } }"
  >
    <span class="category-card__name">{{ displayName }}</span>
    <span class="category-card__count">{{ t('home.toolCount', { count }) }}</span>
  </RouterLink>
</template>

<style scoped>
.category-card {
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: var(--tb-space-2) var(--tb-space-3);
  background: var(--tb-bg-elevated);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-card);
  transition: border-color var(--tb-transition);
}

.category-card:hover {
  border-color: var(--tb-primary);
  text-decoration: none;
}

.category-card__name {
  font-size: var(--tb-font-sm);
  font-weight: 600;
  color: var(--tb-text);
}

.category-card__count {
  font-size: var(--tb-font-xs);
  color: var(--tb-text-tertiary);
}
</style>
