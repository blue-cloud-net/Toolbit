<script setup lang="ts">
/**
 * 吊销列表解析工具页。
 *
 * 材料内容只存在于组件局部状态；离开页面即丢弃，
 * **不写**任何浏览器存储（AGENTS.md §4.11）。
 */
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import CertTextInput from '@/components/tool/CertTextInput.vue'
import DataTable, { type DataTableColumn, type DataTableRow } from '@/components/tool/DataTable.vue'
import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import StatGrid, { type StatGridItem } from '@/components/tool/StatGrid.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useToolRun } from '@/composables/useToolRun'
import { DEFAULT_EXAMPLE } from '@/tools/certificates/crl-parse/example'

/** 一条吊销记录。 */
interface RevokedItem {
  serial: string
  revokedAt: string
  reasonCode: number
  reasonText: string
}

/** 后端 crl-parse 工具的响应。 */
interface CRLParseResult {
  issuer: { field: string; value: string }[]
  issuerText: string
  thisUpdate: string
  nextUpdate: string
  signatureAlgorithm: string
  signatureAlgorithmOid: string
  number: number
  revoked: RevokedItem[]
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const input = ref('')
const inputFormat = ref('auto')

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
  bindInput,
  useAsInput,
} = useToolRun<CRLParseResult>({ slug: 'crl-parse' })

/** 结果区的「作为新输入」：把结果回填到主输入框（仅内存，不写存储）。 */
bindInput((text) => {
  input.value = text
})

const running = computed(() => state.value === 'loading')

/** 结果文本。 */
const resultText = computed(() => {
  const result = data.value
  if (!result) {
    return ''
  }
  const head = [
    `${t('tools.crl-parse.results.issuer')}: ${result.issuerText}`,
    `${t('tools.crl-parse.results.thisUpdate')}: ${result.thisUpdate}`,
    `${t('tools.crl-parse.results.nextUpdate')}: ${result.nextUpdate}`,
    `${t('tools.crl-parse.results.signatureAlgorithm')}: ${result.signatureAlgorithm} (${result.signatureAlgorithmOid})`,
    `${t('tools.crl-parse.results.number')}: ${result.number}`,
    '',
    t('tools.crl-parse.results.revoked'),
  ]
  for (const item of result.revoked) {
    head.push(`  ${item.serial || '?'}  ${item.revokedAt}  ${item.reasonText || item.reasonCode}`)
  }
  return head.join('\n')
})

/** 统计卡片。 */
const stats = computed<StatGridItem[]>(() => {
  const result = data.value
  if (!result) {
    return []
  }
  return [
    { label: t('tools.crl-parse.results.issuer'), value: result.issuerText },
    { label: t('tools.crl-parse.results.thisUpdate'), value: result.thisUpdate },
    { label: t('tools.crl-parse.results.nextUpdate'), value: result.nextUpdate },
    {
      label: t('tools.crl-parse.results.signatureAlgorithm'),
      value: result.signatureAlgorithm,
      hint: result.signatureAlgorithmOid,
    },
    { label: t('tools.crl-parse.results.number'), value: String(result.number) },
    {
      label: t('tools.crl-parse.results.revokedCount'),
      value: String(result.revoked.length),
      level: result.revoked.length > 0 ? ('warning' as const) : ('success' as const),
    },
  ]
})

/** 吊销条目表格列定义。 */
const revokedColumns = computed<DataTableColumn[]>(() => [
  { key: 'serial', label: t('tools.crl-parse.results.serial'), mono: true },
  { key: 'revokedAt', label: t('tools.crl-parse.results.revokedAt'), mono: true },
  { key: 'reason', label: t('tools.crl-parse.results.reason') },
])

/** 吊销条目表格行。 */
const revokedRows = computed<DataTableRow[]>(() =>
  (data.value?.revoked ?? []).map((item, index) => ({
    key: `${item.serial}-${index}`,
    cells: {
      serial: item.serial || '?',
      revokedAt: item.revokedAt,
      reason: item.reasonText || String(item.reasonCode),
    },
  })),
)

const faq = computed<FaqItem[]>(() => tm('tools.crl-parse.faq') as FaqItem[])

/** 结论徽章：CRL 中含吊销条目时提示。 */
const verdict = computed(() => {
  const result = data.value
  if (!result) {
    return null
  }
  if (result.revoked.length > 0) {
    return {
      level: 'warning' as const,
      label: t('tool.verdict.revokedCount', { count: result.revoked.length }),
    }
  }
  return { level: 'success' as const, label: t('tool.verdict.pass') }
})

/** 运行当前输入。 */
function run(): void {
  void execute({
    input: { data: input.value, format: inputFormat.value },
  })
}

/** 重置表单与结果。 */
function reset(): void {
  input.value = ''
  inputFormat.value = 'auto'
  resetRun()
}

/** 填入示例。 */
function fillExample(): void {
  input.value = DEFAULT_EXAMPLE.data
  inputFormat.value = DEFAULT_EXAMPLE.format
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="crl-parse"
    :name="t('tools.crl-parse.name')"
    name-en="CRL Parser"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <CertTextInput
            v-model="input"
            v-model:format="inputFormat"
            :label="t('tools.crl-parse.fields.input')"
            :rows="8"
          />
        </template>

        <template #footer>
          <RunBar
            :running="running"
            :disabled="input === ''"
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
        :verdict="verdict"
        :title="t('tools.crl-parse.results.info')"
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
            :columns="revokedColumns"
            :rows="revokedRows"
            :copy-column="'serial'"
            :empty="t('tools.crl-parse.results.noRevoked')"
          />
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.crl-parse.description') }}
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
