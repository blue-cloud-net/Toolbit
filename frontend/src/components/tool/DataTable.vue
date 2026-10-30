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
import { useI18n } from 'vue-i18n'

import CopyButton from '@/components/tool/CopyButton.vue'

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
    /** 滚动区最大高度；传空串表示不限高。 */
    maxHeight?: string
    /** 无数据时的提示文案。 */
    empty?: string
  }>(),
  {
    copyable: true,
    copyColumn: undefined,
    maxHeight: '420px',
    empty: '',
  },
)

const { t } = useI18n()

/** 某一行的复制文本。 */
function rowText(row: DataTableRow): string {
  if (props.copyColumn !== undefined) {
    return row.cells[props.copyColumn] ?? ''
  }
  return props.columns
    .map((column) => row.cells[column.key] ?? '')
    .filter((value) => value !== '')
    .join('\t')
}

/** 单元格文本。 */
function cellText(row: DataTableRow, column: DataTableColumn): string {
  return row.cells[column.key] ?? ''
}

/** 单元格语义级别：列默认值，行级可覆盖。 */
function cellLevel(row: DataTableRow, column: DataTableColumn): DataTableCellLevel | undefined {
  return row.level ?? column.level
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
            v-if="copyable"
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
            {{ cellText(row, column) }}
          </td>
          <td
            v-if="copyable"
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
.data-table__wrap {
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

.data-table__empty {
  margin: 0;
  padding: var(--tb-space-4) 0;
  font-size: var(--tb-font-xs);
  color: var(--tb-text-tertiary);
  text-align: center;
}
</style>
