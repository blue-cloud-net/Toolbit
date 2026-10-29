<script setup lang="ts">
/**
 * 字符编码转换工具页。
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
import { useAutoRun } from '@/composables/useAutoRun'
import { useToolRun } from '@/composables/useToolRun'
import { EXAMPLES } from '@/tools/encoding/encoding-convert/example'

/** 后端 encoding-convert 工具的响应。 */
interface EncodingConvertResult {
  output: string
  encoding: string
  detected?: string
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

type TransferEncoding = 'utf8' | 'hex' | 'base64' | 'base64url'

const { t, tm } = useI18n()

const input = ref('')
const inputEncoding = ref<TransferEncoding>('utf8')
const from = ref('auto')
const to = ref('utf-8')
const bom = ref(false)
const outputEncoding = ref<TransferEncoding>('base64')

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
  bindInput,
  useAsInput,
} = useToolRun<EncodingConvertResult>({ slug: 'encoding-convert' })

/** 结果区的「作为新输入」：把结果回填到主输入框（仅内存，不写存储）。 */
bindInput((text) => {
  input.value = text
})

const resultText = computed(() => data.value?.output ?? '')
const running = computed(() => state.value === 'loading')

/** 可选的传输编码。 */
const TRANSFER_ENCODINGS: readonly TransferEncoding[] = ['utf8', 'hex', 'base64', 'base64url']

/** 源字符集；`auto` 表示自动探测。 */
const SOURCE_CHARSETS = [
  'auto',
  'utf-8',
  'utf-16',
  'utf-16le',
  'utf-16be',
  'gbk',
  'big5',
  'shift_jis',
  'latin-1',
] as const

/** 目标字符集。 */
const TARGET_CHARSETS = [
  'utf-8',
  'utf-16',
  'utf-16le',
  'utf-16be',
  'gbk',
  'big5',
  'shift_jis',
  'latin-1',
] as const

const faq = computed<FaqItem[]>(() => tm('tools.encoding-convert.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  void execute({
    input: { data: input.value, encoding: inputEncoding.value },
    from: from.value,
    to: to.value,
    bom: bom.value,
    outputEncoding: outputEncoding.value,
  })
}

/** 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。 */
useAutoRun({
  slug: 'encoding-convert',
  run,
  watch: [input, inputEncoding, from, to, bom, outputEncoding],
  enabled: () => input.value !== '',
})

/** 重置表单与结果。 */
function reset(): void {
  input.value = ''
  inputEncoding.value = 'utf8'
  from.value = 'auto'
  to.value = 'utf-8'
  bom.value = false
  outputEncoding.value = 'base64'
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  input.value = example.input
  inputEncoding.value = example.inputEncoding
  from.value = example.from
  to.value = example.to
  bom.value = example.bom
  outputEncoding.value = example.outputEncoding
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
    slug="encoding-convert"
    :name="t('tools.encoding-convert.name')"
    name-en="Encoding Converter"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.encoding-convert.fields.input') }}</span>
            <textarea
              v-model="input"
              class="field__textarea tb-mono"
              rows="6"
              spellcheck="false"
              :aria-label="t('tools.encoding-convert.fields.input')"
              @keydown="onKeydown"
            />
          </label>

          <div class="field-row">
            <label class="field">
              <span class="field__label">{{ t('tools.encoding-convert.fields.inputEncoding') }}</span>
              <EncodingSelect
                v-model="inputEncoding"
                :options="TRANSFER_ENCODINGS"
              />
            </label>
          </div>

          <p class="field__hint">
            {{ t('tools.encoding-convert.hints.input') }}
          </p>
        </template>

        <template #options>
          <div class="field-row">
            <label class="field">
              <span class="field__label">{{ t('tools.encoding-convert.fields.from') }}</span>
              <select
                v-model="from"
                class="field__select"
              >
                <option
                  v-for="charset in SOURCE_CHARSETS"
                  :key="charset"
                  :value="charset"
                >
                  {{ t(`tools.encoding-convert.options.charset.${charset}`) }}
                </option>
              </select>
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.encoding-convert.fields.to') }}</span>
              <select
                v-model="to"
                class="field__select"
              >
                <option
                  v-for="charset in TARGET_CHARSETS"
                  :key="charset"
                  :value="charset"
                >
                  {{ t(`tools.encoding-convert.options.charset.${charset}`) }}
                </option>
              </select>
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.encoding-convert.fields.outputEncoding') }}</span>
              <EncodingSelect
                v-model="outputEncoding"
                :options="TRANSFER_ENCODINGS"
              />
            </label>
          </div>

          <label class="field field--inline">
            <input
              v-model="bom"
              type="checkbox"
              class="field__checkbox"
            >
            <span>{{ t('tools.encoding-convert.fields.bom') }}</span>
          </label>
          <p class="field__hint">
            {{ t('tools.encoding-convert.hints.bom') }}
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
        :title="t('tools.encoding-convert.results.output')"
        :use-as-input="useAsInput"
        @retry="run"
      >
        <template
          v-if="data"
          #extra
        >
          <dl class="detail">
            <div class="detail__row">
              <dt>{{ t('tools.encoding-convert.results.encoding') }}</dt>
              <dd>{{ data?.encoding }}</dd>
            </div>
            <div
              v-if="data?.detected"
              class="detail__row"
            >
              <dt>{{ t('tools.encoding-convert.results.detected') }}</dt>
              <dd>{{ data?.detected }}</dd>
            </div>
          </dl>
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.encoding-convert.hints.output') }}
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
