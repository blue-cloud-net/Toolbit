<script setup lang="ts">
/**
 * DNSSEC 检测工具页。
 *
 * 检查域名的 DS 记录、RRSIG 签名与逐级向上的信任链。
 */
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import DataTable, { type DataTableColumn, type DataTableRow } from '@/components/tool/DataTable.vue'
import ProductText from '@/components/tool/ProductText.vue'
import ResultPanel, { type ResultView } from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import StatGrid, { type StatGridItem } from '@/components/tool/StatGrid.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useToolRun } from '@/composables/useToolRun'

import { DEFAULT_EXAMPLE } from './example'

/** DNSSEC 检查结果。 */
interface DNSSECInfo {
  supported: boolean
  dnssecEnabled: boolean
  dsPresent: boolean
  sigPresent: boolean
  chainOfTrust: string[]
}

/** 后端 dnssec-check 工具的响应。 */
interface DNSSECCheckResult {
  domain: string
  dnssec: DNSSECInfo | null
}

interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const domain = ref(DEFAULT_EXAMPLE.domain)

const { state, data, meta, error, run: execute, reset: resetRun } =
  useToolRun<DNSSECCheckResult>({ slug: 'dnssec-check' })

const running = computed(() => state.value === 'loading')

/** 原始 JSON（作为「原文」视图）。 */
const rawJson = computed(() =>
  data.value?.dnssec ? JSON.stringify(data.value.dnssec, null, 2) : '',
)

/** 结果文本：逐行给出主要字段。 */
const resultText = computed(() => {
  const dnssec = data.value?.dnssec
  if (!dnssec) {
    return ''
  }
  const lines = [
    `domain: ${data.value?.domain ?? ''}`,
    `supported: ${String(dnssec.supported)}`,
    `dnssecEnabled: ${String(dnssec.dnssecEnabled)}`,
    `dsPresent: ${String(dnssec.dsPresent)}`,
    `sigPresent: ${String(dnssec.sigPresent)}`,
  ]
  for (const item of dnssec.chainOfTrust) {
    lines.push(`  ${item}`)
  }
  return lines.join('\n')
})

const faq = computed<FaqItem[]>(() => tm('tools.dnssec-check.faq') as FaqItem[])

/** 结果页签：摘要 / 原始 JSON。 */
const views = computed<ResultView[]>(() => [
  { id: 'summary', label: t('tools.dnssec-check.views.summary') },
  { id: 'raw', label: t('tools.dnssec-check.views.raw') },
])

/** 当前页签；默认停在「摘要」。 */
const activeView = ref('summary')

/** 摘要统计卡片。 */
const stats = computed<StatGridItem[]>(() => {
  const dnssec = data.value?.dnssec
  if (!dnssec) {
    return []
  }
  return [
    {
      label: t('tools.dnssec-check.results.supported'),
      value: dnssec.supported ? t('common.states.yes') : t('common.states.no'),
      level: dnssec.supported ? ('success' as const) : ('warning' as const),
    },
    {
      label: t('tools.dnssec-check.results.dnssecEnabled'),
      value: dnssec.dnssecEnabled ? t('common.states.yes') : t('common.states.no'),
      level: dnssec.dnssecEnabled ? ('success' as const) : ('warning' as const),
    },
    {
      label: t('tools.dnssec-check.results.dsPresent'),
      value: dnssec.dsPresent ? t('common.states.yes') : t('common.states.no'),
    },
    {
      label: t('tools.dnssec-check.results.sigPresent'),
      value: dnssec.sigPresent ? t('common.states.yes') : t('common.states.no'),
    },
    {
      label: t('tools.dnssec-check.results.chainLevels'),
      value: String(dnssec.chainOfTrust.length),
    },
  ]
})

/** 信任链表格列定义。 */
const chainColumns = computed<DataTableColumn[]>(() => [
  { key: 'level', label: t('tools.dnssec-check.results.chainLevels'), align: 'right' },
  { key: 'detail', label: t('tools.dnssec-check.results.chainDetail'), mono: true },
])

/** 信任链表格行。 */
const chainRows = computed<DataTableRow[]>(() =>
  (data.value?.dnssec?.chainOfTrust ?? []).map((item, index) => ({
    key: `${item}-${index}`,
    cells: { level: String(index + 1), detail: item },
  })),
)

/** 是否可以运行。 */
const canRun = computed(() => domain.value.trim() !== '')

/** 运行检查。 */
function run(): void {
  void execute({ domain: domain.value.trim() })
}

/** 重置表单与结果。 */
function reset(): void {
  domain.value = DEFAULT_EXAMPLE.domain
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="dnssec-check"
    :name="t('tools.dnssec-check.name')"
    name-en="DNSSEC Checker"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.dnssec-check.fields.domain') }}</span>
            <ElInput
              v-model="domain"
              class="field__input"
              spellcheck="false"
              autocomplete="off"
              placeholder="example.com"
            />
          </label>

          <p class="field__hint">
            {{ t('tools.dnssec-check.hints.domain') }}
          </p>
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
        v-model:active-view="activeView"
        :state="state"
        :text="resultText"
        :meta="meta"
        :error="error"
        :title="t('tools.dnssec-check.results.output')"
        :views="views"
        download-filename="dnssec-check.json"
        download-mime="application/json;charset=utf-8"
        @retry="run"
      >
        <template #view="{ viewId }">
          <template v-if="viewId === 'summary' && data">
            <StatGrid :items="stats" />

            <DataTable
              v-if="chainRows.length > 0"
              class="table"
              :columns="chainColumns"
              :rows="chainRows"
              :copy-column="'detail'"
            />
          </template>

          <ProductText
            v-else-if="viewId === 'raw'"
            :label="t('tools.dnssec-check.views.raw')"
            :content="rawJson"
            filename="dnssec-check.json"
            mime="application/json;charset=utf-8"
          />
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.dnssec-check.hints.output') }}
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
