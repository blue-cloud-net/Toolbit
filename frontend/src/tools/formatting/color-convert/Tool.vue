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

import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useAutoRun } from '@/composables/useAutoRun'
import { useToolRun } from '@/composables/useToolRun'
import { EXAMPLES } from '@/tools/formatting/color-convert/example'

/** 后端 color-convert 工具的响应。 */
interface ColorConvertResult {
  /** 输入颜色。 */
  input: string
  /** 输入格式。 */
  inputFormat: string
  /** 转换结果。 */
  colors: Record<string, string>
  /** 对比度（相对白底与黑底）。 */
  contrast: {
    onWhite: number
    onBlack: number
  }
  /** WCAG 等级。 */
  wcag: {
    onWhite: string
    onBlack: string
  }
  /** 是否为纯色。 */
  isValid: boolean
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const input = ref('#3366ff')
const inputFormat = ref('hex')
const outputFormats = ref<string[]>(['hex', 'rgb', 'hsl', 'oklch'])

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<ColorConvertResult>({
  slug: 'color-convert',
})

const running = computed(() => state.value === 'loading')

/** 结果文本：各格式的颜色值。 */
const resultText = computed(() => {
  if (!data.value) {
    return ''
  }
  return Object.entries(data.value.colors)
    .map(([format, value]) => `${format}\t${value}`)
    .join('\n')
})

const faq = computed<FaqItem[]>(() => tm('tools.color-convert.faq') as FaqItem[])

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
        @retry="run"
      >
        <template #extra>
          <dl class="detail">
            <div class="detail__row">
              <dt>{{ t('tools.color-convert.results.contrastOnWhite') }}</dt>
              <dd class="tb-mono">
                {{ data?.contrast.onWhite ?? 0 }} ({{ data?.wcag.onWhite ?? '-' }})
              </dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.color-convert.results.contrastOnBlack') }}</dt>
              <dd class="tb-mono">
                {{ data?.contrast.onBlack ?? 0 }} ({{ data?.wcag.onBlack ?? '-' }})
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
