<script setup lang="ts">
/**
 * 下载按钮。
 *
 * 把文本内容（或 Base64 编码的二进制产物）作为文件下载。
 * 下载在浏览器端完成，内容不经过服务端、不写任何浏览器存储。
 *
 * 二进制产物（PNG 等）在接口里以 Base64 文本传递，用 `base64` 传入即可。
 */
import { ElButton } from 'element-plus/es/components/button/index'
import 'element-plus/es/components/button/style/css'
import { ElMessage } from 'element-plus/es/components/message/index'
import 'element-plus/es/components/message/style/css'
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

import { base64ToBytes } from '@/components/tool/base64'

const props = withDefaults(
  defineProps<{
    /** 文件名。 */
    filename: string
    /** 文件内容（文本）。 */
    content?: string
    /** 文件内容（Base64 编码的二进制）；与 `content` 二选一。 */
    base64?: string
    /** MIME 类型。 */
    mime?: string
    /** 无障碍标签；缺省使用「下载」。 */
    label?: string
  }>(),
  { content: '', base64: '', mime: 'text/plain;charset=utf-8', label: undefined },
)

const { t } = useI18n()

/** 是否有可下载的内容。 */
const hasContent = computed(() => props.content !== '' || props.base64 !== '')

/** 触发下载。 */
function download(): void {
  if (typeof document === 'undefined') {
    return
  }

  let blob: Blob
  try {
    blob =
      props.base64 !== ''
        ? new Blob([base64ToBytes(props.base64)], { type: props.mime })
        : new Blob([props.content], { type: props.mime })
  } catch {
    ElMessage({
      message: t('tool.download.failed'),
      type: 'error',
      duration: 1800,
      grouping: true,
      showClose: false,
    })
    return
  }

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
    :disabled="!hasContent"
    @click="download"
  >
    {{ label ?? t('common.actions.download') }}
  </ElButton>
</template>
