<script setup lang="ts">
/**
 * 证书族材料输入框。
 *
 * 用于证书 / 证书签名请求 / 吊销列表等**证书族**材料的输入：
 * 与 KeyInput 的差别是它**允许一个输入里含多个 PEM 块**（bundle）。
 *
 * 材料内容只存在于组件局部状态；离开页面即丢弃，
 * **不写**任何浏览器存储（AGENTS.md §4.11）。
 */
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { ElOption, ElSelect } from 'element-plus/es/components/select/index'
import 'element-plus/es/components/select/style/css'
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

const value = defineModel<string>({ required: true })
const format = defineModel<string>('format', { required: true })

const props = withDefaults(
  defineProps<{
    /** 可选的格式集合。 */
    formats?: readonly string[]
    /** 无障碍标签。 */
    label?: string
    /** 是否禁用。 */
    disabled?: boolean
    /** 文本框行数。 */
    rows?: number
  }>(),
  {
    formats: () => ['auto', 'pem', 'der', 'base64', 'hex'],
    label: undefined,
    disabled: false,
    rows: 8,
  },
)

const { t } = useI18n()

const FORMAT_LABELS: Record<string, string> = {
  auto: 'tool.key.format.auto',
  pem: 'tool.key.format.pem',
  der: 'tool.key.format.der',
  base64: 'tool.key.format.base64',
  hex: 'tool.key.format.hex',
}

/** 格式名优先取词条，缺失时显示原值。 */
function formatLabel(value: string): string {
  const key = FORMAT_LABELS[value]
  if (!key) {
    return value
  }
  const translated = t(key)
  return translated === key ? value : translated
}

/** 输入里的 PEM 块数；非 PEM 输入返回 0。 */
const blockCount = computed(() => {
  if (!value.value.includes('-----BEGIN ')) {
    return 0
  }
  return value.value.split('-----BEGIN ').length - 1
})

/** 块数提示；没有块时给出格式识别提示。 */
const hint = computed(() => {
  if (blockCount.value > 0) {
    return t('tool.cert.blocks', { count: blockCount.value })
  }
  return t('tool.cert.hint')
})
</script>

<template>
  <div class="cert-input">
    <label class="field">
      <span class="field__label">{{ label ?? t('tool.cert.fields.material') }}</span>
      <ElInput
        v-model="value"
        type="textarea"
        class="tb-mono"
        :rows="props.rows"
        spellcheck="false"
        autocomplete="off"
        :aria-label="label ?? t('tool.cert.fields.material')"
        :disabled="props.disabled"
      />
    </label>

    <label class="field">
      <span class="field__label">{{ t('tool.key.fields.format') }}</span>
      <ElSelect
        v-model="format"
        class="cert-input__format"
        :aria-label="t('tool.key.fields.format')"
        :disabled="props.disabled"
      >
        <ElOption
          v-for="f in props.formats"
          :key="f"
          :label="formatLabel(f)"
          :value="f"
        />
      </ElSelect>
    </label>

    <p class="field__hint">
      {{ hint }}
    </p>
  </div>
</template>

<style scoped>
.cert-input {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-2);
}

.cert-input__format {
  width: 100%;
  max-width: 220px;
}
</style>
