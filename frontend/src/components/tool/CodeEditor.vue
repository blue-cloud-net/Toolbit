<script setup lang="ts">
/**
 * 代码编辑框。
 *
 * 相比普通 `textarea`，多了行号栏、Tab 缩进与大文本友好。
 *
 * 行号**不再单独成框**：它只是编辑区左侧的一条窄轨，与输入区**共用同一个外框**，
 * 中间只用一条细分隔线区分；当前行在窄轨里高亮，并随文本滚动同步。
 *
 * 显示行号时输入区**不自动折行**（`wrap="off"`），否则一个逻辑行会摊成多个视觉行，
 * 行号与内容的对应关系就断了；需要软折行的工具可用 `wrap="soft"`，
 * 此时行号自动隐藏。
 *
 * 内容只存在于组件局部状态：**不写**任何浏览器存储。
 */
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'

/** 行号窄轨最多渲染的行数；超出部分不再编号，避免超大输入拖慢渲染。 */
const MAX_GUTTER_LINES = 20000

const model = defineModel<string>({ required: true })

const { t } = useI18n()

const props = withDefaults(
  defineProps<{
    /** 行数上限；超出后只渲染可见部分。 */
    maxRows?: number
    /** 是否显示行号。 */
    showLineNumbers?: boolean
    /**
     * 折行方式。
     *
     * `off` —— 不折行，长行横向滚动（显示行号时的唯一选择）；
     * `soft` —— 按宽度折行，此时行号自动隐藏。
     */
    wrap?: 'off' | 'soft'
    /** 是否只读。 */
    readonly?: boolean
    /** 无障碍标签。 */
    label?: string
    /** 占位提示。 */
    placeholder?: string
    /** 校验结论；由 useValidation 提供。 */
    validation?: { level: 'success' | 'error'; message: string } | null
  }>(),
  {
    maxRows: 20,
    showLineNumbers: true,
    wrap: 'off',
    readonly: false,
    label: undefined,
    placeholder: undefined,
    validation: null,
  },
)

/** 实际是否渲染行号：软折行时行号会失真，一律隐藏。 */
const withLineNumbers = computed(() => props.showLineNumbers && props.wrap !== 'soft')

const textarea = ref<HTMLTextAreaElement | null>(null)
const gutter = ref<HTMLElement | null>(null)

/** 光标位置（1 基）。 */
const cursorLine = ref(1)
const cursorColumn = ref(1)

/** 总行数。 */
const lineCount = computed(() => model.value.split('\n').length)

/** 行号位数，决定窄轨宽度。 */
const gutterDigits = computed(() => String(lineCount.value).length)

/** 行号文本：合成单段文本，避免逐行生成 DOM 节点。 */
const gutterText = computed(() => {
  const count = Math.min(lineCount.value, MAX_GUTTER_LINES)
  const numbers: string[] = []
  for (let i = 1; i <= count; i += 1) {
    numbers.push(String(i))
  }
  return numbers.join('\n')
})

/** 当前行的行号（1 基）。 */
const gutterCurrent = computed(() => String(cursorLine.value))

/** 字符数（Unicode 码点数）。 */
const charCount = computed(() => Array.from(model.value).length)

/** 字节数（UTF-8）。 */
const byteCount = computed(() => new TextEncoder().encode(model.value).length)

/** 根据光标偏移更新行列号。 */
function updateCursor(): void {
  const el = textarea.value
  if (!el) {
    return
  }
  const offset = el.selectionStart ?? 0
  const before = model.value.slice(0, offset)
  const parts = before.split('\n')
  cursorLine.value = parts.length
  cursorColumn.value = (parts[parts.length - 1] ?? '').length + 1
}

/**
 * 行号窄轨跟随文本域滚动。
 * 两者行高一致（同一 `--tb-editor-line-h`），滚动时行号与内容始终对齐。
 */
function syncGutter(): void {
  const el = textarea.value
  const rail = gutter.value
  if (!el || !rail) {
    return
  }
  rail.scrollTop = el.scrollTop
}

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

onMounted(() => {
  syncGutter()
})

watch(model, () => {
  void nextTick(syncGutter)
})

watch(lineCount, (count) => {
  if (cursorLine.value > count) {
    cursorLine.value = count
  }
})
</script>

<template>
  <div class="code-editor">
    <div
      class="code-editor__frame"
      :class="{ 'code-editor__frame--gutter': withLineNumbers }"
      :style="{ '--tb-editor-digits': gutterDigits }"
    >
      <div
        v-if="withLineNumbers"
        ref="gutter"
        class="code-editor__gutter"
        aria-hidden="true"
      >
        <pre class="code-editor__gutter-numbers">{{ gutterText }}</pre>
        <span
          class="code-editor__gutter-current"
          :style="{ '--tb-editor-cursor-line': cursorLine }"
        >{{ gutterCurrent }}</span>
      </div>

      <textarea
        ref="textarea"
        v-model="model"
        class="code-editor__area tb-mono"
        :rows="maxRows"
        :wrap="wrap"
        spellcheck="false"
        autocomplete="off"
        autocapitalize="off"
        :readonly="readonly"
        :aria-label="label"
        :placeholder="placeholder"
        @keydown="onKeydown"
        @keyup="updateCursor"
        @click="updateCursor"
        @select="updateCursor"
        @scroll="syncGutter"
      />
    </div>

    <footer class="code-editor__status">
      <span class="code-editor__status-item">
        {{ t('tool.editor.line', { line: cursorLine, column: cursorColumn }) }}
      </span>
      <span class="code-editor__status-item">
        {{ t('tool.editor.stats', { chars: charCount, bytes: byteCount, lines: lineCount }) }}
      </span>
    </footer>

    <p
      v-if="validation"
      class="code-editor__validation"
      :class="`code-editor__validation--${validation.level}`"
      role="status"
    >
      {{ validation.message }}
    </p>
  </div>
</template>

<style scoped>
.code-editor {
  display: grid;
  gap: var(--tb-space-2);
}

/*
 * 外框：行号窄轨与输入区共用的**同一个**边框。
 * 窄轨是网格的第一列，输入区是第二列，两者之间只有一条细分隔线。
 */
.code-editor__frame {
  --tb-editor-line-h: calc(var(--tb-font-sm) * 1.6);

  display: grid;
  grid-template-columns: auto minmax(0, 1fr);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-control);
  background: var(--tb-bg-sunken);
  overflow: hidden;
}

/* 不显示行号时，窄轨整列不参与布局。 */
.code-editor__frame:not(.code-editor__frame--gutter) {
  grid-template-columns: minmax(0, 1fr);
}

/* 行号窄轨：与输入区同高，宽度随行号位数自适应。 */
.code-editor__gutter {
  position: relative;
  width: calc(
    var(--tb-editor-digits, 1) * 1ch + var(--tb-space-2) * 2 + 1px
  );
  overflow: hidden;
  padding: var(--tb-space-2) var(--tb-space-2);
  font-family: var(--tb-font-mono);
  font-size: var(--tb-font-sm);
  line-height: var(--tb-editor-line-h);
  color: var(--tb-text-tertiary);
  text-align: right;
  background: var(--tb-bg-sunken);
  border-right: 1px solid var(--tb-border);
  user-select: none;
}

/*
 * 行号文本绝对定位：窄轨因此不贡献网格行高，
 * 编辑区高度只由 `rows`（maxRows）决定。
 */
.code-editor__gutter-numbers {
  position: absolute;
  z-index: 1;
  top: var(--tb-space-2);
  right: var(--tb-space-2);
  left: var(--tb-space-2);
  margin: 0;
  font: inherit;
  white-space: pre;
}

/*
 * 当前行高亮：**盖在行号之上**的一层。
 * 它自己带上当前行号，所以行号不会被高亮条吞掉 ——
 * 底色只是行号背后的一块，行号文字画在底色之上（z-index 更高）。
 */
.code-editor__gutter-current {
  position: absolute;
  z-index: 2;
  top: calc(
    var(--tb-space-2) + (var(--tb-editor-cursor-line) - 1) * var(--tb-editor-line-h)
  );
  right: 0;
  left: 0;
  height: var(--tb-editor-line-h);
  padding-right: var(--tb-space-2);
  color: var(--tb-text);
  font-weight: 600;
  line-height: var(--tb-editor-line-h);
  text-align: right;
  background: var(--tb-bg-active);
  pointer-events: none;
}

.code-editor__area {
  display: block;
  width: 100%;
  min-width: 0;
  padding: var(--tb-space-2) var(--tb-space-3);
  font-family: var(--tb-font-mono);
  font-size: var(--tb-font-sm);
  line-height: var(--tb-editor-line-h);
  color: var(--tb-text);
  background: transparent;
  border: none;
  border-radius: 0;
  resize: vertical;
  tab-size: 2;
}

.code-editor__area:read-only {
  color: var(--tb-text-secondary);
}

/*
 * 聚焦：焦点圈画在**整个外框**上（含行号窄轨），
 * 而不是只给输入区画一圈 —— 否则窄轨被排除在外，看着像两个控件。
 * `outline-offset: -2px` 让焦点圈落在边框内侧，不会溢出外框。
 */
.code-editor__frame:focus-within {
  outline: 2px solid var(--tb-primary);
  outline-offset: -2px;
}

/* 行号窄轨的分隔线随焦点一起变色，窄轨才算在焦点范围内。 */
.code-editor__frame:focus-within .code-editor__gutter {
  border-right-color: var(--tb-primary);
}

/* 焦点圈由外框承担，文本域本身不再单独描边。 */
.code-editor__area:focus-visible {
  outline: none;
}

.code-editor__status {
  display: flex;
  justify-content: space-between;
  gap: var(--tb-space-3);
  padding: 0 var(--tb-space-1);
  font-family: var(--tb-font-mono);
  font-size: var(--tb-font-2xs);
  color: var(--tb-text-tertiary);
}

.code-editor__status-item {
  white-space: nowrap;
}

.code-editor__validation {
  margin: 0;
  font-size: var(--tb-font-2xs);
}

.code-editor__validation--success {
  color: var(--tb-success);
}

.code-editor__validation--error {
  color: var(--tb-danger);
}
</style>
