<script setup lang="ts">
/**
 * 结构化数据格式转换工具页。
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
import { EXAMPLES } from '@/tools/formatting/format-convert/example'

/** 后端 format-convert 工具的响应。 */
interface FormatConvertResult {
  output: string
  from: string
  to: string
  lossy?: boolean
  lossyReason?: string
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

type DataFormat = 'json' | 'yaml' | 'toml' | 'xml'

const { t, tm } = useI18n()

const input = ref('')
const from = ref<DataFormat>('json')
const to = ref<DataFormat>('yaml')
const indent = ref(2)
const pretty = ref(true)

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
  bindInput,
  useAsInput,
} = useToolRun<FormatConvertResult>({
  slug: 'format-convert',
})

/** 结果区的「作为新输入」：把结果回填到主输入框（仅内存，不写存储）。 */
bindInput((text) => {
  input.value = text
})

const resultText = computed(() => data.value?.output ?? '')

/** 摘要卡片：4 项关键字段。 */
const stats = computed(() =>
  buildSummary(
    data.value as unknown as Record<string, unknown> | null,
    [
      { key: 'from', label: t('tools.format-convert.results.from') },
      { key: 'to', label: t('tools.format-convert.results.to') },
      { key: 'lossy', label: t('tools.format-convert.results.lossy'), format: (v) => (v ? t('tools.format-convert.results.lossyYes') : t('tools.format-convert.results.lossyNo')), booleanLevel: { true: 'warning', false: 'muted' } },
      { key: 'lossyReason', label: t('tools.format-convert.results.lossyReason'), format: (v) => String(v || t('tools.format-convert.results.lossyHint')), wide: true },
    ],
    { yes: t('common.states.yes'), no: t('common.states.no') },
  ),
)

/** 结果页签：摘要 / 产物。 */
const views = computed<ResultView[]>(() => [
  { id: 'summary', label: t('tools.format-convert.views.summary') },
  { id: 'output', label: t('tools.format-convert.views.output') },
])

/** 当前页签；默认停在「摘要」。 */
const activeView = ref('summary')
const running = computed(() => state.value === 'loading')

/** 受支持的格式。 */
const DATA_FORMATS: readonly DataFormat[] = ['json', 'yaml', 'toml', 'xml']

/** 常用缩进宽度。 */
const INDENT_OPTIONS = [2, 4, 8] as const

const faq = computed<FaqItem[]>(() => tm('tools.format-convert.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  void execute({
    input: { data: input.value },
    from: from.value,
    to: to.value,
    indent: indent.value,
    pretty: pretty.value,
  })
}

/** 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。 */
useAutoRun({
  slug: 'format-convert',
  run,
  watch: [input, from, to, indent, pretty],
  enabled: () => input.value !== '',
})

/** 重置表单与结果。 */
function reset(): void {
  input.value = ''
  from.value = 'json'
  to.value = 'yaml'
  indent.value = 2
  pretty.value = true
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  input.value = example.input
  from.value = example.from
  to.value = example.to
  indent.value = example.indent
  pretty.value = example.pretty
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="format-convert"
    :name="t('tools.format-convert.name')"
    name-en="Format Converter"
    wide
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.format-convert.fields.input') }}</span>
            <CodeEditor
              v-model="input"
              :label="t('tools.format-convert.fields.input')"
              :max-rows="24"
              @run="run"
            />
          </label>

          <p class="field__hint">
            {{ t('tools.format-convert.hints.input') }}
          </p>
        </template>

        <template #options>
          <div class="field-row">
            <label class="field">
              <span class="field__label">{{ t('tools.format-convert.fields.from') }}</span>
              <ElSelect
                v-model="from"
                class="field__select"
                :aria-label="t('tools.format-convert.fields.from')"
              >
                <ElOption
                  v-for="format in DATA_FORMATS"
                  :key="`from-${format}`"
                  :label="t(`tools.format-convert.options.format.${format}`)"
                  :value="format"
                />
              </ElSelect>
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.format-convert.fields.to') }}</span>
              <ElSelect
                v-model="to"
                class="field__select"
                :aria-label="t('tools.format-convert.fields.to')"
              >
                <ElOption
                  v-for="format in DATA_FORMATS"
                  :key="`to-${format}`"
                  :label="t(`tools.format-convert.options.format.${format}`)"
                  :value="format"
                />
              </ElSelect>
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.format-convert.fields.indent') }}</span>
              <ElSelect
                v-model="indent"
                class="field__select"
                :aria-label="t('tools.format-convert.fields.indent')"
              >
                <ElOption
                  v-for="size in INDENT_OPTIONS"
                  :key="size"
                  :label="String(size)"
                  :value="size"
                />
              </ElSelect>
            </label>
          </div>

          <ElCheckbox
            v-model="pretty"
            class="field__checkbox"
          >
            {{ t('tools.format-convert.fields.pretty') }}
          </ElCheckbox>
          <p class="field__hint">
            {{ t('tools.format-convert.hints.pretty') }}
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
        :title="t('tools.format-convert.results.output')"
        :use-as-input="useAsInput"
        @retry="run"
      >
        <template #view="{ viewId }">
          <template v-if="viewId === 'summary' && data">
            <StatGrid :items="stats" />
          </template>

          <ProductText
            v-else-if="viewId === 'output' && data"
            :label="t('tools.format-convert.views.output')"
            :content="String(data.output ?? '')"
          />
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.format-convert.hints.output') }}
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
