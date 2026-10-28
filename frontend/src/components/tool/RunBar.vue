<script setup lang="ts">
/**
 * 运行栏：运行 / 重置 / 填入示例。
 *
 * 移动端吸底（`position: sticky`），桌面端跟随面板。
 */
import { ElButton } from 'element-plus/es/components/button/index'
import 'element-plus/es/components/button/style/css'
import { useI18n } from 'vue-i18n'

withDefaults(
  defineProps<{
    /** 是否正在运行。 */
    running?: boolean
    /** 是否可运行。 */
    disabled?: boolean
    /** 是否显示示例按钮。 */
    showExample?: boolean
  }>(),
  { running: false, disabled: false, showExample: true },
)

const emit = defineEmits<{
  /** 运行。 */
  run: []
  /** 重置。 */
  reset: []
  /** 填入示例。 */
  example: []
}>()

const { t } = useI18n()
</script>

<template>
  <div class="run-bar">
    <div class="run-bar__hint">
      {{ t('tool.run.hint') }}
    </div>

    <div class="run-bar__actions">
      <ElButton
        size="default"
        @click="emit('reset')"
      >
        {{ t('common.actions.reset') }}
      </ElButton>
      <ElButton
        v-if="showExample"
        size="default"
        @click="emit('example')"
      >
        {{ t('common.actions.example') }}
      </ElButton>
      <ElButton
        type="primary"
        size="default"
        :loading="running"
        :disabled="disabled"
        @click="emit('run')"
      >
        {{ running ? t('tool.run.running') : t('common.actions.run') }}
      </ElButton>
    </div>
  </div>
</template>

<style scoped>
.run-bar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--tb-space-3);
  padding: var(--tb-space-2) var(--tb-space-4);
  background: var(--tb-bg-elevated);
  border-top: 1px solid var(--tb-border);
}

.run-bar__hint {
  font-size: var(--tb-font-xs);
  color: var(--tb-text-tertiary);
}

.run-bar__actions {
  display: flex;
  gap: var(--tb-space-2);
}

@media (max-width: 720px) {
  .run-bar {
    position: sticky;
    bottom: 0;
    z-index: 10;
    flex-direction: column;
    align-items: stretch;
    gap: var(--tb-space-2);
  }

  .run-bar__hint {
    display: none;
  }

  .run-bar__actions {
    display: grid;
    grid-template-columns: 1fr 1fr 1fr;
  }
}
</style>
