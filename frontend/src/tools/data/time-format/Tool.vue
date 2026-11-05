<script setup lang="ts">
/**
 * 时间格式速查工具页。
 *
 * 默认展示**全量**格式列表；键入搜索词或点选分类后即时收敛，
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
import { EXAMPLES } from '@/tools/data/time-format/example'

/** 一个时间格式条目。 */
interface TimeFormatEntry {
  /** 格式标识。 */
  key: string
  /** 格式名称。 */
  label: string
  /** 中文说明。 */
  description: string
  /** 示例值。 */
  example: string
  /** 格式分类。 */
  category: string
}

/** 后端 time-format 工具的响应。 */
interface TimeFormatResult {
  /** 匹配条数。 */
  total: number
  /** 数据集条目数。 */
  datasetTotal: number
  /** 命中的条目。 */
  matches: TimeFormatEntry[]
  /** 数据集版本。 */
  dataVersion: string
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const query = ref('')
const category = ref('')

/** 分类枚举（与后端 `TimeFormatCategories` 对应）。 */
const CATEGORIES = ['epoch', 'standard', 'numeric']

/** 分类 → 语义级别；用于分类标签着色。 */
const CATEGORY_LEVELS: Record<string, DataTableCellLevel> = {
  epoch: 'info',
  standard: 'success',
  numeric: 'warning',
}

/** 条目是否命中查询词。 */
function matchesQuery(entry: TimeFormatEntry, term: string): boolean {
  if (term === '') {
    return true
  }
  const lowered = term.toLowerCase()
  if (entry.key.toLowerCase().includes(lowered)) {
    return true
  }
  if (entry.label.toLowerCase().includes(lowered)) {
    return true
  }
  return entry.description.includes(term)
}

/** 全量数据集 + 本地筛选；进页面自动取数，之后键入即筛选。 */
const { state, data, meta, error, filtered, counts, reload } = useLiveFilter<
  TimeFormatEntry,
  TimeFormatResult
>({
  slug: 'time-format',
  query,
  select: (result) => result.matches,
  normalize: (raw) => raw.trim(),
  match: matchesQuery,
  category: (entry) => entry.category,
  categoryRef: category,
})

/** 占位提示：直接用示例查询词。 */
const placeholder = EXAMPLES.map((example) => example.query).join(' / ')

/** 分类筛选项；带各自条目数。 */
const categoryOptions = computed<FilterChipOption[]>(() => {
  const options: FilterChipOption[] = [
    {
      value: '',
      label: t('tools.time-format.options.category.all'),
      hint: String(data.value?.datasetTotal ?? 0),
    },
  ]
  for (const name of CATEGORIES) {
    options.push({
      value: name,
      label: t(`tools.time-format.options.category.${name}`),
      hint: String(counts.value[name] ?? 0),
    })
  }
  return options
})

/** 结果文本：命中条目逐行给出标识、名称与示例。 */
const resultText = computed(() =>
  filtered.value
    .map((entry) => `${entry.key}\t${entry.label}\t${entry.example}\t${entry.description}`)
    .join('\n'),
)

/** 统计卡片。 */
const stats = computed<StatGridItem[]>(() => {
  const items: StatGridItem[] = [
    { label: t('tools.time-format.results.total'), value: String(filtered.value.length) },
    { label: t('tools.time-format.results.datasetTotal'), value: String(data.value?.datasetTotal ?? 0) },
  ]
  if (category.value !== '') {
    items.push({
      label: t('tools.time-format.results.category'),
      value: t(`tools.time-format.options.category.${category.value}`),
    })
  }
  items.push({
    label: t('tools.time-format.results.dataVersion'),
    value: data.value?.dataVersion ?? '-',
  })
  return items
})

/** 表格列定义。 */
const columns = computed<DataTableColumn[]>(() => [
  { key: 'key', label: t('tools.time-format.results.key'), mono: true },
  { key: 'label', label: t('tools.time-format.results.label') },
  { key: 'category', label: t('tools.time-format.results.category') },
  { key: 'example', label: t('tools.time-format.results.example'), mono: true },
  { key: 'description', label: t('tools.time-format.results.description') },
])

/** 表格行。 */
const rows = computed<DataTableRow[]>(() =>
  filtered.value.map((entry) => ({
    key: entry.key,
    cells: {
      key: entry.key,
      label: entry.label,
      category: t(`tools.time-format.options.category.${entry.category}`),
      example: entry.example,
      description: entry.description,
    },
    badges: { key: CATEGORY_LEVELS[entry.category] ?? 'muted' },
  })),
)

const faq = computed<FaqItem[]>(() => tm('tools.time-format.faq') as FaqItem[])
</script>

<template>
  <ToolPageLayout
    slug="time-format"
    :name="t('tools.time-format.name')"
    name-en="Date & Time Format Reference"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <SearchField
            v-model="query"
            autofocus
            :label="t('tools.time-format.fields.query')"
            :placeholder="placeholder"
          />

          <div class="field">
            <span class="field__label">{{ t('tools.time-format.fields.category') }}</span>
            <FilterChips
              v-model="category"
              :options="categoryOptions"
              :aria-label="t('tools.time-format.fields.category')"
            />
          </div>

          <p class="field__hint">
            {{ t('tools.time-format.hints.query') }}
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
        :title="t('tools.time-format.results.title')"
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
            :copy-columns="['key', 'example']"
            :badge-copy="['key']"
            :empty="t('tools.time-format.results.empty')"
          />
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.time-format.hints.output') }}
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
