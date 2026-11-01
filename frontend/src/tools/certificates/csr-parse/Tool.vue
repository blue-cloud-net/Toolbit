<script setup lang="ts">
/**
 * CSR 解析工具页。
 *
 * CSR 内容只存在于组件局部状态；离开页面即丢弃，
 * **不写**任何浏览器存储（AGENTS.md §4.11）。
 */
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import MaterialInput from '@/components/tool/MaterialInput.vue'
import DataTable, { type DataTableColumn, type DataTableRow } from '@/components/tool/DataTable.vue'
import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import StatGrid, { type StatGridItem } from '@/components/tool/StatGrid.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useToolRun } from '@/composables/useToolRun'
import { DEFAULT_EXAMPLE } from '@/tools/certificates/csr-parse/example'

/** 后端 csr-parse 工具的响应。 */
interface CSRParseResult {
  subject: { field: string; value: string }[]
  subjectText: string
  sans: string[]
  publicKey: { algorithm: string; bits: number; curve: string }
  signatureAlgorithm: string
  signatureAlgorithmOid: string
  extensions: {
    nid: number
    oid: string
    name: string
    critical: boolean
    value: string
  }[]
  fingerprintSha256: string
  challengePasswordPresent: boolean
  signatureValid: boolean
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const csr = ref('')
const csrFormat = ref('auto')

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
} = useToolRun<CSRParseResult>({ slug: 'csr-parse' })

const running = computed(() => state.value === 'loading')

/** 结果文本。 */
const resultText = computed(() => {
  const result = data.value
  if (!result) {
    return ''
  }
  const head = [
    `${t('tools.csr-parse.results.subject')}: ${result.subjectText}`,
    `${t('tools.csr-parse.results.publicKey')}: ${result.publicKey.algorithm} ${result.publicKey.bits}`,
    `${t('tools.csr-parse.results.signatureAlgorithm')}: ${result.signatureAlgorithm}`,
    `${t('tools.csr-parse.results.signatureValid')}: ${result.signatureValid}`,
    `${t('tools.csr-parse.results.fingerprint')}: ${result.fingerprintSha256}`,
    `${t('tools.csr-parse.results.challengePassword')}: ${result.challengePasswordPresent}`,
  ]
  if (result.sans.length > 0) {
    head.push(`${t('tools.csr-parse.results.sans')}: ${result.sans.join(', ')}`)
  }
  if (result.extensions.length > 0) {
    head.push('')
    head.push(t('tools.csr-parse.results.extensions'))
    for (const ext of result.extensions) {
      head.push(`  ${ext.name || ext.oid}${ext.critical ? ' (critical)' : ''}: ${ext.value}`)
    }
  }
  return head.join('\n')
})

const faq = computed<FaqItem[]>(() => tm('tools.csr-parse.faq') as FaqItem[])

/** 摘要统计卡片。 */
const stats = computed<StatGridItem[]>(() => {
  const result = data.value
  if (!result) {
    return []
  }
  return [
    { label: t('tools.csr-parse.results.subject'), value: result.subjectText },
    {
      label: t('tools.csr-parse.results.publicKey'),
      value: `${result.publicKey.algorithm} ${result.publicKey.bits}`,
    },
    { label: t('tools.csr-parse.results.signatureAlgorithm'), value: result.signatureAlgorithm },
    {
      label: t('tools.csr-parse.results.signatureValid'),
      value: result.signatureValid ? t('tool.verdict.pass') : t('tool.verdict.fail'),
      level: result.signatureValid ? ('success' as const) : ('danger' as const),
    },
    { label: t('tools.csr-parse.results.fingerprint'), value: result.fingerprintSha256 },
    {
      label: t('tools.csr-parse.results.challengePassword'),
      value: result.challengePasswordPresent ? t('common.states.yes') : t('common.states.no'),
    },
    {
      label: t('tools.csr-parse.results.sans'),
      value: result.sans.join(', ') || '-',
    },
  ]
})

/** 请求扩展表格列定义。 */
const extensionColumns = computed<DataTableColumn[]>(() => [
  { key: 'name', label: t('tools.csr-parse.results.extensionName'), mono: true },
  { key: 'oid', label: t('tools.csr-parse.results.extensionOid'), mono: true },
  { key: 'critical', label: t('tools.csr-parse.results.extensionCritical') },
  { key: 'value', label: t('tools.csr-parse.results.extensionValue'), mono: true },
])

/** 请求扩展表格行。 */
const extensionRows = computed<DataTableRow[]>(() =>
  (data.value?.extensions ?? []).map((ext, index) => ({
    key: `${ext.oid}-${index}`,
    cells: {
      name: ext.name || ext.oid,
      oid: ext.oid,
      critical: ext.critical ? t('common.states.yes') : t('common.states.no'),
      value: ext.value,
    },
  })),
)

/** 运行当前输入。 */
function run(): void {
  void execute({
    input: { data: csr.value, format: csrFormat.value },
  })
}

/** 重置表单与结果。 */
function reset(): void {
  csr.value = ''
  csrFormat.value = 'auto'
  resetRun()
}

/** 填入示例。 */
function fillExample(): void {
  csr.value = DEFAULT_EXAMPLE.csr
  csrFormat.value = 'auto'
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="csr-parse"
    :name="t('tools.csr-parse.name')"
    name-en="CSR Parser"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <MaterialInput
            v-model="csr"
            v-model:format="csrFormat"
            :label="t('tools.csr-parse.fields.csr')"
          />
        </template>

        <template #footer>
          <RunBar
            :running="running"
            :disabled="csr === ''"
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
        :title="t('tools.csr-parse.results.info')"
        @retry="run"
      >
        <template
          v-if="data"
          #extra
        >
          <StatGrid :items="stats" />

          <DataTable
            class="table"
            :columns="extensionColumns"
            :rows="extensionRows"
            :copy-column="'value'"
            :empty="t('tools.csr-parse.results.noExtensions')"
          />
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.csr-parse.description') }}
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
