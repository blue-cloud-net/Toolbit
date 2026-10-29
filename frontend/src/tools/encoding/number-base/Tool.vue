<script setup lang="ts">
/**
 * 进制转换工具页。
 */
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import QuickReference from '@/components/tool/QuickReference.vue'
import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useAutoRun } from '@/composables/useAutoRun'
import { useToolRun } from '@/composables/useToolRun'
import { DEFAULT_EXAMPLE } from '@/tools/encoding/number-base/example'

/** 后端 number-base 工具的响应。 */
interface NumberBaseResult {
  result: string
  from: number
  to: number
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const value = ref('')
const from = ref(10)
const to = ref(16)
const prefix = ref(false)

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
  bindInput,
  useAsInput,
} = useToolRun<NumberBaseResult>({
  slug: 'number-base',
})

/** 结果区的「作为新输入」：把结果回填到主输入框（仅内存，不写存储）。 */
bindInput((text) => {
  value.value = text
})

const resultText = computed(() => data.value?.result ?? '')
const running = computed(() => state.value === 'loading')

/** 常用进制（2–36 中挑出高频项）。 */
const BASE_OPTIONS = [2, 8, 10, 16, 36] as const

const faq = computed<FaqItem[]>(() => tm('tools.number-base.faq') as FaqItem[])

/** 运行转换。 */
function run(): void {
  void execute({
    value: value.value,
    from: from.value,
    to: to.value,
    prefix: prefix.value,
  })
}

/** 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。 */
/** 快速参考条目：点击后填入输入框。 */
const quickRefItems = [
    { label: '十进制前缀', value: '0x' },
    { label: '二进制前缀', value: '0b' },
    { label: '八进制前缀', value: '0o' },
    { label: '十六进制数字', value: '0123456789ABCDEF' },
    { label: '二进制数字', value: '01' },
  ]

/** 把速查条目填入输入框。 */
function applyQuickRef(text: string): void {
  value.value = text
}

useAutoRun({
  slug: 'number-base',
  run,
  watch: [value, from, to, prefix],
  enabled: () => value.value !== '',
})

/** 重置。 */
function reset(): void {
  value.value = ''
  from.value = 10
  to.value = 16
  prefix.value = false
  resetRun()
}

/** 填入示例。 */
function fillExample(): void {
  value.value = DEFAULT_EXAMPLE.value
  from.value = DEFAULT_EXAMPLE.from
  to.value = DEFAULT_EXAMPLE.to
  prefix.value = DEFAULT_EXAMPLE.prefix
  resetRun()
}

/** Ctrl / ⌘ + Enter 触发运行。 */
function onKeydown(event: KeyboardEvent): void {
  if ((event.ctrlKey || event.metaKey) && event.key === 'Enter') {
    event.preventDefault()
    run()
  }
}
</script>

<template>
  <ToolPageLayout
    slug="number-base"
    :name="t('tools.number-base.name')"
    name-en="Number Base Converter"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.number-base.fields.value') }}</span>
            <input
              v-model="value"
              type="text"
              class="field__input"
              spellcheck="false"
              :aria-label="t('tools.number-base.fields.value')"
              @keydown="onKeydown"
            >
          </label>
          <p class="field__hint">
            {{ t('tools.number-base.hints.value') }}
          </p>
        </template>

        <template #options>
          <div class="field-row">
            <label class="field">
              <span class="field__label">{{ t('tools.number-base.fields.from') }}</span>
              <select
                v-model.number="from"
                class="field__select"
              >
                <option
                  v-for="base in BASE_OPTIONS"
                  :key="`f-${base}`"
                  :value="base"
                >
                  {{ t(`tools.number-base.options.base.${base}`) }}
                </option>
              </select>
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.number-base.fields.to') }}</span>
              <select
                v-model.number="to"
                class="field__select"
              >
                <option
                  v-for="base in BASE_OPTIONS"
                  :key="`t-${base}`"
                  :value="base"
                >
                  {{ t(`tools.number-base.options.base.${base}`) }}
                </option>
              </select>
            </label>
          </div>

          <label class="field field--inline">
            <input
              v-model="prefix"
              type="checkbox"
              class="field__checkbox"
            >
            <span>{{ t('tools.number-base.fields.prefix') }}</span>
          </label>
          <p class="field__hint">
            {{ t('tools.number-base.hints.prefix') }}
          </p>
        </template>

        <template #footer>
          <RunBar
            :running="running"
            @run="run"
            @reset="reset"
            @example="fillExample"
          />
        </template>
      </ToolPanel>
    </template>

    <template #result>
      <ResultPanel
        :state="state"
        :text="resultText"
        :meta="meta"
        :error="error"
        :title="t('tools.number-base.results.output')"
        :use-as-input="useAsInput"
        @retry="run"
      />
    </template>

    <template #reference>
      <QuickReference
        :title="t('tool.sections.reference')"
        :items="quickRefItems"
        :insert="applyQuickRef"
      />
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.number-base.results.outputHint') }}
      </p>
    </template>

    <template #faq>
      <dl class="faq">
        <template
          v-for="(item, index) in faq"
          :key="index"
        >
          <dt class="faq__q">
            {{ item.q }}
          </dt>
          <dd class="faq__a">
            {{ item.a }}
          </dd>
        </template>
      </dl>
    </template>
  </ToolPageLayout>
</template>

<style scoped>
.field {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-1);
}

.field--inline {
  flex-direction: row;
  align-items: center;
  gap: var(--tb-space-2);
}

.field__label {
  font-size: var(--tb-font-xs);
  color: var(--tb-text-secondary);
}

.field__input {
  width: 100%;
  padding: var(--tb-space-2) var(--tb-space-3);
  font-family: var(--tb-font-mono);
  font-size: var(--tb-font-sm);
  color: var(--tb-text);
  background: var(--tb-bg-sunken);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-control);
}

.field__select {
  padding: var(--tb-space-1) var(--tb-space-2);
  font-family: inherit;
  font-size: var(--tb-font-sm);
  color: var(--tb-text);
  background: var(--tb-bg-elevated);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-control);
  cursor: pointer;
}

.field__checkbox {
  width: 16px;
  height: 16px;
  accent-color: var(--tb-primary);
}

.field__hint {
  margin: 0;
  font-size: var(--tb-font-xs);
  color: var(--tb-text-tertiary);
}

.field-row {
  display: flex;
  flex-wrap: wrap;
  gap: var(--tb-space-3);
}

.prose {
  margin: 0;
  font-size: var(--tb-font-sm);
  color: var(--tb-text-secondary);
}

.faq {
  margin: 0;
  font-size: var(--tb-font-sm);
}

.faq__q {
  margin-top: var(--tb-space-2);
  font-weight: 600;
  color: var(--tb-text);
}

.faq__a {
  margin: var(--tb-space-1) 0 0;
  color: var(--tb-text-secondary);
}
</style>
