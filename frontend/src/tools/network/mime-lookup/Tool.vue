<script setup lang="ts">
/**
 * MIME 类型查询工具页。
 *
 * 默认展示**全量**媒体类型列表；输入搜索词或点选顶层类型后即时收敛，
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
import { EXAMPLES } from '@/tools/network/mime-lookup/example'

/** 一个媒体类型条目。 */
interface MimeEntry {
  /** 规范媒体类型。 */
  mimeType: string
  /** 常用扩展名。 */
  extensions: string[]
  /** 顶层类型。 */
  category: string
  /** 是否是文本类媒体类型。 */
  isTextual: boolean
  /** 默认字符集。 */
  charset: string
  /** 用途说明。 */
  description: string
  /** 历史别名与常见变体。 */
  aliases: string[]
}

/** 后端 mime-lookup 工具的响应。 */
interface MimeLookupResult {
  /** 规范化后的查询词。 */
  query: string
  /** 生效的顶层类型过滤。 */
  category: string
  /** 命中的条目。 */
  entries: MimeEntry[]
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

/** 顶层类型枚举（与后端 `MimeCategories` 对应）。 */
const MIME_CATEGORIES = [
  'application',
  'audio',
  'font',
  'image',
  'message',
  'model',
  'multipart',
  'text',
  'video',
]

const { t, tm } = useI18n()

const query = ref('')
const category = ref('')

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<MimeLookupResult>({
  slug: 'mime-lookup',
})

const running = computed(() => state.value === 'loading')

/** 顶层类型筛选项；带各自条目数。 */
const categoryOptions = computed<FilterChipOption[]>(() => {
  const counts = new Map<string, number>()
  for (const entry of data.value?.entries ?? []) {
    counts.set(entry.category, (counts.get(entry.category) ?? 0) + 1)
  }
  const options: FilterChipOption[] = [
    {
      value: '',
      label: t('tools.mime-lookup.options.category.all'),
      hint: String(data.value?.total ?? 0),
    },
  ]
  for (const name of MIME_CATEGORIES) {
    options.push({
      value: name,
      label: t(`tools.mime-lookup.options.category.${name}`),
      hint: String(counts.get(name) ?? 0),
    })
  }
  return options
})

/** 归一化查询词：媒体类型去参数，文件名取扩展名。 */
function normalizeQuery(raw: string): string {
  const value = raw.trim().toLowerCase()
  if (value === '') {
    return ''
  }
  const withoutParams = value.split(';')[0]?.trim() ?? ''
  if (withoutParams.includes('/')) {
    const rest = withoutParams.slice(withoutParams.lastIndexOf('/') + 1)
    if (!rest.includes('.')) {
      return withoutParams
    }
  }
  const file = withoutParams.slice(Math.max(withoutParams.lastIndexOf('/'), withoutParams.lastIndexOf('\\')) + 1)
  const dot = file.lastIndexOf('.')
  return dot >= 0 ? file.slice(dot + 1) : file
}

/** 条目是否命中查询词。 */
function matchesQuery(entry: MimeEntry, term: string): boolean {
  if (term === '') {
    return true
  }
  const haystack = [entry.mimeType, entry.category, entry.description, ...entry.extensions, ...entry.aliases]
  return haystack.some((value) => value.toLowerCase().includes(term))
}

/** 命中档位：数字越小越靠前。 */
function matchRank(entry: MimeEntry, term: string): number {
  if (term === '') {
    return 2
  }
  if (entry.mimeType === term) {
    return 0
  }
  if (entry.aliases.includes(term)) {
    return 0
  }
  if (entry.extensions.includes(term)) {
    return 1
  }
  return 2
}

/** 本地过滤结果：搜索词与顶层类型叠加。 */
const filtered = computed<MimeEntry[]>(() => {
  const entries = data.value?.entries ?? []
  const term = normalizeQuery(query.value)
  const hits = entries.filter((entry) => {
    if (category.value !== '' && entry.category !== category.value) {
      return false
    }
    return matchesQuery(entry, term)
  })
  return hits
    .map((entry, index) => ({ entry, rank: matchRank(entry, term), index }))
    .sort((a, b) => (a.rank === b.rank ? a.index - b.index : a.rank - b.rank))
    .map((item) => item.entry)
})

/** 结果文本：命中条目逐行给出媒体类型与扩展名。 */
const resultText = computed(() =>
  filtered.value
    .map((entry) => {
      const extensions = entry.extensions.length > 0 ? entry.extensions.map((ext) => `.${ext}`).join(' ') : '-'
      return `${entry.mimeType}\t${extensions}\t${entry.description}`
    })
    .join('\n'),
)

/** 统计卡片。 */
const stats = computed<StatGridItem[]>(() => {
  const items: StatGridItem[] = [
    { label: t('tools.mime-lookup.results.total'), value: String(filtered.value.length) },
    { label: t('tools.mime-lookup.results.datasetTotal'), value: String(data.value?.total ?? 0) },
  ]
  if (category.value !== '') {
    items.push({
      label: t('tools.mime-lookup.results.category'),
      value: t(`tools.mime-lookup.options.category.${category.value}`),
    })
  }
  items.push({ label: t('tools.mime-lookup.results.dataVersion'), value: data.value?.dataVersion ?? '-' })
  return items
})

/** 表格列定义。 */
const columns = computed<DataTableColumn[]>(() => [
  { key: 'mimeType', label: t('tools.mime-lookup.results.mimeType'), mono: true },
  { key: 'extensions', label: t('tools.mime-lookup.results.extensions'), mono: true },
  { key: 'category', label: t('tools.mime-lookup.results.category') },
  { key: 'charset', label: t('tools.mime-lookup.results.charset'), mono: true },
  { key: 'description', label: t('tools.mime-lookup.results.description') },
])

/** 表格行。 */
const rows = computed<DataTableRow[]>(() =>
  filtered.value.map((entry) => ({
    key: entry.mimeType,
    cells: {
      mimeType: entry.mimeType,
      extensions: entry.extensions.map((ext) => `.${ext}`).join(' '),
      category: t(`tools.mime-lookup.options.category.${entry.category}`),
      charset: entry.charset || t('tools.mime-lookup.results.none'),
      description: entry.description,
    },
  })),
)

const faq = computed<FaqItem[]>(() => tm('tools.mime-lookup.faq') as FaqItem[])

/** 拉取全量数据集。 */
function loadAll(): void {
  void execute({})
}

// 首屏拉全量；catalog 加载完成后触发一次。
useAutoRun({
  slug: 'mime-lookup',
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
    slug="mime-lookup"
    :name="t('tools.mime-lookup.name')"
    name-en="MIME Type Lookup"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.mime-lookup.fields.query') }}</span>
            <ElInput
              v-model="query"
              class="field__input tb-mono"
              autocomplete="off"
              placeholder="png / image/png / report.pdf"
              :aria-label="t('tools.mime-lookup.fields.query')"
            />
          </label>

          <div class="field">
            <span class="field__label">{{ t('tools.mime-lookup.fields.category') }}</span>
            <FilterChips
              v-model="category"
              :options="categoryOptions"
              :aria-label="t('tools.mime-lookup.fields.category')"
            />
          </div>

          <p class="field__hint">
            {{ t('tools.mime-lookup.hints.query') }}
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
        :title="t('tools.mime-lookup.results.list')"
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
            :copy-column="'mimeType'"
            :empty="t('tools.mime-lookup.results.empty')"
          />
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.mime-lookup.hints.output') }}
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
