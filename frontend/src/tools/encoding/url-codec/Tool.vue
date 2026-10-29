<script setup lang="ts">
/**
 * URL 编解码工具页。
 *
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
 */
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useAutoRun } from '@/composables/useAutoRun'
import { useToolRun } from '@/composables/useToolRun'
import { EXAMPLES } from '@/tools/encoding/url-codec/example'

/** URL 解析结果。 */
interface URLParts {
  scheme: string
  host: string
  port: string
  path: string
  query: string
  params: { key: string; value: string }[]
}

/** 后端 url-codec 工具的响应。 */
interface URLCodecResult {
  output: string
  parsed?: URLParts
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

type Operation = 'encode' | 'decode' | 'parse'
type Component = 'full' | 'path' | 'query' | 'fragment'

const { t, tm } = useI18n()

const operation = ref<Operation>('encode')
const input = ref('')
const component = ref<Component>('full')
const plusAsSpace = ref(true)

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
  bindInput,
  useAsInput,
} = useToolRun<URLCodecResult>({ slug: 'url-codec' })

/** 结果区的「作为新输入」：把结果回填到主输入框（仅内存，不写存储）。 */
bindInput((text) => {
  input.value = text
})

const resultText = computed(() => data.value?.output ?? '')
const running = computed(() => state.value === 'loading')

const parsed = computed(() => data.value?.parsed ?? null)

const faq = computed<FaqItem[]>(() => tm('tools.url-codec.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  void execute({
    operation: operation.value,
    input: { data: input.value, encoding: 'utf8' },
    component: component.value,
    plusAsSpace: plusAsSpace.value,
  })
}

/** 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。 */
useAutoRun({
  slug: 'url-codec',
  run,
  watch: [operation, input, component, plusAsSpace],
  enabled: () => input.value !== '',
})

/** 重置表单与结果。 */
function reset(): void {
  operation.value = 'encode'
  input.value = ''
  component.value = 'full'
  plusAsSpace.value = true
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  operation.value = example.operation
  input.value = example.input
  component.value = example.component
  plusAsSpace.value = example.plusAsSpace
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
    slug="url-codec"
    :name="t('tools.url-codec.name')"
    name-en="URL Codec"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.url-codec.fields.input') }}</span>
            <textarea
              v-model="input"
              class="field__textarea tb-mono"
              rows="5"
              spellcheck="false"
              :aria-label="t('tools.url-codec.fields.input')"
              @keydown="onKeydown"
            />
          </label>

          <p class="field__hint">
            {{ t('tools.url-codec.hints.input') }}
          </p>
        </template>

        <template #options>
          <div class="field-row">
            <label class="field">
              <span class="field__label">{{ t('tools.url-codec.fields.operation') }}</span>
              <select
                v-model="operation"
                class="field__select"
              >
                <option value="encode">{{ t('tools.url-codec.options.operation.encode') }}</option>
                <option value="decode">{{ t('tools.url-codec.options.operation.decode') }}</option>
                <option value="parse">{{ t('tools.url-codec.options.operation.parse') }}</option>
              </select>
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.url-codec.fields.component') }}</span>
              <select
                v-model="component"
                class="field__select"
              >
                <option value="full">{{ t('tools.url-codec.options.component.full') }}</option>
                <option value="path">{{ t('tools.url-codec.options.component.path') }}</option>
                <option value="query">{{ t('tools.url-codec.options.component.query') }}</option>
                <option value="fragment">{{ t('tools.url-codec.options.component.fragment') }}</option>
              </select>
            </label>
          </div>

          <label
            v-if="operation === 'decode'"
            class="field field--inline"
          >
            <input
              v-model="plusAsSpace"
              type="checkbox"
              class="field__checkbox"
            >
            <span>{{ t('tools.url-codec.fields.plusAsSpace') }}</span>
          </label>
          <p
            v-if="operation === 'decode'"
            class="field__hint"
          >
            {{ t('tools.url-codec.hints.plusAsSpace') }}
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
        :title="t('tools.url-codec.results.output')"
        :use-as-input="useAsInput"
        @retry="run"
      >
        <template
          v-if="parsed"
          #extra
        >
          <dl class="parts">
            <div class="parts__row">
              <dt>
                {{ t('tools.url-codec.results.parts.scheme') }}
              </dt>
              <dd class="tb-mono">
                {{ parsed.scheme || '—' }}
              </dd>
            </div>
            <div class="parts__row">
              <dt>
                {{ t('tools.url-codec.results.parts.host') }}
              </dt>
              <dd class="tb-mono">
                {{ parsed.host || '—' }}
              </dd>
            </div>
            <div class="parts__row">
              <dt>
                {{ t('tools.url-codec.results.parts.port') }}
              </dt>
              <dd class="tb-mono">
                {{ parsed.port || '—' }}
              </dd>
            </div>
            <div class="parts__row">
              <dt>
                {{ t('tools.url-codec.results.parts.path') }}
              </dt>
              <dd class="tb-mono">
                {{ parsed.path || '—' }}
              </dd>
            </div>
            <div class="parts__row">
              <dt>
                {{ t('tools.url-codec.results.parts.query') }}
              </dt>
              <dd class="tb-mono">
                {{ parsed.query || '—' }}
              </dd>
            </div>
            <div
              v-for="param in parsed.params"
              :key="param.key"
              class="parts__row"
            >
              <dt class="tb-mono">
                {{ param.key }}
              </dt>
              <dd class="tb-mono">
                {{ param.value }}
              </dd>
            </div>
          </dl>
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.url-codec.hints.output') }}
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
.parts {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-1);
  margin: var(--tb-space-3) 0 0;
}

.parts__row {
  display: flex;
  gap: var(--tb-space-2);
  font-size: var(--tb-font-sm);
}

.parts__row dt {
  min-width: 96px;
  color: var(--tb-text-tertiary);
}

.parts__row dd {
  margin: 0;
  word-break: break-all;
}
</style>
