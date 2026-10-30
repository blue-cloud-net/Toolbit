<script setup lang="ts">
/**
 * 域名验证工具页。
 *
 * 三种方式：DNS TXT、DNS CNAME、站点文件。
 */
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { ElOption, ElSelect } from 'element-plus/es/components/select/index'
import 'element-plus/es/components/select/style/css'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import DataTable, { type DataTableColumn, type DataTableRow } from '@/components/tool/DataTable.vue'
import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import StatGrid, { type StatGridItem } from '@/components/tool/StatGrid.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useToolRun } from '@/composables/useToolRun'

import { DEFAULT_EXAMPLE } from './example'

/** 单解析器结论。 */
interface ResolverOutcome {
  resolver: string
  values: string[]
  matched: boolean
  error: { code: string; message: string } | null
}

/** 后端 domain-verify 工具的响应。 */
interface DomainVerifyResult {
  type: string
  domain: string
  expected: string
  matched: boolean
  perResolver: ResolverOutcome[]
  httpStatus: number
  bodyPreview: string
}

interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const type = ref<'dns-txt' | 'dns-cname' | 'http-file'>(DEFAULT_EXAMPLE.type)
const domain = ref(DEFAULT_EXAMPLE.domain)
const expected = ref(DEFAULT_EXAMPLE.expected)
const filePath = ref(DEFAULT_EXAMPLE.filePath)

const { state, data, meta, error, run: execute, reset: resetRun } =
  useToolRun<DomainVerifyResult>({ slug: 'domain-verify' })

const running = computed(() => state.value === 'loading')

const resultText = computed(() => {
  if (!data.value) {
    return ''
  }
  const value = data.value
  const lines = [`matched: ${String(value.matched)}`]
  if (value.type === 'http-file') {
    lines.push(`httpStatus: ${value.httpStatus}`, '', value.bodyPreview)
  }
  for (const outcome of value.perResolver) {
    const state = outcome.error
      ? `! ${outcome.error.code}`
      : outcome.matched
        ? 'match'
        : 'no match'
    lines.push(`# ${outcome.resolver} — ${state}`)
    for (const item of outcome.values) {
      lines.push(`  ${item}`)
    }
  }
  return lines.join('\n')
})

const faq = computed<FaqItem[]>(() => tm('tools.domain-verify.faq') as FaqItem[])

/** 摘要统计卡片。 */
const stats = computed<StatGridItem[]>(() => {
  const result = data.value
  if (!result) {
    return []
  }
  const items: StatGridItem[] = [
    { label: t('tools.domain-verify.results.domain'), value: result.domain },
    { label: t('tools.domain-verify.results.expected'), value: result.expected },
    {
      label: t('tools.domain-verify.results.matched'),
      value: result.matched ? t('tool.verdict.pass') : t('tool.verdict.fail'),
      level: result.matched ? ('success' as const) : ('danger' as const),
    },
    {
      label: t('tools.domain-verify.results.resolvers'),
      value: String(result.perResolver.length),
    },
  ]
  if (result.type === 'http-file') {
    items.push({
      label: t('tools.domain-verify.results.httpStatus'),
      value: String(result.httpStatus),
    })
  }
  return items
})

/** 解析器表格列定义。 */
const resolverColumns = computed<DataTableColumn[]>(() => [
  { key: 'resolver', label: t('tools.domain-verify.results.resolver'), mono: true },
  {
    key: 'state',
    label: t('tools.domain-verify.results.state'),
  },
  { key: 'values', label: t('tools.domain-verify.results.values'), mono: true },
])

/** 解析器表格行。 */
const resolverRows = computed<DataTableRow[]>(() =>
  (data.value?.perResolver ?? []).map((outcome) => ({
    key: outcome.resolver,
    level: outcome.error ? ('danger' as const) : outcome.matched ? ('success' as const) : undefined,
    cells: {
      resolver: outcome.resolver,
      state: outcome.error
        ? outcome.error.code
        : outcome.matched
          ? t('tools.domain-verify.results.match')
          : t('tools.domain-verify.results.noMatch'),
      values: outcome.values.join(' | ') || '-',
    },
  })),
)

/** 汇总：各解析器的成功 / 失败数。 */
const summary = computed(() => {
  const result = data.value
  if (!result || result.perResolver.length === 0) {
    return null
  }
  const succeeded = result.perResolver.filter((item) => item.error === null).length
  return {
    succeeded,
    failed: result.perResolver.length - succeeded,
    total: result.perResolver.length,
  }
})

/** 结论徽章：期望值是否匹配。 */
const verdict = computed(() => {
  const result = data.value
  if (!result) {
    return null
  }
  return result.matched
    ? { level: 'success' as const, label: t('tool.verdict.pass') }
    : { level: 'danger' as const, label: t('tool.verdict.fail') }
})

const canRun = computed(
  () => domain.value.trim() !== '' && expected.value.trim() !== '',
)

function run(): void {
  const payload: Record<string, unknown> = {
    type: type.value,
    domain: domain.value.trim(),
    expected: expected.value,
  }
  if (type.value === 'http-file') {
    payload.filePath = filePath.value.trim()
  }
  void execute(payload)
}

function reset(): void {
  type.value = DEFAULT_EXAMPLE.type
  domain.value = DEFAULT_EXAMPLE.domain
  expected.value = DEFAULT_EXAMPLE.expected
  filePath.value = DEFAULT_EXAMPLE.filePath
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="domain-verify"
    name="域名验证"
    name-en="Domain Verifier"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.domain-verify.fields.type') }}</span>
            <ElSelect
              v-model="type"
              class="field__select"
              :aria-label="t('tools.domain-verify.fields.type')"
            >
              <ElOption
                :label="t('tools.domain-verify.options.dnsTxt')"
                value="dns-txt"
              />
              <ElOption
                :label="t('tools.domain-verify.options.dnsCname')"
                value="dns-cname"
              />
              <ElOption
                :label="t('tools.domain-verify.options.httpFile')"
                value="http-file"
              />
            </ElSelect>
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.domain-verify.fields.domain') }}</span>
            <ElInput
              v-model="domain"
              class="field__input"
              spellcheck="false"
              autocomplete="off"
            />
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.domain-verify.fields.expected') }}</span>
            <ElInput
              v-model="expected"
              class="field__input"
              spellcheck="false"
              autocomplete="off"
            />
          </label>
        </template>

        <template #options>
          <label
            v-if="type === 'http-file'"
            class="field"
          >
            <span class="field__label">{{ t('tools.domain-verify.fields.filePath') }}</span>
            <ElInput
              v-model="filePath"
              class="field__input"
              spellcheck="false"
              autocomplete="off"
            />
          </label>
        </template>

        <template #footer>
          <RunBar
            :running="running"
            :disabled="!canRun"
            @run="run"
            @reset="reset"
            @example="reset"
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
        :title="t('tools.domain-verify.results.output')"
        @retry="run"
      >
        <template
          v-if="summary"
          #summary
        >
          <span>{{ t('tool.result.summary.succeeded') }} {{ summary.succeeded }} / {{ summary.total }}</span>
          <span v-if="summary.failed > 0">{{ t('tool.result.summary.failed') }} {{ summary.failed }}</span>
        </template>

        <template
          v-if="data"
          #extra
        >
          <StatGrid :items="stats" />

          <DataTable
            class="table"
            :columns="resolverColumns"
            :rows="resolverRows"
            :copy-column="'values'"
            :empty="t('tools.domain-verify.results.empty')"
          />

          <pre
            v-if="data.type === 'http-file' && data.bodyPreview !== ''"
            class="product"
          ><code>{{ data.bodyPreview }}</code></pre>
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.domain-verify.description') }}
      </p>
    </template>

    <template #faq>
      <dl class="faq">
        <template
          v-for="item in faq"
          :key="item.q"
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
.table {
  margin-top: var(--tb-space-3);
}

.product {
  margin: var(--tb-space-3) 0 0;
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
