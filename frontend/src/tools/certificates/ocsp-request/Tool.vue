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

import MaterialInput from '@/components/tool/MaterialInput.vue'
import ResultPanel, { type ResultView } from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import StatGrid, { type StatGridItem } from '@/components/tool/StatGrid.vue'
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

/** 结果页签：摘要 / 请求内容。 */
const views = computed<ResultView[]>(() => [
  { id: 'summary', label: t('tools.ocsp-request.views.summary') },
  { id: 'der', label: t('tools.ocsp-request.views.der') },
])

/** 当前页签。 */
const activeView = ref('summary')

/** 摘要统计卡片。 */
const stats = computed<StatGridItem[]>(() => {
  const result = data.value
  if (!result) {
    return []
  }
  return [
    { label: t('tools.ocsp-request.results.hash'), value: result.hash },
    { label: t('tools.ocsp-request.results.length'), value: `${result.length} B` },
  ]
})

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
          <MaterialInput
            v-model="cert"
            v-model:format="certFormat"
            :label="t('tools.ocsp-request.fields.cert')"
            :rows="6"
          />
          <MaterialInput
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
        v-model:active-view="activeView"
        :state="state"
        :text="resultText"
        :meta="meta"
        :error="error"
        :title="t('tools.ocsp-request.results.output')"
        :views="views"
        @retry="run"
      >
        <template #view="{ viewId }">
          <StatGrid
            v-if="viewId === 'summary'"
            :items="stats"
          />
          <pre
            v-else-if="data"
            class="product"
          ><code>{{ data.derHex }}</code></pre>
        </template>
      </ResultPanel>
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

<style scoped>
.product {
  margin: 0;
  padding: var(--tb-space-3);
  overflow-x: auto;
  font-family: var(--tb-font-mono);
  font-size: var(--tb-font-sm);
  line-height: 1.6;
  color: var(--tb-text);
  white-space: pre-wrap;
  word-break: break-all;
  background: var(--tb-bg-sunken);
  border-radius: var(--tb-radius-control);
}
</style>
