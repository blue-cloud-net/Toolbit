<script setup lang="ts">
/**
 * 国家代码工具页。
 *
 * 默认展示**全量**国家与地区列表；键入搜索词或点选首字母后即时收敛，
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
import { EXAMPLES } from '@/tools/data/country-code/example'

/** 一个国家与地区条目。 */
interface Country {
  /** 两位字母代码。 */
  alpha2: string
  /** 三位字母代码。 */
  alpha3: string
  /** 三位数字代码。 */
  numeric: string
  /** 中文名。 */
  nameZh: string
  /** 英文名。 */
  nameEn: string
}

/** 后端 country-code 工具的响应。 */
interface CountryCodeResult {
  /** 匹配的国家与地区。 */
  countries: Country[]
  /** 匹配条数。 */
  total: number
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
const initial = ref('')

/** 条目是否命中查询词。 */
function matchesQuery(country: Country, term: string): boolean {
  if (term === '') {
    return true
  }
  const lowered = term.toLowerCase()
  return (
    country.alpha2.toLowerCase().includes(lowered) ||
    country.alpha3.toLowerCase().includes(lowered) ||
    country.numeric.includes(term) ||
    country.nameZh.includes(term) ||
    country.nameEn.toLowerCase().includes(lowered)
  )
}

/** 命中档位：数字越小越靠前。 */
function matchRank(country: Country, term: string): number {
  if (term === '') {
    return 2
  }
  const lowered = term.toLowerCase()
  if (
    country.alpha2.toLowerCase() === lowered ||
    country.alpha3.toLowerCase() === lowered ||
    country.numeric === term ||
    country.nameZh === term ||
    country.nameEn.toLowerCase() === lowered
  ) {
    return 0
  }
  if (
    country.alpha2.toLowerCase().startsWith(lowered) ||
    country.alpha3.toLowerCase().startsWith(lowered) ||
    country.nameEn.toLowerCase().startsWith(lowered)
  ) {
    return 1
  }
  return 2
}

/** 全量数据集 + 本地筛选；进页面自动取数，之后键入即筛选。 */
const { state, data, meta, error, filtered, counts, reload } = useLiveFilter<
  Country,
  CountryCodeResult
>({
  slug: 'country-code',
  query,
  select: (result) => result.countries,
  normalize: (raw) => raw.trim(),
  match: matchesQuery,
  rank: matchRank,
  category: (country) => country.alpha2.slice(0, 1),
  categoryRef: initial,
})

/** 占位提示：直接用示例查询词。 */
const placeholder = EXAMPLES.map((example) => example.query).join(' / ')

/** 首字母筛选项；带各自条目数。 */
const initialOptions = computed<FilterChipOption[]>(() => {
  const options: FilterChipOption[] = [
    {
      value: '',
      label: t('tools.country-code.options.initial.all'),
      hint: String(data.value?.total ?? 0),
    },
  ]
  for (const letter of Object.keys(counts.value).sort()) {
    options.push({
      value: letter,
      label: letter,
      hint: String(counts.value[letter] ?? 0),
    })
  }
  return options
})

/** 结果文本：命中条目逐行给出代码与名称。 */
const resultText = computed(() =>
  filtered.value
    .map(
      (country) =>
        `${country.alpha2}\t${country.alpha3}\t${country.numeric}\t${country.nameZh}\t${country.nameEn}`,
    )
    .join('\n'),
)

/** 统计卡片。 */
const stats = computed<StatGridItem[]>(() => {
  const items: StatGridItem[] = [
    { label: t('tools.country-code.results.total'), value: String(filtered.value.length) },
    { label: t('tools.country-code.results.datasetTotal'), value: String(data.value?.total ?? 0) },
  ]
  if (initial.value !== '') {
    items.push({
      label: t('tools.country-code.results.initial'),
      value: initial.value,
    })
  }
  items.push({
    label: t('tools.country-code.results.dataVersion'),
    value: data.value?.dataVersion ?? '-',
  })
  return items
})

/** 表格列定义。 */
const columns = computed<DataTableColumn[]>(() => [
  { key: 'alpha2', label: t('tools.country-code.results.alpha2'), mono: true },
  { key: 'alpha3', label: t('tools.country-code.results.alpha3'), mono: true },
  { key: 'numeric', label: t('tools.country-code.results.numeric'), mono: true },
  { key: 'nameZh', label: t('tools.country-code.results.nameZh') },
  { key: 'nameEn', label: t('tools.country-code.results.nameEn') },
])

/** 两位码的标签色：按首字母轮转，用语义令牌区分。 */
const INITIAL_LEVELS: DataTableCellLevel[] = ['info', 'success', 'warning', 'muted']

/** 表格行。 */
const rows = computed<DataTableRow[]>(() =>
  filtered.value.map((country) => ({
    key: country.alpha2,
    cells: {
      alpha2: country.alpha2,
      alpha3: country.alpha3,
      numeric: country.numeric,
      nameZh: country.nameZh,
      nameEn: country.nameEn,
    },
    badges: {
      alpha2: INITIAL_LEVELS[country.alpha2.charCodeAt(0) % INITIAL_LEVELS.length] ?? 'muted',
    },
  })),
)

const faq = computed<FaqItem[]>(() => tm('tools.country-code.faq') as FaqItem[])
</script>

<template>
  <ToolPageLayout
    slug="country-code"
    :name="t('tools.country-code.name')"
    name-en="Country Codes"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <SearchField
            v-model="query"
            autofocus
            :label="t('tools.country-code.fields.query')"
            :placeholder="placeholder"
          />

          <div class="field">
            <span class="field__label">{{ t('tools.country-code.fields.initial') }}</span>
            <FilterChips
              v-model="initial"
              :options="initialOptions"
              :aria-label="t('tools.country-code.fields.initial')"
            />
          </div>

          <p class="field__hint">
            {{ t('tools.country-code.hints.query') }}
          </p>
          <p class="field__hint">
            {{ t('tools.country-code.hints.initial') }}
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
        :title="t('tools.country-code.results.title')"
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
            :copy-columns="['alpha3', 'numeric']"
            :badge-copy="['alpha2']"
            :empty="t('tools.country-code.results.empty')"
          />
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.country-code.hints.output') }}
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
