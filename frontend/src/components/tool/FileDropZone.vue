<script setup lang="ts">
/**
 * 文件拖放区。
 *
 * 支持点击选择与拖放；**单文件**。
 *
 * ⚠️ 文件只存在于组件局部状态：离开页面即丢弃，
 * **不写**任何浏览器存储（AGENTS.md §4.11）。
 */
import { ElButton } from 'element-plus/es/components/button/index'
import 'element-plus/es/components/button/style/css'
import { ElUpload, type UploadFile } from 'element-plus/es/components/upload/index'
import 'element-plus/es/components/upload/style/css'
import { computed } from 'vue'
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

/** `ElUpload` 选中文件后回写。 */
function onChange(upload: UploadFile): void {
  file.value = upload.raw ?? null
}

/** 清除当前文件。 */
function clear(): void {
  file.value = null
}
</script>

<template>
  <div class="drop-zone">
    <ElUpload
      drag
      action="#"
      :auto-upload="false"
      :show-file-list="false"
      :accept="props.accept"
      :disabled="props.disabled"
      :on-change="onChange"
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
    </ElUpload>

    <ElButton
      v-if="fileName"
      :disabled="props.disabled"
      @click="clear"
    >
      {{ t('tool.file.clear') }}
    </ElButton>
  </div>
</template>

<style scoped>
.drop-zone {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-2);
}

.drop-zone__hint {
  margin: 0;
  font-size: var(--tb-font-sm);
  color: var(--tb-text-secondary);
}

.drop-zone__file {
  display: flex;
  align-items: baseline;
  justify-content: center;
  gap: var(--tb-space-2);
  margin: var(--tb-space-1) 0 0;
  font-size: var(--tb-font-xs);
}

.drop-zone__name {
  font-family: var(--tb-font-mono);
  color: var(--tb-text);
  overflow-wrap: anywhere;
}

.drop-zone__size {
  color: var(--tb-text-tertiary);
}
</style>
