<script setup lang="ts">
/**
 * 证书详情对话框。
 *
 * 以浮层展示单张证书的完整字段。
 */
import { ElDialog } from 'element-plus/es/components/dialog/index'
import 'element-plus/es/components/dialog/style/css'
import { useI18n } from 'vue-i18n'

import CopyButton from '@/components/tool/CopyButton.vue'
import KeyValueTable, { type KeyValueRow } from '@/components/tool/KeyValueTable.vue'

export interface CertDetail {
  /** 证书字段（键值对）。 */
  fields: KeyValueRow[]
  /** 证书 PEM 正文；给出时提供复制入口。 */
  pem?: string
}

defineProps<{
  /** 详情；为 null 时不渲染。 */
  detail: CertDetail | null
  /** 关闭回调。 */
  close: () => void
}>()

const { t } = useI18n()
</script>

<template>
  <ElDialog
    :model-value="detail !== null"
    :title="t('tool.cert.detail')"
    width="640px"
    append-to-body
    @close="close()"
  >
    <template v-if="detail">
      <KeyValueTable :rows="detail.fields" />

      <div
        v-if="detail.pem"
        class="cert-dialog__pem"
      >
        <div class="cert-dialog__pem-head">
          <span class="cert-dialog__pem-label">PEM</span>
          <CopyButton :text="detail.pem" />
        </div>
        <pre class="cert-dialog__pem-body"><code>{{ detail.pem }}</code></pre>
      </div>
    </template>
  </ElDialog>
</template>

<style scoped>
.cert-dialog__pem {
  margin-top: var(--tb-space-3);
}

.cert-dialog__pem-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--tb-space-2);
  margin-bottom: var(--tb-space-1);
}

.cert-dialog__pem-label {
  font-size: var(--tb-font-xs);
  color: var(--tb-text-tertiary);
}

.cert-dialog__pem-body {
  margin: 0;
  padding: var(--tb-space-3);
  max-height: 260px;
  overflow: auto;
  font-family: var(--tb-font-mono);
  font-size: var(--tb-font-2xs);
  color: var(--tb-text);
  white-space: pre-wrap;
  word-break: break-all;
  background: var(--tb-bg-sunken);
  border-radius: var(--tb-radius-control);
}
</style>
