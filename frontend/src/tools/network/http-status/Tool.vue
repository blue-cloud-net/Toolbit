<script setup lang="ts">
/**
 * HTTP 状态码查询工具页。
 *
 * 默认展示**全量**状态码列表；输入搜索词或点选类别后即时收敛，
 * 过滤在**前端本地**完成（数据集已一次性取回），因此输入零延迟。
 *
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
 */
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import DataTable, { type DataTableColumn, type DataTableRow } from '@/components/tool/DataTable.vue'
import FilterChips, { type FilterChipOption } from '@/components/tool/FilterChips.vue'
import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import StatGrid, { type StatGridItem } from '@/components/tool/StatGrid.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useAutoRun } from '@/composables/useAutoRun'
import { useToolRun } from '@/composables/useToolRun'
import { EXAMPLES } from '@/tools/network/http-status/example'

/** 一个状态码条目。 */
interface HttpStatusEntry {
  /** 状态码。 */
  code: number
  /** 标准原因短语。 */
  reason: string
  /** 类别：informational / success / redirect / clientError / serverError。 */
  category: string
  /** 中文说明。 */
  description: string
  /** 是否已废弃。 */
  deprecated: boolean
}

/** 后端 http-status 工具的响应。 */
interface HttpStatusResult {
  /** 规范化后的查询词。 */
  query: string
  /** 查询方式：code / keyword / all。 */
  kind: string
  /** 是否命中标准状态码。 */
  matched: boolean
  /** 状态码。 */
  code: number
  /** 标准原因短语。 */
  reason: string
  /** 类别。 */
  category: string
  /** 中文说明。 */
  description: string
  /** 是否已废弃。 */
  deprecated: boolean
  /** 命中的条目。 */
  matches: HttpStatusEntry[]
  /** 命中条数。 */
  total: number
  /** 数据集版本。 */
  dataVersion: string
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

/** 状态码的合法区间。 */
const MIN_CODE = 100
const MAX_CODE = 599

/** 类别枚举（与后端 `StatusCategories` 对应）。 */
const CATEGORIES = ['informational', 'success', 'redirect', 'clientError', 'serverError']

const { t, tm } = useI18n()

const query = ref('')
const category = ref('')

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<HttpStatusResult>({
  slug: 'http-status',
})

const running = computed(() => state.value === 'loading')

/** 状态码模式下的解析结果；不是纯数字时为 null。 */
const parsedCode = computed<number | null>(() => {
  const text = query.value.trim()
  if (!/^\d+$/.test(text)) {
    return null
  }
  return Number.parseInt(text, 10)
})

/** 输入校验提示：状态码必须是 100–599 的整数。 */
const validation = computed<{ level: 'success' | 'error'; message: string } | null>(() => {
  if (parsedCode.value === null) {
    return null
  }
  if (parsedCode.value < MIN_CODE || parsedCode.value > MAX_CODE) {
    return {
      level: 'error',
      message: t('tools.http-status.hints.invalidCode', { min: MIN_CODE, max: MAX_CODE }),
    }
  }
  return { level: 'success', message: t('tools.http-status.hints.validCode') }
})

/** 类别筛选项；带各自条目数。 */
const categoryOptions = computed<FilterChipOption[]>(() => {
  const counts = new Map<string, number>()
  for (const entry of data.value?.matches ?? []) {
    counts.set(entry.category, (counts.get(entry.category) ?? 0) + 1)
  }
  const options: FilterChipOption[] = [
    {
      value: '',
      label: t('tools.http-status.options.category.all'),
      hint: String(data.value?.total ?? 0),
    },
  ]
  for (const name of CATEGORIES) {
    options.push({
      value: name,
      label: t(`tools.http-status.options.category.${name}`),
      hint: String(counts.get(name) ?? 0),
    })
  }
  return options
})

/** 条目是否命中查询词。 */
function matchesQuery(entry: HttpStatusEntry, term: string): boolean {
  if (term === '') {
    return true
  }
  if (String(entry.code) === term) {
    return true
  }
  const lowered = term.toLowerCase()
  if (entry.reason.toLowerCase().includes(lowered)) {
    return true
  }
  return entry.description.includes(term)
}

/** 本地过滤结果：搜索词与类别叠加。 */
const filtered = computed<HttpStatusEntry[]>(() => {
  const entries = data.value?.matches ?? []
  const term = query.value.trim()
  return entries.filter((entry) => {
    if (category.value !== '' && entry.category !== category.value) {
      return false
    }
    return matchesQuery(entry, term)
  })
})

/** 结果文本：命中条目逐行给出状态码、原因短语与含义。 */
const resultText = computed(() =>
  filtered.value
    .map((entry) => `${entry.code}\t${entry.reason}\t${entry.description}`)
    .join('\n'),
)

/** 统计卡片。 */
const stats = computed<StatGridItem[]>(() => {
  const items: StatGridItem[] = [
    { label: t('tools.http-status.results.total'), value: String(filtered.value.length) },
    { label: t('tools.http-status.results.datasetTotal'), value: String(data.value?.total ?? 0) },
  ]
  const deprecated = filtered.value.filter((entry) => entry.deprecated).length
  if (deprecated > 0) {
    items.push({
      label: t('tools.http-status.results.deprecated'),
      value: String(deprecated),
      level: 'warning',
    })
  }
  items.push({
    label: t('tools.http-status.results.dataVersion'),
    value: data.value?.dataVersion ?? '-',
  })
  return items
})

/** 表格列定义。 */
const columns = computed<DataTableColumn[]>(() => [
  { key: 'code', label: t('tools.http-status.results.code'), mono: true, align: 'right' },
  { key: 'reason', label: t('tools.http-status.results.reason'), mono: true },
  { key: 'category', label: t('tools.http-status.results.category') },
  { key: 'description', label: t('tools.http-status.results.description') },
])

/** 表格行。 */
const rows = computed<DataTableRow[]>(() =>
  filtered.value.map((entry) => ({
    key: String(entry.code),
    level: entry.deprecated ? ('warning' as const) : undefined,
    cells: {
      code: String(entry.code),
      reason: entry.reason,
      category: t(`tools.http-status.options.category.${entry.category}`),
      description: entry.description,
    },
  })),
)

const faq = computed<FaqItem[]>(() => tm('tools.http-status.faq') as FaqItem[])

/** 拉取全量数据集。 */
function loadAll(): void {
  void execute({})
}

// 首屏拉全量；catalog 加载完成后触发一次。
useAutoRun({
  slug: 'http-status',
  run: loadAll,
  watch: [],
  immediate: true,
})

/** 重置表单与结果。 */
function reset(): void {
  query.value = ''
  category.value = ''
  resetRun()
  loadAll()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  query.value = example.query
  category.value = ''
}
</script>

<template>
  <ToolPageLayout
    slug="http-status"
    :name="t('tools.http-status.name')"
    name-en="HTTP Status Code Lookup"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.http-status.fields.query') }}</span>
            <ElInput
              v-model="query"
              class="field__input tb-mono"
              autocomplete="off"
              placeholder="404 / not found / 重定向"
              :aria-label="t('tools.http-status.fields.query')"
            />
          </label>

          <div class="field">
            <span class="field__label">{{ t('tools.http-status.fields.category') }}</span>
            <FilterChips
              v-model="category"
              :options="categoryOptions"
              :aria-label="t('tools.http-status.fields.category')"
            />
          </div>

          <ValidationHint :validation="validation" />

          <p class="field__hint">
            {{ t('tools.http-status.hints.query') }}
          </p>
        </template>

        <template #footer>
          <RunBar
            :running="running"
            @run="loadAll"
            @reset="reset"
            @example="fillExample(0)"
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
        :title="t('tools.http-status.results.title')"
        @retry="loadAll"
      >
        <template
          v-if="data"
          #extra
        >
          <StatGrid :items="stats" />

          <DataTable
            class="table"
            :columns="columns"
            :rows="rows"
            :copy-column="'reason'"
            :empty="t('tools.http-status.results.empty')"
          />
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.http-status.hints.output') }}
      </p>
    </template>

    <template #faq>
      <dl class="faq">
        <template
          v-for="(item, index) in faq"
          :key="index"
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
