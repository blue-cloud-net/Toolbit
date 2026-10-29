<script setup lang="ts">
/**
 * OCSP 请求生成工具页。
 *
 * 证书只存在于组件局部状态；离开页面即丢弃，**不写**浏览器存储。
 */
import { ElOption, ElSelect } from 'element-plus/es/components/select/index'
import 'element-plus/es/components/select/style/css'
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

/** 后端 ocsp-request 工具的响应。 */
interface OCSPRequestResult {
  hash: string
  derBase64: string
  derHex: string
  length: number
}

interface FaqItem {
  q: string
  a: string
}

const HASHES = ['sha1', 'sha256', 'sm3']

const { t, tm } = useI18n()

const cert = ref('')
const certFormat = ref('auto')
const issuer = ref('')
const issuerFormat = ref('auto')
const hash = ref('sha256')

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
} = useToolRun<OCSPRequestResult>({ slug: 'ocsp-request' })

const running = computed(() => state.value === 'loading')

const resultText = computed(() => {
  const result = data.value
  if (!result) {
    return ''
  }
  return [
    `${t('tools.ocsp-request.results.hash')}: ${result.hash}`,
    `${t('tools.ocsp-request.results.length')}: ${result.length}`,
    '',
    result.derHex,
  ].join('\n')
})

const faq = computed<FaqItem[]>(() => tm('tools.ocsp-request.faq') as FaqItem[])

function run(): void {
  void execute({
    cert: { data: cert.value, format: certFormat.value },
    issuer: { data: issuer.value, format: issuerFormat.value },
    hash: hash.value,
  })
}

function reset(): void {
  cert.value = ''
  certFormat.value = 'auto'
  issuer.value = ''
  issuerFormat.value = 'auto'
  hash.value = 'sha256'
  resetRun()
}

function fillExample(): void {
  reset()
  cert.value = TEST_LEAF_CERT_PEM
  issuer.value = TEST_CA_CERT_PEM
}
</script>

<template>
  <ToolPageLayout
    slug="ocsp-request"
    name="OCSP 请求生成"
    name-en="OCSP Request Builder"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <CertTextInput
            v-model="cert"
            v-model:format="certFormat"
            :label="t('tools.ocsp-request.fields.cert')"
            :rows="6"
          />
          <CertTextInput
            v-model="issuer"
            v-model:format="issuerFormat"
            :label="t('tools.ocsp-request.fields.issuer')"
            :rows="6"
          />
        </template>

        <template #options>
          <label class="field">
            <span class="field__label">{{ t('tools.ocsp-request.fields.hash') }}</span>
            <ElSelect
              v-model="hash"
              class="field__select"
              :aria-label="t('tools.ocsp-request.fields.hash')"
            >
              <ElOption
                v-for="name in HASHES"
                :key="name"
                :label="name.toUpperCase()"
                :value="name"
              />
            </ElSelect>
          </label>
        </template>

        <template #footer>
          <RunBar
            :running="running"
            :disabled="cert === '' || issuer === ''"
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
        :title="t('tools.ocsp-request.results.output')"
        @retry="run"
      />
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.ocsp-request.description') }}
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
