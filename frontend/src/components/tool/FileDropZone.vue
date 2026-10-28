<script setup lang="ts">
/**
 * 文件拖放区。
 *
 * 支持点击选择与拖放；**单文件**。
 *
 * ⚠️ 文件只存在于组件局部状态：离开页面即丢弃，
 * **不写**任何浏览器存储（AGENTS.md §4.11）。
 */
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

const file = defineModel<File | null>({ required: true })

const props = withDefaults(
  defineProps<{
    /** 无障碍标签。 */
    label?: string
    /** 接受的文件类型（accept 属性）。 */
    accept?: string
    /** 是否禁用。 */
    disabled?: boolean
  }>(),
  { label: undefined, accept: undefined, disabled: false },
)

const { t } = useI18n()

const input = ref<HTMLInputElement | null>(null)
const dragging = ref(false)

const fileName = computed(() => file.value?.name ?? '')

/** 可读的文件大小。 */
const sizeText = computed(() => {
  if (!file.value) {
    return ''
  }
  const bytes = file.value.size
  if (bytes < 1024) {
    return `${bytes} B`
  }
  if (bytes < 1024 * 1024) {
    return `${(bytes / 1024).toFixed(1)} KB`
  }
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
})

/** 打开系统文件选择器。 */
function open(): void {
  if (!props.disabled) {
    input.value?.click()
  }
}

/** 处理选中的文件。 */
function onSelect(event: Event): void {
  const target = event.target as HTMLInputElement
  const picked = target.files?.[0] ?? null
  file.value = picked
  // 重置 value：连续选同一个文件也要能触发 change。
  target.value = ''
}

/** 处理拖放。 */
function onDrop(event: DragEvent): void {
  dragging.value = false
  if (props.disabled) {
    return
  }
  file.value = event.dataTransfer?.files?.[0] ?? null
}

/** 清除当前文件。 */
function clear(): void {
  file.value = null
}
</script>

<template>
  <div class="drop-zone">
    <div
      class="drop-zone__target"
      :class="{ 'drop-zone__target--active': dragging, 'drop-zone__target--disabled': props.disabled }"
      role="button"
      tabindex="0"
      :aria-label="props.label ?? t('tool.file.dropzone')"
      @click="open"
      @keydown.enter.prevent="open"
      @keydown.space.prevent="open"
      @dragover.prevent="dragging = true"
      @dragleave.prevent="dragging = false"
      @drop.prevent="onDrop"
    >
      <p class="drop-zone__hint">
        {{ t('tool.file.dropzone') }}
      </p>
      <p
        v-if="fileName"
        class="drop-zone__file"
      >
        <span class="drop-zone__name">{{ fileName }}</span>
        <span class="drop-zone__size">{{ sizeText }}</span>
      </p>
    </div>

    <input
      ref="input"
      class="drop-zone__input"
      type="file"
      :accept="props.accept"
      :disabled="props.disabled"
      :aria-label="props.label ?? t('tool.file.choose')"
      @change="onSelect"
    >

    <button
      v-if="fileName"
      class="drop-zone__clear"
      type="button"
      :disabled="props.disabled"
      @click="clear"
    >
      {{ t('tool.file.clear') }}
    </button>
  </div>
</template>

<style scoped>
.drop-zone {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-2);
}

.drop-zone__target {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-1);
  padding: var(--tb-space-4);
  text-align: center;
  cursor: pointer;
  background: var(--tb-bg-elevated);
  border: 1px dashed var(--tb-border);
  border-radius: var(--tb-radius-panel);
  transition: border-color 0.15s ease, background-color 0.15s ease;
}

.drop-zone__target:hover,
.drop-zone__target:focus-visible {
  border-color: var(--tb-primary);
}

.drop-zone__target--active {
  border-color: var(--tb-primary);
  border-style: solid;
  background: var(--tb-bg-sunken);
}

.drop-zone__target--disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

.drop-zone__hint {
  margin: 0;
  font-size: var(--tb-font-sm);
  color: var(--tb-text-secondary);
}

.drop-zone__file {
  display: flex;
  gap: var(--tb-space-2);
  align-items: baseline;
  justify-content: center;
  margin: 0;
  font-size: var(--tb-font-sm);
}

.drop-zone__name {
  font-family: var(--tb-font-mono);
  color: var(--tb-text);
  word-break: break-all;
}

.drop-zone__size {
  color: var(--tb-text-secondary);
  white-space: nowrap;
}

.drop-zone__input {
  position: absolute;
  width: 1px;
  height: 1px;
  padding: 0;
  margin: -1px;
  overflow: hidden;
  clip: rect(0, 0, 0, 0);
  white-space: nowrap;
  border: 0;
}

.drop-zone__clear {
  align-self: flex-start;
  padding: var(--tb-space-1) var(--tb-space-2);
  font-family: inherit;
  font-size: var(--tb-font-sm);
  color: var(--tb-text-secondary);
  cursor: pointer;
  background: transparent;
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-control);
}

.drop-zone__clear:hover:not(:disabled) {
  color: var(--tb-text);
}

.drop-zone__clear:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}
</style>
