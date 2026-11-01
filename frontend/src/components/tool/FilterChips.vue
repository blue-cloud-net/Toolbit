<script setup lang="ts">
/**
 * 筛选标签组。
 *
 * 单选：用于「顶层类型」「状态码类别」这类有限枚举的筛选，
 * 选中项高亮，未选中项次要色；重复点击当前项即取消选择。
 *
 * 多选：传入 `selected`（值集合）后启用，用于「DNS 服务器预设」这类
 * 需要同时勾选多个枚举项的场景；点击切换该项的选中状态。
 *
 * 用 `button` + `aria-pressed` 表达选中态，键盘可聚焦。
 */
import { computed } from 'vue'

/** 一个筛选项。 */
export interface FilterChipOption {
  /** 选项值。 */
  value: string
  /** 显示文案（已本地化）。 */
  label: string
  /** 附加说明，如条目数。 */
  hint?: string
  /** 是否禁用该项（如已达数量上限）。 */
  disabled?: boolean
}

const props = withDefaults(
  defineProps<{
    /** 单选模式：当前选中值；空串表示「全部」。 */
    modelValue?: string
    /** 多选模式：当前选中的值集合；传入后启用多选。 */
    selected?: readonly string[]
    /** 选项列表。 */
    options: FilterChipOption[]
    /** 无障碍标签。 */
    ariaLabel?: string
  }>(),
  { modelValue: undefined, selected: undefined, ariaLabel: undefined },
)

const emit = defineEmits<{
  /** 单选模式：选中项变化。 */
  'update:modelValue': [value: string]
  /** 多选模式：选中项集合变化。 */
  'update:selected': [value: string[]]
}>()

/** 是否处于多选模式（由 `selected` 是否传入决定）。 */
const multiple = computed(() => props.selected !== undefined)

/** 某项是否处于选中态。 */
function isActive(value: string): boolean {
  if (multiple.value) {
    return (props.selected ?? []).includes(value)
  }
  return props.modelValue === value
}

/** 选中某项；单选模式下重复点击当前项则取消选择（回到「全部」）。 */
function select(value: string): void {
  if (multiple.value) {
    const current = [...(props.selected ?? [])]
    const index = current.indexOf(value)
    if (index >= 0) {
      current.splice(index, 1)
    } else {
      current.push(value)
    }
    emit('update:selected', current)
    return
  }
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
      :class="{ 'filter-chips__item--active': isActive(option.value) }"
      :aria-pressed="isActive(option.value)"
      :disabled="option.disabled"
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

.filter-chips__item:disabled {
  cursor: not-allowed;
  opacity: 0.45;
}

.filter-chips__item:disabled:hover {
  color: var(--tb-text-secondary);
  border-color: var(--tb-border);
}

.filter-chips__hint {
  font-size: var(--tb-font-2xs);
  color: var(--tb-text-tertiary);
}

.filter-chips__item--active .filter-chips__hint {
  color: var(--tb-primary);
}
</style>
