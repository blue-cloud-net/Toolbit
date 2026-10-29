<script setup lang="ts">
/**
 * 敏感输入框。
 *
 * 密钥、口令、令牌类输入统一走这里：默认遮蔽，
 * 提供「显示 / 隐藏」切换。输入内容只存在于组件局部状态，
 * **不写**任何浏览器存储（AGENTS.md §4.11）。
 */
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import AppIcon from '@/components/common/AppIcon.vue'

const model = defineModel<string>({ required: true })

const props = withDefaults(
  defineProps<{
    /** 无障碍标签。 */
    label?: string
    /** 占位提示。 */
    placeholder?: string
    /** 是否可切换遮蔽；为 false 时始终明文。 */
    toggleable?: boolean
    /** 行数；大于 1 时渲染为多行文本框。 */
    rows?: number
  }>(),
  { label: undefined, placeholder: undefined, toggleable: true, rows: 1 },
)

const { t } = useI18n()
const revealed = ref(false)

/** 当前是否明文显示。 */
const visible = computed(() => !props.toggleable || revealed.value)
</script>

<template>
  <div class="sensitive-input">
    <textarea
      v-if="rows > 1"
      v-model="model"
      class="sensitive-input__area tb-mono"
      :rows="rows"
      :readonly="!visible"
      :aria-label="label"
      :placeholder="placeholder"
      spellcheck="false"
      autocomplete="off"
    />
    <input
      v-else
      v-model="model"
      class="sensitive-input__field tb-mono"
      :type="visible ? 'text' : 'password'"
      :aria-label="label"
      :placeholder="placeholder"
      spellcheck="false"
      autocomplete="off"
    >

    <button
      v-if="toggleable"
      type="button"
      class="sensitive-input__toggle"
      :aria-label="visible ? t('common.actions.hideSecret') : t('common.actions.showSecret')"
      :aria-pressed="revealed"
      :title="visible ? t('common.actions.hideSecret') : t('common.actions.showSecret')"
      @click="revealed = !revealed"
    >
      <AppIcon
        :name="visible ? 'lock' : 'external'"
        :size="15"
      />
    </button>
  </div>
</template>

<style scoped>
.sensitive-input {
  display: flex;
  align-items: stretch;
  gap: var(--tb-space-1);
}

.sensitive-input__field,
.sensitive-input__area {
  flex: 1;
  min-width: 0;
  padding: var(--tb-space-2) var(--tb-space-3);
  font-size: var(--tb-font-sm);
  color: var(--tb-text);
  background: var(--tb-bg-sunken);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-control);
  outline: none;
  transition:
    border-color var(--tb-transition),
    box-shadow var(--tb-transition);
}

.sensitive-input__area {
  resize: vertical;
}

.sensitive-input__field:focus,
.sensitive-input__area:focus {
  border-color: var(--tb-primary);
  box-shadow: 0 0 0 3px color-mix(in srgb, var(--tb-primary) 22%, transparent);
}

.sensitive-input__toggle {
  display: grid;
  place-items: center;
  flex: none;
  width: 36px;
  color: var(--tb-text-tertiary);
  background: var(--tb-bg-sunken);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-control);
  cursor: pointer;
  transition: color var(--tb-transition);
}

.sensitive-input__toggle:hover {
  color: var(--tb-primary);
}
</style>
