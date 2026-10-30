<script setup lang="ts">
/**
 * PKCS#7 提取工具页。
 *
 * 容器只存在于组件局部状态；离开页面即丢弃，**不写**浏览器存储。
 */
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import CertTextInput from '@/components/tool/CertTextInput.vue'
import DataTable, { type DataTableColumn, type DataTableRow } from '@/components/tool/DataTable.vue'
import ResultPanel, { type ResultView } from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import StatGrid, { type StatGridItem } from '@/components/tool/StatGrid.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useToolRun } from '@/composables/useToolRun'
import { TEST_P7B_BASE64 } from '@/tools/certificates/fixtures-v060'

/** 后端 p7b-extract 工具的响应。 */
interface P7BExtractResult {
  certificates: { subjectText: string; issuerText: string; serial: string }[]
  count: number
  pems: string[]
}

interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const container = ref('')
const containerFormat = ref('auto')

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
} = useToolRun<P7BExtractResult>({ slug: 'p7b-extract' })

const running = computed(() => state.value === 'loading')

const resultText = computed(() => {
  const result = data.value
  if (!result) {
    return ''
  }
  return [`${t('tools.p7b-extract.results.count')}: ${result.count}`, '', ...result.pems].join('\n')
})

const faq = computed<FaqItem[]>(() => tm('tools.p7b-extract.faq') as FaqItem[])

/** 结果页签：摘要 / 证书清单 / PEM。 */
const views = computed<ResultView[]>(() => [
  { id: 'summary', label: t('tools.p7b-extract.views.summary') },
  { id: 'list', label: t('tools.p7b-extract.views.list') },
  { id: 'pem', label: t('tools.p7b-extract.views.pem') },
])

/** 当前页签。 */
const activeView = ref('summary')

/** 摘要统计卡片。 */
const stats = computed<StatGridItem[]>(() => {
  const result = data.value
  if (!result) {
    return []
  }
  return [
    { label: t('tools.p7b-extract.results.count'), value: String(result.count) },
  ]
})

/** 证书表格列定义。 */
const certColumns = computed<DataTableColumn[]>(() => [
  { key: 'subject', label: t('tools.p7b-extract.results.subject'), mono: true },
  { key: 'issuer', label: t('tools.p7b-extract.results.issuer'), mono: true },
  { key: 'serial', label: t('tools.p7b-extract.results.serial'), mono: true },
])

/** 证书表格行。 */
const certRows = computed<DataTableRow[]>(() =>
  (data.value?.certificates ?? []).map((cert, index) => ({
    key: `${cert.serial}-${index}`,
    cells: {
      subject: cert.subjectText,
      issuer: cert.issuerText,
      serial: cert.serial,
    },
  })),
)

function run(): void {
  void execute({
    container: { data: container.value, format: containerFormat.value },
  })
}

function reset(): void {
  container.value = ''
  containerFormat.value = 'auto'
  resetRun()
}

function fillExample(): void {
  reset()
  container.value = TEST_P7B_BASE64
  containerFormat.value = 'base64'
}
</script>

<template>
  <ToolPageLayout
    slug="p7b-extract"
    name="PKCS#7 提取"
    name-en="PKCS#7 Extractor"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <CertTextInput
            v-model="container"
            v-model:format="containerFormat"
            :label="t('tools.p7b-extract.fields.container')"
            :rows="8"
          />
        </template>

        <template #footer>
          <RunBar
            :running="running"
            :disabled="container === ''"
            @run="run"
            @reset="reset"
            @example="fillExample"
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
        :title="t('tools.p7b-extract.results.output')"
        :views="views"
        @retry="run"
      >
        <template #view="{ viewId }">
          <StatGrid
            v-if="viewId === 'summary'"
            :items="stats"
          />
          <DataTable
            v-else-if="viewId === 'list'"
            :columns="certColumns"
            :rows="certRows"
            :copy-column="'serial'"
            :empty="t('tools.p7b-extract.results.empty')"
          />
          <pre
            v-else-if="data"
            class="product"
          ><code>{{ data.pems.join('\n') }}</code></pre>
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.p7b-extract.description') }}
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
.product {
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
