<script setup lang="ts">
/**
 * CSR 生成工具页。
 *
 * 私钥与主体信息只存在于组件局部状态；离开页面即丢弃，
 * **不写**任何浏览器存储（AGENTS.md §4.11）。
 */
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import KeyInput from '@/components/tool/KeyInput.vue'
import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useToolRun } from '@/composables/useToolRun'
import { TEST_LEAF_KEY_PEM } from '@/tools/certificates/fixtures-v060'

/** 后端 csr-gen 工具的响应。 */
interface CSRGenResult {
  csrPem: string
  privateKeyPem: string
  keyType: string
  keyBits: number
  keyCurve: string
  info: {
    subjectText: string
    sans: string[]
    signatureAlgorithm: string
    challengePasswordPresent: boolean
    signatureValid: boolean
    fingerprintSha256: string
  }
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const TYPES = ['server', 'client', 'email', 'code-signing']
const ALGORITHMS = ['EC', 'RSA', 'SM2']
const CURVES = ['P-256', 'P-384', 'P-521']
const RSA_BITS = [2048, 3072, 4096]

const { t, tm } = useI18n()

const certType = ref('server')
const cn = ref('csr.example.com')
const org = ref('Toolbit Test')
const country = ref('CN')
const sans = ref('DNS:csr.example.com')
const challengePassword = ref('')

const keyType = ref('EC')
const keyCurve = ref('P-256')
const keyBits = ref(2048)

const useExisting = ref(false)
const existingKey = ref('')
const existingFormat = ref('pem')
const existingPassphrase = ref('')

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
} = useToolRun<CSRGenResult>({ slug: 'csr-gen' })

const running = computed(() => state.value === 'loading')

/** SAN 文本 → 数组。 */
function sanList(): string[] {
  return sans.value
    .split(/[\s,]+/)
    .map((item) => item.trim())
    .filter((item) => item !== '')
}

/** 结果文本。 */
const resultText = computed(() => {
  const result = data.value
  if (!result) {
    return ''
  }
  const head = [
    `${t('tools.csr-gen.results.subject')}: ${result.info.subjectText}`,
    `${t('tools.csr-gen.results.sans')}: ${result.info.sans.join(', ') || '-'}`,
    `${t('tools.csr-gen.results.key')}: ${result.keyType} ${result.keyCurve || result.keyBits}`,
    `${t('tools.csr-gen.results.signatureValid')}: ${result.info.signatureValid}`,
    `${t('tools.csr-gen.results.fingerprint')}: ${result.info.fingerprintSha256}`,
    '',
    result.csrPem,
  ]
  if (result.privateKeyPem) {
    head.push('', result.privateKeyPem)
  }
  return head.join('\n')
})

const faq = computed<FaqItem[]>(() => tm('tools.csr-gen.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  void execute({
    certType: certType.value,
    subject: { CN: cn.value, O: org.value, C: country.value },
    key: { type: keyType.value, bits: keyBits.value, curve: keyCurve.value },
    ...(useExisting.value
      ? {
          existingPrivateKey: {
            data: existingKey.value,
            format: existingFormat.value,
            passphrase: existingPassphrase.value,
          },
        }
      : {}),
    sans: sanList(),
    challengePassword: challengePassword.value,
  })
}

/** 重置表单与结果。 */
function reset(): void {
  certType.value = 'server'
  cn.value = 'csr.example.com'
  org.value = 'Toolbit Test'
  country.value = 'CN'
  sans.value = 'DNS:csr.example.com'
  challengePassword.value = ''
  keyType.value = 'EC'
  keyCurve.value = 'P-256'
  keyBits.value = 2048
  useExisting.value = false
  existingKey.value = ''
  existingFormat.value = 'pem'
  existingPassphrase.value = ''
  resetRun()
}

/** 填入示例。 */
function fillExample(): void {
  reset()
  useExisting.value = true
  existingKey.value = TEST_LEAF_KEY_PEM
}
</script>

<template>
  <ToolPageLayout
    slug="csr-gen"
    :name="t('tools.csr-gen.name')"
    name-en="CSR Generator"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.csr-gen.fields.type') }}</span>
            <select
              v-model="certType"
              class="field__select"
            >
              <option
                v-for="name in TYPES"
                :key="name"
                :value="name"
              >
                {{ t(`tools.csr-gen.options.types.${name}`) }}
              </option>
            </select>
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.csr-gen.fields.cn') }}</span>
            <input
              v-model="cn"
              type="text"
              class="field__input"
              spellcheck="false"
              autocomplete="off"
            >
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.csr-gen.fields.org') }}</span>
            <input
              v-model="org"
              type="text"
              class="field__input"
              spellcheck="false"
              autocomplete="off"
            >
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.csr-gen.fields.country') }}</span>
            <input
              v-model="country"
              type="text"
              class="field__input"
              maxlength="2"
              spellcheck="false"
              autocomplete="off"
            >
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.csr-gen.fields.sans') }}</span>
            <textarea
              v-model="sans"
              class="field__textarea tb-mono"
              rows="3"
              spellcheck="false"
            />
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.csr-gen.fields.challengePassword') }}</span>
            <input
              v-model="challengePassword"
              type="text"
              class="field__input"
              autocomplete="off"
            >
          </label>

          <label class="field field--inline">
            <input
              v-model="useExisting"
              type="checkbox"
            >
            <span>{{ t('tools.csr-gen.fields.useExisting') }}</span>
          </label>

          <template v-if="!useExisting">
            <label class="field">
              <span class="field__label">{{ t('tools.csr-gen.fields.algorithm') }}</span>
              <select
                v-model="keyType"
                class="field__select"
              >
                <option
                  v-for="name in ALGORITHMS"
                  :key="name"
                  :value="name"
                >
                  {{ name }}
                </option>
              </select>
            </label>

            <label
              v-if="keyType === 'EC'"
              class="field"
            >
              <span class="field__label">{{ t('tools.csr-gen.fields.curve') }}</span>
              <select
                v-model="keyCurve"
                class="field__select"
              >
                <option
                  v-for="name in CURVES"
                  :key="name"
                  :value="name"
                >
                  {{ name }}
                </option>
              </select>
            </label>

            <label
              v-if="keyType === 'RSA'"
              class="field"
            >
              <span class="field__label">{{ t('tools.csr-gen.fields.bits') }}</span>
              <select
                v-model.number="keyBits"
                class="field__select"
              >
                <option
                  v-for="bits in RSA_BITS"
                  :key="bits"
                  :value="bits"
                >
                  {{ bits }}
                </option>
              </select>
            </label>
          </template>

          <KeyInput
            v-else
            v-model="existingKey"
            v-model:format="existingFormat"
            v-model:passphrase="existingPassphrase"
            :label="t('tools.csr-gen.fields.existingPrivateKey')"
            :formats="['pem', 'base64', 'hex']"
            show-passphrase
          />
        </template>

        <template #footer>
          <RunBar
            :running="running"
            :disabled="cn === ''"
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
        :title="t('tools.csr-gen.results.output')"
        @retry="run"
      />
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.csr-gen.description') }}
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
