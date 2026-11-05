<script setup lang="ts">
/**
 * WHOIS 查询工具页。
 *
 * 查域名或 IP 的注册信息：优先 RDAP，失败时回退 whois 43 端口。
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

/** whois 注册信息。 */
interface WhoisInfo {
  source: string
  registrar: string
  creationDate: string
  expirationDate: string
  updatedDate: string
  nameServers: string[]
  status: string[]
  registrantCountry: string
  registries: string[]
}

/** 后端 dns-whois 工具的响应。 */
interface DNSWhoisResult {
  query: string
  whois: WhoisInfo | null
}

interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const domain = ref(DEFAULT_EXAMPLE.domain)
const ip = ref('')

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<DNSWhoisResult>({
  slug: 'dns-whois',
})

const running = computed(() => state.value === 'loading')

/** 原始 JSON（作为「原文」视图）。 */
const rawJson = computed(() =>
  data.value?.whois ? JSON.stringify(data.value.whois, null, 2) : '',
)

/** 结果文本：逐行给出主要字段。 */
const resultText = computed(() => {
  const whois = data.value?.whois
  if (!whois) {
    return ''
  }
  return [
    `source: ${whois.source}`,
    `registrar: ${whois.registrar}`,
    `creationDate: ${whois.creationDate}`,
    `expirationDate: ${whois.expirationDate}`,
    `updatedDate: ${whois.updatedDate}`,
    `registrantCountry: ${whois.registrantCountry}`,
    `nameServers: ${whois.nameServers.join(', ')}`,
    `status: ${whois.status.join(', ')}`,
    `registries: ${whois.registries.join(', ')}`,
  ].join('\n')
})

const faq = computed<FaqItem[]>(() => tm('tools.dns-whois.faq') as FaqItem[])

/** 结果页签：摘要 / 原始 JSON。 */
const views = computed<ResultView[]>(() => [
  { id: 'summary', label: t('tools.dns-whois.views.summary') },
  { id: 'raw', label: t('tools.dns-whois.views.raw') },
])

/** 当前页签；默认停在「摘要」。 */
const activeView = ref('summary')

/** 摘要统计卡片。 */
const stats = computed<StatGridItem[]>(() => {
  const whois = data.value?.whois
  if (!whois) {
    return []
  }
  return [
    { label: t('tools.dns-whois.results.source'), value: whois.source },
    { label: t('tools.dns-whois.results.registrar'), value: whois.registrar || '-' },
    { label: t('tools.dns-whois.results.creationDate'), value: whois.creationDate || '-' },
    { label: t('tools.dns-whois.results.expirationDate'), value: whois.expirationDate || '-' },
    { label: t('tools.dns-whois.results.updatedDate'), value: whois.updatedDate || '-' },
    {
      label: t('tools.dns-whois.results.registrantCountry'),
      value: whois.registrantCountry || '-',
    },
    { label: t('tools.dns-whois.results.nameServers'), value: String(whois.nameServers.length) },
    { label: t('tools.dns-whois.results.statusCount'), value: String(whois.status.length) },
  ]
})

/** 名称服务器表格列定义。 */
const nameServerColumns = computed<DataTableColumn[]>(() => [
  { key: 'server', label: t('tools.dns-whois.results.nameServer'), mono: true },
])

/** 名称服务器表格行。 */
const nameServerRows = computed<DataTableRow[]>(() =>
  (data.value?.whois?.nameServers ?? []).map((server, index) => ({
    key: `${server}-${index}`,
    cells: { server },
  })),
)

/** 状态表格列定义。 */
const statusColumns = computed<DataTableColumn[]>(() => [
  { key: 'status', label: t('tools.dns-whois.results.statusItem'), mono: true },
])

/** 状态表格行。 */
const statusRows = computed<DataTableRow[]>(() =>
  (data.value?.whois?.status ?? []).map((item, index) => ({
    key: `${item}-${index}`,
    cells: { status: item },
  })),
)

/** 注册局表格列定义。 */
const registryColumns = computed<DataTableColumn[]>(() => [
  { key: 'registry', label: t('tools.dns-whois.results.registry'), mono: true },
])

/** 注册局表格行。 */
const registryRows = computed<DataTableRow[]>(() =>
  (data.value?.whois?.registries ?? []).map((item, index) => ({
    key: `${item}-${index}`,
    cells: { registry: item },
  })),
)

/** 是否可以运行：域名与地址至少给一个。 */
const canRun = computed(() => domain.value.trim() !== '' || ip.value.trim() !== '')

/** 运行查询。 */
function run(): void {
  const payload: Record<string, unknown> = {}
  if (domain.value.trim() !== '') {
    payload.domain = domain.value.trim()
  }
  if (ip.value.trim() !== '') {
    payload.ip = ip.value.trim()
  }
  void execute(payload)
}

/** 重置表单与结果。 */
function reset(): void {
  domain.value = DEFAULT_EXAMPLE.domain
  ip.value = ''
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="dns-whois"
    :name="t('tools.dns-whois.name')"
    name-en="WHOIS Lookup"
    wide
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.dns-whois.fields.domain') }}</span>
            <ElInput
              v-model="domain"
              class="field__input"
              spellcheck="false"
              autocomplete="off"
              placeholder="example.com"
            />
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.dns-whois.fields.ip') }}</span>
            <ElInput
              v-model="ip"
              class="field__input"
              spellcheck="false"
              autocomplete="off"
              placeholder="8.8.8.8"
            />
          </label>

          <p class="field__hint">
            {{ t('tools.dns-whois.hints.query') }}
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
        :title="t('tools.dns-whois.results.output')"
        :views="views"
        download-filename="dns-whois.json"
        download-mime="application/json;charset=utf-8"
        @retry="run"
      >
        <template #view="{ viewId }">
          <template v-if="viewId === 'summary' && data">
            <StatGrid :items="stats" />

            <DataTable
              v-if="nameServerRows.length > 0"
              class="table"
              :columns="nameServerColumns"
              :rows="nameServerRows"
              :copy-column="'server'"
            />

            <DataTable
              v-if="statusRows.length > 0"
              class="table"
              :columns="statusColumns"
              :rows="statusRows"
              :copy-column="'status'"
            />

            <DataTable
              v-if="registryRows.length > 0"
              class="table"
              :columns="registryColumns"
              :rows="registryRows"
              :copy-column="'registry'"
            />
          </template>

          <ProductText
            v-else-if="viewId === 'raw'"
            :label="t('tools.dns-whois.views.raw')"
            :content="rawJson"
            filename="dns-whois.json"
            mime="application/json;charset=utf-8"
          />
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.dns-whois.hints.output') }}
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
