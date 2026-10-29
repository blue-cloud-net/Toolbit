<script setup lang="ts">
/**
 * 运行栏：运行 / 重置 / 填入示例。
 *
 * 移动端吸底（`position: sticky`），桌面端跟随面板。
 *
 * 本组件**统一提供** `Ctrl / ⌘ + Enter` 运行快捷键（左侧提示文案即此键）：
 * 在 `window` 上以捕获阶段监听并 `stopPropagation()`，因此工具页与 `CodeEditor`
 * 里遗留的同名实现不会再执行，一次按键只触发一次运行。
 */
import { ElButton } from 'element-plus/es/components/button/index'
import 'element-plus/es/components/button/style/css'
import { onBeforeUnmount, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'

const props = withDefaults(
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

/** 处理按键：`Ctrl / ⌘ + Enter` 触发运行（焦点在工具页任意控件内均有效）。 */
function onKeydown(event: KeyboardEvent): void {
  if (event.key !== 'Enter' || !(event.ctrlKey || event.metaKey)) {
    return
  }
  // 输入法组合中、正在运行或不可运行时都不触发。
  if (event.isComposing || props.running || props.disabled) {
    return
  }
  // 弹窗（如证书详情）内的按键不触发运行。
  const target = event.target
  if (target instanceof HTMLElement && target.closest('[role="dialog"]') !== null) {
    return
  }
  // 抢占该组合键，避免页面内遗留的实现重复运行。
  event.preventDefault()
  event.stopPropagation()
  emit('run')
}

onMounted(() => window.addEventListener('keydown', onKeydown, true))
onBeforeUnmount(() => window.removeEventListener('keydown', onKeydown, true))
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
