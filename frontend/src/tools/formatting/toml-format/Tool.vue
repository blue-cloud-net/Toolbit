<script setup lang="ts">
/**
 * TOML 格式化 工具页。
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
import { useAutoRun } from '@/composables/useAutoRun'
import { useToolRun } from '@/composables/useToolRun'
import { EXAMPLES } from '@/tools/formatting/toml-format/example'

/** 后端 toml-format 工具的响应。 */
interface TomlFormatResult {
  output: string
  valid: boolean
  error?: SyntaxErrorInfo
  /** 标记注释是否丢失。 */
  commentsLost?: boolean
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

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
  bindInput,
  useAsInput,
} = useToolRun<TomlFormatResult>({
  slug: 'toml-format',
})

/** 结果区的「作为新输入」：把结果回填到主输入框（仅内存，不写存储）。 */
bindInput((text) => {
  input.value = text
})

const resultText = computed(() => data.value?.output ?? '')
const running = computed(() => state.value === 'loading')

/** 常用缩进宽度。 */
const INDENT_OPTIONS = [2, 4, 8] as const

const faq = computed<FaqItem[]>(() => tm('tools.toml-format.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  void execute({
    input: { data: input.value },
    action: action.value,
    indent: indent.value,
  })
}

/** 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。 */
useAutoRun({
  slug: 'toml-format',
  run,
  watch: [input, action, indent],
  enabled: () => input.value !== '',
})

/** 重置表单与结果。 */
function reset(): void {
  input.value = ''
  action.value = 'format'
  indent.value = 2
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
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="toml-format"
    :name="t('tools.toml-format.name')"
    name-en="TOML Formatter"
    wide
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.toml-format.fields.input') }}</span>
            <CodeEditor
              v-model="input"
              :label="t('tools.toml-format.fields.input')"
              :max-rows="24"
              @run="run"
            />
          </label>

          <p class="field__hint">
            {{ t('tools.toml-format.hints.input') }}
          </p>
        </template>

        <template #options>
          <div class="field-row">
            <label class="field">
              <span class="field__label">{{ t('tools.toml-format.fields.action') }}</span>
              <select
                v-model="action"
                class="field__select"
              >
                <option value="format">{{ t('tools.toml-format.options.action.format') }}</option>
                <option value="minify">{{ t('tools.toml-format.options.action.minify') }}</option>
                <option value="validate">{{ t('tools.toml-format.options.action.validate') }}</option>
              </select>
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.toml-format.fields.indent') }}</span>
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
        :title="t('tools.toml-format.results.output')"
        :use-as-input="useAsInput"
        @retry="run"
      >
        <template
          v-if="data"
          #extra
        >
          <dl class="detail">
            <div class="detail__row">
              <dt>{{ t('tools.toml-format.results.validity') }}</dt>
              <dd :class="data?.valid ? 'detail__ok' : 'detail__bad'">
                {{ data?.valid ? t('tools.toml-format.results.valid') : t('tools.toml-format.results.invalid') }}
              </dd>
            </div>
            <div
              v-if="data?.error"
              class="detail__row"
            >
              <dt>{{ t('tools.toml-format.results.position') }}</dt>
              <dd class="tb-mono">
                {{ data?.error?.line }}:{{ data?.error?.column }}
              </dd>
            </div>
            <div
              v-if="data?.error"
              class="detail__row"
            >
              <dt>{{ t('tools.toml-format.results.reason') }}</dt>
              <dd>{{ data?.error?.message }}</dd>
            </div>
            <div
              v-if="data?.commentsLost"
              class="detail__row"
            >
              <dt>{{ t('tools.toml-format.results.commentsLost') }}</dt>
              <dd>{{ t('tools.toml-format.results.commentsLostHint') }}</dd>
            </div>
          </dl>
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.toml-format.hints.output') }}
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
