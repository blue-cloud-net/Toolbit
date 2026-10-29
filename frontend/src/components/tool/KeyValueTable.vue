<script setup lang="ts">
/**
 * 键值对表格。
 *
 * 结构化结果的通用展示：每行一个字段，可按字段单独复制
 * （分片复制）。
 */
import { ElDescriptions, ElDescriptionsItem } from 'element-plus/es/components/descriptions/index'
import 'element-plus/es/components/descriptions/style/css'
import { useI18n } from 'vue-i18n'

import CopyButton from '@/components/tool/CopyButton.vue'

export interface KeyValueRow {
  /** 字段名。 */
  label: string
  /** 字段值。 */
  value: string
  /** 语义级别；用于给值上色。 */
  level?: 'success' | 'danger' | 'warning' | 'info' | 'muted'
}

withDefaults(
  defineProps<{
    /** 字段行。 */
    rows: KeyValueRow[]
    /** 是否显示逐字段复制按钮。 */
    copyable?: boolean
    /** 无数据时的提示。 */
    empty?: string
  }>(),
  { copyable: true, empty: '' },
)

const { t } = useI18n()
</script>

<template>
  <ElDescriptions
    v-if="rows.length > 0"
    class="kv-table"
    :column="1"
    border
  >
    <ElDescriptionsItem
      v-for="row in rows"
      :key="row.label"
      :label="row.label"
    >
      <span
        class="kv-table__value"
        :class="row.level ? `kv-table__value--${row.level}` : ''"
      >{{ row.value }}</span>
      <CopyButton
        v-if="copyable && row.value !== ''"
        :text="row.value"
        size="small"
      />
    </ElDescriptionsItem>
  </ElDescriptions>
  <p
    v-else
    class="kv-table__empty"
  >
    {{ empty || t('common.states.empty') }}
  </p>
</template>

<style scoped>
.kv-table {
  width: 100%;
}

.kv-table__value {
  margin-right: var(--tb-space-2);
  overflow-wrap: anywhere;
  font-family: var(--tb-font-mono);
  font-size: var(--tb-font-xs);
  color: var(--tb-text);
}

.kv-table__value--success {
  color: var(--tb-success);
}

.kv-table__value--danger {
  color: var(--tb-danger);
}

.kv-table__value--warning {
  color: var(--tb-warning);
}

.kv-table__value--info {
  color: var(--tb-info);
}

.kv-table__value--muted {
  color: var(--tb-text-tertiary);
}

.kv-table__empty {
  margin: 0;
  font-size: var(--tb-font-xs);
  color: var(--tb-text-tertiary);
}
</style>
