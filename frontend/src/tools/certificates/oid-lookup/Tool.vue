<script setup lang="ts">
/**
 * OID 查询工具页。
 *
 * 左侧是可展开的 OID 树，右侧是选中节点的详情（面包屑、说明、子节点）。
 *
 * 三种交互：
 *   - **树**：点节点名 = 选中并展开；点箭头 = 展开 / 收起（未加载时懒加载）；
 *   - **搜索**：键入即查（防抖 300ms），按 OID、名称与说明匹配；
 *   - **快捷入口**：一键定位到常用分支。
 *
 * 查询词与选中项只存在于组件局部状态；离开页面即丢弃，
 * **不写**任何浏览器存储（AGENTS.md §4.11）。
 */
import { computed, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'

import DataTable, { type DataTableColumn, type DataTableRow } from '@/components/tool/DataTable.vue'
import FilterChips, { type FilterChipOption } from '@/components/tool/FilterChips.vue'
import ResultPanel from '@/components/tool/ResultPanel.vue'
import SearchField from '@/components/tool/SearchField.vue'
import StatGrid, { type StatGridItem } from '@/components/tool/StatGrid.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useToolRun } from '@/composables/useToolRun'
import OidTreeNode from '@/tools/certificates/oid-lookup/OidTreeNode.vue'
import { EXAMPLES } from '@/tools/certificates/oid-lookup/example'
import {
  buildBreadcrumb,
  isAbsoluteOid,
  isValidOid,
  mergeChildList,
  normalizeOid,
  type OidCrumb,
} from '@/tools/certificates/oid-lookup/oidTree'
import type { OidLookupResult, OidNode } from '@/tools/certificates/oid-lookup/types'

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

/** 搜索防抖（毫秒）。 */
const SEARCH_DEBOUNCE_MS = 300

/** 单次加载的子节点条数。 */
const CHILD_PAGE_SIZE = 100

const { t, tm, locale } = useI18n()

const query = ref('')
const selectedOid = ref('')

/** 树节点表：OID → 节点。 */
const treeNodes = ref(new Map<string, OidNode>())
/** 子节点表：父 OID → 子节点 OID 列表（根节点挂在空串下）。 */
const treeChildren = ref(new Map<string, string[]>())
/** 展开态。 */
const expanded = ref(new Set<string>())
/** 各节点是否还有更多子节点。 */
const hasMore = ref(new Map<string, boolean>())
/** 正在加载子节点的 OID。 */
const loadingOid = ref('')
/** 搜索命中的节点。 */
const searchMatches = ref<OidNode[]>([])

/**
 * 两套运行实例。
 *
 * `detail` 驱动右侧详情面板；`aux` 服务于树的懒加载与搜索。
 * 分开是为了**不让**展开一个树节点就把右侧详情面板刷成「加载中」。
 */
const {
  state: detailState,
  data: detailData,
  meta: detailMeta,
  error: detailError,
  run: detailRun,
} = useToolRun<OidLookupResult>({ slug: 'oid-lookup' })
const { state: auxState, data: auxData, run: auxRun } = useToolRun<OidLookupResult>({ slug: 'oid-lookup' })

/** 根节点列表。 */
const rootNodes = computed<OidNode[]>(() => {
  const list = treeChildren.value.get('') ?? []
  return list
    .map((oid) => treeNodes.value.get(oid))
    .filter((node): node is OidNode => node !== undefined)
})

/** 取某节点的子节点。 */
function childrenOf(oid: string): readonly OidNode[] {
  const list = treeChildren.value.get(oid) ?? []
  return list
    .map((child) => treeNodes.value.get(child))
    .filter((node): node is OidNode => node !== undefined)
}

/** 某节点是否展开。 */
function expandedOf(oid: string): boolean {
  return expanded.value.has(oid)
}

/** 某节点是否还有更多子节点。 */
function hasMoreOf(oid: string): boolean {
  return hasMore.value.get(oid) ?? false
}

/** 按 OID 取节点。 */
function nodeByOid(oid: string): OidNode | null {
  return treeNodes.value.get(oid) ?? null
}

/** 记录节点。 */
function remember(node: OidNode): void {
  treeNodes.value.set(node.oid, node)
}

/** 记录一批节点。 */
function rememberAll(nodes: readonly OidNode[]): void {
  for (const node of nodes) {
    remember(node)
  }
}

/** 把一批子节点并进子节点表。 */
function attachChildren(parent: string, nodes: readonly OidNode[]): void {
  rememberAll(nodes)
  const existing = treeChildren.value.get(parent) ?? []
  treeChildren.value.set(parent, mergeChildList(existing, nodes.map((node) => node.oid)))
}

/** 取某节点的祖先链（从根到父节点）。 */
function pathOf(oid: string): OidNode[] {
  const chain: OidNode[] = []
  let current = oid
  while (current.includes('.')) {
    current = current.slice(0, current.lastIndexOf('.'))
    const node = treeNodes.value.get(current)
    if (node) {
      chain.unshift(node)
    }
  }
  return chain
}

/** 当前选中节点。 */
const currentNode = computed<OidNode | null>(() => nodeByOid(selectedOid.value))

/** 当前选中节点的祖先链。 */
const currentPath = computed<OidNode[]>(() => {
  const node = currentNode.value
  return node ? pathOf(node.oid) : []
})

/** 面包屑。 */
const breadcrumb = computed<OidCrumb[]>(() =>
  buildBreadcrumb(currentPath.value, (node) => displayName(node)),
)

/** 按当前语言取名称：中文界面优先中文，英文界面优先官方英文名。 */
function displayName(node: OidNode): string {
  return locale.value === 'zh-CN'
    ? node.nameZh || node.name
    : node.name || node.nameZh
}

/** 按当前语言取说明：英文界面不把中文说明当作英文说明展示。 */
function displayDescription(node: OidNode): string {
  return locale.value === 'zh-CN' ? node.descriptionZh || node.description : node.description
}

/** 状态标签文案。 */
function statusLabel(status: string): string {
  const key = `tools.oid-lookup.options.status.${status}`
  const label = t(key)
  return label === key ? status : label
}

/** 把一次响应并进树里。 */
function applyToTree(result: OidLookupResult): void {
  rememberAll(result.path)
  for (const [parent, children] of Object.entries(result.pathChildren)) {
    attachChildren(parent, children)
  }
  if (result.node) {
    remember(result.node)
  }
  if (result.children.length > 0 || result.matched) {
    attachChildren(result.oid, result.children)
    hasMore.value.set(result.oid, result.hasMore)
  }
}

/** 载入根节点。 */
async function loadRoots(): Promise<void> {
  await auxRun({})
  const result = auxData.value
  if (!result) {
    return
  }
  treeChildren.value.set('', mergeChildList([], result.children.map((node) => node.oid)))
  rememberAll(result.children)
}

/**
 * 选中节点：取详情、铺开路径、展开祖先。
 *
 * 后端在 `pathChildren` 里一次性给出祖先各层的子节点，
 * 因此点一次就能把从根到目标的整条路径铺好，不需要逐层请求。
 */
async function selectNode(node: OidNode): Promise<void> {
  selectedOid.value = node.oid
  await detailRun({ oid: node.oid })
  const result = detailData.value
  if (!result) {
    return
  }
  applyToTree(result)
  // 展开从根到目标节点的整条路径。
  for (const ancestor of result.path) {
    expanded.value.add(ancestor.oid)
  }
  expanded.value.add(result.oid)
}

/** 按 OID 选中节点；节点不在树里时忽略。 */
async function selectByOid(oid: string): Promise<void> {
  const node = nodeByOid(oid)
  if (node) {
    await selectNode(node)
  }
}

/**
 * 定位一个 OID：已知节点直接选中；未收录则展示最近祖先。
 *
 * 用于快捷入口、搜索框粘贴的绝对 OID，以及任何「按 OID 导航」的入口。
 */
async function locateOid(oid: string): Promise<void> {
  const known = nodeByOid(oid)
  if (known) {
    await selectNode(known)
    return
  }
  await detailRun({ oid })
  const result = detailData.value
  if (!result) {
    return
  }
  // 未命中时没有可展示的节点，但仍把路径铺进树里，方便看出它落在哪个分支。
  selectedOid.value = result.matched ? result.oid : ''
  applyToTree(result)
  for (const ancestor of result.path) {
    expanded.value.add(ancestor.oid)
  }
  if (result.matched) {
    expanded.value.add(result.oid)
  }
}

/** 展开 / 收起节点；未加载时先懒加载子节点。 */
async function toggleNode(node: OidNode): Promise<void> {
  if (expanded.value.has(node.oid)) {
    expanded.value.delete(node.oid)
    return
  }
  if (!treeChildren.value.has(node.oid)) {
    loadingOid.value = node.oid
    await auxRun({ oid: node.oid, limit: CHILD_PAGE_SIZE })
    const result = auxData.value
    if (result) {
      attachChildren(result.oid, result.children)
      hasMore.value.set(result.oid, result.hasMore)
    }
    loadingOid.value = ''
  }
  expanded.value.add(node.oid)
}

/** 加载某节点的更多子节点。 */
async function loadMore(node: OidNode): Promise<void> {
  const current = treeChildren.value.get(node.oid)?.length ?? 0
  loadingOid.value = node.oid
  await auxRun({ oid: node.oid, offset: current, limit: CHILD_PAGE_SIZE })
  const result = auxData.value
  if (result) {
    attachChildren(result.oid, result.children)
    hasMore.value.set(result.oid, result.hasMore)
  }
  loadingOid.value = ''
}

/** 搜索防抖句柄。 */
let searchTimer: ReturnType<typeof setTimeout> | null = null

/** 输入变化：绝对 OID 直接导航，其余按名称 / 说明搜索。 */
watch(query, (value) => {
  if (searchTimer !== null) {
    clearTimeout(searchTimer)
    searchTimer = null
  }
  const term = value.trim()
  if (term === '') {
    searchMatches.value = []
    return
  }
  const normalized = normalizeOid(term)
  if (isAbsoluteOid(normalized)) {
    searchMatches.value = []
    searchTimer = setTimeout(() => {
      void locateOid(normalized)
    }, SEARCH_DEBOUNCE_MS)
    return
  }
  searchTimer = setTimeout(() => {
    void auxRun({ query: term }).then(() => {
      const result = auxData.value
      if (!result) {
        return
      }
      searchMatches.value = result.matches.map((match) => match.node)
      // 搜索命中的节点也要进树表，这样点开时详情能立刻显示。
      rememberAll(searchMatches.value)
    })
  }, SEARCH_DEBOUNCE_MS)
})

/** 快捷入口：OID 来自示例表，标签取自词条，保证双语一致。 */
const quickOptions = computed<FilterChipOption[]>(() => {
  const labels = new Map<string, string>()
  for (const item of tm('tools.oid-lookup.examples') as { label: string; value: string }[]) {
    labels.set(item.value, item.label)
  }
  return EXAMPLES.map((example) => ({
    value: example.oid,
    label: labels.get(example.oid) ?? example.oid,
    hint: example.oid,
  }))
})

/** 搜索框占位提示：直接给出可粘贴的示例 OID。 */
const placeholder = EXAMPLES.map((example) => example.oid).join(' / ')

/** 当前选中的快捷入口。 */
const quick = ref('')

/** 点击快捷入口：直接定位到该 OID。 */
async function onQuickChange(value: string): Promise<void> {
  quick.value = value
  const oid = normalizeOid(value)
  if (oid === '' || !isValidOid(oid)) {
    return
  }
  await locateOid(oid)
}

/** 搜索结果表格列。 */
const matchColumns = computed<DataTableColumn[]>(() => [
  { key: 'oid', label: t('tools.oid-lookup.results.oid'), mono: true },
  { key: 'nameZh', label: t('tools.oid-lookup.results.nameZh') },
  { key: 'name', label: t('tools.oid-lookup.results.name') },
  { key: 'description', label: t('tools.oid-lookup.results.description') },
])

/** 搜索结果表格行。 */
const matchRows = computed<DataTableRow[]>(() =>
  searchMatches.value.map((node) => ({
    key: node.oid,
    cells: {
      oid: node.oid,
      nameZh: node.nameZh || '—',
      name: node.name,
      description: node.description || '—',
    },
  })),
)

/** 点击搜索结果：选中该节点。 */
function onMatchClick(row: DataTableRow): void {
  const node = nodeByOid(row.cells.oid ?? '')
  if (node) {
    void selectNode(node)
  }
}

/** 子节点表格列。 */
const childColumns = computed<DataTableColumn[]>(() => [
  { key: 'arc', label: t('tools.oid-lookup.results.arc'), mono: true },
  { key: 'oid', label: t('tools.oid-lookup.results.oid'), mono: true },
  { key: 'nameZh', label: t('tools.oid-lookup.results.nameZh') },
  { key: 'name', label: t('tools.oid-lookup.results.name') },
  { key: 'status', label: t('tools.oid-lookup.results.status') },
])

/** 子节点表格行。 */
const childRows = computed<DataTableRow[]>(() =>
  childrenOf(selectedOid.value).map((node) => ({
    key: node.oid,
    cells: {
      arc: node.arc,
      oid: node.oid,
      nameZh: node.nameZh || '—',
      name: node.name,
      status: statusLabel(node.status),
    },
  })),
)

/** 点击子节点：选中。 */
function onChildClick(row: DataTableRow): void {
  const node = nodeByOid(row.cells.oid ?? '')
  if (node) {
    void selectNode(node)
  }
}

/** 统计卡片。 */
const stats = computed<StatGridItem[]>(() => {
  const node = currentNode.value
  if (!node) {
    return []
  }
  return [
    { label: t('tools.oid-lookup.results.oid'), value: node.oid, wide: true },
    {
      label: t('tools.oid-lookup.results.kind'),
      value: t(`tools.oid-lookup.options.kind.${node.kind}`),
    },
    {
      label: t('tools.oid-lookup.results.childCount'),
      value: String(node.childCount),
    },
    {
      label: t('tools.oid-lookup.results.status'),
      value: statusLabel(node.status),
      level: node.status === 'current' ? 'success' : 'warning',
    },
  ]
})

/** 未命中时的提示。 */
const notFound = computed(() => {
  const result = detailData.value
  if (!result || result.matched || result.mode !== 'node') {
    return null
  }
  return result
})

/** 未命中时缺失的弧后缀。 */
const notFoundSuffix = computed(() => notFound.value?.missingSuffix ?? '')

/** 未命中时的最近祖先。 */
const notFoundNearest = computed(() => notFound.value?.nearestOid ?? '')

/** 可复制的文本：选中节点的 OID、名称与说明。 */
const resultText = computed(() => {
  const node = currentNode.value
  if (!node) {
    return ''
  }
  return [node.oid, displayName(node), displayDescription(node)].join('\t')
})

const faq = computed<FaqItem[]>(() => tm('tools.oid-lookup.faq') as FaqItem[])

onMounted(async () => {
  await loadRoots()
  // 首屏选中 `iso` 根节点：右侧立即有内容，树也只展开一层，不会一上来就铺满整屏。
  const first = nodeByOid('1')
  if (first) {
    await selectNode(first)
  }
})
</script>

<template>
  <ToolPageLayout
    slug="oid-lookup"
    :name="t('tools.oid-lookup.name')"
    name-en="OID Lookup"
    wide
  >
    <template #panel>
      <ToolPanel :input-title="t('tools.oid-lookup.sections.tree')">
        <template #input>
          <SearchField
            v-model="query"
            autofocus
            :label="t('tools.oid-lookup.fields.search')"
            :placeholder="placeholder"
          />

          <div class="field">
            <span class="field__label">{{ t('tools.oid-lookup.fields.quick') }}</span>
            <FilterChips
              :model-value="quick"
              :options="quickOptions"
              :aria-label="t('tools.oid-lookup.fields.quick')"
              @update:model-value="onQuickChange"
            />
          </div>

          <p class="field__hint">
            {{ t('tools.oid-lookup.hints.search') }}
          </p>

          <div
            v-if="searchMatches.length > 0"
            class="oid-matches"
          >
            <h3 class="oid-matches__title">
              {{ t('tools.oid-lookup.results.matches', { count: searchMatches.length }) }}
            </h3>
            <DataTable
              :columns="matchColumns"
              :rows="matchRows"
              :copy-column="'oid'"
              clickable
              max-height="240px"
              @row-click="onMatchClick"
            />
          </div>

          <div class="oid-tree">
            <ul class="oid-tree__list">
              <OidTreeNode
                v-for="node in rootNodes"
                :key="node.oid"
                :node="node"
                :children-of="childrenOf"
                :expanded-of="expandedOf"
                :has-more-of="hasMoreOf"
                :loading-oid="loadingOid"
                :selected-oid="selectedOid"
                @select="selectNode"
                @toggle="toggleNode"
                @load-more="loadMore"
              />
            </ul>
            <p
              v-if="rootNodes.length === 0 && auxState === 'loading'"
              class="oid-tree__hint"
            >
              {{ t('tools.oid-lookup.hints.loading') }}
            </p>
          </div>
        </template>
      </ToolPanel>
    </template>

    <template #result>
      <ResultPanel
        :state="detailState"
        :text="resultText"
        :meta="detailMeta"
        :error="detailError"
        :title="t('tools.oid-lookup.results.detail')"
        hide-text
      >
        <template
          v-if="currentNode"
          #extra
        >
          <nav
            v-if="breadcrumb.length > 0"
            class="oid-crumb"
            :aria-label="t('tools.oid-lookup.results.breadcrumb')"
          >
            <button
              v-for="crumb in breadcrumb"
              :key="crumb.oid"
              type="button"
              class="oid-crumb__item tb-mono"
              @click="selectByOid(crumb.oid)"
            >
              {{ crumb.oid }}
            </button>
          </nav>

          <StatGrid :items="stats" />

          <dl class="oid-detail">
            <dt>{{ t('tools.oid-lookup.results.nameZh') }}</dt>
            <dd>{{ currentNode.nameZh || '—' }}</dd>

            <dt>{{ t('tools.oid-lookup.results.name') }}</dt>
            <dd class="tb-mono">
              {{ currentNode.name }}
            </dd>

            <dt>{{ t('tools.oid-lookup.results.descriptionZh') }}</dt>
            <dd>{{ currentNode.descriptionZh || '—' }}</dd>

            <dt>{{ t('tools.oid-lookup.results.description') }}</dt>
            <dd>{{ currentNode.description || '—' }}</dd>

            <dt>{{ t('tools.oid-lookup.results.source') }}</dt>
            <dd>
              <a
                v-if="currentNode.sourceUrl"
                :href="currentNode.sourceUrl"
                target="_blank"
                rel="noreferrer noopener"
              >{{ currentNode.source }}</a>
              <span v-else>{{ currentNode.source || '—' }}</span>
            </dd>

            <template v-if="currentNode.aliases.length > 0">
              <dt>{{ t('tools.oid-lookup.results.aliases') }}</dt>
              <dd class="tb-mono">
                {{ currentNode.aliases.join(' · ') }}
              </dd>
            </template>
          </dl>

          <DataTable
            v-if="childRows.length > 0"
            class="oid-children"
            :columns="childColumns"
            :rows="childRows"
            :copy-column="'oid'"
            clickable
            @row-click="onChildClick"
          />
        </template>

        <template
          v-else-if="notFound"
          #extra
        >
          <p class="oid-missing">
            {{ t('tools.oid-lookup.hints.notFound', { suffix: notFoundSuffix }) }}
          </p>
          <p
            v-if="notFoundNearest"
            class="oid-missing__nearest"
          >
            {{ t('tools.oid-lookup.results.nearest') }}：
            <button
              type="button"
              class="oid-missing__link tb-mono"
              @click="selectByOid(notFoundNearest)"
            >
              {{ notFoundNearest }}
            </button>
          </p>
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.oid-lookup.hints.usage') }}
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
.oid-tree {
  margin-top: var(--tb-space-3);
}

.oid-tree__list {
  margin: 0;
  padding: 0;
  list-style: none;
  font-size: var(--tb-font-sm);
}

.oid-tree__hint {
  margin: 0;
  font-size: var(--tb-font-xs);
  color: var(--tb-text-tertiary);
}

.oid-matches {
  margin-top: var(--tb-space-3);
}

.oid-matches__title {
  margin: 0 0 var(--tb-space-2);
  font-size: var(--tb-font-xs);
  font-weight: 600;
  color: var(--tb-text-secondary);
}

.oid-crumb {
  display: flex;
  flex-wrap: wrap;
  gap: var(--tb-space-1);
  margin-bottom: var(--tb-space-3);
  font-size: var(--tb-font-xs);
}

.oid-crumb__item {
  padding: 1px var(--tb-space-1);
  font: inherit;
  color: var(--tb-text-secondary);
  cursor: pointer;
  background: var(--tb-bg-sunken);
  border: none;
  border-radius: var(--tb-radius-control);
}

.oid-detail {
  display: grid;
  grid-template-columns: auto 1fr;
  gap: var(--tb-space-2) var(--tb-space-3);
  margin: var(--tb-space-3) 0 0;
  font-size: var(--tb-font-sm);
}

.oid-detail dt {
  color: var(--tb-text-tertiary);
  white-space: nowrap;
}

.oid-detail dd {
  margin: 0;
  overflow-wrap: anywhere;
}

.oid-children {
  margin-top: var(--tb-space-3);
}

.oid-missing {
  margin: 0;
  font-size: var(--tb-font-sm);
}

.oid-missing__nearest {
  margin: var(--tb-space-2) 0 0;
  font-size: var(--tb-font-xs);
  color: var(--tb-text-secondary);
}

.oid-missing__link {
  padding: 0;
  font: inherit;
  color: var(--tb-primary);
  cursor: pointer;
  background: none;
  border: 0;
}

.oid-missing__link:hover {
  text-decoration: underline;
}
</style>
