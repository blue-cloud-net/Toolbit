<script setup lang="ts">
/**
 * 文件文本输入。
 *
 * 支持拖放 / 选择文件，把文本内容读入输入框。
 * 内容只存在于组件局部状态，**不写**任何浏览器存储。
 */
import { ElUpload, type UploadFile } from 'element-plus/es/components/upload/index'
import 'element-plus/es/components/upload/style/css'
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

/** `ElUpload` 选中文件后读取文本。 */
function onChange(upload: UploadFile): void {
  if (upload.raw) {
    void readFile(upload.raw)
  }
}
</script>

<template>
  <div class="file-text">
    <ElUpload
      class="file-text__picker"
      drag
      action="#"
      :auto-upload="false"
      :show-file-list="false"
      :accept="accept"
      :on-change="onChange"
    >
      <span class="file-text__hint">{{ t('tool.file.dropHint') }}</span>
    </ElUpload>

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

.file-text__hint {
  font-size: var(--tb-font-xs);
  color: var(--tb-text-tertiary);
}

.file-text__error {
  margin: 0;
  font-size: var(--tb-font-2xs);
  color: var(--tb-danger);
}
</style>
