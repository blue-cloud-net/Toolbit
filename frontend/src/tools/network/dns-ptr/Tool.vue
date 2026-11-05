<script setup lang="ts">
/**
 * PTR 反向解析工具页。
 *
 * 只给域名时先正向解析取第一个地址再反查；
 * 反查出的名字会再做一次正向解析，用于暴露「反向与正向不一致」。
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

/** 反向解析结果。 */
interface PTRInfo {
  reverse: string
  addresses: string[]
  names: string[]
}

/** 后端 dns-ptr 工具的响应。 */
interface DNSPTRResult {
  query: string
  ptr: PTRInfo | null
}

interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const ip = ref(DEFAULT_EXAMPLE.ip)
const domain = ref('')

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<DNSPTRResult>({
  slug: 'dns-ptr',
})

const running = computed(() => state.value === 'loading')

/** 原始 JSON（作为「原文」视图）。 */
const rawJson = computed(() => (data.value?.ptr ? JSON.stringify(data.value.ptr, null, 2) : ''))

/** 结果文本：逐行给出主要字段。 */
const resultText = computed(() => {
  const ptr = data.value?.ptr
  if (!ptr) {
    return ''
  }
  return [
    `reverse: ${ptr.reverse}`,
    `addresses: ${ptr.addresses.join(', ')}`,
    `names: ${ptr.names.join(', ')}`,
  ].join('\n')
})

const faq = computed<FaqItem[]>(() => tm('tools.dns-ptr.faq') as FaqItem[])

/** 结果页签：摘要 / 原始 JSON。 */
const views = computed<ResultView[]>(() => [
  { id: 'summary', label: t('tools.dns-ptr.views.summary') },
  { id: 'raw', label: t('tools.dns-ptr.views.raw') },
])

/** 当前页签；默认停在「摘要」。 */
const activeView = ref('summary')

/** 摘要统计卡片。 */
const stats = computed<StatGridItem[]>(() => {
  const ptr = data.value?.ptr
  if (!ptr) {
    return []
  }
  return [
    { label: t('tools.dns-ptr.results.reverse'), value: ptr.reverse },
    { label: t('tools.dns-ptr.results.addresses'), value: ptr.addresses.join(', ') || '-' },
    { label: t('tools.dns-ptr.results.names'), value: ptr.names.join(', ') || '-' },
  ]
})

/** 名称表格列定义。 */
const nameColumns = computed<DataTableColumn[]>(() => [
  { key: 'name', label: t('tools.dns-ptr.results.name'), mono: true },
])

/** 名称表格行。 */
const nameRows = computed<DataTableRow[]>(() =>
  (data.value?.ptr?.names ?? []).map((name, index) => ({
    key: `${name}-${index}`,
    cells: { name },
  })),
)

/** 地址表格列定义。 */
const addressColumns = computed<DataTableColumn[]>(() => [
  { key: 'address', label: t('tools.dns-ptr.results.address'), mono: true },
])

/** 地址表格行。 */
const addressRows = computed<DataTableRow[]>(() =>
  (data.value?.ptr?.addresses ?? []).map((address, index) => ({
    key: `${address}-${index}`,
    cells: { address },
  })),
)

/** 是否可以运行：地址与域名至少给一个。 */
const canRun = computed(() => ip.value.trim() !== '' || domain.value.trim() !== '')

/** 运行查询。 */
function run(): void {
  const payload: Record<string, unknown> = {}
  if (ip.value.trim() !== '') {
    payload.ip = ip.value.trim()
  }
  if (domain.value.trim() !== '') {
    payload.domain = domain.value.trim()
  }
  void execute(payload)
}

/** 重置表单与结果。 */
function reset(): void {
  ip.value = DEFAULT_EXAMPLE.ip
  domain.value = ''
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="dns-ptr"
    :name="t('tools.dns-ptr.name')"
    name-en="PTR Reverse Lookup"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.dns-ptr.fields.ip') }}</span>
            <ElInput
              v-model="ip"
              class="field__input"
              spellcheck="false"
              autocomplete="off"
              placeholder="8.8.8.8"
            />
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.dns-ptr.fields.domain') }}</span>
            <ElInput
              v-model="domain"
              class="field__input"
              spellcheck="false"
              autocomplete="off"
              placeholder="example.com"
            />
          </label>

          <p class="field__hint">
            {{ t('tools.dns-ptr.hints.query') }}
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
        :title="t('tools.dns-ptr.results.output')"
        :views="views"
        download-filename="dns-ptr.json"
        download-mime="application/json;charset=utf-8"
        @retry="run"
      >
        <template #view="{ viewId }">
          <template v-if="viewId === 'summary' && data">
            <StatGrid :items="stats" />

            <DataTable
              v-if="nameRows.length > 0"
              class="table"
              :columns="nameColumns"
              :rows="nameRows"
              :copy-column="'name'"
            />

            <DataTable
              v-if="addressRows.length > 0"
              class="table"
              :columns="addressColumns"
              :rows="addressRows"
              :copy-column="'address'"
            />
          </template>

          <ProductText
            v-else-if="viewId === 'raw'"
            :label="t('tools.dns-ptr.views.raw')"
            :content="rawJson"
            filename="dns-ptr.json"
            mime="application/json;charset=utf-8"
          />
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.dns-ptr.hints.output') }}
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
