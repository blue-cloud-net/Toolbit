<script setup lang="ts">
/**
 * 筛选标签组。
 *
 * 单选的可点击标签组：用于「顶层类型」「状态码类别」「服务器预设」这类
 * 有限枚举的筛选。选中项高亮，未选中项次要色。
 *
 * 用 `button` + `aria-pressed` 表达选中态，键盘可聚焦。
 */

/** 一个筛选项。 */
export interface FilterChipOption {
  /** 选项值。 */
  value: string
  /** 显示文案（已本地化）。 */
  label: string
  /** 附加说明，如条目数。 */
  hint?: string
}

const props = withDefaults(
  defineProps<{
    /** 当前选中值；空串表示「全部」。 */
    modelValue: string
    /** 选项列表。 */
    options: FilterChipOption[]
    /** 无障碍标签。 */
    ariaLabel?: string
  }>(),
  { ariaLabel: undefined },
)

const emit = defineEmits<{
  /** 选中项变化。 */
  'update:modelValue': [value: string]
}>()

/** 选中某项；重复点击当前项则取消选择（回到「全部」）。 */
function select(value: string): void {
  emit('update:modelValue', props.modelValue === value ? '' : value)
}
</script>

<template>
  <div
    class="filter-chips"
    role="group"
    :aria-label="ariaLabel"
  >
    <button
      v-for="option in options"
      :key="option.value"
      type="button"
      class="filter-chips__item"
      :class="{ 'filter-chips__item--active': modelValue === option.value }"
      :aria-pressed="modelValue === option.value"
      @click="select(option.value)"
    >
      <span class="filter-chips__label">{{ option.label }}</span>
      <span
        v-if="option.hint"
        class="filter-chips__hint"
      >{{ option.hint }}</span>
    </button>
  </div>
</template>

<style scoped>
.filter-chips {
  display: flex;
  flex-wrap: wrap;
  gap: var(--tb-space-1) var(--tb-space-2);
}

.filter-chips__item {
  display: inline-flex;
  align-items: baseline;
  gap: var(--tb-space-1);
  padding: var(--tb-space-1) var(--tb-space-2);
  font-size: var(--tb-font-xs);
  color: var(--tb-text-secondary);
  cursor: pointer;
  background: var(--tb-bg-sunken);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-pill);
}

.filter-chips__item:hover {
  color: var(--tb-primary);
  border-color: var(--tb-primary);
}

.filter-chips__item:focus-visible {
  outline: 2px solid var(--tb-primary);
  outline-offset: 1px;
}

.filter-chips__item--active {
  font-weight: 600;
  color: var(--tb-primary);
  background: var(--tb-primary-soft);
  border-color: var(--tb-primary);
}

.filter-chips__hint {
  font-size: var(--tb-font-2xs);
  color: var(--tb-text-tertiary);
}

.filter-chips__item--active .filter-chips__hint {
  color: var(--tb-primary);
}
</style>
