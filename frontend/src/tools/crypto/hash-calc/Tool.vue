<script setup lang="ts">
/**
 * 摘要计算工具页。
 *
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
 */
import { ElCheckbox } from 'element-plus/es/components/checkbox/index'
import 'element-plus/es/components/checkbox/style/css'
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import EncodingSelect from '@/components/tool/EncodingSelect.vue'
import DataTable, { type DataTableColumn, type DataTableRow } from '@/components/tool/DataTable.vue'
import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import StatGrid, { type StatGridItem } from '@/components/tool/StatGrid.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useAutoRun } from '@/composables/useAutoRun'
import { useToolRun } from '@/composables/useToolRun'
import { DEFAULT_EXAMPLE } from '@/tools/crypto/hash-calc/example'

/** 后端 hash-calc 工具的响应。 */
interface HashCalcResult {
  digests: { algorithm: string; value: string }[]
  inputBytes: number
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

type InputEncoding = 'utf8' | 'hex'

const { t, tm } = useI18n()

const ALGORITHMS = ['SM3', 'SHA256', 'SHA384', 'SHA512', 'SHA224', 'SHA1', 'MD5'] as const

const algorithms = ref<string[]>(['SM3', 'SHA256'])
const input = ref('')
const inputEncoding = ref<InputEncoding>('utf8')
const outputEncoding = ref('hex')

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
  bindInput,
  useAsInput,
} = useToolRun<HashCalcResult>({ slug: 'hash-calc' })

/** 结果区的「作为新输入」：把结果回填到主输入框（仅内存，不写存储）。 */
bindInput((text) => {
  input.value = text
})

const running = computed(() => state.value === 'loading')

/** 结果文本：每行一个「算法: 值」。 */
const resultText = computed(() =>
  (data.value?.digests ?? []).map((d) => `${d.algorithm}: ${d.value}`).join('\n'),
)

const faq = computed<FaqItem[]>(() => tm('tools.hash-calc.faq') as FaqItem[])

/** 摘要表格列定义。 */
const digestColumns = computed<DataTableColumn[]>(() => [
  { key: 'algorithm', label: t('tools.hash-calc.results.algorithm'), mono: true },
  { key: 'length', label: t('tools.hash-calc.results.length'), mono: true, align: 'right' },
  { key: 'value', label: t('tools.hash-calc.results.value'), mono: true },
])

/** 摘要表格行。 */
const digestRows = computed<DataTableRow[]>(() =>
  (data.value?.digests ?? []).map((digest) => ({
    key: digest.algorithm,
    cells: {
      algorithm: digest.algorithm,
      length: `${digest.value.length / 2} B`,
      value: digest.value,
    },
  })),
)

/** 摘要统计卡片。 */
const stats = computed<StatGridItem[]>(() => {
  const result = data.value
  if (!result) {
    return []
  }
  return [
    { label: t('tools.hash-calc.results.count'), value: String(result.digests.length) },
    { label: t('tools.hash-calc.results.inputBytes'), value: `${result.inputBytes} B` },
  ]
})

/** 切换某个算法是否参与计算。 */
function toggleAlgorithm(name: string): void {
  const index = algorithms.value.indexOf(name)
  if (index >= 0) {
    algorithms.value.splice(index, 1)
  } else {
    algorithms.value.push(name)
  }
}

/** 运行当前输入。 */
function run(): void {
  void execute({
    algorithms: algorithms.value,
    input: { data: input.value, encoding: inputEncoding.value },
    outputEncoding: outputEncoding.value,
  })
}

/** 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。 */
useAutoRun({
  slug: 'hash-calc',
  run,
  watch: [input, algorithms, inputEncoding, outputEncoding],
  enabled: () => input.value !== '',
})

/** 重置表单与结果。 */
function reset(): void {
  algorithms.value = ['SM3', 'SHA256']
  input.value = ''
  inputEncoding.value = 'utf8'
  outputEncoding.value = 'hex'
  resetRun()
}

/** 填入示例。 */
function fillExample(): void {
  algorithms.value = [...DEFAULT_EXAMPLE.algorithms]
  input.value = DEFAULT_EXAMPLE.input
  inputEncoding.value = DEFAULT_EXAMPLE.inputEncoding
  outputEncoding.value = 'hex'
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
    slug="hash-calc"
    :name="t('tools.hash-calc.name')"
    name-en="Hash Calculator"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.hash-calc.fields.input') }}</span>
            <ElInput
              v-model="input"
              type="textarea"
              class="field__textarea tb-mono"
              :rows="6"
              spellcheck="false"
              :aria-label="t('tools.hash-calc.fields.input')"
              @keydown="(evt: Event | KeyboardEvent) => onKeydown(evt as KeyboardEvent)"
            />
          </label>

          <div class="field-row">
            <label class="field">
              <span class="field__label">{{ t('tools.hash-calc.fields.inputEncoding') }}</span>
              <EncodingSelect
                v-model="inputEncoding"
                :options="['utf8', 'hex']"
              />
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.hash-calc.fields.outputEncoding') }}</span>
              <EncodingSelect
                v-model="outputEncoding"
                :options="['hex', 'base64', 'base64url']"
              />
            </label>
          </div>

          <p class="field__hint">
            {{ t('tools.hash-calc.hints.input') }}
          </p>
        </template>

        <template #options>
          <fieldset class="field">
            <legend class="field__label">
              {{ t('tools.hash-calc.fields.algorithms') }}
            </legend>
            <div class="algorithm-list">
              <ElCheckbox
                v-for="name in ALGORITHMS"
                :key="name"
                class="field__checkbox"
                :model-value="algorithms.includes(name)"
                @change="toggleAlgorithm(name)"
              >
                {{ t(`tools.hash-calc.options.algorithms.${name}`) }}
              </ElCheckbox>
            </div>
          </fieldset>
          <p class="field__hint">
            {{ t('tools.hash-calc.hints.algorithms') }}
          </p>
        </template>

        <template #footer>
          <RunBar
            :running="running"
            :disabled="algorithms.length === 0"
            @run="run"
            @reset="reset"
            @example="fillExample"
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
        :title="t('tools.hash-calc.results.digests')"
        :use-as-input="useAsInput"
        @retry="run"
      >
        <template
          v-if="data"
          #extra
        >
          <StatGrid :items="stats" />

          <DataTable
            class="table"
            :columns="digestColumns"
            :rows="digestRows"
            :copy-column="'value'"
            :empty="t('tools.hash-calc.results.empty')"
          />
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.hash-calc.description') }}
      </p>
    </template>

    <template #faq>
      <dl class="faq">
        <template
          v-for="item in faq"
          :key="item.q"
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

.algorithm-list {
  display: flex;
  flex-wrap: wrap;
  gap: var(--tb-space-2) var(--tb-space-4);
}
</style>
