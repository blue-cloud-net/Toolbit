<script setup lang="ts">
/**
 * 吊销状态检查工具页。
 *
 * 证书与吊销列表只存在于组件局部状态；离开页面即丢弃，
 * **不写**任何浏览器存储（AGENTS.md §4.11）。
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
  TEST_CRL_PEM,
  TEST_LEAF_CERT_PEM,
} from '@/tools/certificates/fixtures-v060'

/** 后端 revocation-check 工具的响应。 */
interface RevocationResult {
  status: string
  signatureVerified: boolean
  serial: string
  revokedAt: string
  reasonCode: number
  reasonText: string
  crlNumber: number
  crlThisUpdate: string
  crlNextUpdate: string
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const cert = ref('')
const certFormat = ref('auto')
const crl = ref('')
const crlFormat = ref('auto')
const issuer = ref('')
const issuerFormat = ref('auto')

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
} = useToolRun<RevocationResult>({ slug: 'revocation-check' })

const running = computed(() => state.value === 'loading')

/** 结果文本。 */
const resultText = computed(() => {
  const result = data.value
  if (!result) {
    return ''
  }
  const head = [
    `${t('tools.revocation-check.results.status')}: ${t(`tools.revocation-check.options.status.${result.status}`)}`,
    `${t('tools.revocation-check.results.signatureVerified')}: ${result.signatureVerified}`,
    `${t('tools.revocation-check.results.serial')}: ${result.serial}`,
    `${t('tools.revocation-check.results.crlNumber')}: ${result.crlNumber}`,
    `${t('tools.revocation-check.results.crlNextUpdate')}: ${result.crlNextUpdate}`,
  ]
  if (result.status === 'revoked') {
    head.push(`${t('tools.revocation-check.results.revokedAt')}: ${result.revokedAt}`)
    head.push(`${t('tools.revocation-check.results.reason')}: ${result.reasonText || result.reasonCode}`)
  }
  return head.join('\n')
})

const faq = computed<FaqItem[]>(() => tm('tools.revocation-check.faq') as FaqItem[])

/** 结论徽章：按 CRL 状态给出结论。 */
const verdict = computed(() => {
  const result = data.value
  if (!result) {
    return null
  }
  if (result.status === 'good') {
    return { level: 'success' as const, label: t('tool.verdict.pass') }
  }
  if (result.status === 'revoked') {
    return { level: 'danger' as const, label: t('tool.verdict.fail') }
  }
  return { level: 'warning' as const, label: t('tool.verdict.unknown') }
})

/** 运行当前输入。 */
function run(): void {
  void execute({
    cert: { data: cert.value, format: certFormat.value },
    crl: { data: crl.value, format: crlFormat.value },
    ...(issuer.value.trim()
      ? { issuer: { data: issuer.value, format: issuerFormat.value } }
      : {}),
  })
}

/** 重置表单与结果。 */
function reset(): void {
  cert.value = ''
  certFormat.value = 'auto'
  crl.value = ''
  crlFormat.value = 'auto'
  issuer.value = ''
  issuerFormat.value = 'auto'
  resetRun()
}

/** 填入示例。 */
function fillExample(): void {
  reset()
  cert.value = TEST_LEAF_CERT_PEM
  crl.value = TEST_CRL_PEM
  issuer.value = TEST_CA_CERT_PEM
}
</script>

<template>
  <ToolPageLayout
    slug="revocation-check"
    :name="t('tools.revocation-check.name')"
    name-en="Revocation Checker"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <CertTextInput
            v-model="cert"
            v-model:format="certFormat"
            :label="t('tools.revocation-check.fields.cert')"
            :rows="6"
          />
          <CertTextInput
            v-model="crl"
            v-model:format="crlFormat"
            :label="t('tools.revocation-check.fields.crl')"
            :rows="6"
          />
          <CertTextInput
            v-model="issuer"
            v-model:format="issuerFormat"
            :label="t('tools.revocation-check.fields.issuer')"
            :rows="5"
          />
        </template>

        <template #footer>
          <RunBar
            :running="running"
            :disabled="cert === '' || crl === ''"
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
        :verdict="verdict"
        :title="t('tools.revocation-check.results.output')"
        @retry="run"
      />
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.revocation-check.description') }}
      </p>
      <p
        v-if="data && !data.signatureVerified"
        class="prose"
      >
        {{ t('tools.revocation-check.hints.unverified') }}
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
