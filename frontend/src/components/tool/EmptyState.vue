<script setup lang="ts">
/**
 * 空状态。
 *
 * 结果区 / 列表区在没有数据时的占位：图标 + 标题 + 提示。
 */
import { useI18n } from 'vue-i18n'

import AppIcon from '@/components/common/AppIcon.vue'

withDefaults(
  defineProps<{
    /** 标题。 */
    title: string
    /** 补充提示。 */
    hint?: string
    /** 图标名。 */
    icon?: 'document' | 'search' | 'star' | 'clock'
  }>(),
  { hint: '', icon: 'document' },
)

const { t } = useI18n()
</script>

<template>
  <div class="empty-state">
    <span
      class="empty-state__icon"
      aria-hidden="true"
    >
      <AppIcon
        :name="icon"
        :size="24"
      />
    </span>
    <p class="empty-state__title">
      {{ title }}
    </p>
    <p
      v-if="hint"
      class="empty-state__hint"
    >
      {{ hint }}
    </p>
    <slot />
    <span class="tb-sr-only">{{ t('common.states.empty') }}</span>
  </div>
</template>

<style scoped>
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--tb-space-2);
  padding: var(--tb-space-6) var(--tb-space-4);
  text-align: center;
}

.empty-state__icon {
  display: grid;
  place-items: center;
  width: 44px;
  height: 44px;
  color: var(--tb-text-tertiary);
  background: var(--tb-bg-sunken);
  border-radius: var(--tb-radius-control);
}

.empty-state__title {
  margin: 0;
  font-size: var(--tb-font-sm);
  font-weight: 600;
  color: var(--tb-text-secondary);
}

.empty-state__hint {
  margin: 0;
  font-size: var(--tb-font-xs);
  color: var(--tb-text-tertiary);
}
</style>
