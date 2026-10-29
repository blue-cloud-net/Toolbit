<script setup lang="ts">
/**
 * 敏感输入框。
 *
 * 密钥、口令、令牌类输入统一走这里：默认遮蔽，
 * 提供「显示 / 隐藏」切换。输入内容只存在于组件局部状态，
 * **不写**任何浏览器存储（AGENTS.md §4.11）。
 */
import { ElButton } from 'element-plus/es/components/button/index'
import 'element-plus/es/components/button/style/css'
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

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
  <!-- 多行：Element Plus 的 textarea 不支持遮蔽，改为「只读 + 显式切换编辑」。 -->
  <div
    v-if="rows > 1"
    class="sensitive-input"
  >
    <ElInput
      v-model="model"
      type="textarea"
      :rows="rows"
      :readonly="!visible"
      :aria-label="label"
      :placeholder="placeholder"
      spellcheck="false"
      autocomplete="off"
    />
    <ElButton
      v-if="toggleable"
      class="sensitive-input__toggle"
      :aria-pressed="revealed"
      :title="visible ? t('common.actions.hideSecret') : t('common.actions.showSecret')"
      @click="revealed = !revealed"
    >
      {{ visible ? t('common.actions.hideSecret') : t('common.actions.showSecret') }}
    </ElButton>
  </div>

  <!-- 单行：直接用 Element Plus 的「显示 / 隐藏密码」。 -->
  <ElInput
    v-else
    v-model="model"
    class="sensitive-input"
    :type="toggleable ? 'password' : 'text'"
    :show-password="toggleable"
    :aria-label="label"
    :placeholder="placeholder"
    spellcheck="false"
    autocomplete="off"
  />
</template>

<style scoped>
.sensitive-input {
  font-family: var(--tb-font-mono);
}

.sensitive-input__toggle {
  align-self: flex-start;
}
</style>
