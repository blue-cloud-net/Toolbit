<script setup lang="ts">
/**
 * 吊销列表生成工具页。
 *
 * 证书与私钥只存在于组件局部状态；离开页面即丢弃，
 * **不写**任何浏览器存储（AGENTS.md §4.11）。
 */
import { ElInputNumber } from 'element-plus/es/components/input-number/index'
import 'element-plus/es/components/input-number/style/css'
import { ElOption, ElSelect } from 'element-plus/es/components/select/index'
import 'element-plus/es/components/select/style/css'
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
  TEST_LEAF_CERT_PEM,
} from '@/tools/certificates/fixtures-v060'

/** 后端 crl-gen 工具的响应。 */
interface CRLGenResult {
  crlPem: string
  crlDerBase64: string
  issuerText: string
  number: number
  thisUpdate: string
  nextUpdate: string
  revokedCount: number
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const REASONS = [
  { code: 0, key: 'unspecified' },
  { code: 1, key: 'keyCompromise' },
  { code: 2, key: 'cACompromise' },
  { code: 3, key: 'affiliationChanged' },
  { code: 4, key: 'superseded' },
  { code: 5, key: 'cessationOfOperation' },
  { code: 6, key: 'certificateHold' },
  { code: 8, key: 'removeFromCRL' },
  { code: 9, key: 'privilegeWithdrawn' },
  { code: 10, key: 'aACompromise' },
]

const { t, tm } = useI18n()

const issuerCert = ref('')
const issuerCertFormat = ref('auto')
const issuerKey = ref('')
const issuerKeyFormat = ref('pem')
const issuerPassphrase = ref('')
const revokedCert = ref('')
const revokedCertFormat = ref('auto')
const reasonCode = ref(1)
const nextUpdateDays = ref(30)

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
} = useToolRun<CRLGenResult>({ slug: 'crl-gen' })

const running = computed(() => state.value === 'loading')

/** 结果文本。 */
const resultText = computed(() => {
  const result = data.value
  if (!result) {
    return ''
  }
  return [
    `${t('tools.crl-gen.results.issuer')}: ${result.issuerText}`,
    `${t('tools.crl-gen.results.number')}: ${result.number}`,
    `${t('tools.crl-gen.results.thisUpdate')}: ${result.thisUpdate}`,
    `${t('tools.crl-gen.results.nextUpdate')}: ${result.nextUpdate}`,
    `${t('tools.crl-gen.results.revokedCount')}: ${result.revokedCount}`,
    '',
    result.crlPem,
  ].join('\n')
})

const faq = computed<FaqItem[]>(() => tm('tools.crl-gen.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  const revoked = revokedCert.value.trim()
    ? [
        {
          cert: { data: revokedCert.value, format: revokedCertFormat.value },
          reasonCode: reasonCode.value,
        },
      ]
    : []
  void execute({
    issuerCert: { data: issuerCert.value, format: issuerCertFormat.value },
    issuerKey: {
      data: issuerKey.value,
      format: issuerKeyFormat.value,
      passphrase: issuerPassphrase.value,
    },
    revoked,
  })
}

/** 重置表单与结果。 */
function reset(): void {
  issuerCert.value = ''
  issuerCertFormat.value = 'auto'
  issuerKey.value = ''
  issuerKeyFormat.value = 'pem'
  issuerPassphrase.value = ''
  revokedCert.value = ''
  revokedCertFormat.value = 'auto'
  reasonCode.value = 1
  nextUpdateDays.value = 30
  resetRun()
}

/** 填入示例。 */
function fillExample(): void {
  reset()
  issuerCert.value = TEST_CA_CERT_PEM
  issuerKey.value = TEST_CA_KEY_PEM
  revokedCert.value = TEST_LEAF_CERT_PEM
}
</script>

<template>
  <ToolPageLayout
    slug="crl-gen"
    :name="t('tools.crl-gen.name')"
    name-en="CRL Generator"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <CertTextInput
            v-model="issuerCert"
            v-model:format="issuerCertFormat"
            :label="t('tools.crl-gen.fields.issuerCert')"
            :rows="6"
          />
          <KeyInput
            v-model="issuerKey"
            v-model:format="issuerKeyFormat"
            v-model:passphrase="issuerPassphrase"
            :label="t('tools.crl-gen.fields.issuerKey')"
            :formats="['pem', 'base64', 'hex']"
            show-passphrase
          />
          <CertTextInput
            v-model="revokedCert"
            v-model:format="revokedCertFormat"
            :label="t('tools.crl-gen.fields.revokedCert')"
            :rows="6"
          />
        </template>

        <template #options>
          <label class="field">
            <span class="field__label">{{ t('tools.crl-gen.fields.reasonCode') }}</span>
            <ElSelect
              v-model="reasonCode"
              class="field__select"
              :aria-label="t('tools.crl-gen.fields.reasonCode')"
            >
              <ElOption
                v-for="reason in REASONS"
                :key="reason.code"
                :label="`${reason.code} · ${t(`tools.crl-gen.options.reasons.${reason.key}`)}`"
                :value="reason.code"
              />
            </ElSelect>
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.crl-gen.fields.nextUpdateDays') }}</span>
            <ElInputNumber
              v-model="nextUpdateDays"
              class="field__input"
              :min="1"
              :max="3650"
            />
          </label>
        </template>

        <template #footer>
          <RunBar
            :running="running"
            :disabled="issuerCert === '' || issuerKey === ''"
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
        :title="t('tools.crl-gen.results.output')"
        @retry="run"
      />
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.crl-gen.description') }}
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
