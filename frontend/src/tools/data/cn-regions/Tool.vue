<script setup lang="ts">
/**
 * 中国行政区划工具页。
 *
 * 默认展示**全量**行政区划（约 3400 条）；键入搜索词、点选层级或限定省份后即时收敛，
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

/** 一个行政区划条目。 */
interface Region {
  /** 六位行政区划代码。 */
  code: string
  /** 名称。 */
  name: string
  /** 层级：province / city / district。 */
  level: string
  /** 上级代码；省级为空。 */
  parent?: string
  /** 所属省级代码。 */
  province: string
}

/** 后端 cn-regions 工具的响应。 */
interface CNRegionsResult {
  /** 省级行政区。 */
  provinces: Region[]
  /** 地级 / 县级行政区。 */
  cities: Region[]
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

/** 层级枚举（与后端 `LevelProvince` / `LevelCity` / `LevelDistrict` 对应）。 */
const LEVELS = ['province', 'city', 'district']

const { t, tm } = useI18n()

const query = ref('')
const province = ref('')
const level = ref('')

/** 条目是否命中关键词（名称或六位代码）。 */
function matchesQuery(region: Region, term: string): boolean {
  if (term === '') {
    return true
  }
  return region.name.includes(term) || region.code.includes(term)
}

/** 全量数据集 + 本地筛选；进页面自动取数，之后键入即筛选。 */
const { state, data, meta, error, entries, filtered, reload } = useLiveFilter<Region, CNRegionsResult>({
  slug: 'cn-regions',
  query,
  // 一次性取回全部层级；层级、省份与关键词筛选都在本地完成。
  payload: () => ({ level: 'all' }),
  select: (result) => [...result.provinces, ...result.cities],
  normalize: (raw) => raw.trim(),
  match: matchesQuery,
  category: (region) => region.level,
  categoryRef: level,
})

/** 省级代码 → 省级名称。 */
const provinceNameByCode = computed(() => {
  const map = new Map<string, string>()
  for (const region of entries.value) {
    if (region.level === 'province') {
      map.set(region.code, region.name)
    }
  }
  return map
})

/** 条目是否属于指定省份。 */
function matchesProvince(region: Region, term: string): boolean {
  if (term === '') {
    return true
  }
  if (region.level === 'province') {
    return region.name.includes(term) || region.code.includes(term)
  }
  const owner = provinceNameByCode.value.get(region.province)
  if (owner === undefined) {
    return false
  }
  return owner.includes(term) || region.province.includes(term)
}

/** 省份限定后的条目（层级筛选之前）。 */
const provinceScoped = computed(() => {
  const queryTerm = query.value.trim()
  const provinceTerm = province.value.trim()
  return entries.value.filter(
    (region) => matchesQuery(region, queryTerm) && matchesProvince(region, provinceTerm),
  )
})

/** 最终可见条目：关键词 + 省份 + 层级。 */
const visible = computed(() => {
  const provinceTerm = province.value.trim()
  return filtered.value.filter((region) => matchesProvince(region, provinceTerm))
})

/** 各层级的条目数（受关键词与省份影响，不受层级自身影响）。 */
const levelCounts = computed(() => {
  const result: Record<string, number> = {}
  for (const region of provinceScoped.value) {
    result[region.level] = (result[region.level] ?? 0) + 1
  }
  return result
})

/** 层级筛选项；带各自条目数。 */
const levelOptions = computed<FilterChipOption[]>(() => {
  const options: FilterChipOption[] = [
    {
      value: '',
      label: t('tools.cn-regions.options.level.all'),
      hint: String(provinceScoped.value.length),
    },
  ]
  for (const name of LEVELS) {
    options.push({
      value: name,
      label: t(`tools.cn-regions.options.level.${name}`),
      hint: String(levelCounts.value[name] ?? 0),
    })
  }
  return options
})

/** 层级显示文案。 */
function levelLabel(levelName: string): string {
  const key = `tools.cn-regions.options.level.${levelName}`
  const translated = t(key)
  return translated === key ? levelName : translated
}

/** 结果文本：命中条目逐行给出代码、名称与层级。 */
const resultText = computed(() =>
  visible.value
    .map((region) => `${region.code}\t${region.name}\t${levelLabel(region.level)}`)
    .join('\n'),
)

/** 统计卡片。 */
const stats = computed<StatGridItem[]>(() => {
  const items: StatGridItem[] = [
    { label: t('tools.cn-regions.results.total'), value: String(visible.value.length) },
    {
      label: t('tools.cn-regions.results.datasetTotal'),
      value: String(entries.value.length),
    },
  ]
  if (level.value !== '') {
    items.push({
      label: t('tools.cn-regions.results.level'),
      value: levelLabel(level.value),
    })
  }
  items.push({
    label: t('tools.cn-regions.results.dataVersion'),
    value: data.value?.dataVersion ?? '-',
  })
  return items
})

/** 表格列定义。 */
const columns = computed<DataTableColumn[]>(() => [
  { key: 'code', label: t('tools.cn-regions.results.code'), mono: true },
  { key: 'name', label: t('tools.cn-regions.results.name') },
  { key: 'level', label: t('tools.cn-regions.results.level') },
  { key: 'province', label: t('tools.cn-regions.results.provinceName') },
  { key: 'parent', label: t('tools.cn-regions.results.parent'), mono: true },
])

/** 层级的标签色：省级 / 地级 / 县级各一色。 */
const LEVEL_LEVELS: Record<string, DataTableCellLevel> = {
  province: 'info',
  city: 'success',
  district: 'muted',
}

/** 表格行。 */
const rows = computed<DataTableRow[]>(() =>
  visible.value.map((region) => ({
    // 直筒子市（如东莞市）在快照里同时有地级与县级两条记录、代码相同，
    // 因此行标识**必须**带上层级才唯一，否则 Vue 的 key 冲突会多渲染出行。
    key: `${region.code}-${region.level}`,
    cells: {
      code: region.code,
      name: region.name,
      level: levelLabel(region.level),
      province:
        region.level === 'province'
          ? region.name
          : (provinceNameByCode.value.get(region.province) ?? region.province),
      parent: region.parent ?? '',
    },
    badges: { code: LEVEL_LEVELS[region.level] ?? 'muted' },
  })),
)

const faq = computed<FaqItem[]>(() => tm('tools.cn-regions.faq') as FaqItem[])
</script>

<template>
  <ToolPageLayout
    slug="cn-regions"
    :name="t('tools.cn-regions.name')"
    name-en="China Administrative Regions"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <SearchField
            v-model="query"
            autofocus
            :label="t('tools.cn-regions.fields.query')"
            placeholder="杭州"
          />

          <SearchField
            v-model="province"
            :label="t('tools.cn-regions.fields.province')"
            placeholder="浙江省"
          />

          <div class="field">
            <span class="field__label">{{ t('tools.cn-regions.fields.level') }}</span>
            <FilterChips
              v-model="level"
              :options="levelOptions"
              :aria-label="t('tools.cn-regions.fields.level')"
            />
          </div>

          <p class="field__hint">
            {{ t('tools.cn-regions.hints.query') }}
          </p>
          <p class="field__hint">
            {{ t('tools.cn-regions.hints.province') }}
          </p>
          <p class="field__hint">
            {{ t('tools.cn-regions.hints.level') }}
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
        :title="t('tools.cn-regions.results.title')"
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
            :copyable="false"
            :badge-copy="['code']"
            :empty="t('tools.cn-regions.results.empty')"
          />
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.cn-regions.hints.output') }}
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
