<script setup lang="ts">
/**
 * DNS 服务器目录工具页。
 *
 * 列出常用公共 DNS 的 **UDP / TCP / DoT / DoH** 端点；
 * 键入即筛选（全部在**前端本地**完成，零延迟），每行地址可单独复制，
 * 也可一键跳到「DNS 查询」实测该端点。
 *
 * 本页**没有运行按钮**：目录是一次性取回的静态快照，键入即筛选。
 *
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
 */
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import DataTable, { type DataTableColumn, type DataTableRow } from '@/components/tool/DataTable.vue'
import FilterChips, { type FilterChipOption } from '@/components/tool/FilterChips.vue'
import ResultPanel from '@/components/tool/ResultPanel.vue'
import SearchField from '@/components/tool/SearchField.vue'
import StatGrid, { type StatGridItem } from '@/components/tool/StatGrid.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useLiveFilter } from '@/composables/useLiveFilter'

import { EXAMPLES } from './example'

/** 一条 DNS 服务器端点记录。 */
interface DNSServerEntry {
  /** 服务商中文名。 */
  provider: string
  /** 服务商英文名。 */
  providerEn: string
  /** 归属地区。 */
  region: string
  /** 地址族。 */
  family: string
  /** 查询协议。 */
  proto: string
  /** 服务器地址；DoH 时是完整 URL。 */
  addr: string
  /** 端点说明。 */
  note: string
}

/** 后端 dns-servers 工具的响应。 */
interface DNSServersResult {
  /** 命中的端点。 */
  servers: DNSServerEntry[]
  /** 命中条数。 */
  total: number
  /** 命中的服务商数。 */
  providers: number
  /** 数据集版本。 */
  dataVersion: string
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

/** 地区枚举（与后端 `DNSServerRegions` 对应）。 */
const REGIONS = ['china', 'asia', 'global']

/** 协议枚举（与后端 `DNSProtos` 对应）。 */
const PROTOS = ['udp', 'tcp', 'dot', 'doh']

/** 地址族枚举（与后端 `DNSServerFamilies` 对应）。 */
const FAMILIES = ['ipv4', 'ipv6', 'domain']

const { t, tm } = useI18n()

const query = ref('')
/** 地区：单选，空串表示全部。 */
const region = ref('')
/** 协议：多选，空数组表示全部。 */
const selectedProtos = ref<string[]>([])
/** 地址族：多选，空数组表示全部。 */
const selectedFamilies = ref<string[]>([])

/** 目录一次性取回；文本筛选在本地完成。 */
const { state, data, meta, error, filtered, reload } = useLiveFilter<
  DNSServerEntry,
  DNSServersResult
>({
  slug: 'dns-servers',
  query,
  select: (result) => result.servers,
  match: matchesQuery,
})

/**
 * 地区标签。
 *
 * 用**字面量键**而不是模板字面量：`usage.test.ts` 遇到模板字面量键会**整页跳过**，
 * 这会让本页的其它词条引用全部失去静态校验。
 */
const REGION_LABELS = computed<Record<string, string>>(() => ({
  china: t('tools.dns-servers.options.region.china'),
  asia: t('tools.dns-servers.options.region.asia'),
  global: t('tools.dns-servers.options.region.global'),
}))

/** 协议标签。 */
const PROTO_LABELS = computed<Record<string, string>>(() => ({
  udp: t('tools.dns-servers.options.proto.udp'),
  tcp: t('tools.dns-servers.options.proto.tcp'),
  dot: t('tools.dns-servers.options.proto.dot'),
  doh: t('tools.dns-servers.options.proto.doh'),
}))

/** 地址族标签。 */
const FAMILY_LABELS = computed<Record<string, string>>(() => ({
  ipv4: t('tools.dns-servers.options.family.ipv4'),
  ipv6: t('tools.dns-servers.options.family.ipv6'),
  domain: t('tools.dns-servers.options.family.domain'),
}))

/** 条目是否命中查询词。 */
function matchesQuery(entry: DNSServerEntry, term: string): boolean {
  if (term === '') {
    return true
  }
  const haystack = [entry.provider, entry.providerEn, entry.addr, entry.note, entry.proto]
  return haystack.some((value) => value.toLowerCase().includes(term))
}

/** 三组筛选叠加后的可见条目。 */
const visible = computed(() =>
  filtered.value.filter((entry) => {
    if (region.value !== '' && entry.region !== region.value) {
      return false
    }
    if (selectedProtos.value.length > 0 && !selectedProtos.value.includes(entry.proto)) {
      return false
    }
    if (selectedFamilies.value.length > 0 && !selectedFamilies.value.includes(entry.family)) {
      return false
    }
    return true
  }),
)

/** 按键分组计数（基于**文本筛选后**的条目）。 */
function countBy(entries: readonly DNSServerEntry[], pick: (entry: DNSServerEntry) => string) {
  const counts: Record<string, number> = {}
  for (const entry of entries) {
    const key = pick(entry)
    counts[key] = (counts[key] ?? 0) + 1
  }
  return counts
}

/** 地区计数。 */
const regionCounts = computed(() => countBy(filtered.value, (entry) => entry.region))

/** 协议计数。 */
const protoCounts = computed(() => countBy(filtered.value, (entry) => entry.proto))

/** 地址族计数。 */
const familyCounts = computed(() => countBy(filtered.value, (entry) => entry.family))

/** 地区筛选项；单选，空串表示「全部」。 */
const regionOptions = computed<FilterChipOption[]>(() => [
  {
    value: '',
    label: t('tools.dns-servers.options.region.all'),
    hint: String(filtered.value.length),
  },
  ...REGIONS.map((key) => ({
    value: key,
    label: REGION_LABELS.value[key] ?? key,
    hint: String(regionCounts.value[key] ?? 0),
  })),
])

/** 协议筛选项；多选。 */
const protoOptions = computed<FilterChipOption[]>(() =>
  PROTOS.map((key) => ({
    value: key,
    label: PROTO_LABELS.value[key] ?? key,
    hint: String(protoCounts.value[key] ?? 0),
  })),
)

/** 地址族筛选项；多选。 */
const familyOptions = computed<FilterChipOption[]>(() =>
  FAMILIES.map((key) => ({
    value: key,
    label: FAMILY_LABELS.value[key] ?? key,
    hint: String(familyCounts.value[key] ?? 0),
  })),
)

/** 可见条目里的服务商数（去重）。 */
const providerCount = computed(() => {
  const seen = new Set<string>()
  for (const entry of visible.value) {
    seen.add(entry.provider)
  }
  return seen.size
})

/** 统计卡片。 */
const stats = computed<StatGridItem[]>(() => {
  const dotCount = visible.value.filter((entry) => entry.proto === 'dot').length
  const dohCount = visible.value.filter((entry) => entry.proto === 'doh').length
  return [
    {
      label: t('tools.dns-servers.results.providers'),
      value: String(providerCount.value),
    },
    {
      label: t('tools.dns-servers.results.endpoints'),
      value: String(visible.value.length),
    },
    {
      label: t('tools.dns-servers.results.encrypted'),
      value: String(dotCount + dohCount),
      hint: t('tools.dns-servers.results.encryptedHint'),
    },
    {
      label: t('tools.dns-servers.results.dataVersion'),
      value: data.value?.dataVersion ?? '-',
    },
  ]
})

/** 表格列定义。 */
const columns = computed<DataTableColumn[]>(() => [
  { key: 'provider', label: t('tools.dns-servers.results.provider') },
  { key: 'providerEn', label: t('tools.dns-servers.results.providerEn'), mono: true },
  { key: 'proto', label: t('tools.dns-servers.results.proto') },
  { key: 'addr', label: t('tools.dns-servers.results.addr'), mono: true },
  { key: 'family', label: t('tools.dns-servers.results.family') },
  { key: 'note', label: t('tools.dns-servers.results.note') },
])

/** 协议的语义级别：明文中性，TLS 警告色，HTTPS 成功色。 */
function protoLevel(proto: string): 'success' | 'warning' | 'info' | 'muted' {
  switch (proto) {
    case 'doh':
      return 'success'
    case 'dot':
      return 'warning'
    case 'tcp':
      return 'info'
    default:
      return 'muted'
  }
}

/** 表格行；每行带「实测」跳转。 */
const rows = computed<DataTableRow[]>(() =>
  visible.value.map((entry) => ({
    key: `${entry.provider}-${entry.proto}-${entry.addr}`,
    badges: { proto: protoLevel(entry.proto) },
    to: {
      name: 'tool',
      params: { slug: 'dns-lookup' },
      query: { proto: entry.proto, addr: entry.addr },
    },
    cells: {
      provider: entry.provider,
      providerEn: entry.providerEn,
      proto: PROTO_LABELS.value[entry.proto] ?? entry.proto,
      addr: entry.addr,
      family: FAMILY_LABELS.value[entry.family] ?? entry.family,
      note: entry.note,
    },
  })),
)

/** 结果文本：可见端点逐行给出（制表符分隔）。 */
const resultText = computed(() =>
  visible.value
    .map(
      (entry) =>
        `${entry.provider}\t${entry.proto}\t${entry.addr}\t${entry.family}\t${entry.note}`,
    )
    .join('\n'),
)

/** 占位提示：直接用示例搜索词。 */
const placeholder = EXAMPLES.map((example) => example.query).join(' / ')

const faq = computed<FaqItem[]>(() => tm('tools.dns-servers.faq') as FaqItem[])
</script>

<template>
  <ToolPageLayout
    slug="dns-servers"
    :name="t('tools.dns-servers.name')"
    name-en="DNS Server Directory"
    wide
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <SearchField
            v-model="query"
            autofocus
            :label="t('tools.dns-servers.fields.query')"
            :placeholder="placeholder"
          />

          <div class="field">
            <span class="field__label">{{ t('tools.dns-servers.fields.region') }}</span>
            <FilterChips
              v-model="region"
              :options="regionOptions"
              :aria-label="t('tools.dns-servers.fields.region')"
            />
          </div>

          <div class="field">
            <span class="field__label">{{ t('tools.dns-servers.fields.proto') }}</span>
            <FilterChips
              :selected="selectedProtos"
              :options="protoOptions"
              :aria-label="t('tools.dns-servers.fields.proto')"
              @update:selected="selectedProtos = $event"
            />
          </div>

          <div class="field">
            <span class="field__label">{{ t('tools.dns-servers.fields.family') }}</span>
            <FilterChips
              :selected="selectedFamilies"
              :options="familyOptions"
              :aria-label="t('tools.dns-servers.fields.family')"
              @update:selected="selectedFamilies = $event"
            />
          </div>

          <p class="field__hint">
            {{ t('tools.dns-servers.hints.query') }}
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
        :title="t('tools.dns-servers.results.list')"
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
            :copy-column="'addr'"
            :action-label="t('tools.dns-servers.actions.lookup')"
            :empty="t('tools.dns-servers.results.empty')"
          />
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.dns-servers.hints.output') }}
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
