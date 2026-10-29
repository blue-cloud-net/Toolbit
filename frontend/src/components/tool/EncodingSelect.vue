<script setup lang="ts">
/**
 * 编码选择器。
 *
 * 选项由调用方给出，因为不同工具支持的编码集合不同。
 */
import { ElOption, ElSelect } from 'element-plus/es/components/select/index'
import 'element-plus/es/components/select/style/css'
import { useI18n } from 'vue-i18n'

/** 当前值。 */
const model = defineModel<string>({ required: true })

withDefaults(
  defineProps<{
    /** 可选编码。 */
    options?: readonly string[]
    /** 无障碍标签。 */
    label?: string
    /** 是否禁用。 */
    disabled?: boolean
  }>(),
  { options: () => ['utf8', 'hex', 'base64', 'base64url'], label: undefined, disabled: false },
)

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
  <ElSelect
    v-model="model"
    class="encoding-select"
    :aria-label="label ?? t('tool.fields.encoding')"
    :disabled="disabled"
  >
    <ElOption
      v-for="opt in options"
      :key="opt"
      :label="optionLabel(opt)"
      :value="opt"
    />
  </ElSelect>
</template>

<style scoped>
.encoding-select {
  min-width: 120px;
}
</style>
