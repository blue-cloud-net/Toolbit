<script setup lang="ts">
/**
 * 文件文本输入。
 *
 * 支持拖放 / 选择文件，把文本内容读入输入框。
 * 内容只存在于组件局部状态，**不写**任何浏览器存储。
 */
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'

const model = defineModel<string>({ required: true })

const props = withDefaults(
  defineProps<{
    /** 无障碍标签。 */
    label?: string
    /** 接受的文件类型（input 的 accept 属性）。 */
    accept?: string
    /** 单文件大小上限（字节）。 */
    maxBytes?: number
  }>(),
  { label: undefined, accept: undefined, maxBytes: 1024 * 1024 },
)

const { t } = useI18n()

const error = ref('')
const dragging = ref(false)

/** 读取文件内容。 */
async function readFile(file: File): Promise<void> {
  error.value = ''
  if (file.size > props.maxBytes) {
    error.value = t('tool.file.tooLarge', { size: Math.round(props.maxBytes / 1024) })
    return
  }
  try {
    model.value = await file.text()
  } catch {
    error.value = t('tool.file.readFailed')
  }
}

/** 选择文件。 */
function onChange(event: Event): void {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (file) {
    void readFile(file)
  }
}

/** 拖放文件。 */
function onDrop(event: DragEvent): void {
  dragging.value = false
  const file = event.dataTransfer?.files?.[0]
  if (file) {
    void readFile(file)
  }
}
</script>

<template>
  <div
    class="file-text"
    :class="{ 'file-text--dragging': dragging }"
    @dragover.prevent="dragging = true"
    @dragleave.prevent="dragging = false"
    @drop.prevent="onDrop"
  >
    <label class="file-text__picker">
      <span class="tb-sr-only">{{ label ?? t('common.actions.search') }}</span>
      <input
        type="file"
        class="file-text__input"
        :accept="accept"
        @change="onChange"
      >
      <span class="file-text__hint">{{ t('tool.file.dropHint') }}</span>
    </label>

    <p
      v-if="error"
      class="file-text__error"
      role="alert"
    >
      {{ error }}
    </p>
  </div>
</template>

<style scoped>
.file-text {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-1);
}

.file-text--dragging {
  outline: 2px dashed var(--tb-primary);
  outline-offset: 2px;
  border-radius: var(--tb-radius-control);
}

.file-text__picker {
  display: block;
  padding: var(--tb-space-3);
  font-size: var(--tb-font-xs);
  color: var(--tb-text-tertiary);
  text-align: center;
  background: var(--tb-bg-sunken);
  border: 1px dashed var(--tb-border);
  border-radius: var(--tb-radius-control);
  cursor: pointer;
  transition: border-color var(--tb-transition);
}

.file-text__picker:hover {
  border-color: var(--tb-primary);
}

.file-text__input {
  position: absolute;
  width: 1px;
  height: 1px;
  opacity: 0;
  pointer-events: none;
}

.file-text__error {
  margin: 0;
  font-size: var(--tb-font-2xs);
  color: var(--tb-danger);
}
</style>
