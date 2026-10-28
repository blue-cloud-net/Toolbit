<script setup lang="ts">
/**
 * 数字签名工具页。
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
import { DEFAULT_EXAMPLE } from '@/tools/crypto/signature/example'

/** 后端 signature 工具的响应。 */
interface SignatureResult {
  valid?: boolean
  signature?: string
  algorithm: string
  hash: string
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

type Encoding = 'utf8' | 'hex' | 'base64' | 'base64url'

const { t, tm } = useI18n()

const ALGORITHMS = ['SM2', 'RSA', 'ECDSA', 'ED25519', 'ED448'] as const
/** 各算法支持的摘要（与后端 cryptosvc.SignHashes 保持一致）。 */
const HASHES: Record<string, readonly string[]> = {
  RSA: ['SHA256', 'SHA384', 'SHA512'],
  SM2: [],
  ECDSA: [],
  ED25519: [],
  ED448: [],
}

const algorithm = ref<string>('SM2')
const operation = ref<string>('sign')
const hash = ref<string>('SHA256')
const rsaPadding = ref<string>('pss')
const sm2UserId = ref('')
const message = ref('')
const messageEncoding = ref<Encoding>('utf8')
const key = ref('')
const keyFormat = ref('auto')
const keyPassphrase = ref('')
const publicKey = ref('')
const publicKeyFormat = ref('auto')
const signature = ref('')
const signatureEncoding = ref<Encoding>('base64')

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
} = useToolRun<SignatureResult>({ slug: 'signature' })

const running = computed(() => state.value === 'loading')

/** 当前算法支持的摘要。 */
const availableHashes = computed(() => HASHES[algorithm.value] ?? [])

/** 是否显示 RSA 专属参数。 */
const isRSA = computed(() => algorithm.value === 'RSA')

/** 是否显示 SM2 专属参数。 */
const isSM2 = computed(() => algorithm.value === 'SM2')

/** 结果文本。 */
const resultText = computed(() => {
  const d = data.value
  if (!d) {
    return ''
  }
  if (d.signature) {
    return d.signature
  }
  if (typeof d.valid === 'boolean') {
    return d.valid ? t('tools.signature.results.valid') : t('tools.signature.results.invalid')
  }
  return ''
})

const faq = computed<FaqItem[]>(() => tm('tools.signature.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  const payload: Record<string, unknown> = {
    algorithm: algorithm.value,
    operation: operation.value,
    message: { data: message.value, encoding: messageEncoding.value },
  }
  if (isRSA.value) {
    payload.hash = hash.value
    payload.rsaPadding = rsaPadding.value
  }
  if (isSM2.value && sm2UserId.value !== '') {
    payload.sm2UserId = sm2UserId.value
  }
  if (operation.value === 'sign') {
    payload.key = { data: key.value, format: keyFormat.value, passphrase: keyPassphrase.value }
  } else {
    payload.publicKey = { data: publicKey.value, format: publicKeyFormat.value }
    payload.signature = { data: signature.value, encoding: signatureEncoding.value }
  }
  void execute(payload)
}

/** 重置表单与结果。 */
function reset(): void {
  algorithm.value = 'SM2'
  operation.value = 'sign'
  hash.value = 'SHA256'
  rsaPadding.value = 'pss'
  sm2UserId.value = ''
  message.value = ''
  messageEncoding.value = 'utf8'
  key.value = ''
  keyFormat.value = 'auto'
  keyPassphrase.value = ''
  publicKey.value = ''
  publicKeyFormat.value = 'auto'
  signature.value = ''
  signatureEncoding.value = 'base64'
  resetRun()
}

/** 填入示例。 */
function fillExample(): void {
  algorithm.value = DEFAULT_EXAMPLE.algorithm
  operation.value = DEFAULT_EXAMPLE.operation
  message.value = DEFAULT_EXAMPLE.message
  messageEncoding.value = DEFAULT_EXAMPLE.messageEncoding
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="signature"
    :name="t('tools.signature.name')"
    name-en="Digital Signature"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.signature.fields.message') }}</span>
            <textarea
              v-model="message"
              class="field__textarea tb-mono"
              rows="5"
              spellcheck="false"
              :aria-label="t('tools.signature.fields.message')"
            />
          </label>

          <KeyInput
            v-if="operation === 'sign'"
            v-model="key"
            v-model:format="keyFormat"
            v-model:passphrase="keyPassphrase"
            :label="t('tools.signature.fields.key')"
            show-passphrase
          />

          <template v-else>
            <KeyInput
              v-model="publicKey"
              v-model:format="publicKeyFormat"
              v-model:passphrase="keyPassphrase"
              :label="t('tools.signature.fields.publicKey')"
            />

            <label class="field">
              <span class="field__label">{{ t('tools.signature.fields.signature') }}</span>
              <textarea
                v-model="signature"
                class="field__textarea tb-mono"
                rows="3"
                spellcheck="false"
                :aria-label="t('tools.signature.fields.signature')"
              />
            </label>
          </template>
        </template>

        <template #options>
          <div class="field-row">
            <label class="field">
              <span class="field__label">{{ t('tools.signature.fields.algorithm') }}</span>
              <select
                v-model="algorithm"
                class="field__select"
              >
                <option
                  v-for="name in ALGORITHMS"
                  :key="name"
                  :value="name"
                >
                  {{ t(`tools.signature.options.algorithm.${name}`) }}
                </option>
              </select>
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.signature.fields.operation') }}</span>
              <select
                v-model="operation"
                class="field__select"
              >
                <option value="sign">{{ t('tools.signature.options.operation.sign') }}</option>
                <option value="verify">{{ t('tools.signature.options.operation.verify') }}</option>
              </select>
            </label>

            <label
              v-if="availableHashes.length > 0"
              class="field"
            >
              <span class="field__label">{{ t('tools.signature.fields.hash') }}</span>
              <select
                v-model="hash"
                class="field__select"
              >
                <option
                  v-for="h in availableHashes"
                  :key="h"
                  :value="h"
                >
                  {{ t(`tools.signature.options.hash.${h}`) }}
                </option>
              </select>
            </label>

            <label
              v-if="isRSA"
              class="field"
            >
              <span class="field__label">{{ t('tools.signature.fields.rsaPadding') }}</span>
              <select
                v-model="rsaPadding"
                class="field__select"
              >
                <option value="pss">{{ t('tools.signature.options.rsaPadding.pss') }}</option>
                <option value="pkcs1">{{ t('tools.signature.options.rsaPadding.pkcs1') }}</option>
              </select>
            </label>

            <label
              v-if="isSM2"
              class="field"
            >
              <span class="field__label">{{ t('tools.signature.fields.sm2UserId') }}</span>
              <input
                v-model="sm2UserId"
                type="text"
                class="field__input tb-mono"
                autocomplete="off"
                :aria-label="t('tools.signature.fields.sm2UserId')"
              >
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.signature.fields.messageEncoding') }}</span>
              <EncodingSelect
                v-model="messageEncoding"
                :options="['utf8', 'hex', 'base64', 'base64url']"
              />
            </label>

            <label
              v-if="operation === 'verify'"
              class="field"
            >
              <span class="field__label">{{ t('tools.signature.fields.signatureEncoding') }}</span>
              <EncodingSelect
                v-model="signatureEncoding"
                :options="['base64', 'hex', 'base64url', 'utf8']"
              />
            </label>
          </div>
          <p class="field__hint">
            {{ t('tools.signature.hints.hash') }}
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
        :title="t('tools.signature.results.output')"
        @retry="run"
      />
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.signature.description') }}
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
