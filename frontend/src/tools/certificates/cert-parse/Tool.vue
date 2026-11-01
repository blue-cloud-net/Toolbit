<script setup lang="ts">
/**
 * 证书解析工具页。
 *
 * 证书内容只存在于组件局部状态；离开页面即丢弃，
 * **不写**任何浏览器存储（AGENTS.md §4.11）。
 *
 * 结果区分四个页签：摘要 / 扩展 / 文本视图 / 证书原文；
 * 默认停在「摘要」，产物（文本视图、PEM）各自独立成页签。
 */
import { ElButton } from 'element-plus/es/components/button/index'
import 'element-plus/es/components/button/style/css'
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { ElOption, ElSelect } from 'element-plus/es/components/select/index'
import 'element-plus/es/components/select/style/css'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import CertDetail from '@/components/tool/CertDetail.vue'
import MaterialInput from '@/components/tool/MaterialInput.vue'
import DataTable, { type DataTableColumn, type DataTableRow } from '@/components/tool/DataTable.vue'
import DownloadButton from '@/components/tool/DownloadButton.vue'
import ResultPanel, { type ResultView, type Verdict } from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useAutoRun } from '@/composables/useAutoRun'
import { useToolRun } from '@/composables/useToolRun'
import { hexUpper } from '@/tools/certificates/certDetail'
import { DEFAULT_EXAMPLE } from '@/tools/certificates/cert-parse/example'
import type { CertParseResult } from '@/types/certificate'

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

/** 可选的指纹算法。 */
const FINGERPRINT_ALGORITHMS = ['sha256', 'sha1', 'sm3', 'sha384', 'sha512', 'md5']

const cert = ref('')
const certFormat = ref('auto')
const atTime = ref('')
const fingerprints = ref<string[]>(['sha256'])

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
  bindInput,
  useAsInput,
} = useToolRun<CertParseResult>({ slug: 'cert-parse' })

/**
 * 结果区的「作为新输入」：把**证书 PEM**回填到主输入框（仅内存，不写存储）。
 *
 * 回填的是 PEM 而不是结果摘要：摘要不是可解析的证书材料，
 * 回填后应当能直接再次解析。
 */
bindInput(() => {
  cert.value = data.value?.pem ?? ''
})

const running = computed(() => state.value === 'loading')

/** 结果页签：摘要 / 扩展 / 文本视图 / 证书原文。 */
const views = computed<ResultView[]>(() => [
  { id: 'summary', label: t('tools.cert-parse.views.summary') },
  { id: 'extensions', label: t('tools.cert-parse.views.extensions') },
  { id: 'text', label: t('tools.cert-parse.views.text') },
  { id: 'pem', label: t('tools.cert-parse.views.pem') },
])

/** 当前页签。 */
const activeView = ref('summary')

/** 结论徽章：证书当前是否可用。 */
const verdict = computed<Verdict | null>(() => {
  const result = data.value
  if (!result) {
    return null
  }
  if (result.validity.status === 'expired') {
    return { level: 'danger', label: t('tool.verdict.fail') }
  }
  if (result.validity.status === 'not_yet_valid') {
    return { level: 'warning', label: t('tool.verdict.warn') }
  }
  return { level: 'success', label: t('tool.verdict.pass') }
})

/** 扩展表格列。 */
const extensionColumns = computed<DataTableColumn[]>(() => [
  { key: 'name', label: t('tools.cert-parse.results.extensionName') },
  { key: 'oid', label: t('tools.cert-parse.results.oid'), mono: true },
  { key: 'critical', label: t('tools.cert-parse.results.critical'), align: 'center' },
  { key: 'value', label: t('tools.cert-parse.results.extensionValue'), mono: true },
])

/** 扩展表格行。 */
const extensionRows = computed<DataTableRow[]>(() =>
  (data.value?.certificate.extensions ?? []).map((ext) => ({
    key: `${ext.oid}-${ext.nid}`,
    cells: {
      name: ext.name || ext.oid,
      oid: ext.oid,
      critical: ext.critical ? t('common.states.yes') : '-',
      value: ext.text || hexUpper(ext.value),
    },
  })),
)

/** 结果文本：结构化摘要（复制 / 下载的数据源）。 */
const resultText = computed(() => {
  const result = data.value
  if (!result) {
    return ''
  }
  const info = result.certificate
  const head = [
    `${t('tools.cert-parse.results.subject')}: ${info.subjectText}`,
    `${t('tools.cert-parse.results.issuer')}: ${info.issuerText}`,
    `${t('tools.cert-parse.results.serial')}: ${info.serial}`,
    `${t('tools.cert-parse.results.validity')}: ${result.validity.status} (${result.validity.daysRemaining}d)`,
    `${t('tools.cert-parse.results.notBefore')}: ${info.notBefore}`,
    `${t('tools.cert-parse.results.notAfter')}: ${info.notAfter}`,
    `${t('tools.cert-parse.results.publicKey')}: ${info.publicKey.algorithm} ${info.publicKey.bits}`,
    `${t('tools.cert-parse.results.signatureAlgorithm')}: ${info.signatureAlgorithm}`,
    `${t('tools.cert-parse.results.selfSigned')}: ${result.selfSigned}`,
  ]
  if (info.san.length > 0) {
    head.push(`${t('tools.cert-parse.results.san')}: ${info.san.join(', ')}`)
  }
  if (info.keyUsage.length > 0) {
    head.push(`${t('tools.cert-parse.results.keyUsage')}: ${info.keyUsage.join(', ')}`)
  }
  if (info.extendedKeyUsage.length > 0) {
    head.push(`${t('tools.cert-parse.results.extendedKeyUsage')}: ${info.extendedKeyUsage.join(', ')}`)
  }
  head.push('')
  head.push(t('tools.cert-parse.results.fingerprints'))
  for (const fp of result.fingerprints) {
    head.push(`  ${fp.algorithm}: ${fp.value}`)
  }
  if (info.extensions.length > 0) {
    head.push('')
    head.push(t('tools.cert-parse.results.extensions'))
    for (const ext of info.extensions) {
      head.push(
        `  ${ext.name || ext.oid}${ext.critical ? ' (critical)' : ''}: ${ext.text || hexUpper(ext.value)}`,
      )
    }
  }
  return head.join('\n')
})

const faq = computed<FaqItem[]>(() => tm('tools.cert-parse.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  if (cert.value.trim() === '') {
    return
  }
  void execute({
    input: { data: cert.value, format: certFormat.value },
    atTime: atTime.value,
    fingerprints: fingerprints.value,
  })
}

// 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。
useAutoRun({
  slug: 'cert-parse',
  run,
  watch: [cert, certFormat, atTime, fingerprints],
  enabled: () => cert.value.trim() !== '',
})

/** 重置表单与结果。 */
function reset(): void {
  cert.value = ''
  certFormat.value = 'auto'
  atTime.value = ''
  fingerprints.value = ['sha256']
  resetRun()
}

/** 填入示例。 */
function fillExample(): void {
  cert.value = DEFAULT_EXAMPLE.cert
  certFormat.value = 'auto'
  atTime.value = ''
  fingerprints.value = ['sha256']
  resetRun()
}

/** 把「判断时刻」设为当前时间。 */
function setNow(): void {
  atTime.value = new Date().toISOString()
}
</script>

<template>
  <ToolPageLayout
    slug="cert-parse"
    :name="t('tools.cert-parse.name')"
    name-en="Certificate Parser"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <MaterialInput
            v-model="cert"
            v-model:format="certFormat"
            :label="t('tools.cert-parse.fields.cert')"
            slug="cert-parse"
          />
        </template>

        <template #options>
          <label class="field">
            <span class="field__label">{{ t('tools.cert-parse.fields.fingerprints') }}</span>
            <ElSelect
              v-model="fingerprints"
              class="field__select"
              multiple
              :aria-label="t('tools.cert-parse.fields.fingerprints')"
            >
              <ElOption
                v-for="alg in FINGERPRINT_ALGORITHMS"
                :key="alg"
                :label="String(alg)"
                :value="alg"
              />
            </ElSelect>
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.cert-parse.fields.atTime') }}</span>
            <ElInput
              v-model="atTime"
              class="field__input"
              placeholder="2026-01-01T00:00:00Z"
              spellcheck="false"
              autocomplete="off"
            />
            <ElButton
              size="small"
              @click="setNow"
            >
              {{ t('tools.cert-parse.actions.now') }}
            </ElButton>
          </label>
        </template>

        <template #footer>
          <RunBar
            :running="running"
            :disabled="cert === ''"
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
        :verdict="verdict"
        :title="t('tools.cert-parse.results.info')"
        :views="views"
        :use-as-input="useAsInput"
        download-filename="certificate-summary.txt"
        @retry="run"
      >
        <template #view="{ viewId }">
          <CertDetail
            v-if="viewId === 'summary' && data"
            :info="data.certificate"
            :validity="data.validity"
            :fingerprints="data.fingerprints"
            :public-key="data.publicKey"
          />
          <DataTable
            v-else-if="viewId === 'extensions' && data"
            :columns="extensionColumns"
            :rows="extensionRows"
            :copy-column="'value'"
          />
          <div
            v-else-if="viewId === 'text' && data"
            class="product-wrap"
          >
            <div class="product-wrap__head">
              <span class="product-wrap__label">{{ t('tools.cert-parse.results.textView') }}</span>
              <DownloadButton
                filename="certificate.txt"
                :content="data.textView"
                mime="text/plain;charset=utf-8"
              />
            </div>
            <pre class="product"><code>{{ data.textView }}</code></pre>
          </div>
          <div
            v-else-if="viewId === 'pem' && data"
            class="product-wrap"
          >
            <div class="product-wrap__head">
              <span class="product-wrap__label">PEM</span>
              <DownloadButton
                filename="certificate.pem"
                :content="data.pem"
                mime="application/x-pem-file"
              />
            </div>
            <pre class="product"><code>{{ data.pem }}</code></pre>
          </div>
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.cert-parse.description') }}
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
.product-wrap {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-2);
}

.product-wrap__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--tb-space-2);
}

.product-wrap__label {
  font-size: var(--tb-font-xs);
  color: var(--tb-text-tertiary);
}

.product {
  margin: 0;
  padding: var(--tb-space-3);
  max-height: 520px;
  overflow: auto;
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
