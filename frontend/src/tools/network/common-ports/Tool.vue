<script setup lang="ts">
/**
 * 常用端口查询工具页。
 *
 * 默认展示**全量**常用端口列表；键入搜索词或点选服务分类后即时收敛，
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
import { EXAMPLES } from '@/tools/network/common-ports/example'

/** 一个端口条目。 */
interface CommonPortEntry {
  /** 端口号。 */
  port: number
  /** 服务名称。 */
  label: string
  /** 中文说明。 */
  description: string
  /** 传输协议。 */
  protocols: string[]
  /** 服务分类。 */
  category: string
}

/** 后端 common-ports 工具的响应。 */
interface CommonPortResult {
  /** 匹配条数。 */
  total: number
  /** 数据集条目数。 */
  datasetTotal: number
  /** 命中的条目。 */
  matches: CommonPortEntry[]
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

/** 服务分类枚举（与后端 `CommonPortCategories` 对应）。 */
const CATEGORIES = [
  'web',
  'mail',
  'file',
  'remote',
  'database',
  'messaging',
  'directory',
  'network',
  'devops',
]

/** 分类 → 语义级别；用于分类标签着色。 */
const CATEGORY_LEVELS: Record<string, DataTableCellLevel> = {
  web: 'info',
  mail: 'success',
  file: 'warning',
  remote: 'danger',
  database: 'muted',
  messaging: 'info',
  directory: 'muted',
  network: 'success',
  devops: 'warning',
}

/** 条目是否命中查询词。 */
function matchesQuery(entry: CommonPortEntry, term: string): boolean {
  if (term === '') {
    return true
  }
  const lowered = term.toLowerCase()
  if (String(entry.port) === term) {
    return true
  }
  if (entry.label.toLowerCase().includes(lowered)) {
    return true
  }
  if (entry.description.includes(term)) {
    return true
  }
  return entry.protocols.some((protocol) => protocol.toLowerCase().includes(lowered))
}

/** 全量数据集 + 本地筛选；进页面自动取数，之后键入即筛选。 */
const { state, data, meta, error, filtered, counts, reload } = useLiveFilter<
  CommonPortEntry,
  CommonPortResult
>({
  slug: 'common-ports',
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
      label: t('tools.common-ports.options.category.all'),
      hint: String(data.value?.datasetTotal ?? 0),
    },
  ]
  for (const name of CATEGORIES) {
    options.push({
      value: name,
      label: t(`tools.common-ports.options.category.${name}`),
      hint: String(counts.value[name] ?? 0),
    })
  }
  return options
})

/** 结果文本：命中条目逐行给出端口、服务名与说明。 */
const resultText = computed(() =>
  filtered.value
    .map(
      (entry) =>
        `${entry.port}\t${entry.label}\t${entry.protocols.join('/')}\t${entry.description}`,
    )
    .join('\n'),
)

/** 统计卡片。 */
const stats = computed<StatGridItem[]>(() => {
  const items: StatGridItem[] = [
    { label: t('tools.common-ports.results.total'), value: String(filtered.value.length) },
    { label: t('tools.common-ports.results.datasetTotal'), value: String(data.value?.datasetTotal ?? 0) },
  ]
  if (category.value !== '') {
    items.push({
      label: t('tools.common-ports.results.category'),
      value: t(`tools.common-ports.options.category.${category.value}`),
    })
  }
  items.push({
    label: t('tools.common-ports.results.dataVersion'),
    value: data.value?.dataVersion ?? '-',
  })
  return items
})

/** 表格列定义。 */
const columns = computed<DataTableColumn[]>(() => [
  { key: 'port', label: t('tools.common-ports.results.port'), mono: true, align: 'right' },
  { key: 'label', label: t('tools.common-ports.results.label'), mono: true },
  { key: 'protocols', label: t('tools.common-ports.results.protocols') },
  { key: 'category', label: t('tools.common-ports.results.category') },
  { key: 'description', label: t('tools.common-ports.results.description') },
])

/** 表格行。 */
const rows = computed<DataTableRow[]>(() =>
  filtered.value.map((entry) => ({
    key: String(entry.port),
    cells: {
      port: String(entry.port),
      label: entry.label,
      protocols: entry.protocols.join(' / '),
      category: t(`tools.common-ports.options.category.${entry.category}`),
      description: entry.description,
    },
    badges: { port: CATEGORY_LEVELS[entry.category] ?? 'muted' },
  })),
)

const faq = computed<FaqItem[]>(() => tm('tools.common-ports.faq') as FaqItem[])
</script>

<template>
  <ToolPageLayout
    slug="common-ports"
    :name="t('tools.common-ports.name')"
    name-en="Common Port Lookup"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <SearchField
            v-model="query"
            autofocus
            :label="t('tools.common-ports.fields.query')"
            :placeholder="placeholder"
          />

          <div class="field">
            <span class="field__label">{{ t('tools.common-ports.fields.category') }}</span>
            <FilterChips
              v-model="category"
              :options="categoryOptions"
              :aria-label="t('tools.common-ports.fields.category')"
            />
          </div>

          <p class="field__hint">
            {{ t('tools.common-ports.hints.query') }}
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
        :title="t('tools.common-ports.results.title')"
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
            :copy-columns="['port', 'label']"
            :badge-copy="['port']"
            :empty="t('tools.common-ports.results.empty')"
          />
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.common-ports.hints.output') }}
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
