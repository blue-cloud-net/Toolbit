<script setup lang="ts">
/**
 * 下载按钮。
 *
 * 把文本内容作为文件下载。下载在浏览器端完成，
 * 内容不经过服务端、不写任何浏览器存储。
 */
import { ElButton } from 'element-plus/es/components/button/index'
import 'element-plus/es/components/button/style/css'
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
  <ElButton
    class="download-button"
    :aria-label="label ?? t('common.actions.download')"
    :disabled="content === ''"
    @click="download"
  >
    {{ label ?? t('common.actions.download') }}
  </ElButton>
</template>
