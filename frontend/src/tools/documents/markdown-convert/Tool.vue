<script setup lang="ts">
/**
 * Markdown 转换工具页。
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
import { EXAMPLES } from '@/tools/documents/markdown-convert/example'

/** 后端 markdown-convert 工具的响应。 */
interface MarkdownConvertResult {
  /** 模式。 */
  mode: string
  /** 转换结果。 */
  output: string
  /** 是否启用 GFM。 */
  gfm: boolean
  /** 转换警告。 */
  warnings: string[]
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const mode = ref('toHtml')
const input = ref('# 标题\n\n这是一段 **加粗** 文本。\n\n- 项目一\n- 项目二')
const gfm = ref(true)

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<MarkdownConvertResult>({
  slug: 'markdown-convert',
})

const running = computed(() => state.value === 'loading')

/** 结果文本。 */
const resultText = computed(() => data.value?.output ?? '')

const faq = computed<FaqItem[]>(() => tm('tools.markdown-convert.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  void execute({ mode: mode.value, input: input.value, gfm: gfm.value })
}

// 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。
useAutoRun({
  slug: 'markdown-convert',
  run,
  watch: [mode, input, gfm],
  enabled: () => true,
})

/** 重置表单与结果。 */
function reset(): void {
  mode.value = 'toHtml'
  input.value = '# 标题\n\n这是一段 **加粗** 文本。\n\n- 项目一\n- 项目二'
  gfm.value = true
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
    slug="markdown-convert"
    :name="t('tools.markdown-convert.name')"
    name-en="Markdown Converter"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.markdown-convert.fields.mode') }}</span>
            <ElSelect
              v-model="mode"
              class="field__select"
              :aria-label="t('tools.markdown-convert.fields.mode')"
            >
              <ElOption
                :label="t('tools.markdown-convert.options.mode.toHtml')"
                value="toHtml"
              />
              <ElOption
                :label="t('tools.markdown-convert.options.mode.toMarkdown')"
                value="toMarkdown"
              />
            </ElSelect>
          </label>

          <label class="field field--grow">
            <span class="field__label">{{ t('tools.markdown-convert.fields.input') }}</span>
            <ElInput
              v-model="input"
              class="field__input tb-mono"
              type="textarea"
              :rows="12"
              autocomplete="off"
              :aria-label="t('tools.markdown-convert.fields.input')"
            />
          </label>
        </template>

        <template #options>
          <label class="field field--check">
            <ElCheckbox
              v-model="gfm"
              :aria-label="t('tools.markdown-convert.fields.gfm')"
            >
              {{ t('tools.markdown-convert.fields.gfm') }}
            </ElCheckbox>
          </label>

          <p class="field__hint">
            {{ t('tools.markdown-convert.hints.input') }}
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
        :title="t('tools.markdown-convert.results.output')"
        @retry="run"
      >
        <template #extra>
          <dl class="detail">
            <div
              v-for="(warning, index) in data?.warnings ?? []"
              :key="index"
              class="detail__row"
            >
              <dt>!</dt>
              <dd>{{ warning }}</dd>
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
