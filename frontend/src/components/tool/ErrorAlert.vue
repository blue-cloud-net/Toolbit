<script setup lang="ts">
/**
 * 内联错误提示。
 *
 * 展示错误码 chip + 本地化文案 + 请求标识 + 重试按钮；
 * **禁止**只用弹窗。
 */
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
  <div
    v-if="error"
    class="error-alert"
    role="alert"
  >
    <div class="error-alert__head">
      <strong class="error-alert__title">{{ t('tool.error.title') }}</strong>
      <code class="error-alert__code">{{ error.code }}</code>
    </div>

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
      <button
        v-if="retryable"
        type="button"
        class="error-alert__retry"
        @click="emit('retry')"
      >
        {{ t('tool.error.retry') }}
      </button>
    </div>
  </div>
</template>

<style scoped>
.error-alert {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-2);
  padding: var(--tb-space-3) var(--tb-space-4);
  background: color-mix(in srgb, var(--tb-danger) 8%, var(--tb-bg-elevated));
  border: 1px solid color-mix(in srgb, var(--tb-danger) 40%, var(--tb-border));
  border-radius: var(--tb-radius-card);
}

.error-alert__head {
  display: flex;
  align-items: center;
  gap: var(--tb-space-2);
}

.error-alert__title {
  font-size: var(--tb-font-sm);
  color: var(--tb-danger);
}

.error-alert__code {
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
}

.error-alert__request-id {
  font-size: 11px;
  color: var(--tb-text-tertiary);
}

.error-alert__retry {
  padding: var(--tb-space-1) var(--tb-space-3);
  font-family: inherit;
  font-size: var(--tb-font-xs);
  color: var(--tb-text);
  background: var(--tb-bg-elevated);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-control);
  cursor: pointer;
}

.error-alert__retry:hover {
  border-color: var(--tb-danger);
}
</style>
