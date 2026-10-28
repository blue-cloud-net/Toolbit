<script setup lang="ts">
/**
 * 对称加解密工具页。
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
import { DEFAULT_EXAMPLE } from '@/tools/crypto/symmetric-crypt/example'

/** 后端 symmetric-crypt 工具的响应。 */
interface SymmetricCryptResult {
  output: string
  tag?: string
  iv?: string
  algorithm: string
  mode: string
  padding: string
  keyBytes: number
  ivBytes: number
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

type Encoding = 'utf8' | 'hex' | 'base64' | 'base64url'

const { t, tm } = useI18n()

/** 各算法支持的模式（与后端 cryptosvc.SymModes 保持一致）。 */
const MODES: Record<string, readonly string[]> = {
  SM4: ['ECB', 'CBC', 'CTR', 'OFB', 'CFB', 'GCM'],
  AES: ['ECB', 'CBC', 'CTR', 'GCM'],
}

const algorithm = ref<string>('SM4')
const operation = ref<string>('encrypt')
const mode = ref<string>('CBC')
const padding = ref<string>('pkcs7')
const key = ref('')
const keyEncoding = ref<Encoding>('hex')
const iv = ref('')
const ivEncoding = ref<Encoding>('hex')
const aad = ref('')
const aadEncoding = ref<Encoding>('utf8')
const tag = ref('')
const tagEncoding = ref<Encoding>('hex')
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
} = useToolRun<SymmetricCryptResult>({ slug: 'symmetric-crypt' })

const running = computed(() => state.value === 'loading')

/** 当前算法支持的模式。 */
const availableModes = computed(() => MODES[algorithm.value] ?? [])

/** 填充选项：只有 SM4 的 ECB / CBC 支持 zero / none。 */
const availablePaddings = computed(() => {
  if (algorithm.value === 'SM4' && (mode.value === 'ECB' || mode.value === 'CBC')) {
    return ['pkcs7', 'zero', 'none']
  }
  return ['pkcs7']
})

/** 是否需要 IV。 */
const needsIV = computed(() => mode.value !== 'ECB')

/** 是否需要认证标签。 */
const needsTag = computed(() => mode.value === 'GCM' && operation.value === 'decrypt')

/** 结果文本：密文 + 标签 + 自动生成的 nonce。 */
const resultText = computed(() => {
  const d = data.value
  if (!d) {
    return ''
  }
  const lines = [d.output]
  if (d.tag) {
    lines.push(`${t('tools.symmetric-crypt.results.tag')}: ${d.tag}`)
  }
  if (d.iv) {
    lines.push(`${t('tools.symmetric-crypt.results.iv')}: ${d.iv}`)
  }
  return lines.join('\n')
})

const faq = computed<FaqItem[]>(() => tm('tools.symmetric-crypt.faq') as FaqItem[])

/** 切换算法时校正模式与填充。 */
function onAlgorithmChange(): void {
  if (!availableModes.value.includes(mode.value)) {
    mode.value = availableModes.value[0] ?? 'CBC'
  }
  if (!availablePaddings.value.includes(padding.value)) {
    padding.value = 'pkcs7'
  }
}

/** 运行当前输入。 */
function run(): void {
  const payload: Record<string, unknown> = {
    algorithm: algorithm.value,
    operation: operation.value,
    mode: mode.value,
    padding: padding.value,
    key: { data: key.value, encoding: keyEncoding.value },
    input: { data: input.value, encoding: inputEncoding.value },
    outputEncoding: outputEncoding.value,
  }
  if (needsIV.value && iv.value !== '') {
    payload.iv = { data: iv.value, encoding: ivEncoding.value }
  }
  if (aad.value !== '') {
    payload.aad = { data: aad.value, encoding: aadEncoding.value }
  }
  if (needsTag.value && tag.value !== '') {
    payload.tag = { data: tag.value, encoding: tagEncoding.value }
  }
  void execute(payload)
}

/** 重置表单与结果。 */
function reset(): void {
  algorithm.value = 'SM4'
  operation.value = 'encrypt'
  mode.value = 'CBC'
  padding.value = 'pkcs7'
  key.value = ''
  keyEncoding.value = 'hex'
  iv.value = ''
  ivEncoding.value = 'hex'
  aad.value = ''
  aadEncoding.value = 'utf8'
  tag.value = ''
  tagEncoding.value = 'hex'
  input.value = ''
  inputEncoding.value = 'utf8'
  outputEncoding.value = 'base64'
  resetRun()
}

/** 填入示例。 */
function fillExample(): void {
  algorithm.value = DEFAULT_EXAMPLE.algorithm
  operation.value = DEFAULT_EXAMPLE.operation
  mode.value = DEFAULT_EXAMPLE.mode
  padding.value = DEFAULT_EXAMPLE.padding
  key.value = DEFAULT_EXAMPLE.key
  keyEncoding.value = 'hex'
  iv.value = DEFAULT_EXAMPLE.iv
  ivEncoding.value = 'hex'
  input.value = DEFAULT_EXAMPLE.input
  inputEncoding.value = DEFAULT_EXAMPLE.inputEncoding
  outputEncoding.value = 'base64'
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="symmetric-crypt"
    :name="t('tools.symmetric-crypt.name')"
    name-en="Symmetric Encryption"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.symmetric-crypt.fields.key') }}</span>
            <textarea
              v-model="key"
              class="field__textarea tb-mono"
              rows="2"
              spellcheck="false"
              autocomplete="off"
              :aria-label="t('tools.symmetric-crypt.fields.key')"
            />
          </label>

          <label
            v-if="needsIV"
            class="field"
          >
            <span class="field__label">{{ t('tools.symmetric-crypt.fields.iv') }}</span>
            <input
              v-model="iv"
              type="text"
              class="field__input tb-mono"
              autocomplete="off"
              :aria-label="t('tools.symmetric-crypt.fields.iv')"
            >
          </label>

          <label
            v-if="mode === 'GCM'"
            class="field"
          >
            <span class="field__label">{{ t('tools.symmetric-crypt.fields.aad') }}</span>
            <input
              v-model="aad"
              type="text"
              class="field__input tb-mono"
              autocomplete="off"
              :aria-label="t('tools.symmetric-crypt.fields.aad')"
            >
          </label>

          <label
            v-if="needsTag"
            class="field"
          >
            <span class="field__label">{{ t('tools.symmetric-crypt.fields.tag') }}</span>
            <input
              v-model="tag"
              type="text"
              class="field__input tb-mono"
              autocomplete="off"
              :aria-label="t('tools.symmetric-crypt.fields.tag')"
            >
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.symmetric-crypt.fields.input') }}</span>
            <textarea
              v-model="input"
              class="field__textarea tb-mono"
              rows="5"
              spellcheck="false"
              :aria-label="t('tools.symmetric-crypt.fields.input')"
            />
          </label>
        </template>

        <template #options>
          <div class="field-row">
            <label class="field">
              <span class="field__label">{{ t('tools.symmetric-crypt.fields.algorithm') }}</span>
              <select
                v-model="algorithm"
                class="field__select"
                @change="onAlgorithmChange"
              >
                <option value="SM4">{{ t('tools.symmetric-crypt.options.algorithm.SM4') }}</option>
                <option value="AES">{{ t('tools.symmetric-crypt.options.algorithm.AES') }}</option>
              </select>
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.symmetric-crypt.fields.operation') }}</span>
              <select
                v-model="operation"
                class="field__select"
              >
                <option value="encrypt">{{ t('tools.symmetric-crypt.options.operation.encrypt') }}</option>
                <option value="decrypt">{{ t('tools.symmetric-crypt.options.operation.decrypt') }}</option>
              </select>
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.symmetric-crypt.fields.mode') }}</span>
              <select
                v-model="mode"
                class="field__select"
                @change="onAlgorithmChange"
              >
                <option
                  v-for="m in availableModes"
                  :key="m"
                  :value="m"
                >
                  {{ t(`tools.symmetric-crypt.options.mode.${m}`) }}
                </option>
              </select>
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.symmetric-crypt.fields.padding') }}</span>
              <select
                v-model="padding"
                class="field__select"
              >
                <option
                  v-for="p in availablePaddings"
                  :key="p"
                  :value="p"
                >
                  {{ t(`tools.symmetric-crypt.options.padding.${p}`) }}
                </option>
              </select>
            </label>
          </div>

          <div class="field-row">
            <label class="field">
              <span class="field__label">{{ t('tools.symmetric-crypt.fields.keyEncoding') }}</span>
              <EncodingSelect
                v-model="keyEncoding"
                :options="['utf8', 'hex', 'base64', 'base64url']"
              />
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.symmetric-crypt.fields.inputEncoding') }}</span>
              <EncodingSelect
                v-model="inputEncoding"
                :options="['utf8', 'hex', 'base64', 'base64url']"
              />
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.symmetric-crypt.fields.outputEncoding') }}</span>
              <EncodingSelect
                v-model="outputEncoding"
                :options="['base64', 'hex', 'base64url', 'utf8']"
              />
            </label>
          </div>

          <p class="field__hint">
            {{ t('tools.symmetric-crypt.hints.mode') }}
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
        :title="t('tools.symmetric-crypt.results.output')"
        @retry="run"
      />
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.symmetric-crypt.description') }}
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
