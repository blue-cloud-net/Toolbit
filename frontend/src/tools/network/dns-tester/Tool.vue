<script setup lang="ts">
/**
 * DNS 综合检测工具页。
 *
 * 三个检测项：whois（RDAP / whois）、dnssec、ptr。
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

/** DNSSEC 检查结果。 */
interface DNSSECInfo {
  supported: boolean
  dnssecEnabled: boolean
  dsPresent: boolean
  sigPresent: boolean
  chainOfTrust: string[]
}

/** 反向解析结果。 */
interface PTRInfo {
  reverse: string
  addresses: string[]
  names: string[]
}

/** 后端 dns-tester 工具的响应。 */
interface DNSTesterResult {
  action: string
  whois: WhoisInfo | null
  dnssec: DNSSECInfo | null
  ptr: PTRInfo | null
}

interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const action = ref<'whois' | 'dnssec' | 'ptr'>(DEFAULT_EXAMPLE.action)
const domain = ref(DEFAULT_EXAMPLE.domain)
const ip = ref(DEFAULT_EXAMPLE.ip)

const { state, data, meta, error, run: execute, reset: resetRun } =
  useToolRun<DNSTesterResult>({ slug: 'dns-tester' })

const running = computed(() => state.value === 'loading')

/** 当前动作对应的原始 JSON（作为「原文」视图）。 */
const rawJson = computed(() => {
  const payload = data.value?.whois ?? data.value?.dnssec ?? data.value?.ptr
  return payload ? JSON.stringify(payload, null, 2) : ''
})

/** 结果文本：逐行给出主要字段。 */
const resultText = computed(() => {
  if (!data.value) {
    return ''
  }
  const lines: string[] = []
  const whois = data.value.whois
  if (whois) {
    lines.push(`source: ${whois.source}`)
    lines.push(`registrar: ${whois.registrar}`)
    lines.push(`creationDate: ${whois.creationDate}`)
    lines.push(`expirationDate: ${whois.expirationDate}`)
    lines.push(`updatedDate: ${whois.updatedDate}`)
    lines.push(`registrantCountry: ${whois.registrantCountry}`)
    lines.push(`nameServers: ${whois.nameServers.join(', ')}`)
    lines.push(`status: ${whois.status.join(', ')}`)
    lines.push(`registries: ${whois.registries.join(', ')}`)
  }
  const dnssec = data.value.dnssec
  if (dnssec) {
    lines.push(`supported: ${String(dnssec.supported)}`)
    lines.push(`dnssecEnabled: ${String(dnssec.dnssecEnabled)}`)
    lines.push(`dsPresent: ${String(dnssec.dsPresent)}`)
    lines.push(`sigPresent: ${String(dnssec.sigPresent)}`)
    for (const item of dnssec.chainOfTrust) {
      lines.push(`  ${item}`)
    }
  }
  const ptr = data.value.ptr
  if (ptr) {
    lines.push(`reverse: ${ptr.reverse}`)
    lines.push(`addresses: ${ptr.addresses.join(', ')}`)
    lines.push(`names: ${ptr.names.join(', ')}`)
  }
  return lines.join('\n')
})

const faq = computed<FaqItem[]>(() => tm('tools.dns-tester.faq') as FaqItem[])

/** 摘要统计卡片。 */
const stats = computed<StatGridItem[]>(() => {
  const result = data.value
  if (!result) {
    return []
  }
  const whois = result.whois
  if (whois) {
    return [
      { label: t('tools.dns-tester.results.source'), value: whois.source },
      { label: t('tools.dns-tester.results.registrar'), value: whois.registrar || '-' },
      { label: t('tools.dns-tester.results.creationDate'), value: whois.creationDate || '-' },
      { label: t('tools.dns-tester.results.expirationDate'), value: whois.expirationDate || '-' },
      { label: t('tools.dns-tester.results.updatedDate'), value: whois.updatedDate || '-' },
      {
        label: t('tools.dns-tester.results.registrantCountry'),
        value: whois.registrantCountry || '-',
      },
      { label: t('tools.dns-tester.results.nameServers'), value: String(whois.nameServers.length) },
      { label: t('tools.dns-tester.results.statusCount'), value: String(whois.status.length) },
    ]
  }
  const dnssec = result.dnssec
  if (dnssec) {
    return [
      {
        label: t('tools.dns-tester.results.supported'),
        value: dnssec.supported ? t('common.states.yes') : t('common.states.no'),
        level: dnssec.supported ? ('success' as const) : ('warning' as const),
      },
      {
        label: t('tools.dns-tester.results.dnssecEnabled'),
        value: dnssec.dnssecEnabled ? t('common.states.yes') : t('common.states.no'),
        level: dnssec.dnssecEnabled ? ('success' as const) : ('warning' as const),
      },
      {
        label: t('tools.dns-tester.results.dsPresent'),
        value: dnssec.dsPresent ? t('common.states.yes') : t('common.states.no'),
      },
      {
        label: t('tools.dns-tester.results.sigPresent'),
        value: dnssec.sigPresent ? t('common.states.yes') : t('common.states.no'),
      },
      {
        label: t('tools.dns-tester.results.chainLevels'),
        value: String(dnssec.chainOfTrust.length),
      },
    ]
  }
  const ptr = result.ptr
  if (ptr) {
    return [
      { label: t('tools.dns-tester.results.reverse'), value: ptr.reverse },
      { label: t('tools.dns-tester.results.addresses'), value: ptr.addresses.join(', ') || '-' },
      { label: t('tools.dns-tester.results.names'), value: ptr.names.join(', ') || '-' },
    ]
  }
  return []
})

/** 名称服务器表格列定义。 */
const nameServerColumns = computed<DataTableColumn[]>(() => [
  { key: 'server', label: t('tools.dns-tester.results.nameServer'), mono: true },
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
  { key: 'status', label: t('tools.dns-tester.results.statusItem'), mono: true },
])

/** 状态表格行。 */
const statusRows = computed<DataTableRow[]>(() =>
  (data.value?.whois?.status ?? []).map((item, index) => ({
    key: `${item}-${index}`,
    cells: { status: item },
  })),
)

/** 信任链表格列定义。 */
const chainColumns = computed<DataTableColumn[]>(() => [
  { key: 'level', label: t('tools.dns-tester.results.chainLevels'), align: 'right' },
  { key: 'detail', label: t('tools.dns-tester.results.chainDetail'), mono: true },
])

/** 信任链表格行。 */
const chainRows = computed<DataTableRow[]>(() =>
  (data.value?.dnssec?.chainOfTrust ?? []).map((item, index) => ({
    key: `${item}-${index}`,
    cells: { level: String(index + 1), detail: item },
  })),
)

const canRun = computed(() =>
  action.value === 'ptr' ? ip.value.trim() !== '' || domain.value.trim() !== '' : domain.value.trim() !== '',
)

function run(): void {
  const payload: Record<string, unknown> = { action: action.value }
  if (domain.value.trim() !== '') {
    payload.domain = domain.value.trim()
  }
  if (ip.value.trim() !== '') {
    payload.ip = ip.value.trim()
  }
  void execute(payload)
}

function reset(): void {
  action.value = DEFAULT_EXAMPLE.action
  domain.value = DEFAULT_EXAMPLE.domain
  ip.value = DEFAULT_EXAMPLE.ip
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="dns-tester"
    name="DNS 多记录测试"
    name-en="DNS Record Tester"
    wide
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.dns-tester.fields.action') }}</span>
            <ElSelect
              v-model="action"
              class="field__select"
              :aria-label="t('tools.dns-tester.fields.action')"
            >
              <ElOption
                :label="t('tools.dns-tester.options.whois')"
                value="whois"
              />
              <ElOption
                :label="t('tools.dns-tester.options.dnssec')"
                value="dnssec"
              />
              <ElOption
                :label="t('tools.dns-tester.options.ptr')"
                value="ptr"
              />
            </ElSelect>
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.dns-tester.fields.domain') }}</span>
            <ElInput
              v-model="domain"
              class="field__input"
              spellcheck="false"
              autocomplete="off"
            />
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.dns-tester.fields.ip') }}</span>
            <ElInput
              v-model="ip"
              class="field__input"
              spellcheck="false"
              autocomplete="off"
              placeholder="8.8.8.8"
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
        :title="t('tools.dns-tester.results.output')"
        @retry="run"
      >
        <template
          v-if="data"
          #extra
        >
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
            v-if="chainRows.length > 0"
            class="table"
            :columns="chainColumns"
            :rows="chainRows"
            :copy-column="'detail'"
          />

          <details
            v-if="rawJson !== ''"
            class="raw"
          >
            <summary class="raw__summary">
              {{ t('tools.dns-tester.results.rawJson') }}
            </summary>
            <pre class="raw__body"><code>{{ rawJson }}</code></pre>
          </details>
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.dns-tester.description') }}
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

.raw {
  margin-top: var(--tb-space-3);
}

.raw__summary {
  font-size: var(--tb-font-xs);
  color: var(--tb-text-secondary);
  cursor: pointer;
}

.raw__body {
  margin: var(--tb-space-2) 0 0;
  padding: var(--tb-space-3);
  overflow-x: auto;
  font-family: var(--tb-font-mono);
  font-size: var(--tb-font-xs);
  line-height: 1.6;
  color: var(--tb-text);
  white-space: pre-wrap;
  word-break: break-all;
  background: var(--tb-bg-sunken);
  border-radius: var(--tb-radius-control);
}
</style>
