<script setup lang="ts">
/**
 * 文件拖放区。
 *
 * 读入文件并转成 base64；文件**不**落任何浏览器存储，
 * 离开页面即丢弃（AGENTS.md §4.11）。
 */
import { useI18n } from 'vue-i18n'
import { ref } from 'vue'

/** 读入结果。 */
export interface LoadedFile {
  /** 文件名。 */
  name: string
  /** MIME 类型。 */
  mime: string
  /** 字节数。 */
  size: number
  /** base64 内容（不含 data URI 前缀）。 */
  base64: string
}

const emit = defineEmits<{
  /** 文件读入完成。 */
  loaded: [file: LoadedFile]
  /** 读取失败。 */
  error: [message: string]
}>()

const props = withDefaults(
  defineProps<{
    /** 接受的文件类型（accept 属性）。 */
    accept?: string
    /** 单文件字节上限；默认 4 MB。 */
    maxBytes?: number
    /** 是否禁用。 */
    disabled?: boolean
  }>(),
  { accept: '', maxBytes: 4 * 1024 * 1024, disabled: false },
)

const { t } = useI18n()

const dragging = ref(false)
const input = ref<HTMLInputElement | null>(null)

/** 打开文件选择器。 */
function pick(): void {
  if (!props.disabled) {
    input.value?.click()
  }
}

/** 处理拖放事件。 */
function onDrop(event: DragEvent): void {
  dragging.value = false
  if (props.disabled) {
    return
  }
  const file = event.dataTransfer?.files?.[0]
  if (file) {
    void readFile(file)
  }
}

/** 处理文件选择。 */
function onChange(event: Event): void {
  const target = event.target as HTMLInputElement
  const file = target.files?.[0]
  if (file) {
    void readFile(file)
  }
  // 重置以便再次选择同一文件。
  target.value = ''
}

/** 读入文件并转 base64。 */
async function readFile(file: File): Promise<void> {
  if (file.size > props.maxBytes) {
    emit('error', t('tool.dropzone.tooLarge', { max: props.maxBytes }))
    return
  }

  try {
    const buffer = await file.arrayBuffer()
    const bytes = new Uint8Array(buffer)

    let binary = ''
    const chunk = 0x8000
    for (let i = 0; i < bytes.length; i += chunk) {
      binary += String.fromCharCode(...bytes.subarray(i, i + chunk))
    }

    emit('loaded', {
      name: file.name,
      mime: file.type || 'application/octet-stream',
      size: file.size,
      base64: btoa(binary),
    })
  } catch {
    emit('error', t('tool.dropzone.readFailed'))
  }
}
</script>

<template>
  <div
    class="dropzone"
    :class="{ 'dropzone--active': dragging, 'dropzone--disabled': disabled }"
    role="button"
    tabindex="0"
    :aria-disabled="disabled"
    @click="pick"
    @keydown.enter="pick"
    @keydown.space.prevent="pick"
    @dragover.prevent="dragging = true"
    @dragleave.prevent="dragging = false"
    @drop.prevent="onDrop"
  >
    <input
      ref="input"
      type="file"
      class="dropzone__input"
      :accept="accept"
      :disabled="disabled"
      @change="onChange"
    >
    <span class="dropzone__text">{{ t('tool.dropzone.hint') }}</span>
  </div>
</template>

<style scoped>
.dropzone {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 72px;
  padding: var(--tb-space-3);
  font-size: var(--tb-font-sm);
  color: var(--tb-text-secondary);
  text-align: center;
  background: var(--tb-bg-sunken);
  border: 1px dashed var(--tb-border);
  border-radius: var(--tb-radius-control);
  cursor: pointer;
  transition: border-color var(--tb-transition), background var(--tb-transition);
}

.dropzone:hover,
.dropzone--active {
  border-color: var(--tb-primary);
  background: var(--tb-primary-soft);
}

.dropzone--disabled {
  cursor: not-allowed;
  opacity: 0.6;
}

.dropzone__input {
  display: none;
}

.dropzone__text {
  pointer-events: none;
}
</style>
