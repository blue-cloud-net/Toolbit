<script setup lang="ts">
/**
 * HTTP 状态码查询工具页。
 *
 * 默认展示**全量**状态码列表；键入搜索词或点选类别后即时收敛，
 * 过滤在**前端本地**完成（数据集已一次性取回），因此输入零延迟。
 *
 * 本页**没有运行按钮**：没有需要触发的动作 —— 进页面即取数，键入即筛选。
 *
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
 */
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import DataTable, {
  type DataTableCellLevel,
  type DataTableColumn,
  type DataTableRow,
} from '@/components/tool/DataTable.vue'
import FilterChips, { type FilterChipOption } from '@/components/tool/FilterChips.vue'
import ResultPanel from '@/components/tool/ResultPanel.vue'
import SearchField from '@/components/tool/SearchField.vue'
import StatGrid, { type StatGridItem } from '@/components/tool/StatGrid.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useLiveFilter } from '@/composables/useLiveFilter'
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

/** 类别 → 语义级别；用于状态码列的彩色标签。 */
const CATEGORY_LEVELS: Record<string, DataTableCellLevel> = {
  informational: 'muted',
  success: 'success',
  redirect: 'info',
  clientError: 'warning',
  serverError: 'danger',
}

const { t, tm } = useI18n()

const query = ref('')
const category = ref('')

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

/** 全量数据集 + 本地筛选；进页面自动取数，之后键入即筛选。 */
const { state, data, meta, error, filtered, counts, reload } = useLiveFilter<
  HttpStatusEntry,
  HttpStatusResult
>({
  slug: 'http-status',
  query,
  select: (result) => result.matches,
  normalize: (raw) => raw.trim(),
  match: matchesQuery,
  category: (entry) => entry.category,
  categoryRef: category,
})

/** 占位提示：直接用示例查询词。 */
const placeholder = EXAMPLES.map((example) => example.query).join(' / ')

/** 类别筛选项；带各自条目数。 */
const categoryOptions = computed<FilterChipOption[]>(() => {
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
      hint: String(counts.value[name] ?? 0),
    })
  }
  return options
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
    // 废弃条目不与类别色撞色：整行降为次级文字，状态码标签仍保留类别色。
    level: entry.deprecated ? ('muted' as const) : undefined,
    cells: {
      code: String(entry.code),
      reason: entry.reason,
      category: t(`tools.http-status.options.category.${entry.category}`),
      description: entry.deprecated
        ? `${entry.description}（${t('tools.http-status.results.deprecated')}）`
        : entry.description,
    },
    badges: { code: CATEGORY_LEVELS[entry.category] ?? 'muted' },
  })),
)

const faq = computed<FaqItem[]>(() => tm('tools.http-status.faq') as FaqItem[])
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
          <SearchField
            v-model="query"
            autofocus
            :label="t('tools.http-status.fields.query')"
            :placeholder="placeholder"
          />

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
      </ToolPanel>
    </template>

    <template #result>
      <ResultPanel
        :state="state"
        :text="resultText"
        :meta="meta"
        :error="error"
        :title="t('tools.http-status.results.title')"
        hide-text
        @retry="reload"
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
            :copy-columns="['reason', 'description']"
            :badge-copy="['code']"
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
