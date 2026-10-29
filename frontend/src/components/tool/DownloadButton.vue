<script setup lang="ts">
/**
 * 下载按钮。
 *
 * 把文本内容作为文件下载。下载在浏览器端完成，
 * 内容不经过服务端、不写任何浏览器存储。
 */
import { useI18n } from 'vue-i18n'

const props = withDefaults(
  defineProps<{
    /** 文件名。 */
    filename: string
    /** 文件内容。 */
    content: string
    /** MIME 类型。 */
    mime?: string
    /** 无障碍标签；缺省使用「下载」。 */
    label?: string
  }>(),
  { mime: 'text/plain;charset=utf-8', label: undefined },
)

const { t } = useI18n()

/** 触发下载。 */
function download(): void {
  if (typeof document === 'undefined') {
    return
  }
  const blob = new Blob([props.content], { type: props.mime })
  const url = URL.createObjectURL(blob)
  const anchor = document.createElement('a')
  anchor.href = url
  anchor.download = props.filename
  document.body.appendChild(anchor)
  anchor.click()
  document.body.removeChild(anchor)
  URL.revokeObjectURL(url)
}
</script>

<template>
  <button
    type="button"
    class="download-button"
    :aria-label="label ?? t('common.actions.download')"
    :disabled="content === ''"
    @click="download"
  >
    {{ label ?? t('common.actions.download') }}
  </button>
</template>

<style scoped>
.download-button {
  padding: var(--tb-space-1) var(--tb-space-2);
  font-family: inherit;
  font-size: var(--tb-font-xs);
  color: var(--tb-text-secondary);
  background: transparent;
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-control);
  cursor: pointer;
  transition:
    color var(--tb-transition),
    border-color var(--tb-transition);
}

.download-button:hover:not(:disabled) {
  color: var(--tb-primary);
  border-color: var(--tb-primary);
}

.download-button:disabled {
  cursor: not-allowed;
  opacity: 0.5;
}
</style>
