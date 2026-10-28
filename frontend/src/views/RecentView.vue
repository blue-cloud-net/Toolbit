<script setup lang="ts">
/**
 * 最近使用页。
 *
 * 只记录访问过的工具 slug，**不**记录任何输入内容。
 */
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

import ToolCard from '@/components/discovery/ToolCard.vue'
import { useCatalogStore } from '@/stores/catalog'
import { usePrefsStore } from '@/stores/prefs'

const { t } = useI18n()
const catalog = useCatalogStore()
const prefs = usePrefsStore()

/** 按使用顺序解析 slug，跳过目录里已不存在的项。 */
const tools = computed(() =>
  prefs.recent.map((slug) => catalog.getTool(slug)).filter((tool) => tool !== undefined),
)
</script>

<template>
  <div class="list-page">
    <header class="list-page__head">
      <h1 class="list-page__title">
        {{ t('common.nav.recent') }}
      </h1>
      <p class="list-page__count">
        {{ t('home.countTools', { count: tools.length }) }}
      </p>
    </header>

    <p
      v-if="tools.length === 0"
      class="list-page__empty"
    >
      {{ t('home.recentEmpty') }}
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

    <button
      v-if="tools.length > 0"
      type="button"
      class="list-page__clear"
      @click="prefs.clearRecent()"
    >
      {{ t('home.clearRecent') }}
    </button>
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
  padding: var(--tb-space-10) 0;
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

.list-page__clear {
  align-self: flex-start;
  padding: var(--tb-space-1) var(--tb-space-3);
  font-family: inherit;
  font-size: var(--tb-font-xs);
  color: var(--tb-text-tertiary);
  background: transparent;
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-control);
  cursor: pointer;
}

.list-page__clear:hover {
  color: var(--tb-text);
  border-color: var(--tb-border-strong);
}
</style>
