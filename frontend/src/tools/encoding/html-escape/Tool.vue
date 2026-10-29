<script setup lang="ts">
/**
 * HTML 转义工具页。
 *
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
 */
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
import { EXAMPLES } from '@/tools/encoding/html-escape/example'

/** 后端 html-escape 工具的响应。 */
interface HTMLEscapeResult {
  /** 模式。 */
  mode: string
  /** 转义 / 反转义结果。 */
  output: string
  /** 处理的实体数量。 */
  entityCount: number
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const mode = ref('encode')
// 用拼接拼出结束标签，避免在 SFC 中出现字面量的结束标签。
const SAMPLE_INPUT = '<script>alert("xss")</' + 'script> & more'
const input = ref(SAMPLE_INPUT)

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<HTMLEscapeResult>({
  slug: 'html-escape',
})

const running = computed(() => state.value === 'loading')

/** 结果文本。 */
const resultText = computed(() => data.value?.output ?? '')

const faq = computed<FaqItem[]>(() => tm('tools.html-escape.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  void execute({ mode: mode.value, input: input.value })
}

// 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。
useAutoRun({
  slug: 'html-escape',
  run,
  watch: [mode, input],
  enabled: () => true,
})

/** 重置表单与结果。 */
function reset(): void {
  mode.value = 'encode'
  input.value = SAMPLE_INPUT
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
    slug="html-escape"
    :name="t('tools.html-escape.name')"
    name-en="HTML Escape"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.html-escape.fields.mode') }}</span>
            <ElSelect
              v-model="mode"
              class="field__select"
              :aria-label="t('tools.html-escape.fields.mode')"
            >
              <ElOption
                :label="t('tools.html-escape.options.mode.encode')"
                value="encode"
              />
              <ElOption
                :label="t('tools.html-escape.options.mode.decode')"
                value="decode"
              />
            </ElSelect>
          </label>

          <label class="field field--grow">
            <span class="field__label">{{ t('tools.html-escape.fields.input') }}</span>
            <ElInput
              v-model="input"
              class="field__input tb-mono"
              type="textarea"
              :rows="8"
              autocomplete="off"
              :aria-label="t('tools.html-escape.fields.input')"
            />
          </label>

          <p class="field__hint">
            {{ t('tools.html-escape.hints.input') }}
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
        :title="t('tools.html-escape.results.output')"
        @retry="run"
      >
        <template #extra>
          <dl class="detail">
            <div class="detail__row">
              <dt>{{ t('tools.html-escape.results.entityCount') }}</dt>
              <dd class="tb-mono">
                {{ data?.entityCount ?? 0 }}
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
