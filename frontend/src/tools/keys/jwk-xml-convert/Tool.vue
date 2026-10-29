<script setup lang="ts">
/**
 * JWK / PEM / XML 转换工具页。
 *
 * 密钥内容只存在于组件局部状态；离开页面即丢弃，
 * **不写**任何浏览器存储（AGENTS.md §4.11）。
 */
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useToolRun } from '@/composables/useToolRun'
import { DEFAULT_EXAMPLE } from '@/tools/keys/jwk-xml-convert/example'

/** 后端 jwk-xml-convert 工具的响应。 */
interface JWKXMLConvertResult {
  output: { format: string; data: string }
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const FORMATS = ['jwk', 'pem', 'xml'] as const
const KEY_TYPES = ['private', 'public'] as const

const inputFormat = ref<string>('pem')
const inputData = ref('')
const to = ref<string>('jwk')
const keyType = ref<string>('public')

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
  bindInput,
  useAsInput,
} = useToolRun<JWKXMLConvertResult>({ slug: 'jwk-xml-convert' })

/** 结果区的「作为新输入」：把结果回填到主输入框（仅内存，不写存储）。 */
bindInput((text) => {
  inputData.value = text
})

const running = computed(() => state.value === 'loading')

const resultText = computed(() => data.value?.output.data ?? '')

const faq = computed<FaqItem[]>(() => tm('tools.jwk-xml-convert.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  void execute({
    input: { format: inputFormat.value, data: inputData.value },
    to: to.value,
    keyType: keyType.value,
  })
}

/** 重置表单与结果。 */
function reset(): void {
  inputFormat.value = 'pem'
  inputData.value = ''
  to.value = 'jwk'
  keyType.value = 'public'
  resetRun()
}

/** 填入示例。 */
function fillExample(): void {
  inputFormat.value = DEFAULT_EXAMPLE.from
  inputData.value = DEFAULT_EXAMPLE.data
  to.value = DEFAULT_EXAMPLE.to
  keyType.value = DEFAULT_EXAMPLE.keyType
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="jwk-xml-convert"
    :name="t('tools.jwk-xml-convert.name')"
    name-en="JWK & XML Converter"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.jwk-xml-convert.fields.inputFormat') }}</span>
            <select
              v-model="inputFormat"
              class="field__select"
            >
              <option
                v-for="name in FORMATS"
                :key="name"
                :value="name"
              >
                {{ name }}
              </option>
            </select>
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.jwk-xml-convert.fields.inputData') }}</span>
            <textarea
              v-model="inputData"
              class="field__textarea tb-mono"
              rows="10"
              spellcheck="false"
              autocomplete="off"
              :aria-label="t('tools.jwk-xml-convert.fields.inputData')"
            />
          </label>
        </template>

        <template #options>
          <div class="field-row">
            <label class="field">
              <span class="field__label">{{ t('tools.jwk-xml-convert.fields.to') }}</span>
              <select
                v-model="to"
                class="field__select"
              >
                <option
                  v-for="name in FORMATS"
                  :key="name"
                  :value="name"
                >
                  {{ name }}
                </option>
              </select>
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.jwk-xml-convert.fields.keyType') }}</span>
              <select
                v-model="keyType"
                class="field__select"
              >
                <option
                  v-for="name in KEY_TYPES"
                  :key="name"
                  :value="name"
                >
                  {{ name }}
                </option>
              </select>
            </label>
          </div>

          <p class="field__hint">
            {{ t('tools.jwk-xml-convert.hints.support') }}
          </p>
        </template>

        <template #footer>
          <RunBar
            :running="running"
            :disabled="inputData === ''"
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
        :title="t('tools.jwk-xml-convert.results.output')"
        :use-as-input="useAsInput"
        @retry="run"
      />
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.jwk-xml-convert.description') }}
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
