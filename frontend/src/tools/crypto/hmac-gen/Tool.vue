<script setup lang="ts">
/**
 * HMAC 计算工具页。
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
import { DEFAULT_EXAMPLE } from '@/tools/crypto/hmac-gen/example'

/** 后端 hmac-gen 工具的响应。 */
interface HmacGenResult {
  mac: string
  algorithm: string
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

type Encoding = 'utf8' | 'hex' | 'base64' | 'base64url'

const { t, tm } = useI18n()

const ALGORITHMS = [
  'HMAC-SM3',
  'HMAC-SHA256',
  'HMAC-SHA384',
  'HMAC-SHA512',
  'HMAC-SHA224',
  'HMAC-SHA1',
  'HMAC-MD5',
] as const

const algorithm = ref<string>('HMAC-SM3')
const key = ref('')
const keyEncoding = ref<Encoding>('hex')
const message = ref('')
const messageEncoding = ref<Encoding>('utf8')
const outputEncoding = ref<string>('hex')

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
  bindInput,
  useAsInput,
} = useToolRun<HmacGenResult>({ slug: 'hmac-gen' })

/** 结果区的「作为新输入」：把结果回填到主输入框（仅内存，不写存储）。 */
bindInput((text) => {
  message.value = text
})

const running = computed(() => state.value === 'loading')
const resultText = computed(() => data.value?.mac ?? '')

const faq = computed<FaqItem[]>(() => tm('tools.hmac-gen.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  void execute({
    algorithm: algorithm.value,
    key: { data: key.value, encoding: keyEncoding.value },
    message: { data: message.value, encoding: messageEncoding.value },
    outputEncoding: outputEncoding.value,
  })
}

/** 重置表单与结果。 */
function reset(): void {
  algorithm.value = 'HMAC-SM3'
  key.value = ''
  keyEncoding.value = 'hex'
  message.value = ''
  messageEncoding.value = 'utf8'
  outputEncoding.value = 'hex'
  resetRun()
}

/** 填入示例。 */
function fillExample(): void {
  algorithm.value = DEFAULT_EXAMPLE.algorithm
  key.value = DEFAULT_EXAMPLE.key
  keyEncoding.value = DEFAULT_EXAMPLE.keyEncoding
  message.value = DEFAULT_EXAMPLE.message
  messageEncoding.value = DEFAULT_EXAMPLE.messageEncoding
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
    slug="hmac-gen"
    :name="t('tools.hmac-gen.name')"
    name-en="HMAC Generator"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.hmac-gen.fields.key') }}</span>
            <textarea
              v-model="key"
              class="field__textarea tb-mono"
              rows="3"
              spellcheck="false"
              autocomplete="off"
              :aria-label="t('tools.hmac-gen.fields.key')"
              @keydown="onKeydown"
            />
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.hmac-gen.fields.message') }}</span>
            <textarea
              v-model="message"
              class="field__textarea tb-mono"
              rows="6"
              spellcheck="false"
              :aria-label="t('tools.hmac-gen.fields.message')"
              @keydown="onKeydown"
            />
          </label>

          <div class="field-row">
            <label class="field">
              <span class="field__label">{{ t('tools.hmac-gen.fields.keyEncoding') }}</span>
              <EncodingSelect
                v-model="keyEncoding"
                :options="['utf8', 'hex', 'base64', 'base64url']"
              />
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.hmac-gen.fields.messageEncoding') }}</span>
              <EncodingSelect
                v-model="messageEncoding"
                :options="['utf8', 'hex', 'base64', 'base64url']"
              />
            </label>
          </div>
        </template>

        <template #options>
          <div class="field-row">
            <label class="field">
              <span class="field__label">{{ t('tools.hmac-gen.fields.algorithm') }}</span>
              <select
                v-model="algorithm"
                class="field__select"
              >
                <option
                  v-for="name in ALGORITHMS"
                  :key="name"
                  :value="name"
                >
                  {{ t(`tools.hmac-gen.options.algorithm.${name}`) }}
                </option>
              </select>
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.hmac-gen.fields.outputEncoding') }}</span>
              <EncodingSelect
                v-model="outputEncoding"
                :options="['hex', 'base64', 'base64url']"
              />
            </label>
          </div>
          <p class="field__hint">
            {{ t('tools.hmac-gen.hints.algorithm') }}
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
        :title="t('tools.hmac-gen.results.mac')"
        :use-as-input="useAsInput"
        @retry="run"
      />
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.hmac-gen.hints.key') }}
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
