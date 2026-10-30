<script setup lang="ts">
/**
 * 统计卡片网格。
 *
 * 结构化结果里的「少量关键指标」展示：CIDR 的网段 / 掩码 / 可用数、
 * 证书链的校验结论、DNS 查询的耗时与成败台数等。
 *
 * 每张卡片一个字段；值用等宽字体，语义级别决定颜色。
 * 纯 CSS 网格（`auto-fit`），不引任何图表库。
 */
/** 卡片语义级别。 */
export type StatGridLevel = 'success' | 'danger' | 'warning' | 'info' | 'muted'

/** 一张卡片。 */
export interface StatGridItem {
  /** 指标名（已本地化）。 */
  label: string
  /** 指标值；已格式化的字符串。 */
  value: string
  /** 补充说明；如单位、来源、换算。 */
  hint?: string
  /** 语义级别。 */
  level?: StatGridLevel
  /** 是否用等宽字体；默认 `true`。 */
  mono?: boolean
}

withDefaults(
  defineProps<{
    /** 指标卡片。 */
    items: StatGridItem[]
    /** 每张卡片的最小宽度（CSS 长度）；默认 `9rem`。 */
    minWidth?: string
  }>(),
  { minWidth: '9rem' },
)
</script>

<template>
  <div
    v-if="items.length > 0"
    class="stat-grid"
    :style="{ '--stat-grid-min': minWidth }"
  >
    <div
      v-for="item in items"
      :key="item.label"
      class="stat-grid__card"
      :class="item.level ? `stat-grid__card--${item.level}` : ''"
    >
      <span class="stat-grid__label">{{ item.label }}</span>
      <span
        class="stat-grid__value"
        :class="item.mono === false ? '' : 'tb-mono'"
      >{{ item.value }}</span>
      <span
        v-if="item.hint"
        class="stat-grid__hint"
      >{{ item.hint }}</span>
    </div>
  </div>
</template>

<style scoped>
.stat-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(var(--stat-grid-min), 1fr));
  gap: var(--tb-space-2);
}

.stat-grid__card {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-1);
  min-width: 0;
  padding: var(--tb-space-2) var(--tb-space-3);
  background: var(--tb-bg-sunken);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-control);
  border-left: 3px solid var(--tb-border);
}

.stat-grid__label {
  font-size: var(--tb-font-2xs);
  color: var(--tb-text-tertiary);
}

.stat-grid__value {
  overflow-wrap: anywhere;
  font-size: var(--tb-font-base);
  font-weight: 600;
  color: var(--tb-text);
}

.stat-grid__hint {
  font-size: var(--tb-font-2xs);
  color: var(--tb-text-tertiary);
}

.stat-grid__card--success {
  border-left-color: var(--tb-success);
}

.stat-grid__card--success .stat-grid__value {
  color: var(--tb-success);
}

.stat-grid__card--danger {
  border-left-color: var(--tb-danger);
}

.stat-grid__card--danger .stat-grid__value {
  color: var(--tb-danger);
}

.stat-grid__card--warning {
  border-left-color: var(--tb-warning);
}

.stat-grid__card--warning .stat-grid__value {
  color: var(--tb-warning);
}

.stat-grid__card--info {
  border-left-color: var(--tb-info);
}

.stat-grid__card--info .stat-grid__value {
  color: var(--tb-info);
}

.stat-grid__card--muted .stat-grid__value {
  color: var(--tb-text-tertiary);
}
</style>
