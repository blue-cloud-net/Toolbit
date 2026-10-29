<script setup lang="ts">
/**
 * 内联错误提示。
 *
 * 展示错误码 chip + 本地化文案 + 请求标识 + 重试按钮；
 * **禁止**只用弹窗。
 */
import { ElAlert } from 'element-plus/es/components/alert/index'
import 'element-plus/es/components/alert/style/css'
import { ElButton } from 'element-plus/es/components/button/index'
import 'element-plus/es/components/button/style/css'
import { useI18n } from 'vue-i18n'

import type { LocalizedError } from '@/i18n/errors'

withDefaults(
  defineProps<{
    /** 本地化后的错误。 */
    error: LocalizedError | null
    /** 是否显示重试按钮。 */
    retryable?: boolean
  }>(),
  { retryable: true },
)

const emit = defineEmits<{
  /** 用户点击重试。 */
  retry: []
}>()

const { t } = useI18n()
</script>

<template>
  <ElAlert
    v-if="error"
    class="error-alert"
    type="error"
    :closable="false"
    show-icon
  >
    <template #title>
      <span class="error-alert__title">{{ t('tool.error.title') }}</span>
      <code class="error-alert__code">{{ error.code }}</code>
    </template>

    <p class="error-alert__text">
      {{ error.text }}
    </p>

    <div class="error-alert__foot">
      <span
        v-if="error.requestId"
        class="error-alert__request-id"
      >
        {{ t('tool.error.requestId') }}: <code>{{ error.requestId }}</code>
      </span>
      <ElButton
        v-if="retryable"
        size="small"
        @click="emit('retry')"
      >
        {{ t('tool.error.retry') }}
      </ElButton>
    </div>
  </ElAlert>
</template>

<style scoped>
.error-alert__title {
  font-size: var(--tb-font-sm);
  font-weight: 600;
}

.error-alert__code {
  margin-left: var(--tb-space-2);
  padding: 0 var(--tb-space-1);
  font-family: var(--tb-font-mono);
  font-size: 11px;
  color: var(--tb-danger);
  background: color-mix(in srgb, var(--tb-danger) 12%, transparent);
  border-radius: var(--tb-radius-pill);
}

.error-alert__text {
  margin: 0;
  font-size: var(--tb-font-sm);
  color: var(--tb-text);
}

.error-alert__foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--tb-space-2);
  margin-top: var(--tb-space-2);
}

.error-alert__request-id {
  font-size: 11px;
  color: var(--tb-text-tertiary);
}
</style>
