<script setup lang="ts">
/**
 * 颜色格式转换工具页。
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
import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import StatGrid, { type StatGridItem } from '@/components/tool/StatGrid.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useAutoRun } from '@/composables/useAutoRun'
import { useToolRun } from '@/composables/useToolRun'
import { EXAMPLES } from '@/tools/formatting/color-convert/example'

/** 后端 color-convert 工具的响应（扁平字段）。 */
interface ColorConvertResult {
  /** 规范化后的输入值。 */
  input: string
  /** 识别出的输入格式。 */
  inputFormat: string
  /** 十六进制表示。 */
  hex: string
  /** RGB 表示。 */
  rgb: string
  /** HSL 表示。 */
  hsl: string
  /** OKLCH 表示。 */
  oklch: string
  /** 与白色的对比度。 */
  contrastOnWhite: number
  /** 与黑色的对比度。 */
  contrastOnBlack: number
  /** 是否是深色。 */
  isDark: boolean
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

/** 结果里可能出现的格式键（与后端字段同名）。 */
const FORMATS = ['hex', 'rgb', 'hsl', 'oklch'] as const

type ColorFormat = (typeof FORMATS)[number]

const input = ref('#3366ff')
const inputFormat = ref('hex')
const outputFormats = ref<string[]>(['hex', 'rgb', 'hsl', 'oklch'])

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<ColorConvertResult>({
  slug: 'color-convert',
})

const running = computed(() => state.value === 'loading')

/** 结果文本：各格式的颜色值。 */
const resultText = computed(() => {
  const result = data.value
  if (!result) {
    return ''
  }
  return FORMATS.filter((format) => result[format] !== '')
    .map((format) => `${format}\t${result[format]}`)
    .join('\n')
})

const faq = computed<FaqItem[]>(() => tm('tools.color-convert.faq') as FaqItem[])

/** 统计卡片：对比度与深色标记。 */
const stats = computed<StatGridItem[]>(() => [
  {
    label: t('tools.color-convert.results.contrastOnWhite'),
    value: String(data.value?.contrastOnWhite ?? 0),
  },
  {
    label: t('tools.color-convert.results.contrastOnBlack'),
    value: String(data.value?.contrastOnBlack ?? 0),
  },
  {
    label: t('tools.color-convert.results.isDark'),
    value: data.value?.isDark ? t('common.states.yes') : t('common.states.no'),
  },
])

/** 表格列定义。 */
const columns = computed<DataTableColumn[]>(() => [
  { key: 'format', label: t('tools.color-convert.results.formatName'), mono: true },
  { key: 'value', label: t('tools.color-convert.results.formatValue'), mono: true },
])

/** 表格行；格式名按 `options.format.*` 本地化，请求未包含的格式不出现。 */
const rows = computed<DataTableRow[]>(() => {
  const result = data.value
  if (!result) {
    return []
  }
  return FORMATS.filter((format: ColorFormat) => result[format] !== '').map((format) => ({
    key: format,
    cells: {
      format: t(`tools.color-convert.options.format.${format}`),
      value: result[format],
    },
  }))
})

/** 运行当前输入。 */
function run(): void {
  if (input.value.trim() === '') {
    return
  }
  void execute({
    input: input.value,
    inputFormat: inputFormat.value,
    outputFormats: outputFormats.value,
  })
}

// 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。
useAutoRun({
  slug: 'color-convert',
  run,
  watch: [input, inputFormat, outputFormats],
  enabled: () => input.value.trim() !== '',
})

/** 重置表单与结果。 */
function reset(): void {
  input.value = '#3366ff'
  inputFormat.value = 'hex'
  outputFormats.value = ['hex', 'rgb', 'hsl', 'oklch']
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
    slug="color-convert"
    :name="t('tools.color-convert.name')"
    name-en="Color Converter"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field field--grow">
            <span class="field__label">{{ t('tools.color-convert.fields.input') }}</span>
            <ElInput
              v-model="input"
              class="field__input tb-mono"
              autocomplete="off"
              :aria-label="t('tools.color-convert.fields.input')"
              @keydown.enter.prevent="run"
            />
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.color-convert.fields.inputFormat') }}</span>
            <ElSelect
              v-model="inputFormat"
              class="field__select"
              :aria-label="t('tools.color-convert.fields.inputFormat')"
            >
              <ElOption
                v-for="name in ['hex', 'rgb', 'hsl', 'oklch']"
                :key="name"
                :label="t(`tools.color-convert.options.format.${name}`)"
                :value="name"
              />
            </ElSelect>
          </label>
        </template>

        <template #options>
          <div class="field field--grow">
            <span class="field__label">{{ t('tools.color-convert.fields.outputFormats') }}</span>
            <ElCheckboxGroup
              v-model="outputFormats"
              class="field__checkgroup"
              :aria-label="t('tools.color-convert.fields.outputFormats')"
            >
              <ElCheckbox value="hex">
                {{ t('tools.color-convert.options.format.hex') }}
              </ElCheckbox>
              <ElCheckbox value="rgb">
                {{ t('tools.color-convert.options.format.rgb') }}
              </ElCheckbox>
              <ElCheckbox value="hsl">
                {{ t('tools.color-convert.options.format.hsl') }}
              </ElCheckbox>
              <ElCheckbox value="oklch">
                {{ t('tools.color-convert.options.format.oklch') }}
              </ElCheckbox>
            </ElCheckboxGroup>
          </div>

          <p class="field__hint">
            {{ t('tools.color-convert.hints.input') }}
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
        :title="t('tools.color-convert.results.colors')"
        hide-text
        @retry="run"
      >
        <template #extra>
          <StatGrid :items="stats" />

          <DataTable
            class="table"
            :columns="columns"
            :rows="rows"
            :copy-column="'value'"
            :empty="t('tools.color-convert.results.empty')"
          />
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

<style scoped>
.table {
  margin-top: var(--tb-space-3);
}
</style>
