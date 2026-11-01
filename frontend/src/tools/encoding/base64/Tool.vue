<script setup lang="ts">
/**
 * Base64 编解码工具页。
 *
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
 */
import { ElButton } from 'element-plus/es/components/button/index'
import 'element-plus/es/components/button/style/css'
import { ElCheckbox } from 'element-plus/es/components/checkbox/index'
import 'element-plus/es/components/checkbox/style/css'
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { ElOption, ElSelect } from 'element-plus/es/components/select/index'
import 'element-plus/es/components/select/style/css'
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'

import AppIcon from '@/components/common/AppIcon.vue'
import EncodingSelect from '@/components/tool/EncodingSelect.vue'
import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import ValidationHint from '@/components/tool/ValidationHint.vue'
import { useAutoRun } from '@/composables/useAutoRun'
import { useToolRun } from '@/composables/useToolRun'
import { useValidation } from '@/composables/useValidation'
import { EXAMPLES } from '@/tools/encoding/base64/example'
import {
  buildPayload,
  normalizeForm,
  reverseConversion,
  type Base64FormState,
  type Base64InputEncoding,
  type Base64Operation,
  type Base64SourceVariant,
  type Base64Variant,
} from '@/tools/encoding/base64/form'

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

type Operation = Base64Operation
type Variant = Base64Variant
type SourceVariant = Base64SourceVariant
type InputEncoding = Base64InputEncoding

const { t, tm } = useI18n()

const operation = ref<Operation>('encode')
const input = ref('')
const inputEncoding = ref<InputEncoding>('utf8')
const variant = ref<SourceVariant>('standard')
const padding = ref(true)
const outputEncoding = ref<InputEncoding>('utf8')
const targetVariant = ref<Variant>('url')

/** 当前表单状态。 */
const form = computed<Base64FormState>(() => ({
  operation: operation.value,
  input: input.value,
  inputEncoding: inputEncoding.value,
  variant: variant.value,
  padding: padding.value,
  outputEncoding: outputEncoding.value,
  targetVariant: targetVariant.value,
}))

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

const inputEncodings = computed<readonly InputEncoding[]>(() =>
  operation.value === 'convert' ? ['utf8'] : ['utf8', 'hex'],
)

const faq = computed<FaqItem[]>(() => tm('tools.base64.faq') as FaqItem[])

/**
 * 切换操作时归一化互斥的选项。
 *
 * `auto` 变体只对 convert 有意义，`hex` 输入编码只对 encode / decode 有意义；
 * 切走时回落为各自的默认值，避免发出后端不接受的组合。
 */
watch(operation, () => {
  const next = normalizeForm(form.value)
  variant.value = next.variant
  inputEncoding.value = next.inputEncoding
})

/** 运行当前输入。 */
function run(): void {
  void execute(buildPayload(form.value))
}

/** 反向转换：把结果作为新输入再来一次（标准 ⇄ URL 安全）。 */
function reverse(): void {
  const next = reverseConversion(form.value, data.value?.output ?? '')
  variant.value = next.variant
  targetVariant.value = next.targetVariant
  input.value = next.input
}

/** 输入即时校验：纯前端语法检查，不发请求、不写存储。 */
const { validation } = useValidation(input, {
  format: computed(() => {
    if (inputEncoding.value === 'hex') {
      return 'hex'
    }
    // convert 的输入本身就是 Base64 文本，其余操作按所选变体检查字符集。
    return 'base64'
  }),
})

/** 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。 */
useAutoRun({
  slug: 'base64',
  run,
  watch: [operation, input, inputEncoding, variant, padding, outputEncoding, targetVariant],
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
  targetVariant.value = 'url'
  resetRun()
}

/**
 * 填入示例。
 *
 * 优先给**当前操作**对应的示例（encode / decode / convert 各一），
 * 没有匹配时回落到第一个示例。
 */
function fillExample(): void {
  const example =
    EXAMPLES.find((item) => item.operation === operation.value) ?? EXAMPLES[0]
  if (!example) {
    return
  }
  operation.value = example.operation
  input.value = example.input
  inputEncoding.value = example.inputEncoding
  variant.value = example.variant
  padding.value = example.padding
  outputEncoding.value = example.outputEncoding
  targetVariant.value = example.targetVariant
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
            <ElInput
              v-model="input"
              type="textarea"
              class="field__textarea tb-mono"
              :rows="6"
              spellcheck="false"
              :aria-label="t('tools.base64.fields.input')"
              @keydown="(event) => onKeydown(event as KeyboardEvent)"
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
              <ElSelect
                v-model="operation"
                class="field__select"
                :aria-label="t('tools.base64.fields.operation')"
              >
                <ElOption
                  :label="t('tools.base64.options.operation.encode')"
                  value="encode"
                />
                <ElOption
                  :label="t('tools.base64.options.operation.decode')"
                  value="decode"
                />
                <ElOption
                  :label="t('tools.base64.options.operation.convert')"
                  value="convert"
                />
              </ElSelect>
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.base64.fields.variant') }}</span>
              <ElSelect
                v-model="variant"
                class="field__select"
                :aria-label="t('tools.base64.fields.variant')"
              >
                <ElOption
                  v-if="operation === 'convert'"
                  :label="t('tools.base64.options.variant.auto')"
                  value="auto"
                />
                <ElOption
                  :label="t('tools.base64.options.variant.standard')"
                  value="standard"
                />
                <ElOption
                  :label="t('tools.base64.options.variant.url')"
                  value="url"
                />
              </ElSelect>
            </label>

            <label
              v-if="operation === 'convert'"
              class="field"
            >
              <span class="field__label">{{ t('tools.base64.fields.targetVariant') }}</span>
              <ElSelect
                v-model="targetVariant"
                class="field__select"
                :aria-label="t('tools.base64.fields.targetVariant')"
              >
                <ElOption
                  :label="t('tools.base64.options.variant.standard')"
                  value="standard"
                />
                <ElOption
                  :label="t('tools.base64.options.variant.url')"
                  value="url"
                />
              </ElSelect>
            </label>

            <ElButton
              v-if="operation === 'convert'"
              class="field__reverse"
              :aria-label="t('tools.base64.actions.reverse')"
              :title="t('tools.base64.hints.reverse')"
              @click="reverse"
            >
              <AppIcon
                name="swap"
                :size="16"
              />
            </ElButton>

            <label
              v-if="operation === 'decode'"
              class="field"
            >
              <span class="field__label">{{ t('tools.base64.fields.outputEncoding') }}</span>
              <ElSelect
                v-model="outputEncoding"
                class="field__select"
                :aria-label="t('tools.base64.fields.outputEncoding')"
              >
                <ElOption
                  :label="t('tools.base64.options.outputEncoding.utf8')"
                  value="utf8"
                />
                <ElOption
                  :label="t('tools.base64.options.outputEncoding.hex')"
                  value="hex"
                />
              </ElSelect>
            </label>
          </div>

          <ElCheckbox
            v-model="padding"
            class="field__checkbox"
          >
            {{ t('tools.base64.fields.padding') }}
          </ElCheckbox>
          <p class="field__hint">
            {{ t('tools.base64.hints.padding') }}
          </p>
          <p
            v-if="operation === 'convert'"
            class="field__hint"
          >
            {{ t('tools.base64.hints.convert') }}
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
}

.field__select {
  width: 100%;
}

.field__reverse {
  align-self: flex-end;
  margin-bottom: var(--tb-space-1);
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
