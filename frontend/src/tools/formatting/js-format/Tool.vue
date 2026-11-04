<script setup lang="ts">
/**
 * JavaScript 格式化 工具页。
 *
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
 */
import { ElCheckbox } from 'element-plus/es/components/checkbox/index'
import 'element-plus/es/components/checkbox/style/css'
import { ElOption, ElSelect } from 'element-plus/es/components/select/index'
import 'element-plus/es/components/select/style/css'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import CodeEditor from '@/components/tool/CodeEditor.vue'
import ProductText from '@/components/tool/ProductText.vue'
import ResultPanel, { type ResultView } from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import StatGrid from '@/components/tool/StatGrid.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { buildSummary } from '@/components/tool/summary'
import { useAutoRun } from '@/composables/useAutoRun'
import { useToolRun } from '@/composables/useToolRun'
import { EXAMPLES } from '@/tools/formatting/js-format/example'

/** 后端 js-format 工具的响应。 */
interface JsFormatResult {
  output: string
  valid: boolean
  error?: SyntaxErrorInfo
}

/** 语法错误的位置与原因。 */
interface SyntaxErrorInfo {
  line: number
  column: number
  message: string
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

type Action = 'format' | 'minify' | 'validate'

const { t, tm } = useI18n()

const input = ref('')
const action = ref<Action>('format')
const indent = ref(2)
const quotes = ref<'double' | 'single'>('double')
const semi = ref(true)

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
  bindInput,
  useAsInput,
} = useToolRun<JsFormatResult>({
  slug: 'js-format',
})

/** 结果区的「作为新输入」：把结果回填到主输入框（仅内存，不写存储）。 */
bindInput((text) => {
  input.value = text
})

const resultText = computed(() => data.value?.output ?? '')

/** 摘要卡片：3 项关键字段。 */
const stats = computed(() =>
  buildSummary(
    data.value as unknown as Record<string, unknown> | null,
    [
      { key: 'valid', label: t('tools.js-format.results.validity'), format: (v) => (v ? t('tools.js-format.results.valid') : t('tools.js-format.results.invalid')), booleanLevel: { true: 'success', false: 'danger' } },
      { key: 'error', label: t('tools.js-format.results.position'), format: (v) => { const p = v as { line?: number; column?: number } | null; return p && p.line !== undefined ? String(p.line) + ':' + String(p.column) : '' } },
      { key: 'error.message', label: t('tools.js-format.results.reason') },
    ],
    { yes: t('common.states.yes'), no: t('common.states.no') },
  ),
)

/** 结果页签：摘要 / 产物。 */
const views = computed<ResultView[]>(() => [
  { id: 'summary', label: t('tools.js-format.views.summary') },
  { id: 'output', label: t('tools.js-format.views.output') },
])

/** 当前页签；默认停在「摘要」。 */
const activeView = ref('summary')
const running = computed(() => state.value === 'loading')

/** 常用缩进宽度。 */
const INDENT_OPTIONS = [2, 4, 8] as const

const faq = computed<FaqItem[]>(() => tm('tools.js-format.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  void execute({
    input: { data: input.value },
    action: action.value,
    indent: indent.value,
    quotes: quotes.value,
    semi: semi.value,
  })
}

/** 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。 */
useAutoRun({
  slug: 'js-format',
  run,
  watch: [input, action, indent, quotes, semi],
  enabled: () => input.value !== '',
})

/** 重置表单与结果。 */
function reset(): void {
  input.value = ''
  action.value = 'format'
  indent.value = 2
  quotes.value = 'double'
  semi.value = true
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  input.value = example.input
  action.value = example.action
  indent.value = example.indent
  quotes.value = example.quotes
  semi.value = example.semi
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="js-format"
    :name="t('tools.js-format.name')"
    name-en="JavaScript Formatter"
    wide
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.js-format.fields.input') }}</span>
            <CodeEditor
              v-model="input"
              :label="t('tools.js-format.fields.input')"
              :max-rows="24"
              @run="run"
            />
          </label>

          <p class="field__hint">
            {{ t('tools.js-format.hints.input') }}
          </p>
        </template>

        <template #options>
          <div class="field-row">
            <label class="field">
              <span class="field__label">{{ t('tools.js-format.fields.action') }}</span>
              <ElSelect
                v-model="action"
                class="field__select"
                :aria-label="t('tools.js-format.fields.action')"
              >
                <ElOption
                  :label="t('tools.js-format.options.action.format')"
                  value="format"
                />
                <ElOption
                  :label="t('tools.js-format.options.action.minify')"
                  value="minify"
                />
                <ElOption
                  :label="t('tools.js-format.options.action.validate')"
                  value="validate"
                />
              </ElSelect>
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.js-format.fields.indent') }}</span>
              <ElSelect
                v-model="indent"
                class="field__select"
                :aria-label="t('tools.js-format.fields.indent')"
              >
                <ElOption
                  v-for="size in INDENT_OPTIONS"
                  :key="size"
                  :label="String(size)"
                  :value="size"
                />
              </ElSelect>
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.js-format.fields.quotes') }}</span>
              <ElSelect
                v-model="quotes"
                class="field__select"
                :aria-label="t('tools.js-format.fields.quotes')"
              >
                <ElOption
                  :label="t('tools.js-format.options.quotes.double')"
                  value="double"
                />
                <ElOption
                  :label="t('tools.js-format.options.quotes.single')"
                  value="single"
                />
              </ElSelect>
            </label>
          </div>

          <ElCheckbox
            v-model="semi"
            class="field__checkbox"
          >
            {{ t('tools.js-format.fields.semi') }}
          </ElCheckbox>
          <p class="field__hint">
            {{ t('tools.js-format.hints.semi') }}
          </p>
        </template>

        <template #footer>
          <RunBar
            :running="running"
            :examples="EXAMPLES"
            @run="run"
            @reset="reset"
            @fill="fillExample"
          />
        </template>
      </ToolPanel>
    </template>

    <template #result>
      <ResultPanel
        v-model:active-view="activeView"
        :state="state"
        :text="resultText"
        :meta="meta"
        :error="error"
        :views="views"
        :title="t('tools.js-format.results.output')"
        :use-as-input="useAsInput"
        @retry="run"
      >
        <template #view="{ viewId }">
          <template v-if="viewId === 'summary' && data">
            <StatGrid :items="stats" />
          </template>

          <ProductText
            v-else-if="viewId === 'output' && data"
            :label="t('tools.js-format.views.output')"
            :content="String(data.output ?? '')"
          />
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.js-format.hints.output') }}
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
