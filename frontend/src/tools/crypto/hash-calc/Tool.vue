<script setup lang="ts">
/**
 * 摘要计算工具页。
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
import { DEFAULT_EXAMPLE } from '@/tools/crypto/hash-calc/example'

/** 后端 hash-calc 工具的响应。 */
interface HashCalcResult {
  digests: { algorithm: string; value: string }[]
  inputBytes: number
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

type InputEncoding = 'utf8' | 'hex'

const { t, tm } = useI18n()

const ALGORITHMS = ['SM3', 'SHA256', 'SHA384', 'SHA512', 'SHA224', 'SHA1', 'MD5'] as const

const algorithms = ref<string[]>(['SM3', 'SHA256'])
const input = ref('')
const inputEncoding = ref<InputEncoding>('utf8')
const outputEncoding = ref('hex')

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
} = useToolRun<HashCalcResult>({ slug: 'hash-calc' })

const running = computed(() => state.value === 'loading')

/** 结果文本：每行一个「算法: 值」。 */
const resultText = computed(() =>
  (data.value?.digests ?? []).map((d) => `${d.algorithm}: ${d.value}`).join('\n'),
)

const faq = computed<FaqItem[]>(() => tm('tools.hash-calc.faq') as FaqItem[])

/** 切换某个算法是否参与计算。 */
function toggleAlgorithm(name: string): void {
  const index = algorithms.value.indexOf(name)
  if (index >= 0) {
    algorithms.value.splice(index, 1)
  } else {
    algorithms.value.push(name)
  }
}

/** 运行当前输入。 */
function run(): void {
  void execute({
    algorithms: algorithms.value,
    input: { data: input.value, encoding: inputEncoding.value },
    outputEncoding: outputEncoding.value,
  })
}

/** 重置表单与结果。 */
function reset(): void {
  algorithms.value = ['SM3', 'SHA256']
  input.value = ''
  inputEncoding.value = 'utf8'
  outputEncoding.value = 'hex'
  resetRun()
}

/** 填入示例。 */
function fillExample(): void {
  algorithms.value = [...DEFAULT_EXAMPLE.algorithms]
  input.value = DEFAULT_EXAMPLE.input
  inputEncoding.value = DEFAULT_EXAMPLE.inputEncoding
  outputEncoding.value = 'hex'
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
    slug="hash-calc"
    :name="t('tools.hash-calc.name')"
    name-en="Hash Calculator"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.hash-calc.fields.input') }}</span>
            <textarea
              v-model="input"
              class="field__textarea tb-mono"
              rows="6"
              spellcheck="false"
              :aria-label="t('tools.hash-calc.fields.input')"
              @keydown="onKeydown"
            />
          </label>

          <div class="field-row">
            <label class="field">
              <span class="field__label">{{ t('tools.hash-calc.fields.inputEncoding') }}</span>
              <EncodingSelect
                v-model="inputEncoding"
                :options="['utf8', 'hex']"
              />
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.hash-calc.fields.outputEncoding') }}</span>
              <EncodingSelect
                v-model="outputEncoding"
                :options="['hex', 'base64', 'base64url']"
              />
            </label>
          </div>

          <p class="field__hint">
            {{ t('tools.hash-calc.hints.input') }}
          </p>
        </template>

        <template #options>
          <fieldset class="field">
            <legend class="field__label">
              {{ t('tools.hash-calc.fields.algorithms') }}
            </legend>
            <div class="algorithm-list">
              <label
                v-for="name in ALGORITHMS"
                :key="name"
                class="field field--inline"
              >
                <input
                  type="checkbox"
                  class="field__checkbox"
                  :checked="algorithms.includes(name)"
                  @change="toggleAlgorithm(name)"
                >
                <span>{{ t(`tools.hash-calc.options.algorithms.${name}`) }}</span>
              </label>
            </div>
          </fieldset>
          <p class="field__hint">
            {{ t('tools.hash-calc.hints.algorithms') }}
          </p>
        </template>

        <template #footer>
          <RunBar
            :running="running"
            :disabled="algorithms.length === 0"
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
        :title="t('tools.hash-calc.results.digests')"
        @retry="run"
      />
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.hash-calc.description') }}
      </p>
    </template>

    <template #faq>
      <dl class="faq">
        <template
          v-for="item in faq"
          :key="item.q"
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
.algorithm-list {
  display: flex;
  flex-wrap: wrap;
  gap: var(--tb-space-2) var(--tb-space-4);
}
</style>
