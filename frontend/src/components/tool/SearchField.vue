<script setup lang="ts">
/**
 * 搜索框：键入即筛选页面的输入控件。
 *
 * 与普通输入框的差别：
 *   - **自带清空按钮**（`clearable`）—— 一键回到全量，无需另设「重置」；
 *   - 可选进入页面即聚焦（`autofocus`），让用户落地就能直接打字；
 *   - 组件自带字段标签与 `<label>` 包裹，内部 `<input>` 通过**隐式标签关联**
 *     获得无障碍名称，调用方**不需要**再传 `aria-label`；
 *   - 查询词只存在于组件局部状态，**不写**任何浏览器存储。
 */
import { ElInput, type InputInstance } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { onMounted, ref } from 'vue'

const props = withDefaults(
  defineProps<{
    /** 当前查询词。 */
    modelValue: string
    /** 字段标签（可见文本，必填）。 */
    label: string
    /** 占位提示。 */
    placeholder?: string
    /** 进入页面是否自动聚焦。 */
    autofocus?: boolean
  }>(),
  { placeholder: undefined, autofocus: false },
)

const emit = defineEmits<{
  /** 查询词变化。 */
  'update:modelValue': [value: string]
}>()

const inputRef = ref<InputInstance | null>(null)

/**
 * 进入页面即聚焦。
 *
 * 不用原生 `autofocus` 属性：Vue 挂载时属性晚于元素插入，浏览器不保证生效，
 * 故显式调用组件暴露的 `focus()`。
 */
onMounted(() => {
  if (props.autofocus) {
    inputRef.value?.focus()
  }
})
</script>

<template>
  <label class="field">
    <span class="field__label">{{ label }}</span>
    <ElInput
      ref="inputRef"
      :model-value="modelValue"
      class="field__input tb-mono"
      clearable
      autocomplete="off"
      :placeholder="placeholder"
      @update:model-value="emit('update:modelValue', $event)"
    />
  </label>
</template>
