<script setup lang="ts">
/**
 * 分类卡片：显示分类图标、分类名与该分类下的工具数。
 */
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

import AppIcon from '@/components/common/AppIcon.vue'
import { useCatalogStore } from '@/stores/catalog'
import { categoryIcon } from '@/tools/categories'
import { categoryTone, toneStyle } from '@/tools/tones'
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

/** 分类图标名。 */
const iconName = computed(() => categoryIcon(props.category.icon))

/** 分类色调。 */
const tone = computed(() => categoryTone(props.category.slug))

/** 色调对应的 CSS 变量（前景色 + 底色）。 */
const toneVars = computed(() => toneStyle(tone.value))
</script>

<template>
  <RouterLink
    class="category-card"
    :style="toneVars"
    :to="{ name: 'category', params: { category: category.slug } }"
  >
    <span
      class="category-card__icon"
      aria-hidden="true"
    >
      <AppIcon
        :name="iconName"
        :size="16"
      />
    </span>
    <span class="category-card__body">
      <span class="category-card__name">{{ displayName }}</span>
      <span class="category-card__count">{{ t('home.toolCount', { count }) }}</span>
    </span>
  </RouterLink>
</template>

<style scoped>
.category-card {
  display: flex;
  align-items: center;
  gap: var(--tb-space-2);
  padding: var(--tb-space-2) var(--tb-space-3);
  background: var(--tb-bg-elevated);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-card);
  transition: border-color var(--tb-transition);
}

.category-card:hover {
  border-color: var(--tb-tone, var(--tb-primary));
  text-decoration: none;
}

/* 分类图标胶囊：每个分类一个颜色 */
.category-card__icon {
  display: grid;
  place-items: center;
  flex: none;
  width: 26px;
  height: 26px;
  color: var(--tb-tone, var(--tb-primary));
  background: var(--tb-tone-soft, var(--tb-primary-soft));
  border-radius: var(--tb-radius-control);
}

.category-card__body {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
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
