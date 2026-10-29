<script setup lang="ts">
/**
 * HTTP 状态码查询工具页。
 *
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
 */
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { ElOption, ElSelect } from 'element-plus/es/components/select/index'
import 'element-plus/es/components/select/style/css'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import ValidationHint from '@/components/tool/ValidationHint.vue'
import { useAutoRun } from '@/composables/useAutoRun'
import { useToolRun } from '@/composables/useToolRun'
import { EXAMPLES, type StatusQueryKind } from '@/tools/network/http-status/example'

/** 一个状态码条目。 */
interface HttpStatusEntry {
  /** 状态码。 */
  code: number
  /** 标准原因短语。 */
  reason: string
  /** 类别：informational / success / redirect / clientError / serverError。 */
  category: string
  /** 中文说明。 */
  description: string
  /** 是否已废弃。 */
  deprecated: boolean
}

/** 后端 http-status 工具的响应。 */
interface HttpStatusResult {
  /** 规范化后的查询词。 */
  query: string
  /** 查询方式：code / keyword。 */
  kind: string
  /** 是否命中标准状态码。 */
  matched: boolean
  /** 状态码。 */
  code: number
  /** 标准原因短语。 */
  reason: string
  /** 类别。 */
  category: string
  /** 中文说明。 */
  description: string
  /** 是否已废弃。 */
  deprecated: boolean
  /** 命中的条目。 */
  matches: HttpStatusEntry[]
  /** 命中条数。 */
  total: number
  /** 数据集版本。 */
  dataVersion: string
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

/** 状态码的合法区间。 */
const MIN_CODE = 100
const MAX_CODE = 599

const { t, tm } = useI18n()

const kind = ref<StatusQueryKind>('code')
const input = ref('404')

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<HttpStatusResult>({
  slug: 'http-status',
})

const running = computed(() => state.value === 'loading')

/** 状态码模式下的解析结果；不是纯数字时为 null。 */
const parsedCode = computed<number | null>(() => {
  if (kind.value !== 'code') {
    return null
  }
  const text = input.value.trim()
  if (!/^\d+$/.test(text)) {
    return null
  }
  return Number.parseInt(text, 10)
})

/** 输入校验提示：状态码必须是 100–599 的整数。 */
const validation = computed<{ level: 'success' | 'error'; message: string } | null>(() => {
  if (kind.value !== 'code' || input.value.trim() === '') {
    return null
  }
  const value = parsedCode.value
  if (value === null || value < MIN_CODE || value > MAX_CODE) {
    return {
      level: 'error',
      message: t('tools.http-status.hints.invalidCode', { min: MIN_CODE, max: MAX_CODE }),
    }
  }
  return { level: 'success', message: t('tools.http-status.hints.validCode') }
})

/** 结果文本：状态码与原因短语。 */
const resultText = computed(() => {
  if (!data.value?.matched) {
    return ''
  }
  return `${data.value.code} ${data.value.reason}`
})

/** 结论徽章：命中与否。 */
const verdict = computed(() => {
  if (!data.value) {
    return null
  }
  return data.value.matched
    ? { level: 'success' as const, label: t('tools.http-status.results.matched') }
    : { level: 'info' as const, label: t('tools.http-status.results.notMatched') }
})

/** 命中条目。 */
const matches = computed<HttpStatusEntry[]>(() => data.value?.matches ?? [])

const faq = computed<FaqItem[]>(() => tm('tools.http-status.faq') as FaqItem[])

/** 类别的显示文案。 */
function categoryLabel(value: string): string {
  const key = `tools.http-status.options.category.${value}`
  const translated = t(key)
  return translated === key ? value : translated
}

/** 运行当前输入。 */
function run(): void {
  const text = input.value.trim()
  if (text === '') {
    return
  }
  if (kind.value === 'code') {
    const code = parsedCode.value
    if (code === null) {
      return
    }
    void execute({ code })
    return
  }
  void execute({ query: text })
}

// 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。
useAutoRun({
  slug: 'http-status',
  run,
  watch: [kind, input],
  enabled: () => input.value.trim() !== '',
})

/** 重置表单与结果。 */
function reset(): void {
  kind.value = 'code'
  input.value = '404'
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  kind.value = example.kind
  input.value = example.kind === 'code' ? example.code : example.keyword
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="http-status"
    :name="t('tools.http-status.name')"
    name-en="HTTP Status Code Lookup"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <div class="row">
            <label class="field field--kind">
              <span class="field__label">{{ t('tools.http-status.fields.kind') }}</span>
              <ElSelect
                v-model="kind"
                class="field__input"
                :aria-label="t('tools.http-status.fields.kind')"
              >
                <ElOption
                  :label="t('tools.http-status.options.kind.code')"
                  value="code"
                />
                <ElOption
                  :label="t('tools.http-status.options.kind.keyword')"
                  value="keyword"
                />
              </ElSelect>
            </label>

            <label class="field field--grow">
              <span class="field__label">{{ t('tools.http-status.fields.query') }}</span>
              <ElInput
                v-model="input"
                class="field__input tb-mono"
                autocomplete="off"
                :placeholder="kind === 'code' ? '404' : 'not found'"
                :aria-label="t('tools.http-status.fields.query')"
                @keydown.enter.prevent="run"
              />
            </label>
          </div>

          <ValidationHint :validation="validation" />

          <p class="field__hint">
            {{ t('tools.http-status.hints.query') }}
          </p>
        </template>

        <template #footer>
          <RunBar
            :running="running"
            @run="run"
            @reset="reset"
            @example="fillExample(0)"
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
        :verdict="verdict"
        :title="t('tools.http-status.results.title')"
        @retry="run"
      >
        <template
          v-if="data"
          #extra
        >
          <dl class="detail">
            <div
              v-if="data.category"
              class="detail__row"
            >
              <dt>{{ t('tools.http-status.results.category') }}</dt>
              <dd>
                <span class="badge">{{ categoryLabel(data.category) }}</span>
              </dd>
            </div>
            <div
              v-if="data.matched"
              class="detail__row"
            >
              <dt>{{ t('tools.http-status.results.description') }}</dt>
              <dd>
                {{ data.description }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.http-status.results.total') }}</dt>
              <dd class="tb-mono">
                {{ data.total }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.http-status.results.dataVersion') }}</dt>
              <dd class="tb-mono">
                {{ data.dataVersion }}
              </dd>
            </div>
          </dl>

          <div
            v-if="matches.length"
            class="table-wrap"
          >
            <table class="table">
              <thead>
                <tr>
                  <th scope="col">
                    {{ t('tools.http-status.results.code') }}
                  </th>
                  <th scope="col">
                    {{ t('tools.http-status.results.reason') }}
                  </th>
                  <th scope="col">
                    {{ t('tools.http-status.results.category') }}
                  </th>
                  <th scope="col">
                    {{ t('tools.http-status.results.description') }}
                  </th>
                </tr>
              </thead>
              <tbody>
                <tr
                  v-for="item in matches"
                  :key="item.code"
                >
                  <td class="tb-mono">
                    {{ item.code }}
                  </td>
                  <td class="tb-mono">
                    {{ item.reason }}
                  </td>
                  <td>
                    {{ categoryLabel(item.category) }}
                  </td>
                  <td>
                    {{ item.description }}
                  </td>
                </tr>
              </tbody>
            </table>
          </div>
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.http-status.hints.output') }}
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
.row {
  display: flex;
  flex-wrap: wrap;
  gap: var(--tb-space-3);
}

.field--kind {
  min-width: 160px;
}

.field--grow {
  flex: 1 1 200px;
}

.detail {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-1);
  margin: var(--tb-space-3) 0 0;
}

.detail__row {
  display: flex;
  gap: var(--tb-space-2);
  font-size: var(--tb-font-sm);
}

.detail__row dt {
  min-width: 96px;
  color: var(--tb-text-tertiary);
}

.detail__row dd {
  margin: 0;
}

.badge {
  padding: 0 var(--tb-space-2);
  font-size: var(--tb-font-xs);
  color: var(--tb-text-secondary);
  background: var(--tb-bg-sunken);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-pill);
}

.table-wrap {
  max-height: 420px;
  margin-top: var(--tb-space-3);
  overflow: auto;
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-control);
}

.table {
  width: 100%;
  font-size: var(--tb-font-sm);
  border-collapse: collapse;
}

.table th,
.table td {
  padding: var(--tb-space-1) var(--tb-space-3);
  text-align: left;
  border-bottom: 1px solid var(--tb-border);
}

.table th {
  position: sticky;
  top: 0;
  font-weight: 600;
  color: var(--tb-text-secondary);
  background: var(--tb-bg-sunken);
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
