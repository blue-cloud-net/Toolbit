<script setup lang="ts">
/**
 * CAA 记录生成工具页。
 *
 * 纯文本拼装，不涉及任何密码学运算。
 */
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { ElInputNumber } from 'element-plus/es/components/input-number/index'
import 'element-plus/es/components/input-number/style/css'
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

/** 后端 caa-gen 工具的响应。 */
interface CAAGenResult {
  domain: string
  records: { flags: number; tag: string; value: string; zone: string }[]
  zoneLines: string[]
  count: number
}

interface FaqItem {
  q: string
  a: string
}

const TAGS = ['issue', 'issuewild', 'iodef']

const { t, tm } = useI18n()

const domain = ref('example.com')
const tag = ref('issue')
const value = ref('letsencrypt.org')
const ttl = ref(3600)

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
} = useToolRun<CAAGenResult>({ slug: 'caa-gen' })

const running = computed(() => state.value === 'loading')

const resultText = computed(() =>
  data.value ? data.value.zoneLines.join('\n') : '',
)

const faq = computed<FaqItem[]>(() => tm('tools.caa-gen.faq') as FaqItem[])

/** 摘要统计卡片。 */
const stats = computed<StatGridItem[]>(() => {
  const result = data.value
  if (!result) {
    return []
  }
  return [
    { label: t('tools.caa-gen.results.domain'), value: result.domain },
    { label: t('tools.caa-gen.results.count'), value: String(result.count) },
  ]
})

/** 记录表格列定义。 */
const recordColumns = computed<DataTableColumn[]>(() => [
  { key: 'tag', label: t('tools.caa-gen.results.tag'), mono: true },
  { key: 'value', label: t('tools.caa-gen.results.value'), mono: true },
  { key: 'flags', label: t('tools.caa-gen.results.flags'), mono: true, align: 'right' },
  { key: 'zone', label: t('tools.caa-gen.results.zone'), mono: true },
])

/** 记录表格行。 */
const recordRows = computed<DataTableRow[]>(() =>
  (data.value?.records ?? []).map((record, index) => ({
    key: `${record.tag}-${index}`,
    cells: {
      tag: record.tag,
      value: record.value,
      flags: String(record.flags),
      zone: record.zone,
    },
  })),
)

function run(): void {
  void execute({
    domain: domain.value,
    ttl: ttl.value,
    records: [{ tag: tag.value, value: value.value }],
  })
}

function reset(): void {
  domain.value = 'example.com'
  tag.value = 'issue'
  value.value = 'letsencrypt.org'
  ttl.value = 3600
  resetRun()
}

function fillExample(): void {
  reset()
}
</script>

<template>
  <ToolPageLayout
    slug="caa-gen"
    name="CAA 生成"
    name-en="CAA Generator"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.caa-gen.fields.domain') }}</span>
            <ElInput
              v-model="domain"
              class="field__input"
              spellcheck="false"
              autocomplete="off"
            />
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.caa-gen.fields.tag') }}</span>
            <ElSelect
              v-model="tag"
              class="field__select"
              :aria-label="t('tools.caa-gen.fields.tag')"
            >
              <ElOption
                v-for="name in TAGS"
                :key="name"
                :label="String(name)"
                :value="name"
              />
            </ElSelect>
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.caa-gen.fields.value') }}</span>
            <ElInput
              v-model="value"
              class="field__input"
              spellcheck="false"
              autocomplete="off"
            />
          </label>
        </template>

        <template #options>
          <label class="field">
            <span class="field__label">{{ t('tools.caa-gen.fields.ttl') }}</span>
            <ElInputNumber
              v-model="ttl"
              class="field__input"
              :min="0"
            />
          </label>
        </template>

        <template #footer>
          <RunBar
            :running="running"
            :disabled="domain === '' || value === ''"
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
        :title="t('tools.caa-gen.results.output')"
        @retry="run"
      >
        <template
          v-if="data"
          #extra
        >
          <StatGrid :items="stats" />

          <DataTable
            class="table"
            :columns="recordColumns"
            :rows="recordRows"
            :copy-column="'value'"
            :empty="t('tools.caa-gen.results.empty')"
          />
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.caa-gen.description') }}
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
