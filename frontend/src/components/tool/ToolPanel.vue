<script setup lang="ts">
/**
 * 工具面板：输入区 + 参数区 + 运行栏。
 *
 * 所有工具页**必须**复用本组件，不允许各写各的。
 */
import { useI18n } from 'vue-i18n'

const { t } = useI18n()

withDefaults(
  defineProps<{
    /** 输入区标题。 */
    inputTitle?: string
    /** 参数区标题。 */
    optionsTitle?: string
  }>(),
  { inputTitle: undefined, optionsTitle: undefined },
)
</script>

<template>
  <section class="tool-panel">
    <div class="tool-panel__body">
      <div class="tool-panel__input">
        <h2 class="tool-panel__title">
          {{ inputTitle ?? t('tool.sections.input') }}
        </h2>
        <slot name="input" />
      </div>

      <div
        v-if="$slots.options"
        class="tool-panel__options"
      >
        <h2 class="tool-panel__title">
          {{ optionsTitle ?? t('tool.sections.options') }}
        </h2>
        <slot name="options" />
      </div>
    </div>

    <slot name="footer" />
  </section>
</template>

<style scoped>
.tool-panel {
  display: flex;
  flex-direction: column;
  background: var(--tb-bg-elevated);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-panel);
  overflow: hidden;
}

.tool-panel__body {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-4);
  padding: var(--tb-space-4);
}

.tool-panel__title {
  margin: 0 0 var(--tb-space-2);
  font-size: var(--tb-font-sm);
  font-weight: 600;
  color: var(--tb-text-secondary);
}

.tool-panel__options {
  display: flex;
  flex-direction: column;
}
</style>
