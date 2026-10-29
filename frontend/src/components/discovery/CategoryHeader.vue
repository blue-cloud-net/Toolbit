<script setup lang="ts">
/**
 * 分类分组标题。
 *
 * 首页与分类页按分类分组展示工具，分组标题**可点击**，进入该分类的工具列表页。
 * 标题左侧的图标胶囊使用分类自己的色调（每个分类一个颜色）。
 */
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

import AppIcon from '@/components/common/AppIcon.vue'
import { categoryIcon } from '@/tools/categories'
import { categoryTone, toneStyle } from '@/tools/tones'
import type { CategoryMeta } from '@/types/tools'

const props = defineProps<{
  /** 分类元信息。 */
  category: CategoryMeta
  /** 该分类下当前展示的工具数。 */
  count: number
}>()

const { t, locale } = useI18n()

const displayName = computed(() =>
  locale.value === 'en-US' ? props.category.nameEn : props.category.name,
)

/** 分类图标名。 */
const iconName = computed(() => categoryIcon(props.category.icon))

/** 分类色调。 */
const tone = computed(() => categoryTone(props.category.slug))

/** 色调对应的 CSS 变量（前景色 + 底色）。 */
const toneVars = computed(() => toneStyle(tone.value))
</script>

<template>
  <RouterLink
    class="category-header"
    :style="toneVars"
    :to="{ name: 'category', params: { category: category.slug } }"
  >
    <span
      class="category-header__icon"
      aria-hidden="true"
    >
      <AppIcon
        :name="iconName"
        :size="16"
      />
    </span>
    <h3 class="category-header__title">
      {{ displayName }}
    </h3>
    <span class="category-header__count">{{ t('home.countTools', { count }) }}</span>
    <span class="category-header__more">
      {{ t('home.viewAll') }}
      <AppIcon
        name="chevron"
        :size="14"
      />
    </span>
  </RouterLink>
</template>

<style scoped>
.category-header {
  display: flex;
  align-items: center;
  gap: var(--tb-space-2);
  color: var(--tb-text);
  text-decoration: none;
}

/* 分类图标胶囊：每个分类一个颜色 */
.category-header__icon {
  display: grid;
  place-items: center;
  flex: none;
  width: 28px;
  height: 28px;
  color: var(--tb-tone, var(--tb-primary));
  background: var(--tb-tone-soft, var(--tb-primary-soft));
  border-radius: var(--tb-radius-control);
}

.category-header__title {
  margin: 0;
  font-size: var(--tb-font-lg);
  font-weight: 600;
  color: var(--tb-text);
  transition: color var(--tb-transition);
}

.category-header:hover .category-header__title {
  color: var(--tb-tone, var(--tb-primary));
}

.category-header__count {
  font-size: var(--tb-font-xs);
  color: var(--tb-text-tertiary);
}

.category-header__more {
  display: inline-flex;
  align-items: center;
  gap: 2px;
  margin-left: auto;
  font-size: var(--tb-font-xs);
  color: var(--tb-text-tertiary);
  transition: color var(--tb-transition);
}

.category-header:hover .category-header__more {
  color: var(--tb-tone, var(--tb-primary));
}
</style>