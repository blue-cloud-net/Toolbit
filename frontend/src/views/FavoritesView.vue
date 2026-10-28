<script setup lang="ts">
/**
 * 我的收藏页。
 *
 * 收藏只存 slug，**不**存任何工具输入。
 */
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

import ToolCard from '@/components/discovery/ToolCard.vue'
import { useCatalogStore } from '@/stores/catalog'
import { usePrefsStore } from '@/stores/prefs'

const { t } = useI18n()
const catalog = useCatalogStore()
const prefs = usePrefsStore()

/** 按收藏顺序解析 slug，跳过目录里已不存在的项。 */
const tools = computed(() =>
  prefs.favorites.map((slug) => catalog.getTool(slug)).filter((tool) => tool !== undefined),
)
</script>

<template>
  <div class="list-page">
    <header class="list-page__head">
      <h1 class="list-page__title">
        {{ t('common.nav.favorites') }}
      </h1>
      <p class="list-page__count">
        {{ t('home.countTools', { count: tools.length }) }}
      </p>
    </header>

    <p
      v-if="tools.length === 0"
      class="list-page__empty"
    >
      {{ t('home.favoritesEmpty') }}
    </p>

    <div
      v-else
      class="list-page__grid"
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
.list-page {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-3);
}

.list-page__head {
  display: flex;
  align-items: baseline;
  gap: var(--tb-space-2);
}

.list-page__title {
  margin: 0;
  font-size: var(--tb-font-2xl);
  font-weight: 700;
}

.list-page__count {
  margin: 0;
  font-size: var(--tb-font-xs);
  color: var(--tb-text-tertiary);
}

.list-page__empty {
  margin: 0;
  padding: var(--tb-space-10) var(--tb-space-4);
  font-size: var(--tb-font-sm);
  color: var(--tb-text-tertiary);
  text-align: center;
  background: var(--tb-bg-elevated);
  border: 1px dashed var(--tb-border);
  border-radius: var(--tb-radius-panel);
}

.list-page__grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(220px, 1fr));
  gap: var(--tb-space-3);
}
</style>
