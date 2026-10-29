<script setup lang="ts">
/**
 * 键值对表格。
 *
 * 结构化结果的通用展示：每行一个字段，可按字段单独复制
 * （分片复制）。
 */
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
  <div class="kv-table">
    <p
      v-if="rows.length === 0"
      class="kv-table__empty"
    >
      {{ empty || t('common.states.empty') }}
    </p>
    <dl
      v-else
      class="kv-table__list"
    >
      <div
        v-for="row in rows"
        :key="row.label"
        class="kv-table__row"
      >
        <dt class="kv-table__label">
          {{ row.label }}
        </dt>
        <dd
          class="kv-table__value"
          :class="row.level ? `kv-table__value--${row.level}` : ''"
        >
          <span class="kv-table__text">{{ row.value }}</span>
          <CopyButton
            v-if="copyable && row.value !== ''"
            :text="row.value"
            size="small"
          />
        </dd>
      </div>
    </dl>
  </div>
</template>

<style scoped>
.kv-table {
  width: 100%;
}

.kv-table__list {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-1);
  margin: 0;
}

.kv-table__row {
  display: grid;
  grid-template-columns: minmax(96px, 28%) minmax(0, 1fr);
  gap: var(--tb-space-2);
  align-items: baseline;
  padding: var(--tb-space-1) var(--tb-space-2);
  border-radius: var(--tb-radius-control);
}

.kv-table__row:nth-child(odd) {
  background: var(--tb-bg-sunken);
}

.kv-table__label {
  margin: 0;
  font-size: var(--tb-font-xs);
  color: var(--tb-text-tertiary);
}

.kv-table__value {
  display: flex;
  align-items: baseline;
  gap: var(--tb-space-2);
  margin: 0;
  min-width: 0;
}

.kv-table__text {
  flex: 1;
  min-width: 0;
  overflow-wrap: anywhere;
  font-family: var(--tb-font-mono);
  font-size: var(--tb-font-xs);
  color: var(--tb-text);
}

.kv-table__value--success .kv-table__text {
  color: var(--tb-success);
}

.kv-table__value--danger .kv-table__text {
  color: var(--tb-danger);
}

.kv-table__value--warning .kv-table__text {
  color: var(--tb-warning);
}

.kv-table__value--info .kv-table__text {
  color: var(--tb-info);
}

.kv-table__value--muted .kv-table__text {
  color: var(--tb-text-tertiary);
}

.kv-table__empty {
  margin: 0;
  font-size: var(--tb-font-xs);
  color: var(--tb-text-tertiary);
}

@media (max-width: 640px) {
  .kv-table__row {
    grid-template-columns: minmax(0, 1fr);
  }
}
</style>
