<script setup lang="ts">
/**
 * 提示列表。
 *
 * 工具结果里的**提示条目**——生成配置的警告、规则集的备注、格式转换的告警——
 * 本质是「一串短文本，逐条可复制」。
 *
 * 这类记录在四个工具里重复出现（`docker-compose-gen` / `nginx-config` /
 * `firewall-rules` / `markdown-convert`），因此共用本组件，
 * 避免各工具各写一份列定义与表格样板。
 *
 * 语义色与表格样式全部来自 `DataTable`，本组件不新增任何颜色。
 */
import DataTable, { type DataTableColumn, type DataTableRow } from '@/components/tool/DataTable.vue'
import { computed } from 'vue'

const props = withDefaults(
  defineProps<{
    /** 提示条目。 */
    items: readonly string[] | null | undefined
    /** 列标题（已本地化）。 */
    label: string
    /** 无数据时的提示文案；传空串则不渲染空状态。 */
    empty?: string
  }>(),
  { empty: '' },
)

/** 列表列定义：单列，等宽字体。 */
const columns = computed<DataTableColumn[]>(() => [{ key: 'note', label: props.label, mono: true }])

/** 列表行：每条提示一行，行键用下标 + 内容避免重复。 */
const rows = computed<DataTableRow[]>(() =>
  (props.items ?? []).map((note, index) => ({
    key: `${index}-${note}`,
    cells: { note },
  })),
)
</script>

<template>
  <DataTable
    v-if="rows.length > 0"
    class="table"
    :columns="columns"
    :rows="rows"
    copy-column="note"
    :empty="empty"
  />
</template>
