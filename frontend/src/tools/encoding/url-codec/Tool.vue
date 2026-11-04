<script setup lang="ts">
/**
 * URL 编解码工具页。
 *
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
 */
import { ElCheckbox } from 'element-plus/es/components/checkbox/index'
import 'element-plus/es/components/checkbox/style/css'
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { ElOption, ElSelect } from 'element-plus/es/components/select/index'
import 'element-plus/es/components/select/style/css'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import DataTable, { type DataTableColumn, type DataTableRow } from '@/components/tool/DataTable.vue'
import ProductText from '@/components/tool/ProductText.vue'
import ResultPanel, { type ResultView } from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import StatGrid from '@/components/tool/StatGrid.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { buildSummary } from '@/components/tool/summary'
import { useAutoRun } from '@/composables/useAutoRun'
import { useToolRun } from '@/composables/useToolRun'
import { EXAMPLES } from '@/tools/encoding/url-codec/example'

/** URL 解析结果。 */
interface URLParts {
  scheme: string
  host: string
  port: string
  path: string
  query: string
  params: { key: string; value: string }[]
}

/** 后端 url-codec 工具的响应。 */
interface URLCodecResult {
  output: string
  parsed?: URLParts
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

type Operation = 'encode' | 'decode' | 'parse'
type Component = 'full' | 'path' | 'query' | 'fragment'

const { t, tm } = useI18n()

const operation = ref<Operation>('encode')
const input = ref('')
const component = ref<Component>('full')
const plusAsSpace = ref(true)

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
  bindInput,
  useAsInput,
} = useToolRun<URLCodecResult>({ slug: 'url-codec' })

/** 结果区的「作为新输入」：把结果回填到主输入框（仅内存，不写存储）。 */
bindInput((text) => {
  input.value = text
})

const resultText = computed(() => data.value?.output ?? '')

/** 摘要卡片：5 项关键字段。 */
const stats = computed(() =>
  buildSummary(
    data.value as unknown as Record<string, unknown> | null,
    [
      { key: 'parsed.scheme', label: t('tools.url-codec.results.parts.scheme') },
      { key: 'parsed.host', label: t('tools.url-codec.results.parts.host') },
      { key: 'parsed.port', label: t('tools.url-codec.results.parts.port') },
      { key: 'parsed.path', label: t('tools.url-codec.results.parts.path') },
      { key: 'parsed.query', label: t('tools.url-codec.results.parts.query'), wide: true },
    ],
    { yes: t('common.states.yes'), no: t('common.states.no') },
  ),
)

/** 结果页签：摘要 / 产物。 */
const views = computed<ResultView[]>(() => [
  { id: 'summary', label: t('tools.url-codec.views.summary') },
  { id: 'output', label: t('tools.url-codec.views.output') },
])

/** 当前页签；默认停在「摘要」。 */
const activeView = ref('summary')
const running = computed(() => state.value === 'loading')

const parsed = computed(() => data.value?.parsed ?? null)

/** 查询参数明细表列定义。 */
const paramColumns: DataTableColumn[] = [
  { key: 'key', label: t('tools.url-codec.results.parts.key'), mono: true },
  { key: 'value', label: t('tools.url-codec.results.parts.value'), mono: true },
]

/** 查询参数明细行。 */
const paramRows = computed<DataTableRow[]>(() =>
  (parsed.value?.params ?? []).map((param) => ({
    key: param.key,
    cells: { key: param.key, value: param.value },
  })),
)

const faq = computed<FaqItem[]>(() => tm('tools.url-codec.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  void execute({
    operation: operation.value,
    input: { data: input.value, encoding: 'utf8' },
    component: component.value,
    plusAsSpace: plusAsSpace.value,
  })
}

/** 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。 */
useAutoRun({
  slug: 'url-codec',
  run,
  watch: [operation, input, component, plusAsSpace],
  enabled: () => input.value !== '',
})

/** 重置表单与结果。 */
function reset(): void {
  operation.value = 'encode'
  input.value = ''
  component.value = 'full'
  plusAsSpace.value = true
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  operation.value = example.operation
  input.value = example.input
  component.value = example.component
  plusAsSpace.value = example.plusAsSpace
  resetRun()
}

/** Ctrl / ⌘ + Enter 触发运行。 */
function onKeydown(event: KeyboardEvent): void {
  if ((event.ctrlKey || event.metaKey) && event.key === 'Enter') {
    event.preventDefault()
    run()
  }
}
</script>

<template>
  <ToolPageLayout
    slug="url-codec"
    :name="t('tools.url-codec.name')"
    name-en="URL Codec"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.url-codec.fields.input') }}</span>
            <ElInput
              v-model="input"
              type="textarea"
              class="field__textarea tb-mono"
              :rows="5"
              spellcheck="false"
              :aria-label="t('tools.url-codec.fields.input')"
              @keydown="(event) => onKeydown(event as KeyboardEvent)"
            />
          </label>

          <p class="field__hint">
            {{ t('tools.url-codec.hints.input') }}
          </p>
        </template>

        <template #options>
          <div class="field-row">
            <label class="field">
              <span class="field__label">{{ t('tools.url-codec.fields.operation') }}</span>
              <ElSelect
                v-model="operation"
                class="field__select"
                :aria-label="t('tools.url-codec.fields.operation')"
              >
                <ElOption
                  :label="t('tools.url-codec.options.operation.encode')"
                  value="encode"
                />
                <ElOption
                  :label="t('tools.url-codec.options.operation.decode')"
                  value="decode"
                />
                <ElOption
                  :label="t('tools.url-codec.options.operation.parse')"
                  value="parse"
                />
              </ElSelect>
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.url-codec.fields.component') }}</span>
              <ElSelect
                v-model="component"
                class="field__select"
                :aria-label="t('tools.url-codec.fields.component')"
              >
                <ElOption
                  :label="t('tools.url-codec.options.component.full')"
                  value="full"
                />
                <ElOption
                  :label="t('tools.url-codec.options.component.path')"
                  value="path"
                />
                <ElOption
                  :label="t('tools.url-codec.options.component.query')"
                  value="query"
                />
                <ElOption
                  :label="t('tools.url-codec.options.component.fragment')"
                  value="fragment"
                />
              </ElSelect>
            </label>
          </div>

          <ElCheckbox
            v-if="operation === 'decode'"
            v-model="plusAsSpace"
            class="field__checkbox"
          >
            {{ t('tools.url-codec.fields.plusAsSpace') }}
          </ElCheckbox>
          <p
            v-if="operation === 'decode'"
            class="field__hint"
          >
            {{ t('tools.url-codec.hints.plusAsSpace') }}
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
        :title="t('tools.url-codec.results.output')"
        :use-as-input="useAsInput"
        :views="views"
        @retry="run"
      >
        <template #view="{ viewId }">
          <template v-if="viewId === 'summary' && data">
            <StatGrid :items="stats" />

            <DataTable
              v-if="paramRows.length > 0"
              class="table"
              :columns="paramColumns"
              :rows="paramRows"
              copy-column="value"
            />
          </template>

          <ProductText
            v-else-if="viewId === 'output' && data"
            :label="t('tools.url-codec.views.output')"
            :content="String(data.output ?? '')"
          />
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.url-codec.hints.output') }}
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
.table {
  margin-top: var(--tb-space-3);
}
</style>
