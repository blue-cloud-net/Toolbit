<script setup lang="ts">
/**
 * XML 格式化 工具页。
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
import { useToolRun } from '@/composables/useToolRun'
import { EXAMPLES } from '@/tools/formatting/xml-format/example'

/** 后端 xml-format 工具的响应。 */
interface XmlFormatResult {
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
const keepComments = ref(false)

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<XmlFormatResult>({
  slug: 'xml-format',
})

const resultText = computed(() => data.value?.output ?? '')
const running = computed(() => state.value === 'loading')

/** 常用缩进宽度。 */
const INDENT_OPTIONS = [2, 4, 8] as const

const faq = computed<FaqItem[]>(() => tm('tools.xml-format.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  void execute({
    input: { data: input.value },
    action: action.value,
    indent: indent.value,
    keepComments: keepComments.value,
  })
}

/** 重置表单与结果。 */
function reset(): void {
  input.value = ''
  action.value = 'format'
  indent.value = 2
  keepComments.value = false
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
  keepComments.value = example.keepComments
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="xml-format"
    :name="t('tools.xml-format.name')"
    name-en="XML Formatter"
    wide
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.xml-format.fields.input') }}</span>
            <CodeEditor
              v-model="input"
              :label="t('tools.xml-format.fields.input')"
              :max-rows="24"
              @run="run"
            />
          </label>

          <p class="field__hint">
            {{ t('tools.xml-format.hints.input') }}
          </p>
        </template>

        <template #options>
          <div class="field-row">
            <label class="field">
              <span class="field__label">{{ t('tools.xml-format.fields.action') }}</span>
              <select
                v-model="action"
                class="field__select"
              >
                <option value="format">{{ t('tools.xml-format.options.action.format') }}</option>
                <option value="minify">{{ t('tools.xml-format.options.action.minify') }}</option>
                <option value="validate">{{ t('tools.xml-format.options.action.validate') }}</option>
              </select>
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.xml-format.fields.indent') }}</span>
              <select
                v-model.number="indent"
                class="field__select"
              >
                <option
                  v-for="size in INDENT_OPTIONS"
                  :key="size"
                  :value="size"
                >
                  {{ size }}
                </option>
              </select>
            </label>
          </div>

          <label class="field field--inline">
            <input
              v-model="keepComments"
              type="checkbox"
              class="field__checkbox"
            >
            <span>{{ t('tools.xml-format.fields.keepComments') }}</span>
          </label>
          <p class="field__hint">
            {{ t('tools.xml-format.hints.keepComments') }}
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
        :title="t('tools.xml-format.results.output')"
        @retry="run"
      >
        <template
          v-if="data"
          #extra
        >
          <dl class="detail">
            <div class="detail__row">
              <dt>{{ t('tools.xml-format.results.validity') }}</dt>
              <dd :class="data?.valid ? 'detail__ok' : 'detail__bad'">
                {{ data?.valid ? t('tools.xml-format.results.valid') : t('tools.xml-format.results.invalid') }}
              </dd>
            </div>
            <div
              v-if="data?.error"
              class="detail__row"
            >
              <dt>{{ t('tools.xml-format.results.position') }}</dt>
              <dd class="tb-mono">
                {{ data?.error?.line }}:{{ data?.error?.column }}
              </dd>
            </div>
            <div
              v-if="data?.error"
              class="detail__row"
            >
              <dt>{{ t('tools.xml-format.results.reason') }}</dt>
              <dd>{{ data?.error?.message }}</dd>
            </div>
          </dl>
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.xml-format.hints.output') }}
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

.detail__ok {
  color: var(--tb-success);
}

.detail__bad {
  color: var(--tb-danger);
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
