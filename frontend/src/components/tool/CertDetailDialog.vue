<script setup lang="ts">
/**
 * 证书详情对话框。
 *
 * 以浮层展示单张证书的完整字段。
 */
import { onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'

import AppIcon from '@/components/common/AppIcon.vue'
import CopyButton from '@/components/tool/CopyButton.vue'
import KeyValueTable, { type KeyValueRow } from '@/components/tool/KeyValueTable.vue'

export interface CertDetail {
  /** 证书字段（键值对）。 */
  fields: KeyValueRow[]
  /** 证书 PEM 正文；给出时提供复制入口。 */
  pem?: string
}

const props = defineProps<{
  /** 详情；为 null 时不渲染。 */
  detail: CertDetail | null
  /** 关闭回调。 */
  close: () => void
}>()

const { t } = useI18n()
const panel = ref<HTMLElement | null>(null)

/** Esc 关闭。 */
function onKeydown(event: KeyboardEvent): void {
  if (event.key === 'Escape') {
    event.preventDefault()
    props.close()
  }
}

watch(
  () => props.detail,
  (next) => {
    if (typeof document === 'undefined') {
      return
    }
    if (next) {
      document.addEventListener('keydown', onKeydown)
    } else {
      document.removeEventListener('keydown', onKeydown)
    }
  },
)

onMounted(() => {
  if (typeof document !== 'undefined') {
    document.addEventListener('keydown', onKeydown)
  }
})

onBeforeUnmount(() => {
  if (typeof document !== 'undefined') {
    document.removeEventListener('keydown', onKeydown)
  }
})
</script>

<template>
  <div
    v-if="detail"
    class="cert-dialog"
    @click="close()"
  >
    <div
      ref="panel"
      class="cert-dialog__panel"
      role="dialog"
      aria-modal="true"
      :aria-label="t('tool.cert.detail')"
      @click.stop
    >
      <header class="cert-dialog__head">
        <strong class="cert-dialog__title">{{ t('tool.cert.detail') }}</strong>
        <button
          type="button"
          class="cert-dialog__close"
          :aria-label="t('common.actions.close')"
          @click="close()"
        >
          <AppIcon
            name="close"
            :size="16"
          />
        </button>
      </header>

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
    </div>
  </div>
</template>

<style scoped>
.cert-dialog {
  position: fixed;
  inset: 0;
  z-index: 2050;
  display: grid;
  place-items: center;
  padding: var(--tb-space-4);
  background: rgb(0 0 0 / 45%);
}

.cert-dialog__panel {
  width: 100%;
  max-width: 640px;
  max-height: 80vh;
  overflow-y: auto;
  padding: var(--tb-space-4);
  background: var(--tb-bg-elevated);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-panel);
  box-shadow: var(--tb-shadow-lg);
}

.cert-dialog__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--tb-space-2);
  margin-bottom: var(--tb-space-3);
}

.cert-dialog__title {
  font-size: var(--tb-font-base);
}

.cert-dialog__close {
  display: grid;
  place-items: center;
  width: 28px;
  height: 28px;
  color: var(--tb-text-secondary);
  background: transparent;
  border: 1px solid transparent;
  border-radius: var(--tb-radius-control);
  cursor: pointer;
}

.cert-dialog__close:hover {
  color: var(--tb-text);
  background: var(--tb-bg-hover);
}

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
