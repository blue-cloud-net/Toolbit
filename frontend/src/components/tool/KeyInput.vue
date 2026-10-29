<script setup lang="ts">
/**
 * 密钥输入框。
 *
 * 密钥内容只存在于组件局部状态；离开页面即丢弃，
 * **不写**任何浏览器存储（AGENTS.md §4.11）。
 */
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { ElOption, ElSelect } from 'element-plus/es/components/select/index'
import 'element-plus/es/components/select/style/css'
import { useI18n } from 'vue-i18n'

import SensitiveInput from '@/components/tool/SensitiveInput.vue'

const value = defineModel<string>({ required: true })
const format = defineModel<string>('format', { required: true })
const passphrase = defineModel<string>('passphrase', { default: '' })

const props = withDefaults(
  defineProps<{
    /** 可选的格式集合。 */
    formats?: readonly string[]
    /** 无障碍标签。 */
    label?: string
    /** 是否显示口令输入框。 */
    showPassphrase?: boolean
    /** 是否禁用。 */
    disabled?: boolean
  }>(),
  {
    formats: () => ['auto', 'pem', 'der', 'base64', 'hex'],
    label: undefined,
    showPassphrase: false,
    disabled: false,
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
</script>

<template>
  <div class="key-input">
    <label class="field">
      <span class="field__label">{{ label ?? t('tool.key.fields.key') }}</span>
      <ElInput
        v-model="value"
        type="textarea"
        class="tb-mono"
        :rows="5"
        spellcheck="false"
        autocomplete="off"
        :aria-label="label ?? t('tool.key.fields.key')"
        :disabled="props.disabled"
      />
    </label>

    <label class="field">
      <span class="field__label">{{ t('tool.key.fields.format') }}</span>
      <ElSelect
        v-model="format"
        class="key-input__format"
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

    <label
      v-if="props.showPassphrase"
      class="field"
    >
      <span class="field__label">{{ t('tool.key.fields.passphrase') }}</span>
      <SensitiveInput
        v-model="passphrase"
        :label="t('tool.key.fields.passphrase')"
      />
    </label>

    <p class="field__hint">
      {{ t('tool.key.hints.format') }}
    </p>
  </div>
</template>

<style scoped>
.key-input {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-2);
}

.key-input__format {
  width: 100%;
  max-width: 220px;
}
</style>
