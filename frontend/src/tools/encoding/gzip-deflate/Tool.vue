<script setup lang="ts">
/**
 * GZip / Deflate 压缩工具页。
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
import { useToolRun } from '@/composables/useToolRun'
import { EXAMPLES } from '@/tools/encoding/gzip-deflate/example'

/** 后端 gzip-deflate 工具的响应。 */
interface GzipDeflateResult {
  output: string
  encoding: string
  ratio: number
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

type Operation = 'compress' | 'decompress'
type ContainerFormat = 'gzip' | 'deflate' | 'raw'
type TransferEncoding = 'utf8' | 'hex' | 'base64' | 'base64url'

const { t, tm } = useI18n()

const operation = ref<Operation>('compress')
const format = ref<ContainerFormat>('gzip')
const input = ref('')
const inputEncoding = ref<TransferEncoding>('utf8')
const level = ref(6)
const outputEncoding = ref<TransferEncoding>('base64')

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<GzipDeflateResult>({
  slug: 'gzip-deflate',
})

const resultText = computed(() => data.value?.output ?? '')
const running = computed(() => state.value === 'loading')

/** 可选的传输编码。 */
const TRANSFER_ENCODINGS: readonly TransferEncoding[] = ['utf8', 'hex', 'base64', 'base64url']

/** 压缩级别。 */
const LEVEL_OPTIONS = [0, 1, 2, 3, 4, 5, 6, 7, 8, 9] as const

const faq = computed<FaqItem[]>(() => tm('tools.gzip-deflate.faq') as FaqItem[])

/** 压缩比；百分数。 */
const ratioText = computed(() => {
  const ratio = data.value?.ratio
  if (ratio === undefined) {
    return ''
  }
  return `${(ratio * 100).toFixed(1)}%`
})

/** 运行当前输入。 */
function run(): void {
  void execute({
    operation: operation.value,
    format: format.value,
    input: { data: input.value, encoding: inputEncoding.value },
    level: level.value,
    outputEncoding: outputEncoding.value,
  })
}

/** 重置表单与结果。 */
function reset(): void {
  operation.value = 'compress'
  format.value = 'gzip'
  input.value = ''
  inputEncoding.value = 'utf8'
  level.value = 6
  outputEncoding.value = 'base64'
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  operation.value = example.operation
  format.value = example.format
  input.value = example.input
  inputEncoding.value = example.inputEncoding
  level.value = example.level
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
    slug="gzip-deflate"
    :name="t('tools.gzip-deflate.name')"
    name-en="GZip &amp; Deflate"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.gzip-deflate.fields.input') }}</span>
            <textarea
              v-model="input"
              class="field__textarea tb-mono"
              rows="6"
              spellcheck="false"
              :aria-label="t('tools.gzip-deflate.fields.input')"
              @keydown="onKeydown"
            />
          </label>

          <div class="field-row">
            <label class="field">
              <span class="field__label">{{ t('tools.gzip-deflate.fields.inputEncoding') }}</span>
              <EncodingSelect
                v-model="inputEncoding"
                :options="TRANSFER_ENCODINGS"
              />
            </label>
          </div>

          <p class="field__hint">
            {{ t('tools.gzip-deflate.hints.input') }}
          </p>
        </template>

        <template #options>
          <div class="field-row">
            <label class="field">
              <span class="field__label">{{ t('tools.gzip-deflate.fields.operation') }}</span>
              <select
                v-model="operation"
                class="field__select"
              >
                <option value="compress">{{ t('tools.gzip-deflate.options.operation.compress') }}</option>
                <option value="decompress">{{ t('tools.gzip-deflate.options.operation.decompress') }}</option>
              </select>
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.gzip-deflate.fields.format') }}</span>
              <select
                v-model="format"
                class="field__select"
              >
                <option value="gzip">{{ t('tools.gzip-deflate.options.format.gzip') }}</option>
                <option value="deflate">{{ t('tools.gzip-deflate.options.format.deflate') }}</option>
                <option value="raw">{{ t('tools.gzip-deflate.options.format.raw') }}</option>
              </select>
            </label>

            <label
              v-if="operation === 'compress'"
              class="field"
            >
              <span class="field__label">{{ t('tools.gzip-deflate.fields.level') }}</span>
              <select
                v-model.number="level"
                class="field__select"
              >
                <option
                  v-for="item in LEVEL_OPTIONS"
                  :key="item"
                  :value="item"
                >
                  {{ item }}
                </option>
              </select>
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.gzip-deflate.fields.outputEncoding') }}</span>
              <EncodingSelect
                v-model="outputEncoding"
                :options="TRANSFER_ENCODINGS"
              />
            </label>
          </div>

          <p class="field__hint">
            {{ t('tools.gzip-deflate.hints.format') }}
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
        :title="t('tools.gzip-deflate.results.output')"
        @retry="run"
      >
        <template
          v-if="data"
          #extra
        >
          <dl class="detail">
            <div class="detail__row">
              <dt>{{ t('tools.gzip-deflate.results.encoding') }}</dt>
              <dd>{{ data?.encoding }}</dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.gzip-deflate.results.ratio') }}</dt>
              <dd>{{ ratioText }}</dd>
            </div>
          </dl>
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.gzip-deflate.hints.output') }}
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
