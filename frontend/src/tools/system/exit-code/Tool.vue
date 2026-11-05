<script setup lang="ts">
/**
 * 退出码查询工具页。
 *
 * 默认展示**全量**退出码列表；键入搜索词或点选类别后即时收敛，
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
import { EXAMPLES } from '@/tools/system/exit-code/example'

/** 一个退出码条目。 */
interface ExitCodeEntry {
  /** 退出码。 */
  code: number
  /** 简短名称。 */
  label: string
  /** 中文说明。 */
  description: string
  /** 类别：success / failure / signal。 */
  type: string
}

/** 后端 exit-code 工具的响应。 */
interface ExitCodeResult {
  /** 匹配条数。 */
  total: number
  /** 数据集条目数。 */
  datasetTotal: number
  /** 命中的条目。 */
  matches: ExitCodeEntry[]
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
const type = ref('')

/** 类别枚举（与后端 `ExitCodeTypes` 对应）。 */
const TYPES = ['success', 'failure', 'signal']

/** 类别 → 语义级别；用于退出码标签着色。 */
const TYPE_LEVELS: Record<string, DataTableCellLevel> = {
  success: 'success',
  failure: 'danger',
  signal: 'warning',
}

/** 条目是否命中查询词。 */
function matchesQuery(entry: ExitCodeEntry, term: string): boolean {
  if (term === '') {
    return true
  }
  const lowered = term.toLowerCase()
  if (String(entry.code) === term) {
    return true
  }
  if (entry.label.toLowerCase().includes(lowered)) {
    return true
  }
  return entry.description.includes(term)
}

/** 全量数据集 + 本地筛选；进页面自动取数，之后键入即筛选。 */
const { state, data, meta, error, filtered, counts, reload } = useLiveFilter<
  ExitCodeEntry,
  ExitCodeResult
>({
  slug: 'exit-code',
  query,
  select: (result) => result.matches,
  normalize: (raw) => raw.trim(),
  match: matchesQuery,
  category: (entry) => entry.type,
  categoryRef: type,
})

/** 占位提示：直接用示例查询词。 */
const placeholder = EXAMPLES.map((example) => example.query).join(' / ')

/** 类别筛选项；带各自条目数。 */
const typeOptions = computed<FilterChipOption[]>(() => {
  const options: FilterChipOption[] = [
    {
      value: '',
      label: t('tools.exit-code.options.type.all'),
      hint: String(data.value?.datasetTotal ?? 0),
    },
  ]
  for (const name of TYPES) {
    options.push({
      value: name,
      label: t(`tools.exit-code.options.type.${name}`),
      hint: String(counts.value[name] ?? 0),
    })
  }
  return options
})

/** 结果文本：命中条目逐行给出退出码、名称与说明。 */
const resultText = computed(() =>
  filtered.value
    .map((entry) => `${entry.code}\t${entry.label}\t${entry.description}`)
    .join('\n'),
)

/** 统计卡片。 */
const stats = computed<StatGridItem[]>(() => {
  const items: StatGridItem[] = [
    { label: t('tools.exit-code.results.total'), value: String(filtered.value.length) },
    { label: t('tools.exit-code.results.datasetTotal'), value: String(data.value?.datasetTotal ?? 0) },
  ]
  if (type.value !== '') {
    items.push({
      label: t('tools.exit-code.results.type'),
      value: t(`tools.exit-code.options.type.${type.value}`),
    })
  }
  items.push({
    label: t('tools.exit-code.results.dataVersion'),
    value: data.value?.dataVersion ?? '-',
  })
  return items
})

/** 表格列定义。 */
const columns = computed<DataTableColumn[]>(() => [
  { key: 'code', label: t('tools.exit-code.results.code'), mono: true, align: 'right' },
  { key: 'label', label: t('tools.exit-code.results.label') },
  { key: 'type', label: t('tools.exit-code.results.type') },
  { key: 'description', label: t('tools.exit-code.results.description') },
])

/** 表格行。 */
const rows = computed<DataTableRow[]>(() =>
  filtered.value.map((entry) => ({
    key: String(entry.code),
    cells: {
      code: String(entry.code),
      label: entry.label,
      type: t(`tools.exit-code.options.type.${entry.type}`),
      description: entry.description,
    },
    badges: { code: TYPE_LEVELS[entry.type] ?? 'muted' },
  })),
)

const faq = computed<FaqItem[]>(() => tm('tools.exit-code.faq') as FaqItem[])
</script>

<template>
  <ToolPageLayout
    slug="exit-code"
    :name="t('tools.exit-code.name')"
    name-en="Exit Code Lookup"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <SearchField
            v-model="query"
            autofocus
            :label="t('tools.exit-code.fields.query')"
            :placeholder="placeholder"
          />

          <div class="field">
            <span class="field__label">{{ t('tools.exit-code.fields.type') }}</span>
            <FilterChips
              v-model="type"
              :options="typeOptions"
              :aria-label="t('tools.exit-code.fields.type')"
            />
          </div>

          <p class="field__hint">
            {{ t('tools.exit-code.hints.query') }}
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
        :title="t('tools.exit-code.results.title')"
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
            :copy-columns="['code', 'label']"
            :badge-copy="['code']"
            :empty="t('tools.exit-code.results.empty')"
          />
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.exit-code.hints.output') }}
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
