<script setup lang="ts">
/**
 * CSV 转换工具页。
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

import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useAutoRun } from '@/composables/useAutoRun'
import { useToolRun } from '@/composables/useToolRun'
import { EXAMPLES } from '@/tools/data/csv-convert/example'

/** 后端 csv-convert 工具的响应。 */
interface CSVConvertResult {
  /** 转换方向。 */
  operation: string
  /** 转换结果。 */
  output: string
  /** 数据行数。 */
  rowCount: number
  /** 列数。 */
  columnCount: number
  /** 列名。 */
  columns: string[]
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const operation = ref('csvToJson')
const input = ref('name,age\nAlice,30\nBob,25')
const delimiter = ref(',')
const hasHeader = ref(true)
const pretty = ref(false)

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<CSVConvertResult>({
  slug: 'csv-convert',
})

const running = computed(() => state.value === 'loading')

/** 结果文本：转换结果。 */
const resultText = computed(() => data.value?.output ?? '')

const faq = computed<FaqItem[]>(() => tm('tools.csv-convert.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  if (input.value.trim() === '') {
    return
  }
  void execute({
    operation: operation.value,
    input: input.value,
    delimiter: delimiter.value,
    hasHeader: hasHeader.value,
    pretty: pretty.value,
  })
}

// 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。
useAutoRun({
  slug: 'csv-convert',
  run,
  watch: [operation, input, delimiter, hasHeader, pretty],
  enabled: () => input.value.trim() !== '',
})

/** 重置表单与结果。 */
function reset(): void {
  operation.value = 'csvToJson'
  input.value = 'name,age\nAlice,30\nBob,25'
  delimiter.value = ','
  hasHeader.value = true
  pretty.value = false
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  input.value = example.input
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="csv-convert"
    :name="t('tools.csv-convert.name')"
    name-en="CSV Converter"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.csv-convert.fields.operation') }}</span>
            <ElSelect
              v-model="operation"
              class="field__select"
              :aria-label="t('tools.csv-convert.fields.operation')"
            >
              <ElOption
                :label="t('tools.csv-convert.options.operation.csvToJson')"
                value="csvToJson"
              />
              <ElOption
                :label="t('tools.csv-convert.options.operation.jsonToCsv')"
                value="jsonToCsv"
              />
            </ElSelect>
          </label>

          <label class="field field--grow">
            <span class="field__label">{{ t('tools.csv-convert.fields.input') }}</span>
            <ElInput
              v-model="input"
              class="field__input tb-mono"
              type="textarea"
              :rows="8"
              autocomplete="off"
              :aria-label="t('tools.csv-convert.fields.input')"
            />
          </label>
        </template>

        <template #options>
          <label class="field">
            <span class="field__label">{{ t('tools.csv-convert.fields.delimiter') }}</span>
            <ElInput
              v-model="delimiter"
              class="field__input tb-mono"
              maxlength="1"
              autocomplete="off"
              :aria-label="t('tools.csv-convert.fields.delimiter')"
            />
          </label>

          <label class="field field--check">
            <ElCheckbox
              v-model="hasHeader"
              :aria-label="t('tools.csv-convert.fields.hasHeader')"
            >
              {{ t('tools.csv-convert.fields.hasHeader') }}
            </ElCheckbox>
          </label>

          <label class="field field--check">
            <ElCheckbox
              v-model="pretty"
              :aria-label="t('tools.csv-convert.fields.pretty')"
            >
              {{ t('tools.csv-convert.fields.pretty') }}
            </ElCheckbox>
          </label>

          <p class="field__hint">
            {{ t('tools.csv-convert.hints.input') }}
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
        :title="t('tools.csv-convert.results.output')"
        @retry="run"
      >
        <template #extra>
          <dl class="detail">
            <div class="detail__row">
              <dt>{{ t('tools.csv-convert.results.rowCount') }}</dt>
              <dd class="tb-mono">
                {{ data?.rowCount ?? 0 }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.csv-convert.results.columnCount') }}</dt>
              <dd class="tb-mono">
                {{ data?.columnCount ?? 0 }}
              </dd>
            </div>
            <div
              v-if="data && data.columns.length > 0"
              class="detail__row"
            >
              <dt>{{ t('tools.csv-convert.results.columns') }}</dt>
              <dd class="tb-mono">
                {{ data.columns.join(', ') }}
              </dd>
            </div>
          </dl>
        </template>
      </ResultPanel>
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
