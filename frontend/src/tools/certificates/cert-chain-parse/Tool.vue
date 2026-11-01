<script setup lang="ts">
/**
 * 证书链解析工具页。
 *
 * 证书内容只存在于组件局部状态；离开页面即丢弃，
 * **不写**任何浏览器存储（AGENTS.md §4.11）。
 */
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import DataTable, { type DataTableColumn, type DataTableRow } from '@/components/tool/DataTable.vue'
import ResultPanel from '@/components/tool/ResultPanel.vue'
import MaterialInput from '@/components/tool/MaterialInput.vue'
import RunBar from '@/components/tool/RunBar.vue'
import StatGrid, { type StatGridItem } from '@/components/tool/StatGrid.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useToolRun } from '@/composables/useToolRun'
import { DEFAULT_EXAMPLE } from '@/tools/certificates/cert-chain-parse/example'

/** 链上一张证书。 */
interface ChainCert {
  pem: string
  info: {
    subjectText: string
    issuerText: string
    serial: string
    notAfter: string
    isCa: boolean
  }
}

/** 后端 cert-chain-parse 工具的响应。 */
interface ChainParseResult {
  count: number
  chain: ChainCert[]
  ordered: boolean
  missingRoot: boolean
  selfSigned: boolean
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const bundle = ref('')
const bundleFormat = ref('auto')

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
  bindInput,
  useAsInput,
} = useToolRun<ChainParseResult>({ slug: 'cert-chain-parse' })

/** 结果区的「作为新输入」：把结果回填到主输入框（仅内存，不写存储）。 */
bindInput((text) => {
  bundle.value = text
})

const running = computed(() => state.value === 'loading')

/** 结果文本。 */
const resultText = computed(() => {
  const result = data.value
  if (!result) {
    return ''
  }
  const head = [
    `${t('tools.cert-chain-parse.results.count')}: ${result.count}`,
    `${t('tools.cert-chain-parse.results.ordered')}: ${result.ordered}`,
    `${t('tools.cert-chain-parse.results.missingRoot')}: ${result.missingRoot}`,
    `${t('tools.cert-chain-parse.results.selfSigned')}: ${result.selfSigned}`,
    '',
  ]
  result.chain.forEach((cert, index) => {
    head.push(`[${index}] ${cert.info.subjectText}`)
    head.push(`    issuer: ${cert.info.issuerText}`)
    head.push(`    serial: ${cert.info.serial}  notAfter: ${cert.info.notAfter}`)
  })
  return head.join('\n')
})

const faq = computed<FaqItem[]>(() => tm('tools.cert-chain-parse.faq') as FaqItem[])

/** 摘要统计卡片。 */
const stats = computed<StatGridItem[]>(() => {
  const result = data.value
  if (!result) {
    return []
  }
  return [
    { label: t('tools.cert-chain-parse.results.count'), value: String(result.count) },
    {
      label: t('tools.cert-chain-parse.results.ordered'),
      value: result.ordered ? t('common.states.yes') : t('common.states.no'),
      level: result.ordered ? ('success' as const) : ('warning' as const),
    },
    {
      label: t('tools.cert-chain-parse.results.missingRoot'),
      value: result.missingRoot ? t('common.states.yes') : t('common.states.no'),
      level: result.missingRoot ? ('warning' as const) : ('success' as const),
    },
    {
      label: t('tools.cert-chain-parse.results.selfSigned'),
      value: result.selfSigned ? t('common.states.yes') : t('common.states.no'),
    },
  ]
})

/** 链上证书表格列定义。 */
const chainColumns = computed<DataTableColumn[]>(() => [
  { key: 'index', label: t('tools.cert-chain-parse.results.index'), align: 'right' },
  { key: 'subject', label: t('tools.cert-chain-parse.results.subject'), mono: true },
  { key: 'issuer', label: t('tools.cert-chain-parse.results.issuer'), mono: true },
  { key: 'serial', label: t('tools.cert-chain-parse.results.serial'), mono: true },
  { key: 'notAfter', label: t('tools.cert-chain-parse.results.notAfter'), mono: true },
])

/** 链上证书表格行。 */
const chainRows = computed<DataTableRow[]>(() =>
  (data.value?.chain ?? []).map((cert, index) => ({
    key: `${cert.info.serial}-${index}`,
    cells: {
      index: String(index),
      subject: cert.info.subjectText,
      issuer: cert.info.issuerText,
      serial: cert.info.serial,
      notAfter: cert.info.notAfter,
    },
  })),
)

/** 运行当前输入。 */
function run(): void {
  void execute({
    bundle: { data: bundle.value, format: bundleFormat.value },
  })
}

/** 重置表单与结果。 */
function reset(): void {
  bundle.value = ''
  bundleFormat.value = 'auto'
  resetRun()
}

/** 填入示例。 */
function fillExample(): void {
  bundle.value = DEFAULT_EXAMPLE.bundle
  bundleFormat.value = 'auto'
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="cert-chain-parse"
    :name="t('tools.cert-chain-parse.name')"
    name-en="Certificate Chain Parser"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <MaterialInput
            v-model="bundle"
            v-model:format="bundleFormat"
            :label="t('tools.cert-chain-parse.fields.bundle')"
            :rows="12"
            slug="cert-chain-parse"
            append
          />
        </template>

        <template #footer>
          <RunBar
            :running="running"
            :disabled="bundle === ''"
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
        :title="t('tools.cert-chain-parse.results.chain')"
        :use-as-input="useAsInput"
        @retry="run"
      >
        <template
          v-if="data"
          #extra
        >
          <StatGrid :items="stats" />

          <DataTable
            class="table"
            :columns="chainColumns"
            :rows="chainRows"
            :copy-column="'subject'"
            :empty="t('tools.cert-chain-parse.results.empty')"
          />
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.cert-chain-parse.description') }}
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
</style>
