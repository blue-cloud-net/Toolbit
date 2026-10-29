<script setup lang="ts">
/**
 * Base64 编解码工具页。
 *
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
 */
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import EncodingSelect from '@/components/tool/EncodingSelect.vue'
import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import ValidationHint from '@/components/tool/ValidationHint.vue'
import { useAutoRun } from '@/composables/useAutoRun'
import { useToolRun } from '@/composables/useToolRun'
import { useValidation } from '@/composables/useValidation'
import { DEFAULT_EXAMPLE } from '@/tools/encoding/base64/example'

/** 后端 base64 工具的响应。 */
interface Base64Result {
  output: string
  encoding: string
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

type Operation = 'encode' | 'decode'
type Variant = 'standard' | 'url'
type InputEncoding = 'utf8' | 'hex'

const { t, tm } = useI18n()

const operation = ref<Operation>('encode')
const input = ref('')
const inputEncoding = ref<InputEncoding>('utf8')
const variant = ref<Variant>('standard')
const padding = ref(true)
const outputEncoding = ref<InputEncoding>('utf8')

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
  bindInput,
  useAsInput,
} = useToolRun<Base64Result>({ slug: 'base64' })

/** 结果区的「作为新输入」：把结果回填到主输入框（仅内存，不写存储）。 */
bindInput((text) => {
  input.value = text
})

const resultText = computed(() => data.value?.output ?? '')
const running = computed(() => state.value === 'loading')

const inputEncodings: readonly InputEncoding[] = ['utf8', 'hex']
const outputEncodings = computed<readonly InputEncoding[]>(() =>
  operation.value === 'encode' ? ['utf8'] : ['utf8', 'hex'],
)

const faq = computed<FaqItem[]>(() => tm('tools.base64.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  void execute({
    operation: operation.value,
    input: { data: input.value, encoding: inputEncoding.value },
    variant: variant.value,
    padding: padding.value,
    outputEncoding: outputEncoding.value,
  })
}

/** 输入即时校验：纯前端语法检查，不发请求、不写存储。 */
const { validation } = useValidation(input, {
  format: computed(() => (inputEncoding.value === 'hex' ? 'hex' : 'base64')),
})

/** 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。 */
useAutoRun({
  slug: 'base64',
  run,
  watch: [operation, input, inputEncoding, variant, padding, outputEncoding],
  enabled: () => input.value !== '',
})

/** 重置表单与结果。 */
function reset(): void {
  operation.value = 'encode'
  input.value = ''
  inputEncoding.value = 'utf8'
  variant.value = 'standard'
  padding.value = true
  outputEncoding.value = 'utf8'
  resetRun()
}

/** 填入示例。 */
function fillExample(): void {
  operation.value = DEFAULT_EXAMPLE.operation
  input.value = DEFAULT_EXAMPLE.input
  inputEncoding.value = DEFAULT_EXAMPLE.inputEncoding
  variant.value = DEFAULT_EXAMPLE.variant
  padding.value = DEFAULT_EXAMPLE.padding
  outputEncoding.value = DEFAULT_EXAMPLE.outputEncoding
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
    slug="base64"
    :name="t('tools.base64.name')"
    name-en="Base64 Codec"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.base64.fields.input') }}</span>
            <textarea
              v-model="input"
              class="field__textarea tb-mono"
              rows="6"
              spellcheck="false"
              :aria-label="t('tools.base64.fields.input')"
              @keydown="onKeydown"
            />
            <ValidationHint :validation="validation" />
          </label>

          <div class="field-row">
            <label class="field">
              <span class="field__label">{{ t('tools.base64.fields.inputEncoding') }}</span>
              <EncodingSelect
                v-model="inputEncoding"
                :options="inputEncodings"
              />
            </label>
          </div>

          <p class="field__hint">
            {{ t('tools.base64.hints.input') }}
          </p>
        </template>

        <template #options>
          <div class="field-row">
            <label class="field">
              <span class="field__label">{{ t('tools.base64.fields.operation') }}</span>
              <select
                v-model="operation"
                class="field__select"
              >
                <option value="encode">{{ t('tools.base64.options.operation.encode') }}</option>
                <option value="decode">{{ t('tools.base64.options.operation.decode') }}</option>
              </select>
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.base64.fields.variant') }}</span>
              <select
                v-model="variant"
                class="field__select"
              >
                <option value="standard">{{ t('tools.base64.options.variant.standard') }}</option>
                <option value="url">{{ t('tools.base64.options.variant.url') }}</option>
              </select>
            </label>

            <label
              v-if="operation === 'decode'"
              class="field"
            >
              <span class="field__label">{{ t('tools.base64.fields.outputEncoding') }}</span>
              <select
                v-model="outputEncoding"
                class="field__select"
              >
                <option
                  v-for="enc in outputEncodings"
                  :key="enc"
                  :value="enc"
                >
                  {{ t(`tools.base64.options.outputEncoding.${enc}`) }}
                </option>
              </select>
            </label>
          </div>

          <label class="field field--inline">
            <input
              v-model="padding"
              type="checkbox"
              class="field__checkbox"
            >
            <span>{{ t('tools.base64.fields.padding') }}</span>
          </label>
          <p class="field__hint">
            {{ t('tools.base64.hints.padding') }}
          </p>
        </template>

        <template #footer>
          <RunBar
            :running="running"
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
        :title="t('tools.base64.results.output')"
        :use-as-input="useAsInput"
        @retry="run"
      />
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.base64.hints.output') }}
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
.field {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-1);
}

.field--inline {
  flex-direction: row;
  align-items: center;
  gap: var(--tb-space-2);
}

.field__label {
  font-size: var(--tb-font-xs);
  color: var(--tb-text-secondary);
}

.field__textarea {
  width: 100%;
  padding: var(--tb-space-2) var(--tb-space-3);
  font-family: var(--tb-font-mono);
  font-size: var(--tb-font-sm);
  line-height: 1.6;
  color: var(--tb-text);
  background: var(--tb-bg-sunken);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-control);
  resize: vertical;
}

.field__select {
  padding: var(--tb-space-1) var(--tb-space-2);
  font-family: inherit;
  font-size: var(--tb-font-sm);
  color: var(--tb-text);
  background: var(--tb-bg-elevated);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-control);
  cursor: pointer;
}

.field__checkbox {
  width: 16px;
  height: 16px;
  accent-color: var(--tb-primary);
}

.field__hint {
  margin: 0;
  font-size: var(--tb-font-xs);
  color: var(--tb-text-tertiary);
}

.field-row {
  display: flex;
  flex-wrap: wrap;
  gap: var(--tb-space-3);
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
