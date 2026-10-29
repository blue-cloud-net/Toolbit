<script setup lang="ts">
/**
 * 结构化数据格式转换工具页。
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
import { EXAMPLES } from '@/tools/formatting/format-convert/example'

/** 后端 format-convert 工具的响应。 */
interface FormatConvertResult {
  output: string
  from: string
  to: string
  lossy?: boolean
  lossyReason?: string
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

type DataFormat = 'json' | 'yaml' | 'toml' | 'xml'

const { t, tm } = useI18n()

const input = ref('')
const from = ref<DataFormat>('json')
const to = ref<DataFormat>('yaml')
const indent = ref(2)
const pretty = ref(true)

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
  bindInput,
  useAsInput,
} = useToolRun<FormatConvertResult>({
  slug: 'format-convert',
})

/** 结果区的「作为新输入」：把结果回填到主输入框（仅内存，不写存储）。 */
bindInput((text) => {
  input.value = text
})

const resultText = computed(() => data.value?.output ?? '')
const running = computed(() => state.value === 'loading')

/** 受支持的格式。 */
const DATA_FORMATS: readonly DataFormat[] = ['json', 'yaml', 'toml', 'xml']

/** 常用缩进宽度。 */
const INDENT_OPTIONS = [2, 4, 8] as const

const faq = computed<FaqItem[]>(() => tm('tools.format-convert.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  void execute({
    input: { data: input.value },
    from: from.value,
    to: to.value,
    indent: indent.value,
    pretty: pretty.value,
  })
}

/** 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。 */
useAutoRun({
  slug: 'format-convert',
  run,
  watch: [input, from, to, indent, pretty],
  enabled: () => input.value !== '',
})

/** 重置表单与结果。 */
function reset(): void {
  input.value = ''
  from.value = 'json'
  to.value = 'yaml'
  indent.value = 2
  pretty.value = true
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  input.value = example.input
  from.value = example.from
  to.value = example.to
  indent.value = example.indent
  pretty.value = example.pretty
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="format-convert"
    :name="t('tools.format-convert.name')"
    name-en="Format Converter"
    wide
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.format-convert.fields.input') }}</span>
            <CodeEditor
              v-model="input"
              :label="t('tools.format-convert.fields.input')"
              :max-rows="24"
              @run="run"
            />
          </label>

          <p class="field__hint">
            {{ t('tools.format-convert.hints.input') }}
          </p>
        </template>

        <template #options>
          <div class="field-row">
            <label class="field">
              <span class="field__label">{{ t('tools.format-convert.fields.from') }}</span>
              <select
                v-model="from"
                class="field__select"
              >
                <option
                  v-for="format in DATA_FORMATS"
                  :key="`from-${format}`"
                  :value="format"
                >
                  {{ t(`tools.format-convert.options.format.${format}`) }}
                </option>
              </select>
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.format-convert.fields.to') }}</span>
              <select
                v-model="to"
                class="field__select"
              >
                <option
                  v-for="format in DATA_FORMATS"
                  :key="`to-${format}`"
                  :value="format"
                >
                  {{ t(`tools.format-convert.options.format.${format}`) }}
                </option>
              </select>
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.format-convert.fields.indent') }}</span>
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
              v-model="pretty"
              type="checkbox"
              class="field__checkbox"
            >
            <span>{{ t('tools.format-convert.fields.pretty') }}</span>
          </label>
          <p class="field__hint">
            {{ t('tools.format-convert.hints.pretty') }}
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
        :title="t('tools.format-convert.results.output')"
        :use-as-input="useAsInput"
        @retry="run"
      >
        <template
          v-if="data"
          #extra
        >
          <dl class="detail">
            <div class="detail__row">
              <dt>{{ t('tools.format-convert.results.route') }}</dt>
              <dd>{{ data?.from }} → {{ data?.to }}</dd>
            </div>
            <div
              v-if="data?.lossy"
              class="detail__row"
            >
              <dt>{{ t('tools.format-convert.results.lossy') }}</dt>
              <dd>{{ data?.lossyReason || t('tools.format-convert.results.lossyHint') }}</dd>
            </div>
          </dl>
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.format-convert.hints.output') }}
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
