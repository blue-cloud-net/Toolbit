<script setup lang="ts">
/**
 * OCSP 响应解析工具页。
 *
 * 证书与响应只存在于组件局部状态；离开页面即丢弃，**不写**浏览器存储。
 */
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import CertTextInput from '@/components/tool/CertTextInput.vue'
import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useToolRun } from '@/composables/useToolRun'
import {
  TEST_CA_CERT_PEM,
  TEST_LEAF_CERT_PEM,
} from '@/tools/certificates/fixtures-v060'

/** 后端 ocsp-response 工具的响应。 */
interface OCSPResponseResult {
  status: number
  statusText: string
  producedAt: string
  certStatus: string
  revocationTime: string
  reasonCode: number
  reasonText: string
  thisUpdate: string
  nextUpdate: string
  responderCount: number
  signatureVerified: boolean
}

interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const response = ref('')
const responseFormat = ref('auto')
const cert = ref('')
const certFormat = ref('auto')
const issuer = ref('')
const issuerFormat = ref('auto')
const trustedRoots = ref('')
const trustedRootsFormat = ref('auto')

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
} = useToolRun<OCSPResponseResult>({ slug: 'ocsp-response' })

const running = computed(() => state.value === 'loading')

const resultText = computed(() => {
  const result = data.value
  if (!result) {
    return ''
  }
  const head = [
    `${t('tools.ocsp-response.results.status')}: ${result.statusText} (${result.status})`,
    `${t('tools.ocsp-response.results.certStatus')}: ${t(`tools.ocsp-response.options.certStatus.${result.certStatus}`)}`,
    `${t('tools.ocsp-response.results.signatureVerified')}: ${result.signatureVerified}`,
    `${t('tools.ocsp-response.results.producedAt')}: ${result.producedAt}`,
    `${t('tools.ocsp-response.results.thisUpdate')}: ${result.thisUpdate}`,
    `${t('tools.ocsp-response.results.nextUpdate')}: ${result.nextUpdate}`,
    `${t('tools.ocsp-response.results.responderCount')}: ${result.responderCount}`,
  ]
  if (result.certStatus === 'revoked') {
    head.push(`${t('tools.ocsp-response.results.revocationTime')}: ${result.revocationTime}`)
    head.push(`${t('tools.ocsp-response.results.reason')}: ${result.reasonText || result.reasonCode}`)
  }
  return head.join('\n')
})

const faq = computed<FaqItem[]>(() => tm('tools.ocsp-response.faq') as FaqItem[])

function run(): void {
  const roots = trustedRoots.value.trim()
    ? [{ data: trustedRoots.value, format: trustedRootsFormat.value }]
    : []
  void execute({
    response: { data: response.value, format: responseFormat.value },
    cert: { data: cert.value, format: certFormat.value },
    issuer: { data: issuer.value, format: issuerFormat.value },
    trustedRoots: roots,
  })
}

function reset(): void {
  response.value = ''
  responseFormat.value = 'auto'
  cert.value = ''
  certFormat.value = 'auto'
  issuer.value = ''
  issuerFormat.value = 'auto'
  trustedRoots.value = ''
  trustedRootsFormat.value = 'auto'
  resetRun()
}

function fillExample(): void {
  reset()
  cert.value = TEST_LEAF_CERT_PEM
  issuer.value = TEST_CA_CERT_PEM
  trustedRoots.value = TEST_CA_CERT_PEM
}
</script>

<template>
  <ToolPageLayout
    slug="ocsp-response"
    name="OCSP 响应解析"
    name-en="OCSP Response Parser"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <CertTextInput
            v-model="response"
            v-model:format="responseFormat"
            :label="t('tools.ocsp-response.fields.response')"
            :rows="6"
          />
          <CertTextInput
            v-model="cert"
            v-model:format="certFormat"
            :label="t('tools.ocsp-response.fields.cert')"
            :rows="5"
          />
          <CertTextInput
            v-model="issuer"
            v-model:format="issuerFormat"
            :label="t('tools.ocsp-response.fields.issuer')"
            :rows="5"
          />
          <CertTextInput
            v-model="trustedRoots"
            v-model:format="trustedRootsFormat"
            :label="t('tools.ocsp-response.fields.trustedRoots')"
            :rows="5"
          />
        </template>

        <template #footer>
          <RunBar
            :running="running"
            :disabled="response === '' || cert === '' || issuer === ''"
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
        :title="t('tools.ocsp-response.results.output')"
        @retry="run"
      />
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.ocsp-response.description') }}
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
