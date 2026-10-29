<script setup lang="ts">
/**
 * 非对称加解密工具页。
 *
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
 */
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import EncodingSelect from '@/components/tool/EncodingSelect.vue'
import KeyInput from '@/components/tool/KeyInput.vue'
import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useToolRun } from '@/composables/useToolRun'
import { DEFAULT_EXAMPLE } from '@/tools/crypto/asymmetric-crypt/example'

/** 后端 asymmetric-crypt 工具的响应。 */
interface AsymmetricCryptResult {
  output: string
  algorithm: string
  order?: string
  format?: string
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

type Encoding = 'utf8' | 'hex' | 'base64' | 'base64url'

const { t, tm } = useI18n()

const algorithm = ref<string>('SM2')
const operation = ref<string>('encrypt')
const rsaPadding = ref<string>('oaep')
const oaepHash = ref<string>('SHA256')
const key = ref('')
const keyFormat = ref('auto')
const keyPassphrase = ref('')
const input = ref('')
const inputEncoding = ref<Encoding>('utf8')
const outputEncoding = ref<string>('base64')

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
  bindInput,
  useAsInput,
} = useToolRun<AsymmetricCryptResult>({ slug: 'asymmetric-crypt' })

/** 结果区的「作为新输入」：把结果回填到主输入框（仅内存，不写存储）。 */
bindInput((text) => {
  input.value = text
})

const running = computed(() => state.value === 'loading')
const isRSA = computed(() => algorithm.value === 'RSA')
const resultText = computed(() => data.value?.output ?? '')

const faq = computed<FaqItem[]>(() => tm('tools.asymmetric-crypt.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  const payload: Record<string, unknown> = {
    algorithm: algorithm.value,
    operation: operation.value,
    key: { data: key.value, format: keyFormat.value, passphrase: keyPassphrase.value },
    input: { data: input.value, encoding: inputEncoding.value },
    outputEncoding: outputEncoding.value,
  }
  if (isRSA.value) {
    payload.rsaPadding = rsaPadding.value
    payload.oaepHash = oaepHash.value
  }
  void execute(payload)
}

/** 重置表单与结果。 */
function reset(): void {
  algorithm.value = 'SM2'
  operation.value = 'encrypt'
  rsaPadding.value = 'oaep'
  oaepHash.value = 'SHA256'
  key.value = ''
  keyFormat.value = 'auto'
  keyPassphrase.value = ''
  input.value = ''
  inputEncoding.value = 'utf8'
  outputEncoding.value = 'base64'
  resetRun()
}

/** 填入示例。 */
function fillExample(): void {
  algorithm.value = DEFAULT_EXAMPLE.algorithm
  operation.value = DEFAULT_EXAMPLE.operation
  rsaPadding.value = DEFAULT_EXAMPLE.rsaPadding
  oaepHash.value = DEFAULT_EXAMPLE.oaepHash
  input.value = DEFAULT_EXAMPLE.input
  inputEncoding.value = DEFAULT_EXAMPLE.inputEncoding
  outputEncoding.value = 'base64'
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="asymmetric-crypt"
    :name="t('tools.asymmetric-crypt.name')"
    name-en="Asymmetric Encryption"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <KeyInput
            v-model="key"
            v-model:format="keyFormat"
            v-model:passphrase="keyPassphrase"
            :label="t('tools.asymmetric-crypt.fields.key')"
            show-passphrase
          />

          <label class="field">
            <span class="field__label">{{ t('tools.asymmetric-crypt.fields.input') }}</span>
            <textarea
              v-model="input"
              class="field__textarea tb-mono"
              rows="5"
              spellcheck="false"
              :aria-label="t('tools.asymmetric-crypt.fields.input')"
            />
          </label>
        </template>

        <template #options>
          <div class="field-row">
            <label class="field">
              <span class="field__label">{{ t('tools.asymmetric-crypt.fields.algorithm') }}</span>
              <select
                v-model="algorithm"
                class="field__select"
              >
                <option value="SM2">{{ t('tools.asymmetric-crypt.options.algorithm.SM2') }}</option>
                <option value="RSA">{{ t('tools.asymmetric-crypt.options.algorithm.RSA') }}</option>
              </select>
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.asymmetric-crypt.fields.operation') }}</span>
              <select
                v-model="operation"
                class="field__select"
              >
                <option value="encrypt">{{ t('tools.asymmetric-crypt.options.operation.encrypt') }}</option>
                <option value="decrypt">{{ t('tools.asymmetric-crypt.options.operation.decrypt') }}</option>
              </select>
            </label>

            <label
              v-if="isRSA"
              class="field"
            >
              <span class="field__label">{{ t('tools.asymmetric-crypt.fields.rsaPadding') }}</span>
              <select
                v-model="rsaPadding"
                class="field__select"
              >
                <option value="oaep">{{ t('tools.asymmetric-crypt.options.rsaPadding.oaep') }}</option>
                <option value="pkcs1">{{ t('tools.asymmetric-crypt.options.rsaPadding.pkcs1') }}</option>
              </select>
            </label>

            <label
              v-if="isRSA"
              class="field"
            >
              <span class="field__label">{{ t('tools.asymmetric-crypt.fields.oaepHash') }}</span>
              <select
                v-model="oaepHash"
                class="field__select"
              >
                <option value="SHA256">{{ t('tools.asymmetric-crypt.options.oaepHash.SHA256') }}</option>
                <option value="SHA384">{{ t('tools.asymmetric-crypt.options.oaepHash.SHA384') }}</option>
                <option value="SHA512">{{ t('tools.asymmetric-crypt.options.oaepHash.SHA512') }}</option>
              </select>
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.asymmetric-crypt.fields.inputEncoding') }}</span>
              <EncodingSelect
                v-model="inputEncoding"
                :options="['utf8', 'hex', 'base64', 'base64url']"
              />
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.asymmetric-crypt.fields.outputEncoding') }}</span>
              <EncodingSelect
                v-model="outputEncoding"
                :options="['base64', 'hex', 'base64url', 'utf8']"
              />
            </label>
          </div>
          <p class="field__hint">
            {{ t('tools.asymmetric-crypt.hints.algorithm') }}
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
        :title="t('tools.asymmetric-crypt.results.output')"
        :use-as-input="useAsInput"
        @retry="run"
      />
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.asymmetric-crypt.description') }}
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
