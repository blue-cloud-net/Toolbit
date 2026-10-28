<script setup lang="ts">
/**
 * 编码选择器。
 *
 * 选项由调用方给出，因为不同工具支持的编码集合不同。
 */
import { useI18n } from 'vue-i18n'

withDefaults(
  defineProps<{
    /** 当前值。 */
    modelValue: string
    /** 可选编码。 */
    options?: readonly string[]
    /** 无障碍标签。 */
    label?: string
    /** 是否禁用。 */
    disabled?: boolean
  }>(),
  { options: () => ['utf8', 'hex', 'base64', 'base64url'], label: undefined, disabled: false },
)

const emit = defineEmits<{
  /** 值变更。 */
  'update:modelValue': [value: string]
}>()

const { t } = useI18n()

const LABEL_KEYS: Record<string, string> = {
  utf8: 'encoding.utf8',
  hex: 'encoding.hex',
  base64: 'encoding.base64',
  base64url: 'encoding.base64url',
}

/** 编码名优先取词条，缺失时显示原值。 */
function optionLabel(value: string): string {
  const key = LABEL_KEYS[value]
  if (!key) {
    return value
  }
  const translated = t(`tool.${key}`)
  return translated === `tool.${key}` ? value : translated
}
</script>

<template>
  <select
    class="encoding-select"
    :aria-label="label ?? t('tool.fields.encoding')"
    :value="modelValue"
    :disabled="disabled"
    @change="emit('update:modelValue', ($event.target as HTMLSelectElement).value)"
  >
    <option
      v-for="opt in options"
      :key="opt"
      :value="opt"
    >
      {{ optionLabel(opt) }}
    </option>
  </select>
</template>

<style scoped>
.encoding-select {
  padding: var(--tb-space-1) var(--tb-space-2);
  font-family: inherit;
  font-size: var(--tb-font-sm);
  color: var(--tb-text);
  background: var(--tb-bg-elevated);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-control);
  cursor: pointer;
}

.encoding-select:disabled {
  cursor: not-allowed;
  opacity: 0.6;
}
</style>
