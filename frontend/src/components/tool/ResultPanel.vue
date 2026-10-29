<script setup lang="ts">
/**
 * 结果面板。
 *
 * 四态齐备；成功时滚动到视野内并通过 `aria-live` 播报。
 */
import { ElAlert } from 'element-plus/es/components/alert/index'
import 'element-plus/es/components/alert/style/css'
import { ElButton } from 'element-plus/es/components/button/index'
import 'element-plus/es/components/button/style/css'
import { ElTabPane, ElTabs } from 'element-plus/es/components/tabs/index'
import 'element-plus/es/components/tabs/style/css'
import { computed, nextTick, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'

import CopyButton from '@/components/tool/CopyButton.vue'
import DownloadButton from '@/components/tool/DownloadButton.vue'
import EmptyState from '@/components/tool/EmptyState.vue'
import ErrorAlert from '@/components/tool/ErrorAlert.vue'
import LoadingSkeleton from '@/components/tool/LoadingSkeleton.vue'
import AppIcon from '@/components/common/AppIcon.vue'
import { localizeWarning, type LocalizedError } from '@/i18n/errors'
import type { Meta } from '@/types/api'
import type { RunState } from '@/composables/useToolRun'

/** 结论徽章。 */
export interface Verdict {
  /** 语义级别。 */
  level: 'success' | 'danger' | 'warning' | 'info'
  /** 结论文案。 */
  label: string
}

/** 结构化结果的一个视图。 */
export interface ResultView {
  /** 视图标识。 */
  id: string
  /** 页签标题。 */
  label: string
}

/** 当前选中的视图标识。 */
const activeView = defineModel<string>('activeView', { default: '' })

/** 实际展示的视图：优先取选中项，否则取第一个。 */
const currentView = computed(() => {
  const list = props.views
  if (list.length === 0) {
    return ''
  }
  const found = list.find((item) => item.id === activeView.value)
  return (found ?? list[0]!).id
})

/** 选中视图。 */
function selectView(id: string): void {
  activeView.value = id
}

/** `ElTabs` 页签切换。 */
function onTabChange(name: string | number): void {
  selectView(String(name))
}

/** 结论级别 → 图标名。 */
const VERDICT_ICONS = {
  success: 'check',
  danger: 'close',
  warning: 'alert',
  info: 'document',
} as const

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
    /** 结论徽章；由可验证类工具传入。 */
    verdict?: Verdict | null
    /** 「作为新输入」回调；工具页用 useToolRun 的 useAsInput 传入。 */
    useAsInput?: ((text: string) => boolean) | null
    /** 下载文件名；给出时在标题栏渲染下载按钮。 */
    downloadFilename?: string
    /** 下载内容的 MIME 类型。 */
    downloadMime?: string
    /** 结构化视图；给出时在结果区渲染可切换的页签。 */
    views?: ResultView[]
  }>(),
  {
    text: '',
    title: undefined,
    meta: null,
    error: null,
    showWarnings: true,
    verdict: null,
    useAsInput: null,
    downloadFilename: undefined,
    downloadMime: 'text/plain;charset=utf-8',
    views: () => [],
  },
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

/** 把结果回填到输入区。 */
function onUseAsInput(): void {
  props.useAsInput?.(props.text)
}

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
        <DownloadButton
          v-if="downloadFilename && state === 'success' && text !== ''"
          :filename="downloadFilename"
          :content="text"
          :mime="downloadMime"
        />
        <ElButton
          v-if="useAsInput && state === 'success' && text !== ''"
          size="small"
          :title="t('tool.result.useAsInput')"
          @click="onUseAsInput"
        >
          {{ t('tool.result.useAsInput') }}
        </ElButton>
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

    <div
      v-if="verdict && state === 'success'"
      class="result__verdict"
      :class="`result__verdict--${verdict.level}`"
      role="status"
    >
      <AppIcon
        :name="VERDICT_ICONS[verdict.level]"
        :size="20"
      />
      <span class="result__verdict-label">{{ verdict.label }}</span>
    </div>

    <div
      v-if="$slots.summary && state === 'success'"
      class="result__summary"
    >
      <slot name="summary" />
    </div>

    <EmptyState
      v-if="state === 'idle'"
      :title="t('tool.result.idle')"
      :hint="t('tool.result.idleHint')"
      icon="document"
    />

    <LoadingSkeleton
      v-else-if="state === 'loading'"
      :rows="4"
      variant="block"
    />

    <ErrorAlert
      v-else-if="state === 'error'"
      :error="error ?? null"
      @retry="emit('retry')"
    />

    <template v-else>
      <ElAlert
        v-if="showWarnings && warnings.length > 0"
        type="warning"
        :closable="false"
        show-icon
      >
        <ul class="result__warnings">
          <li
            v-for="(w, i) in warnings"
            :key="i"
          >
            {{ w }}
          </li>
        </ul>
      </ElAlert>

      <ElTabs
        v-if="views.length > 0"
        class="result__tabs"
        :model-value="currentView"
        @tab-change="onTabChange"
      >
        <ElTabPane
          v-for="view in views"
          :key="view.id"
          :label="view.label"
          :name="view.id"
        >
          <slot
            v-if="view.id === currentView"
            name="view"
            :view-id="view.id"
          />
        </ElTabPane>
      </ElTabs>

      <template v-else>
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

      <slot name="extra" />
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
  /* 顶部主色带：明确「以下是输出」，与输入区区分。 */
  border-top: 3px solid var(--tb-primary);
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
  margin: 0;
  padding-left: var(--tb-space-4);
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

/* 结论徽章：大字结论，一眼可见。 */
.result__verdict {
  display: flex;
  align-items: center;
  gap: var(--tb-space-2);
  padding: var(--tb-space-3);
  font-size: var(--tb-font-lg);
  font-weight: 600;
  border: 1px solid;
  border-radius: var(--tb-radius-control);
}

.result__verdict--success {
  color: var(--tb-success);
  background: color-mix(in srgb, var(--tb-success) 12%, var(--tb-bg-elevated));
  border-color: color-mix(in srgb, var(--tb-success) 40%, var(--tb-border));
}

.result__verdict--danger {
  color: var(--tb-danger);
  background: color-mix(in srgb, var(--tb-danger) 12%, var(--tb-bg-elevated));
  border-color: color-mix(in srgb, var(--tb-danger) 40%, var(--tb-border));
}

.result__verdict--warning {
  color: var(--tb-warning);
  background: color-mix(in srgb, var(--tb-warning) 12%, var(--tb-bg-elevated));
  border-color: color-mix(in srgb, var(--tb-warning) 40%, var(--tb-border));
}

.result__verdict--info {
  color: var(--tb-info);
  background: color-mix(in srgb, var(--tb-info) 12%, var(--tb-bg-elevated));
  border-color: color-mix(in srgb, var(--tb-info) 40%, var(--tb-border));
}

.result__verdict-label {
  min-width: 0;
}

/* 汇总统计条：网络类工具的成败计数。 */
.result__summary {
  display: flex;
  flex-wrap: wrap;
  gap: var(--tb-space-2) var(--tb-space-4);
  padding: var(--tb-space-2) var(--tb-space-3);
  font-size: var(--tb-font-xs);
  background: var(--tb-bg-sunken);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-control);
}

/* 多视图页签：间距交给 ElTabs，这里只保留外层间距。 */
.result__tabs {
  margin-top: var(--tb-space-1);
}
</style>
