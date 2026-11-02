<script setup lang="ts">
/**
 * DNS 查询工具页。
 *
 * **多服务器**对比查询：预设解析器多选 + 自定义服务器逐台追加，
 * 结果用页签呈现 —— 第一个页签是「汇总」（成败台数、最快最慢、
 * TTL 区间、记录一致性差异），其后每台服务器一个页签。
 *
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
 */
import { ElAlert } from 'element-plus/es/components/alert/index'
import 'element-plus/es/components/alert/style/css'
import { ElButton } from 'element-plus/es/components/button/index'
import 'element-plus/es/components/button/style/css'
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { ElOption, ElSelect } from 'element-plus/es/components/select/index'
import 'element-plus/es/components/select/style/css'
import { computed, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useRoute, useRouter } from 'vue-router'

import DataTable, { type DataTableColumn, type DataTableRow } from '@/components/tool/DataTable.vue'
import FilterChips from '@/components/tool/FilterChips.vue'
import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import StatGrid, { type StatGridItem } from '@/components/tool/StatGrid.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useToolRun } from '@/composables/useToolRun'
import {
  DEFAULT_PRESET_IDS,
  DNS_PRESET_GROUPS,
  DNS_SERVER_PRESETS,
  MAX_SERVERS,
  findDNSPreset,
  type DNSServerPreset,
} from '@/tools/network/dns-lookup/presets'

import { compareAnswers, serverLabel, type DNSServerResult } from './compare'
import { DEFAULT_EXAMPLE } from './example'

/** 后端 dns-lookup 工具的响应。 */
interface DNSLookupResult {
  host: string
  type: string
  perServer: DNSServerResult[]
}

/** 一台自定义服务器。 */
interface CustomServer {
  /** 内部标识。 */
  id: number
  /** 查询协议。 */
  proto: string
  /** 服务器地址。 */
  addr: string
}

interface FaqItem {
  q: string
  a: string
}

const RECORD_TYPES = [
  'A',
  'AAAA',
  'CNAME',
  'MX',
  'TXT',
  'NS',
  'SOA',
  'SRV',
  'CAA',
  'PTR',
  'ANY',
]
const PROTOS = ['udp', 'tcp', 'dot', 'doh']

const { t, tm } = useI18n()

/** 当前路由；用于读取「DNS 服务器大全」带来的预填参数。 */
const route = useRoute()
const router = useRouter()

const host = ref(DEFAULT_EXAMPLE.host)
const type = ref(DEFAULT_EXAMPLE.type)
const selectedPresets = ref<string[]>([...DEFAULT_PRESET_IDS])
const customServers = ref<CustomServer[]>([])
const customProto = ref('udp')
const customAddr = ref('')

const { state, data, meta, error, run: execute, reset: resetRun } =
  useToolRun<DNSLookupResult>({ slug: 'dns-lookup' })

/** 结果页签选中项；仅组件局部状态，不写存储。 */
const activeView = ref('summary')

const running = computed(() => state.value === 'loading')

/** 当前选中的预设服务器。 */
const activePresets = computed<DNSServerPreset[]>(() =>
  DNS_SERVER_PRESETS.filter((preset) => selectedPresets.value.includes(preset.id)),
)

/** 服务器总数（预设 + 自定义）。 */
const serverCount = computed(() => activePresets.value.length + customServers.value.length)

/** 是否还能再加服务器。 */
const canAddServer = computed(() => serverCount.value < MAX_SERVERS)

/**
 * 地区标签。
 *
 * 用**字面量键**而不是模板字面量：`usage.test.ts` 遇到模板字面量键会**整页跳过**，
 * 这会让本页的其它词条引用全部失去静态校验。
 *
 * 放在 `computed` 里，切换语言时才会重新求值。
 */
const REGION_LABELS = computed<Record<string, string>>(() => ({
  global: t('tools.dns-lookup.options.region.global'),
  switzerland: t('tools.dns-lookup.options.region.switzerland'),
  china: t('tools.dns-lookup.options.region.china'),
  usa: t('tools.dns-lookup.options.region.usa'),
}))

/**
 * 分组标题。
 *
 * 与 `REGION_LABELS` 同理：用**字面量键**而不是模板字面量，
 * 否则 `usage.test.ts` 会整页跳过本页的静态校验。
 */
const GROUP_LABELS = computed<Record<string, string>>(() => ({
  primary: t('tools.dns-lookup.options.group.primary'),
  backup: t('tools.dns-lookup.options.group.backup'),
  tcp: t('tools.dns-lookup.options.group.tcp'),
  dot: t('tools.dns-lookup.options.group.dot'),
  doh: t('tools.dns-lookup.options.group.doh'),
}))

/** 分组后的预设筛选项；已达数量上限时禁用未选中的项。 */
const presetGroups = computed(() =>
  DNS_PRESET_GROUPS.map((group) => ({
    key: group,
    label: GROUP_LABELS.value[group] ?? group,
    options: DNS_SERVER_PRESETS.filter((preset) => preset.group === group).map((preset) => ({
      value: preset.id,
      label: preset.label,
      hint: REGION_LABELS.value[preset.regionKey] ?? preset.regionKey,
      disabled: !selectedPresets.value.includes(preset.id) && !canAddServer.value,
    })),
  })),
)

/** 实际要查询的服务器列表；同一协议 + 同一地址只查一次。 */
const servers = computed(() => {
  const seen = new Set<string>()
  const list: { proto: string; addr: string }[] = []
  const candidates = [
    ...activePresets.value.map((preset) => ({ proto: preset.proto, addr: preset.addr })),
    ...customServers.value
      .filter((item) => item.addr.trim() !== '')
      .map((item) => ({ proto: item.proto, addr: item.addr.trim() })),
  ]
  for (const item of candidates) {
    const key = `${item.proto} ${item.addr}`
    if (seen.has(key)) {
      continue
    }
    seen.add(key)
    list.push(item)
  }
  return list
})

/** 单台服务器页签的标识前缀。 */
const SERVER_VIEW_PREFIX = 'server-'

/** 单台服务器页签的标识。 */
function serverViewId(index: number): string {
  return `${SERVER_VIEW_PREFIX}${index}`
}

/** 结果页签：汇总 + 全部服务器 + 每台服务器。 */
const views = computed(() => {
  const list = [
    { id: 'summary', label: t('tools.dns-lookup.views.summary') },
    { id: 'all', label: t('tools.dns-lookup.views.all') },
  ]
  const perServer = data.value?.perServer ?? []
  perServer.forEach((server, index) => {
    // 标识用**下标**而不是地址：同一个地址既有明文 UDP 又有 TCP 通道，
    // 用地址当标识会让两个页签撞在一起、显示成同一台服务器。
    list.push({ id: serverViewId(index), label: serverLabel(server) })
  })
  return list
})

/** 成功的服务器。 */
const succeeded = computed(() =>
  (data.value?.perServer ?? []).filter((item) => item.error === null),
)

/** 汇总：成败台数、最快最慢、TTL 区间。 */
const summary = computed(() => {
  const servers = data.value?.perServer ?? []
  if (servers.length === 0) {
    return null
  }
  const ok = succeeded.value
  let fastest = ''
  let fastestMs = Number.POSITIVE_INFINITY
  let slowest = ''
  let slowestMs = -1
  for (const item of ok) {
    if (item.latencyMs < fastestMs) {
      fastestMs = item.latencyMs
      fastest = `${item.latencyMs}ms (${serverLabel(item)})`
    }
    if (item.latencyMs > slowestMs) {
      slowestMs = item.latencyMs
      slowest = `${item.latencyMs}ms (${serverLabel(item)})`
    }
  }
  let ttlMin = Number.POSITIVE_INFINITY
  let ttlMax = 0
  let hasTtl = false
  for (const item of ok) {
    for (const answer of item.answers) {
      hasTtl = true
      ttlMin = Math.min(ttlMin, answer.ttl)
      ttlMax = Math.max(ttlMax, answer.ttl)
    }
  }
  return {
    succeeded: ok.length,
    failed: servers.length - ok.length,
    total: servers.length,
    fastest,
    slowest,
    ttlMin: hasTtl ? ttlMin : null,
    ttlMax: hasTtl ? ttlMax : null,
  }
})

/**
 * 记录一致性报告。
 *
 * 结论有三种：`na`（成功返回的服务器不足两台，没有可比对象）、
 * `consistent`（记录集合完全相同）与 `inconsistent`（存在差异）。
 */
const consistency = computed(() => compareAnswers(data.value?.perServer ?? []))

/** 汇总统计卡片。 */
const summaryStats = computed<StatGridItem[]>(() => {
  const value = summary.value
  if (!value) {
    return []
  }
  const items: StatGridItem[] = [
    { label: t('tools.dns-lookup.results.servers'), value: String(value.total) },
    {
      label: t('tool.result.summary.succeeded'),
      value: String(value.succeeded),
      level: 'success',
    },
  ]
  if (value.failed > 0) {
    items.push({
      label: t('tool.result.summary.failed'),
      value: String(value.failed),
      level: 'danger',
    })
  }
  if (value.fastest) {
    items.push({ label: t('tools.dns-lookup.results.fastest'), value: value.fastest })
  }
  if (value.slowest) {
    items.push({ label: t('tool.result.summary.slowest'), value: value.slowest })
  }
  if (value.ttlMin !== null && value.ttlMax !== null) {
    items.push({
      label: t('tools.dns-lookup.results.ttl'),
      value: value.ttlMin === value.ttlMax ? `${value.ttlMin}s` : `${value.ttlMin}–${value.ttlMax}s`,
    })
  }
  // 一致性只有一台成功服务器时**没有可比对象**，
  // 必须报「不适用」而不是「存在差异」。
  if (consistency.value.state === 'na') {
    items.push({
      label: t('tools.dns-lookup.results.consistency'),
      value: t('tools.dns-lookup.results.notApplicable'),
      level: 'muted',
      hint: t('tools.dns-lookup.results.consistencyNa'),
    })
  } else if (consistency.value.state === 'consistent') {
    items.push({
      label: t('tools.dns-lookup.results.consistency'),
      value: t('tools.dns-lookup.results.consistent'),
      level: 'success',
      hint: t('tools.dns-lookup.results.consistencyHint'),
    })
  } else {
    items.push({
      label: t('tools.dns-lookup.results.consistency'),
      value: t('tools.dns-lookup.results.inconsistent'),
      level: 'warning',
      hint: t('tools.dns-lookup.results.consistencyHint'),
    })
  }
  return items
})

/** 差异表格。 */
const differenceRows = computed<DataTableRow[]>(() =>
  consistency.value.differences.map((row) => ({
    key: row.value,
    cells: {
      value: row.value,
      missing: row.missing.join('、'),
    },
  })),
)

/** 差异表格的空态文案。 */
const differenceEmptyText = computed(() =>
  consistency.value.state === 'na'
    ? t('tools.dns-lookup.results.emptyNa')
    : t('tools.dns-lookup.results.empty'),
)

/** 差异表格列定义。 */
const differenceColumns = computed<DataTableColumn[]>(() => [
  { key: 'value', label: t('tools.dns-lookup.results.record'), mono: true },
  { key: 'missing', label: t('tools.dns-lookup.results.missing') },
])

/** 记录表格列定义。 */
const answerColumns = computed<DataTableColumn[]>(() => [
  { key: 'name', label: t('tools.dns-lookup.results.name'), mono: true },
  { key: 'type', label: t('tools.dns-lookup.results.type'), mono: true },
  { key: 'ttl', label: t('tools.dns-lookup.results.ttlColumn'), mono: true, align: 'right' },
  { key: 'value', label: t('tools.dns-lookup.results.value'), mono: true },
])

/** 「全部服务器」表里每台服务器最多预览的记录条数。 */
const ANSWER_PREVIEW_LIMIT = 4

/** 「全部服务器」表的列定义。 */
const allServerColumns = computed<DataTableColumn[]>(() => [
  { key: 'server', label: t('tools.dns-lookup.results.server'), mono: true },
  { key: 'proto', label: t('tools.dns-lookup.results.protocol'), mono: true },
  { key: 'latency', label: t('tools.dns-lookup.results.latency'), mono: true, align: 'right' },
  { key: 'status', label: t('tools.dns-lookup.results.status') },
  { key: 'records', label: t('tools.dns-lookup.results.recordCount'), align: 'right' },
  { key: 'answers', label: t('tools.dns-lookup.results.answers'), mono: true },
])

/** 「全部服务器」表的行：每台服务器的耗时、状态与结果摘要。 */
const allServerRows = computed<DataTableRow[]>(() =>
  (data.value?.perServer ?? []).map((server, index) => {
    const answers = server.answers.map((answer) => `${answer.type} ${answer.value}`)
    const preview = answers.slice(0, ANSWER_PREVIEW_LIMIT).join('、')
    const more =
      answers.length > ANSWER_PREVIEW_LIMIT ? ` …(${answers.length})` : ''
    return {
      // 行标识要唯一：同一地址的 udp / tcp 两行地址文本相同。
      key: `${index}-${server.proto}-${server.server}`,
      level: server.error ? ('danger' as const) : undefined,
      badges: {
        status: server.error ? ('danger' as const) : ('success' as const),
      },
      cells: {
        server: server.server,
        proto: server.proto,
        latency: `${server.latencyMs}ms`,
        status: server.error ? server.error.code : t('tools.dns-lookup.results.ok'),
        records: String(server.answers.length),
        answers: preview + more,
      },
    }
  }),
)

/** 当前页签对应的服务器。 */
const activeServer = computed<DNSServerResult | null>(() => {
  const id = activeView.value
  if (!id.startsWith(SERVER_VIEW_PREFIX)) {
    return null
  }
  const index = Number(id.slice(SERVER_VIEW_PREFIX.length))
  return data.value?.perServer?.[index] ?? null
})

/** 某台服务器的记录表格。 */
function answerRows(server: DNSServerResult): DataTableRow[] {
  return server.answers.map((answer, index) => ({
    key: `${answer.type}-${answer.value}-${index}`,
    cells: {
      name: answer.name,
      type: answer.type,
      ttl: `${answer.ttl}s`,
      value: answer.value,
    },
  }))
}

/** 某台服务器的统计卡片。 */
function serverStats(server: DNSServerResult): StatGridItem[] {
  const items: StatGridItem[] = [
    { label: t('tools.dns-lookup.results.protocol'), value: server.proto },
    { label: t('tools.dns-lookup.results.latency'), value: `${server.latencyMs}ms` },
    { label: t('tools.dns-lookup.results.recordCount'), value: String(server.answers.length) },
  ]
  if (server.error) {
    items.push({
      label: t('tools.dns-lookup.results.status'),
      value: server.error.code,
      level: 'danger',
    })
  }
  return items
}

/** 结果文本：逐台服务器输出。 */
const resultText = computed(() => {
  const servers = data.value?.perServer ?? []
  if (servers.length === 0) {
    return ''
  }
  const lines: string[] = []
  for (const server of servers) {
    lines.push(`# ${server.server} (${server.proto}) ${server.latencyMs}ms`)
    if (server.error) {
      lines.push(`  ! ${server.error.code}: ${server.error.message}`)
    }
    for (const answer of server.answers) {
      lines.push(`  ${answer.name} ${answer.ttl}s ${answer.type} ${answer.value}`)
    }
    lines.push('')
  }
  return lines.join('\n').trimEnd()
})

const faq = computed<FaqItem[]>(() => tm('tools.dns-lookup.faq') as FaqItem[])

/** 应用预设服务器的新选择；超出数量上限时忽略。 */
function updateSelectedPresets(value: string[]): void {
  if (value.length > MAX_SERVERS - customServers.value.length) {
    return
  }
  selectedPresets.value = value
}

/** 追加一台自定义服务器。 */
function addCustomServer(): void {
  const addr = customAddr.value.trim()
  if (addr === '' || !canAddServer.value) {
    return
  }
  customServers.value.push({
    id: Date.now() + customServers.value.length,
    proto: customProto.value,
    addr,
  })
  customAddr.value = ''
}

/** 删除一台自定义服务器。 */
function removeCustomServer(id: number): void {
  const index = customServers.value.findIndex((item) => item.id === id)
  if (index >= 0) {
    customServers.value.splice(index, 1)
  }
}

/** 执行查询。 */
function run(): void {
  const payload: Record<string, unknown> = { host: host.value.trim(), type: type.value }
  const list = servers.value
  if (list.length > 0) {
    payload.servers = list
  }
  // 新一轮的服务器集合可能与上一轮不同：回到汇总页签，
  // 否则会停在一个已经不存在的服务器页签上（页签内容空白）。
  activeView.value = 'summary'
  void execute(payload)
}

/** 重置表单与结果。 */
function reset(): void {
  host.value = DEFAULT_EXAMPLE.host
  type.value = DEFAULT_EXAMPLE.type
  selectedPresets.value = [...DEFAULT_PRESET_IDS]
  customServers.value = []
  customProto.value = 'udp'
  customAddr.value = ''
  activeView.value = 'summary'
  resetRun()
}

/**
 * 应用 URL 预填。
 *
 * 「DNS 服务器大全」的行尾「实测」链接会带上 `proto` 与 `addr`；
 * 也支持 `preset` 直接点名某个预设。预填**覆盖**默认预选，
 * 避免与「一次最多 8 台」的上限冲突。
 *
 * 返回是否真的预填了。
 */
function applyPrefill(): boolean {
  const query = route.query
  const presetId = typeof query.preset === 'string' ? query.preset : ''
  if (presetId !== '') {
    const preset = findDNSPreset(presetId)
    if (!preset) {
      return false
    }
    selectedPresets.value = [preset.id]
    customServers.value = []
    return true
  }

  const proto = typeof query.proto === 'string' ? query.proto : ''
  const addr = typeof query.addr === 'string' ? query.addr : ''
  // 协议必须是受支持的取值；非法值直接忽略，由后端校验兼底。
  if (proto === '' || addr === '' || !PROTOS.includes(proto)) {
    return false
  }
  selectedPresets.value = []
  customServers.value = [{ id: 1, proto, addr }]
  customProto.value = proto
  return true
}

// 首次进入即应用预填，保证首帧就是预填后的状态。
const prefilled = applyPrefill()

onMounted(() => {
  if (!prefilled) {
    return
  }
  // 清掉查询参数：刷新页面不应重复套用。
  void router.replace({ name: 'tool', params: { slug: 'dns-lookup' }, query: {} })
})
</script>

<template>
  <ToolPageLayout
    slug="dns-lookup"
    name="DNS 查询"
    name-en="DNS Lookup"
    wide
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.dns-lookup.fields.host') }}</span>
            <ElInput
              v-model="host"
              class="field__input"
              spellcheck="false"
              autocomplete="off"
              @keydown.enter.prevent="run"
            />
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.dns-lookup.fields.type') }}</span>
            <ElSelect
              v-model="type"
              class="field__select"
              :aria-label="t('tools.dns-lookup.fields.type')"
            >
              <ElOption
                v-for="name in RECORD_TYPES"
                :key="name"
                :label="String(name)"
                :value="name"
              />
            </ElSelect>
          </label>
        </template>

        <template #options>
          <div class="field">
            <span class="field__label">{{ t('tools.dns-lookup.fields.presets') }}</span>
            <div
              v-for="group in presetGroups"
              :key="group.key"
              class="presets"
            >
              <span class="presets__title">{{ group.label }}</span>
              <FilterChips
                :selected="selectedPresets"
                :options="group.options"
                :aria-label="group.label"
                @update:selected="updateSelectedPresets"
              />
            </div>
            <p class="field__hint">
              {{ t('tools.dns-lookup.hints.presets') }}
            </p>
          </div>

          <div class="field">
            <span class="field__label">{{ t('tools.dns-lookup.fields.custom') }}</span>
            <div class="custom">
              <ElSelect
                v-model="customProto"
                class="custom__proto"
                :aria-label="t('tools.dns-lookup.fields.proto')"
              >
                <ElOption
                  v-for="name in PROTOS"
                  :key="name"
                  :label="String(name)"
                  :value="name"
                />
              </ElSelect>
              <ElInput
                v-model="customAddr"
                class="custom__addr tb-mono"
                spellcheck="false"
                autocomplete="off"
                placeholder="1.1.1.1:53 / https://dns.example/dns-query"
                :aria-label="t('tools.dns-lookup.fields.addr')"
                @keydown.enter.prevent="addCustomServer"
              />
              <ElButton
                size="small"
                :disabled="customAddr.trim() === '' || !canAddServer"
                @click="addCustomServer"
              >
                {{ t('tools.dns-lookup.actions.add') }}
              </ElButton>
            </div>

            <ul
              v-if="customServers.length > 0"
              class="custom-list"
            >
              <li
                v-for="item in customServers"
                :key="item.id"
                class="custom-list__item"
              >
                <span class="custom-list__proto">{{ item.proto }}</span>
                <span class="custom-list__addr tb-mono">{{ item.addr }}</span>
                <ElButton
                  size="small"
                  text
                  :aria-label="t('tools.dns-lookup.actions.remove')"
                  @click="removeCustomServer(item.id)"
                >
                  ×
                </ElButton>
              </li>
            </ul>

            <p class="field__hint">
              {{ t('tools.dns-lookup.hints.limit', { max: MAX_SERVERS, count: serverCount }) }}
            </p>
          </div>
        </template>

        <template #footer>
          <RunBar
            :running="running"
            :disabled="host.trim() === ''"
            @run="run"
            @reset="reset"
            @example="reset"
          />
        </template>
      </ToolPanel>
    </template>

    <template #result>
      <ResultPanel
        v-model:active-view="activeView"
        :state="state"
        :text="resultText"
        :meta="meta"
        :error="error"
        :title="t('tools.dns-lookup.results.output')"
        :views="views"
        @retry="run"
      >
        <template #view="{ viewId }">
          <template v-if="viewId === 'summary'">
            <StatGrid :items="summaryStats" />

            <DataTable
              class="table"
              :columns="differenceColumns"
              :rows="differenceRows"
              :empty="differenceEmptyText"
            />
          </template>

          <template v-else-if="viewId === 'all'">
            <DataTable
              class="table"
              :columns="allServerColumns"
              :rows="allServerRows"
              :copy-column="'server'"
              :empty="t('tools.dns-lookup.results.noServer')"
            />
          </template>

          <div
            v-else-if="activeServer"
          >
            <StatGrid :items="serverStats(activeServer)" />

            <ElAlert
              v-if="activeServer.error"
              class="alert"
              type="error"
              :closable="false"
              show-icon
            >
              {{ activeServer.error.code }} — {{ activeServer.error.message }}
            </ElAlert>

            <DataTable
              class="table"
              :columns="answerColumns"
              :rows="answerRows(activeServer)"
              :copy-column="'value'"
              :empty="t('tools.dns-lookup.results.noRecord')"
            />
          </div>
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.dns-lookup.description') }}
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
.presets {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-1);
  margin-bottom: var(--tb-space-2);
}

.presets__title {
  font-size: var(--tb-font-2xs);
  color: var(--tb-text-tertiary);
}

.custom {
  display: flex;
  flex-wrap: wrap;
  gap: var(--tb-space-2);
  align-items: center;
}

.custom__proto {
  width: 96px;
}

.custom__addr {
  flex: 1 1 200px;
}

.custom-list {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-1);
  margin: var(--tb-space-2) 0 0;
  padding: 0;
  list-style: none;
}

.custom-list__item {
  display: flex;
  align-items: center;
  gap: var(--tb-space-2);
  padding: var(--tb-space-1) var(--tb-space-2);
  font-size: var(--tb-font-xs);
  background: var(--tb-bg-sunken);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-control);
}

.custom-list__proto {
  flex: none;
  padding: 0 var(--tb-space-2);
  color: var(--tb-primary);
  background: var(--tb-primary-soft);
  border-radius: var(--tb-radius-pill);
}

.custom-list__addr {
  flex: 1 1 auto;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.alert {
  margin-top: var(--tb-space-3);
}

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
  font-weight: 600;
}

.faq__a {
  margin: 0 0 var(--tb-space-2);
  color: var(--tb-text-secondary);
}
</style>
