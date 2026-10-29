<script setup lang="ts">
/**
 * 证书签发工具页。
 *
 * 私钥与证书材料只存在于组件局部状态；离开页面即丢弃，
 * **不写**任何浏览器存储（AGENTS.md §4.11）。
 */
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import CertTextInput from '@/components/tool/CertTextInput.vue'
import KeyInput from '@/components/tool/KeyInput.vue'
import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useToolRun } from '@/composables/useToolRun'
import {
  TEST_CA_CERT_PEM,
  TEST_CA_KEY_PEM,
} from '@/tools/certificates/fixtures-v060'

/** 后端 cert-gen 工具的响应。 */
interface CertGenResult {
  certPem: string
  privateKeyPem: string
  chainPem: string[]
  serial: string
  subjectText: string
  issuerText: string
  notBefore: string
  notAfter: string
  keyType: string
  keyCurve: string
  keyBits: number
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const MODES = ['self', 'ca']
const ALGORITHMS = ['EC', 'RSA', 'SM2']
const CURVES = ['P-256', 'P-384', 'P-521']
const RSA_BITS = [2048, 3072, 4096]

const { t, tm } = useI18n()

const signMode = ref('self')
const cn = ref('new.example.com')
const org = ref('Toolbit Test')
const country = ref('CN')
const sans = ref('DNS:new.example.com')
const notAfterDays = ref(365)
const serialHex = ref('')
const isCa = ref(false)

const keyType = ref('EC')
const keyCurve = ref('P-256')
const keyBits = ref(2048)

const issuerCert = ref('')
const issuerCertFormat = ref('auto')
const issuerKey = ref('')
const issuerKeyFormat = ref('pem')
const issuerPassphrase = ref('')

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
} = useToolRun<CertGenResult>({ slug: 'cert-gen' })

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
    `${t('tools.cert-gen.results.subject')}: ${result.subjectText}`,
    `${t('tools.cert-gen.results.issuer')}: ${result.issuerText}`,
    `${t('tools.cert-gen.results.serial')}: ${result.serial}`,
    `${t('tools.cert-gen.results.notBefore')}: ${result.notBefore}`,
    `${t('tools.cert-gen.results.notAfter')}: ${result.notAfter}`,
    '',
    result.certPem,
  ]
  for (const pem of result.chainPem.slice(1)) {
    head.push('', pem)
  }
  if (result.privateKeyPem) {
    head.push('', result.privateKeyPem)
  }
  return head.join('\n')
})

const faq = computed<FaqItem[]>(() => tm('tools.cert-gen.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  const issuer =
    signMode.value === 'ca'
      ? {
          cert: { data: issuerCert.value, format: issuerCertFormat.value },
          key: {
            data: issuerKey.value,
            format: issuerKeyFormat.value,
            passphrase: issuerPassphrase.value,
          },
        }
      : undefined
  void execute({
    signMode: signMode.value,
    ...(issuer ? { issuer } : {}),
    subject: { CN: cn.value, O: org.value, C: country.value },
    key: { type: keyType.value, bits: keyBits.value, curve: keyCurve.value },
    notAfterDays: notAfterDays.value,
    ...(serialHex.value ? { serial: serialHex.value } : {}),
    sans: sanList(),
    basicConstraints: { isCa: isCa.value, pathLen: -1 },
  })
}

/** 重置表单与结果。 */
function reset(): void {
  signMode.value = 'self'
  cn.value = 'new.example.com'
  org.value = 'Toolbit Test'
  country.value = 'CN'
  sans.value = 'DNS:new.example.com'
  notAfterDays.value = 365
  serialHex.value = ''
  isCa.value = false
  keyType.value = 'EC'
  keyCurve.value = 'P-256'
  keyBits.value = 2048
  issuerCert.value = ''
  issuerCertFormat.value = 'auto'
  issuerKey.value = ''
  issuerKeyFormat.value = 'pem'
  issuerPassphrase.value = ''
  resetRun()
}

/** 填入示例（CA 签发模式）。 */
function fillExample(): void {
  reset()
  signMode.value = 'ca'
  issuerCert.value = TEST_CA_CERT_PEM
  issuerKey.value = TEST_CA_KEY_PEM
}
</script>

<template>
  <ToolPageLayout
    slug="cert-gen"
    :name="t('tools.cert-gen.name')"
    name-en="Certificate Generator"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.cert-gen.fields.signMode') }}</span>
            <select
              v-model="signMode"
              class="field__select"
            >
              <option
                v-for="mode in MODES"
                :key="mode"
                :value="mode"
              >
                {{ t(`tools.cert-gen.options.modes.${mode}`) }}
              </option>
            </select>
          </label>

          <template v-if="signMode === 'ca'">
            <CertTextInput
              v-model="issuerCert"
              v-model:format="issuerCertFormat"
              :label="t('tools.cert-gen.fields.issuerCert')"
              :rows="6"
            />
            <KeyInput
              v-model="issuerKey"
              v-model:format="issuerKeyFormat"
              v-model:passphrase="issuerPassphrase"
              :label="t('tools.cert-gen.fields.issuerKey')"
              :formats="['pem', 'base64', 'hex']"
              show-passphrase
            />
          </template>

          <label class="field">
            <span class="field__label">{{ t('tools.cert-gen.fields.cn') }}</span>
            <input
              v-model="cn"
              type="text"
              class="field__input"
              spellcheck="false"
              autocomplete="off"
            >
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.cert-gen.fields.org') }}</span>
            <input
              v-model="org"
              type="text"
              class="field__input"
              spellcheck="false"
              autocomplete="off"
            >
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.cert-gen.fields.country') }}</span>
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
            <span class="field__label">{{ t('tools.cert-gen.fields.sans') }}</span>
            <textarea
              v-model="sans"
              class="field__textarea tb-mono"
              rows="3"
              spellcheck="false"
            />
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.cert-gen.fields.notAfterDays') }}</span>
            <input
              v-model.number="notAfterDays"
              type="number"
              class="field__input"
              min="1"
              max="3650"
            >
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.cert-gen.fields.serial') }}</span>
            <input
              v-model="serialHex"
              type="text"
              class="field__input"
              spellcheck="false"
              autocomplete="off"
            >
          </label>

          <label class="field field--inline">
            <input
              v-model="isCa"
              type="checkbox"
            >
            <span>{{ t('tools.cert-gen.fields.isCa') }}</span>
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.cert-gen.fields.algorithm') }}</span>
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
            <span class="field__label">{{ t('tools.cert-gen.fields.curve') }}</span>
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
            <span class="field__label">{{ t('tools.cert-gen.fields.bits') }}</span>
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
        :title="t('tools.cert-gen.results.output')"
        @retry="run"
      />
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.cert-gen.description') }}
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
