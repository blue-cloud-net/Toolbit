<script setup lang="ts">
/**
 * 文本命名风格转换工具页。
 *
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
 */
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import CodeEditor from '@/components/tool/CodeEditor.vue'
import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useAutoRun } from '@/composables/useAutoRun'
import { useToolRun } from '@/composables/useToolRun'
import { EXAMPLES, type TextStyle } from '@/tools/text/text-convert/example'

/** 后端 text-convert 工具的响应。 */
interface TextConvertResult {
  /** 转换结果。 */
  output: string
  /** 实际使用的风格。 */
  style: string
  /** 切分出的词元。 */
  words: string[]
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

/** 受支持的命名风格；顺序与服务端一致。 */
const STYLES: readonly TextStyle[] = [
  'upper',
  'lower',
  'camel',
  'pascal',
  'snake',
  'kebab',
  'constant',
  'title',
  'sentence',
  'dot',
  'path',
]

const input = ref('')
const style = ref<TextStyle | ''>('camel')
const separator = ref('_')

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
  bindInput,
  useAsInput,
} = useToolRun<TextConvertResult>({
  slug: 'text-convert',
})

/** 结果区的「作为新输入」：把结果回填到主输入框（仅内存，不写存储）。 */
bindInput((text) => {
  input.value = text
})

const running = computed(() => state.value === 'loading')

/** 结果文本。 */
const resultText = computed(() => data.value?.output ?? '')

/** 切分出的词元。 */
const words = computed<string[]>(() => data.value?.words ?? [])

/** 是否使用自定义分隔符。 */
const custom = computed(() => style.value === '')

const faq = computed<FaqItem[]>(() => tm('tools.text-convert.faq') as FaqItem[])

/** 风格项的显示文案。 */
function styleLabel(value: TextStyle | ''): string {
  const key = value === '' ? 'custom' : value
  return t(`tools.text-convert.options.style.${key}`)
}

/** 服务端返回的风格名 → 显示文案；未知风格按自定义分隔符处理。 */
function styleText(raw: string): string {
  const matched = STYLES.find((item) => item === raw)
  return matched === undefined ? styleLabel('') : styleLabel(matched)
}

/** 运行当前输入。 */
function run(): void {
  void execute({
    input: { data: input.value },
    style: style.value,
    separator: separator.value,
  })
}

/** 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。 */
useAutoRun({
  slug: 'text-convert',
  run,
  watch: [input, style, separator],
  enabled: () => input.value !== '',
})

/** 重置表单与结果。 */
function reset(): void {
  input.value = ''
  style.value = 'camel'
  separator.value = '_'
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  input.value = example.input
  style.value = example.style
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="text-convert"
    :name="t('tools.text-convert.name')"
    name-en="Text Converter"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.text-convert.fields.input') }}</span>
            <CodeEditor
              v-model="input"
              :label="t('tools.text-convert.fields.input')"
              :max-rows="10"
              @run="run"
            />
          </label>

          <p class="field__hint">
            {{ t('tools.text-convert.hints.input') }}
          </p>
        </template>

        <template #options>
          <div class="field-row">
            <label class="field">
              <span class="field__label">{{ t('tools.text-convert.fields.style') }}</span>
              <select
                v-model="style"
                class="field__select"
              >
                <option
                  v-for="item in STYLES"
                  :key="item"
                  :value="item"
                >
                  {{ styleLabel(item) }}
                </option>
                <option value="">
                  {{ styleLabel('') }}
                </option>
              </select>
            </label>

            <label
              v-if="custom"
              class="field"
            >
              <span class="field__label">{{ t('tools.text-convert.fields.separator') }}</span>
              <input
                v-model="separator"
                type="text"
                class="field__input tb-mono"
                autocomplete="off"
                :aria-label="t('tools.text-convert.fields.separator')"
              >
            </label>
          </div>

          <p class="field__hint">
            {{ t('tools.text-convert.hints.style') }}
          </p>
          <p
            v-if="custom"
            class="field__hint"
          >
            {{ t('tools.text-convert.hints.separator') }}
          </p>
        </template>

        <template #footer>
          <RunBar
            :running="running"
            @run="run"
            @reset="reset"
            @example="fillExample(0)"
          />
        </template>
      </ToolPanel>
    </template>

    <template #result>
      <ResultPanel
        :state="state"
        :text="resultText"
        :meta="meta"
        :error="error"
        :title="t('tools.text-convert.results.output')"
        :use-as-input="useAsInput"
        @retry="run"
      >
        <template
          v-if="data"
          #extra
        >
          <dl class="detail">
            <div class="detail__row">
              <dt>{{ t('tools.text-convert.results.style') }}</dt>
              <dd>{{ styleText(data?.style ?? '') }}</dd>
            </div>
          </dl>

          <div
            v-if="words.length > 0"
            class="words"
          >
            <span class="words__label">
              {{ t('tools.text-convert.results.words') }}
            </span>
            <ul class="words__list">
              <li
                v-for="(word, index) in words"
                :key="index"
                class="words__item tb-mono"
              >
                {{ word }}
              </li>
            </ul>
          </div>
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.text-convert.hints.output') }}
      </p>
    </template>

    <template #faq>
      <dl class="faq">
        <template
          v-for="(item, index) in faq"
          :key="index"
        >
          <dt class="faq__q">
            {{ item.q }}
          </dt>
          <dd class="faq__a">
            {{ item.a }}
          </dd>
        </template>
      </dl>
    </template>
  </ToolPageLayout>
</template>

<style scoped>
.detail {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-1);
  margin: var(--tb-space-3) 0 0;
}

.detail__row {
  display: flex;
  gap: var(--tb-space-2);
  font-size: var(--tb-font-sm);
}

.detail__row dt {
  min-width: 96px;
  color: var(--tb-text-tertiary);
}

.detail__row dd {
  margin: 0;
  word-break: break-all;
}

.words {
  margin-top: var(--tb-space-3);
}

.words__label {
  font-size: var(--tb-font-xs);
  color: var(--tb-text-tertiary);
}

.words__list {
  display: flex;
  flex-wrap: wrap;
  gap: var(--tb-space-1) var(--tb-space-2);
  margin: var(--tb-space-1) 0 0;
  padding: 0;
  list-style: none;
}

.words__item {
  padding: 0 var(--tb-space-2);
  font-size: var(--tb-font-xs);
  color: var(--tb-text-secondary);
  background: var(--tb-bg-sunken);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-pill);
}

.prose {
  margin: 0;
  font-size: var(--tb-font-sm);
  color: var(--tb-text-secondary);
}

.faq {
  margin: 0;
  font-size: var(--tb-font-sm);
}

.faq__q {
  margin-top: var(--tb-space-2);
  font-weight: 600;
  color: var(--tb-text);
}

.faq__a {
  margin: var(--tb-space-1) 0 0;
  color: var(--tb-text-secondary);
}
</style>
