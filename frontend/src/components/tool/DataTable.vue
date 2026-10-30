<script setup lang="ts">
/**
 * 数据表格。
 *
 * 结构化结果里的「多记录」展示：一列一字段，一行一条记录。
 * 用于证书链、吊销条目、DNS 记录、状态码全集、算法列表这类列表型结果，
 * 取代此前各工具手写的 `<table class="table">`（重复样式）与
 * `key: value` 文本行（无法逐行复制、语义弱）。
 *
 * **不引入** Element Plus 表格或图表库：自渲染语义 `<table>`，
 * 避免击穿「单工具 chunk < 40 KB gzip」的预算。
 */
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

import CopyButton from '@/components/tool/CopyButton.vue'
import { useCopy } from '@/composables/useCopy'

/** 单元格语义级别。 */
export type DataTableCellLevel = 'success' | 'danger' | 'warning' | 'info' | 'muted'

/** 表格列定义。 */
export interface DataTableColumn {
  /** 列标识；取值为 `row.cells[key]`。 */
  key: string
  /** 列标题（已本地化）。 */
  label: string
  /** 是否用等宽字体（地址、指纹、摘要、媒体类型等）。 */
  mono?: boolean
  /** 该列默认的语义级别。 */
  level?: DataTableCellLevel
  /** 对齐方式；默认左对齐。 */
  align?: 'left' | 'center' | 'right'
}

/** 表格行定义。 */
export interface DataTableRow {
  /** 行标识；缺失时回落到行下标。 */
  key?: string
  /** 行级语义级别（如失败行整行标红）。 */
  level?: DataTableCellLevel
  /** 按列覆盖语义级别；只给这些列的文字着色，其余列不受行级 `level` 影响时为默认色。 */
  cellLevels?: Record<string, DataTableCellLevel>
  /** 需要渲染成彩色标签的列（列键 → 级别）。标签自带底色，行内其余列不动。 */
  badges?: Record<string, DataTableCellLevel>
  /** 各列文本；缺失列渲染为空。 */
  cells: Record<string, string>
}

const props = withDefaults(
  defineProps<{
    /** 列定义。 */
    columns: DataTableColumn[]
    /** 行数据。 */
    rows: DataTableRow[]
    /** 是否显示逐行复制按钮。 */
    copyable?: boolean
    /** 复制哪一列的值；缺省复制整行（制表符分隔）。 */
    copyColumn?: string
    /** 需要内嵌复制按钮的多列（与 `copyColumn` 合并）。 */
    copyColumns?: string[]
    /** 徽标列：点击徽标即复制该单元格的值，不额外渲染按钮。 */
    badgeCopy?: string[]
    /** 滚动区最大高度；传空串表示不限高。 */
    maxHeight?: string
    /** 无数据时的提示文案。 */
    empty?: string
  }>(),
  {
    copyable: true,
    copyColumn: undefined,
    copyColumns: () => [],
    badgeCopy: () => [],
    maxHeight: '420px',
    empty: '',
  },
)

const { t } = useI18n()
const { copy: copyText } = useCopy()

/** 内嵌复制按钮的列键集合（`copyColumn` 与 `copyColumns` 合并）。 */
const inlineCopyKeys = computed<string[]>(() => {
  const keys = [...props.copyColumns]
  if (props.copyColumn !== undefined && !keys.includes(props.copyColumn)) {
    keys.push(props.copyColumn)
  }
  return keys
})

/** 整行复制文本（制表符分隔）。 */
function rowText(row: DataTableRow): string {
  return props.columns
    .map((column) => row.cells[column.key] ?? '')
    .filter((value) => value !== '')
    .join('\t')
}

/** 单元格文本。 */
function cellText(row: DataTableRow, column: DataTableColumn): string {
  return row.cells[column.key] ?? ''
}

/** 单元格语义级别：行级 > 列级，单元格级最高。 */
function cellLevel(row: DataTableRow, column: DataTableColumn): DataTableCellLevel | undefined {
  return row.cellLevels?.[column.key] ?? row.level ?? column.level
}

/** 该单元格是否要渲染成彩色标签。 */
function badgeLevel(row: DataTableRow, column: DataTableColumn): DataTableCellLevel | undefined {
  return row.badges?.[column.key]
}

/** 是否需要尾部的「整行复制」列；指定了内嵌复制列时不再渲染。 */
const showRowAction = computed(() => props.copyable && inlineCopyKeys.value.length === 0)

/** 该列是否内嵌复制按钮（复制它的列值）。 */
function hasInlineCopy(column: DataTableColumn): boolean {
  return props.copyable && inlineCopyKeys.value.includes(column.key)
}

/** 该列的徽标是否点击即复制。 */
function badgeCopies(column: DataTableColumn): boolean {
  return props.badgeCopy.includes(column.key)
}

/** 点击徽标复制该单元格的值。 */
function onBadgeClick(row: DataTableRow, column: DataTableColumn): void {
  void copyText(cellText(row, column))
}
</script>

<template>
  <div
    v-if="rows.length > 0"
    class="data-table__wrap"
    :style="maxHeight === '' ? undefined : { maxHeight }"
  >
    <table class="data-table">
      <thead>
        <tr>
          <th
            v-for="column in columns"
            :key="column.key"
            scope="col"
            :style="column.align ? { textAlign: column.align } : undefined"
          >
            {{ column.label }}
          </th>
          <th
            v-if="showRowAction"
            scope="col"
            class="data-table__action"
          >
            <span class="tb-sr-only">{{ t('common.actions.copy') }}</span>
          </th>
        </tr>
      </thead>
      <tbody>
        <tr
          v-for="(row, index) in rows"
          :key="row.key ?? index"
        >
          <td
            v-for="column in columns"
            :key="column.key"
            :style="column.align ? { textAlign: column.align } : undefined"
            :class="[
              column.mono ? 'tb-mono' : '',
              cellLevel(row, column) ? `data-table__cell--${cellLevel(row, column)}` : '',
            ]"
          >
            <div
              v-if="hasInlineCopy(column)"
              class="data-table__cell"
            >
              <span class="data-table__value">{{ cellText(row, column) }}</span>
              <CopyButton
                :text="cellText(row, column)"
                size="small"
              />
            </div>
            <button
              v-else-if="badgeLevel(row, column) && badgeCopies(column)"
              type="button"
              class="data-table__badge data-table__badge--action"
              :class="`data-table__badge--${badgeLevel(row, column)}`"
              :aria-label="t('common.actions.copy')"
              :title="t('common.actions.copy')"
              @click="onBadgeClick(row, column)"
            >
              {{ cellText(row, column) }}
            </button>
            <span
              v-else-if="badgeLevel(row, column)"
              class="data-table__badge"
              :class="`data-table__badge--${badgeLevel(row, column)}`"
            >
              {{ cellText(row, column) }}
            </span>
            <template v-else>
              {{ cellText(row, column) }}
            </template>
          </td>
          <td
            v-if="showRowAction"
            class="data-table__action"
          >
            <CopyButton
              :text="rowText(row)"
              size="small"
            />
          </td>
        </tr>
      </tbody>
    </table>
  </div>
  <p
    v-else
    class="data-table__empty"
  >
    {{ empty }}
  </p>
</template>

<style scoped>
/*
 * 滚动容器。
 *
 * `display: block` **必须**显式声明：UnoCSS 的 `.table` 工具类是
 * `display: table`，而调用方普遍写 `<DataTable class="table">`，
 * 若不显式声明，外层会变成 table 盒子 —— `max-height` 与 `overflow: auto`
 * 随之失效，长表格（如行政区划全量）会撑开整页而不再内部滚动。
 */
.data-table__wrap {
  display: block;
  overflow: auto;
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-control);
}

.data-table {
  width: 100%;
  font-size: var(--tb-font-sm);
  border-collapse: collapse;
}

.data-table th,
.data-table td {
  padding: var(--tb-space-2) var(--tb-space-3);
  text-align: left;
  vertical-align: top;
  border-bottom: 1px solid var(--tb-border);
}

.data-table th {
  position: sticky;
  top: 0;
  z-index: 1;
  font-size: var(--tb-font-xs);
  font-weight: 600;
  color: var(--tb-text-secondary);
  white-space: nowrap;
  background: var(--tb-bg-sunken);
}

.data-table tbody tr:last-child td {
  border-bottom: none;
}

.data-table tbody tr:hover td {
  background: var(--tb-bg-sunken);
}

.data-table__action {
  width: 1%;
  white-space: nowrap;
}

/* 单元格内嵌复制：值占满剩余宽度，按钮紧贴其右侧 */
.data-table__cell {
  display: flex;
  align-items: center;
  gap: var(--tb-space-1);
}

.data-table__value {
  min-width: 0;
  overflow-wrap: anywhere;
}

.data-table__cell--success {
  color: var(--tb-success);
}

.data-table__cell--danger {
  color: var(--tb-danger);
}

.data-table__cell--warning {
  color: var(--tb-warning);
}

.data-table__cell--info {
  color: var(--tb-info);
}

.data-table__cell--muted {
  color: var(--tb-text-tertiary);
}

/* 彩色标签：浅底 + 同色文字 + 同色描边，与 ResultPanel 的结论徽章同一手法。
   颜色只来自语义令牌，不在此处硬编码色值，也不随主色漂移。 */
.data-table__badge {
  display: inline-block;
  padding: 0 var(--tb-space-2);
  font-size: var(--tb-font-xs);
  font-weight: 600;
  line-height: 1.7;
  white-space: nowrap;
  border: 1px solid;
  border-radius: 999px;
}

.data-table__badge--success {
  color: var(--tb-success);
  background: color-mix(in srgb, var(--tb-success) 12%, var(--tb-bg-elevated));
  border-color: color-mix(in srgb, var(--tb-success) 40%, var(--tb-border));
}

.data-table__badge--danger {
  color: var(--tb-danger);
  background: color-mix(in srgb, var(--tb-danger) 12%, var(--tb-bg-elevated));
  border-color: color-mix(in srgb, var(--tb-danger) 40%, var(--tb-border));
}

.data-table__badge--warning {
  color: var(--tb-warning);
  background: color-mix(in srgb, var(--tb-warning) 12%, var(--tb-bg-elevated));
  border-color: color-mix(in srgb, var(--tb-warning) 40%, var(--tb-border));
}

.data-table__badge--info {
  color: var(--tb-info);
  background: color-mix(in srgb, var(--tb-info) 12%, var(--tb-bg-elevated));
  border-color: color-mix(in srgb, var(--tb-info) 40%, var(--tb-border));
}

.data-table__badge--muted {
  color: var(--tb-text-secondary);
  background: var(--tb-bg-sunken);
  border-color: var(--tb-border);
}

/* 可点击徽标：徽标本身就是复制按钮，不再另加按钮 */
.data-table__badge--action {
  font-family: inherit;
  cursor: pointer;
  appearance: none;
}

.data-table__badge--action:hover {
  border-color: currentcolor;
}

.data-table__badge--action:focus-visible {
  outline: 2px solid var(--tb-primary);
  outline-offset: 2px;
}

.data-table__empty {
  margin: 0;
  padding: var(--tb-space-4) 0;
  font-size: var(--tb-font-xs);
  color: var(--tb-text-tertiary);
  text-align: center;
}
</style>
