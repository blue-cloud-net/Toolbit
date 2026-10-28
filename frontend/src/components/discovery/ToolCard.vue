<script setup lang="ts">
/**
 * 工具卡片。
 *
 * 未实现的工具卡片仍可点击，但标注「即将上线」并在工具页显示说明。
 */
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

import { isImplemented } from '@/tools'
import type { ToolMeta } from '@/types/tools'

const props = defineProps<{
  /** 工具元信息。 */
  tool: ToolMeta
}>()

const { t, locale } = useI18n()

const implemented = computed(() => isImplemented(props.tool.slug))
const displayName = computed(() => (locale.value === 'en-US' ? props.tool.nameEn : props.tool.name))
const displayDesc = computed(() => {
  const key = `tools.${props.tool.slug}.description`
  const translated = t(key)
  return translated === key ? '' : translated
})

/** 无词条时用英文名兜底，但避免与标题重复。 */
const fallbackText = computed(() =>
  props.tool.nameEn === displayName.value ? '' : props.tool.nameEn,
)
</script>

<template>
  <RouterLink
    class="tool-card"
    :class="{ 'tool-card--soon': !implemented }"
    :to="{ name: 'tool', params: { slug: tool.slug } }"
  >
    <div class="tool-card__head">
      <h3 class="tool-card__title">
        {{ displayName }}
      </h3>
      <span
        v-if="!implemented"
        class="tool-card__badge"
      >{{ t('home.soon') }}</span>
    </div>
    <p
      v-if="displayDesc"
      class="tool-card__desc"
    >
      {{ displayDesc }}
    </p>
    <p
      v-else-if="fallbackText"
      class="tool-card__desc tool-card__desc--muted"
    >
      {{ fallbackText }}
    </p>
  </RouterLink>
</template>

<style scoped>
.tool-card {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-1);
  padding: var(--tb-space-3) var(--tb-space-4);
  background: var(--tb-bg-elevated);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-card);
  transition: border-color var(--tb-transition), transform var(--tb-transition);
}

.tool-card:hover {
  border-color: var(--tb-border-strong);
  transform: translateY(-1px);
  text-decoration: none;
}

.tool-card--soon {
  background: var(--tb-bg-sunken);
}

.tool-card__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--tb-space-2);
}

.tool-card__title {
  margin: 0;
  font-size: var(--tb-font-base);
  font-weight: 600;
  color: var(--tb-text);
}

.tool-card--soon .tool-card__title {
  color: var(--tb-text-secondary);
}

.tool-card__badge {
  flex: none;
  padding: 0 var(--tb-space-1);
  font-size: 10px;
  color: var(--tb-text-tertiary);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-pill);
}

.tool-card__desc {
  margin: 0;
  font-size: var(--tb-font-xs);
  line-height: 1.5;
  color: var(--tb-text-secondary);
}

.tool-card__desc--muted {
  color: var(--tb-text-tertiary);
}
</style>
