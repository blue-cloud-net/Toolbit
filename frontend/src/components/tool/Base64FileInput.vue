<script setup lang="ts">
/**
 * 文件拖放区。
 *
 * 读入文件并转成 base64；文件**不**落任何浏览器存储，
 * 离开页面即丢弃（AGENTS.md §4.11）。
 */
import { ElUpload, type UploadFile } from 'element-plus/es/components/upload/index'
import 'element-plus/es/components/upload/style/css'
import { useI18n } from 'vue-i18n'

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

/** `ElUpload` 选中文件后读取。 */
function onChange(upload: UploadFile): void {
  if (props.disabled) {
    return
  }
  if (upload.raw) {
    void readFile(upload.raw)
  }
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
  <ElUpload
    class="dropzone"
    drag
    action="#"
    :auto-upload="false"
    :show-file-list="false"
    :accept="accept"
    :disabled="disabled"
    :on-change="onChange"
  >
    <span class="dropzone__text">{{ t('tool.dropzone.hint') }}</span>
  </ElUpload>
</template>

<style scoped>
.dropzone__text {
  font-size: var(--tb-font-sm);
  color: var(--tb-text-secondary);
}
</style>
