<script setup lang="ts">
/**
 * 结果面板。
 *
 * 四态齐备；成功时滚动到视野内并通过 `aria-live` 播报。
 */
import { computed, nextTick, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'

import CopyButton from '@/components/tool/CopyButton.vue'
import ErrorAlert from '@/components/tool/ErrorAlert.vue'
import { localizeWarning, type LocalizedError } from '@/i18n/errors'
import type { Meta } from '@/types/api'
import type { RunState } from '@/composables/useToolRun'

const props = withDefaults(
  defineProps<{
    /** 四态之一。 */
    state: RunState
    /** 可复制的文本结果。 */
    text?: string
    /** 结果标题；缺省使用「结果」。 */
    title?: string
    /** 执行元信息。 */
    meta?: Meta | null
    /** 错误。 */
    error?: LocalizedError | null
    /** 是否显示警告区。 */
    showWarnings?: boolean
  }>(),
  { text: '', title: undefined, meta: null, error: null, showWarnings: true },
)

const emit = defineEmits<{
  /** 用户点击重试。 */
  retry: []
}>()

const { t } = useI18n()
const root = ref<HTMLElement | null>(null)

const heading = computed(() => props.title ?? t('tool.sections.result'))

/** 状态摘要，供 `aria-live` 播报。 */
const announcement = computed(() => {
  switch (props.state) {
    case 'loading':
      return t('tool.result.loading')
    case 'success':
      return t('tool.result.success')
    case 'error':
      return t('tool.result.error')
    default:
      return ''
  }
})

const warnings = computed(() =>
  (props.meta?.warnings ?? []).map((w) => localizeWarning(w.code, w.params ?? {})),
)

const durationText = computed(() => {
  const ms = props.meta?.durationMs
  return typeof ms === 'number' ? t('tool.result.duration', { ms }) : ''
})

watch(
  () => props.state,
  async (next) => {
    if (next === 'success' || next === 'error') {
      await nextTick()
      root.value?.scrollIntoView({ behavior: 'smooth', block: 'nearest' })
    }
  },
)
</script>

<template>
  <section
    ref="root"
    class="result"
    :aria-label="heading"
  >
    <header class="result__head">
      <h2 class="result__title">
        {{ heading }}
      </h2>
      <div class="result__meta">
        <span
          v-if="durationText"
          class="result__duration"
        >{{ durationText }}</span>
        <CopyButton
          v-if="state === 'success' && text !== ''"
          :text="text"
        />
      </div>
    </header>

    <p
      class="tb-sr-only"
      aria-live="polite"
    >
      {{ announcement }}
    </p>

    <p
      v-if="state === 'idle'"
      class="result__placeholder"
    >
      {{ t('tool.result.idle') }}
      <span class="result__hint">{{ t('tool.result.idleHint') }}</span>
    </p>

    <p
      v-else-if="state === 'loading'"
      class="result__placeholder"
    >
      {{ t('tool.result.loading') }}
    </p>

    <ErrorAlert
      v-else-if="state === 'error'"
      :error="error ?? null"
      @retry="emit('retry')"
    />

    <template v-else>
      <ul
        v-if="showWarnings && warnings.length > 0"
        class="result__warnings"
      >
        <li
          v-for="(w, i) in warnings"
          :key="i"
          class="result__warning"
        >
          {{ w }}
        </li>
      </ul>

      <pre
        v-if="text !== ''"
        class="result__output"
      ><code>{{ text }}</code></pre>
      <p
        v-else
        class="result__placeholder"
      >
        {{ t('tool.result.emptyOutput') }}
      </p>
    </template>
  </section>
</template>

<style scoped>
.result {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-2);
  padding: var(--tb-space-4);
  background: var(--tb-bg-elevated);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-panel);
}

.result__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--tb-space-2);
}

.result__title {
  margin: 0;
  font-size: var(--tb-font-base);
  font-weight: 600;
}

.result__meta {
  display: flex;
  align-items: center;
  gap: var(--tb-space-2);
}

.result__duration {
  font-size: var(--tb-font-xs);
  color: var(--tb-text-tertiary);
}

.result__placeholder {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-1);
  margin: 0;
  padding: var(--tb-space-6) 0;
  font-size: var(--tb-font-sm);
  color: var(--tb-text-tertiary);
  text-align: center;
}

.result__hint {
  font-size: var(--tb-font-xs);
}

.result__warnings {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-1);
  margin: 0;
  padding: var(--tb-space-2) var(--tb-space-3);
  list-style: none;
  background: color-mix(in srgb, var(--tb-warning) 10%, var(--tb-bg-elevated));
  border: 1px solid color-mix(in srgb, var(--tb-warning) 35%, var(--tb-border));
  border-radius: var(--tb-radius-control);
}

.result__warning {
  font-size: var(--tb-font-xs);
  color: var(--tb-text-secondary);
}

.result__output {
  margin: 0;
  padding: var(--tb-space-3);
  overflow-x: auto;
  font-family: var(--tb-font-mono);
  font-size: var(--tb-font-sm);
  line-height: 1.6;
  color: var(--tb-text);
  white-space: pre-wrap;
  word-break: break-all;
  background: var(--tb-bg-sunken);
  border-radius: var(--tb-radius-control);
}
</style>
