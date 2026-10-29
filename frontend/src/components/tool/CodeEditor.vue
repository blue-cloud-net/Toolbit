<script setup lang="ts">
/**
 * 代码编辑框。
 *
 * 相比普通 `textarea`，多了行号栏、Tab 缩进与大文本友好���
 * 内容只存在于组件局部状态：**不写**任何浏览器存储。
 */
import { computed, ref } from 'vue'

const model = defineModel<string>({ required: true })

const props = withDefaults(
  defineProps<{
    /** 行数上限；超出后只渲染可见部分。 */
    maxRows?: number
    /** 是否显示行号。 */
    showLineNumbers?: boolean
    /** 是否只读。 */
    readonly?: boolean
    /** 无障碍标签。 */
    label?: string
    /** 占位提示。 */
    placeholder?: string
  }>(),
  { maxRows: 20, showLineNumbers: true, readonly: false, label: undefined, placeholder: undefined },
)

const textarea = ref<HTMLTextAreaElement | null>(null)

/** 行号列表。 */
const lineNumbers = computed(() => {
  const count = model.value.split('\n').length
  return Array.from({ length: count }, (_, i) => i + 1)
})

/** 处理按键：Tab 缩进、Ctrl / ⌘ + Enter 运行。 */
function onKeydown(event: KeyboardEvent): void {
  if (event.key === 'Tab' && !props.readonly) {
    event.preventDefault()
    insertIndent()
    return
  }

  if ((event.ctrlKey || event.metaKey) && event.key === 'Enter') {
    event.preventDefault()
    emit('run')
  }
}

/** 在光标处插入两个空格。 */
function insertIndent(): void {
  const el = textarea.value
  if (!el) {
    return
  }

  const start = el.selectionStart
  const end = el.selectionEnd
  model.value = `${model.value.slice(0, start)}  ${model.value.slice(end)}`

  // 光标位置在 Vue 渲染后才会更新，这里主动修正。
  void Promise.resolve().then(() => {
    el.selectionStart = start + 2
    el.selectionEnd = start + 2
  })
}

const emit = defineEmits<{
  /** Ctrl / ⌘ + Enter。 */
  run: []
}>()
</script>

<template>
  <div class="code-editor">
    <div
      v-if="showLineNumbers"
      class="code-editor__gutter"
      aria-hidden="true"
    >
      <span
        v-for="n in lineNumbers"
        :key="n"
        class="code-editor__line-no"
      >{{ n }}</span>
    </div>

    <textarea
      ref="textarea"
      v-model="model"
      class="code-editor__area tb-mono"
      :rows="maxRows"
      spellcheck="false"
      autocomplete="off"
      autocapitalize="off"
      :readonly="readonly"
      :aria-label="label"
      :placeholder="placeholder"
      @keydown="onKeydown"
    />
  </div>
</template>

<style scoped>
.code-editor {
  display: flex;
  gap: var(--tb-space-2);
  align-items: stretch;
}

.code-editor__gutter {
  display: flex;
  flex-direction: column;
  padding: var(--tb-space-2) var(--tb-space-1);
  font-family: var(--tb-font-mono);
  font-size: var(--tb-font-sm);
  line-height: 1.6;
  color: var(--tb-text-tertiary);
  text-align: right;
  background: var(--tb-bg-sunken);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-control);
  user-select: none;
  overflow: hidden;
}

.code-editor__line-no {
  display: block;
  white-space: nowrap;
}

.code-editor__area {
  flex: 1;
  min-width: 0;
  padding: var(--tb-space-2) var(--tb-space-3);
  font-family: var(--tb-font-mono);
  font-size: var(--tb-font-sm);
  line-height: 1.6;
  color: var(--tb-text);
  background: var(--tb-bg-sunken);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-control);
  resize: vertical;
  tab-size: 2;
}

.code-editor__area:read-only {
  color: var(--tb-text-secondary);
}
</style>
