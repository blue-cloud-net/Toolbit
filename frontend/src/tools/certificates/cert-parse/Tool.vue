<script setup lang="ts">
/**
 * 证书解析工具页。
 *
 * 证书内容只存在于组件局部状态；离开页面即丢弃，
 * **不写**任何浏览器存储（AGENTS.md §4.11）。
 */
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { ElOption, ElSelect } from 'element-plus/es/components/select/index'
import 'element-plus/es/components/select/style/css'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import CertTextInput from '@/components/tool/CertTextInput.vue'
import KeyValueTable, { type KeyValueRow } from '@/components/tool/KeyValueTable.vue'
import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useToolRun } from '@/composables/useToolRun'
import { DEFAULT_EXAMPLE } from '@/tools/certificates/cert-parse/example'

/** 后端 cert-parse 工具的响应。 */
interface CertParseResult {
  certificate: {
    version: number
    subject: { field: string; value: string }[]
    subjectText: string
    issuer: { field: string; value: string }[]
    issuerText: string
    serial: string
    notBefore: string
    notAfter: string
    publicKey: { algorithm: string; bits: number; curve: string }
    signatureAlgorithm: string
    signatureAlgorithmOid: string
    isCa: boolean
    pathLen: number
    san: string[]
    keyUsage: string[]
    extendedKeyUsage: string[]
    subjectKeyId: string
    authorityKeyId: string
    subjectHash: string
    issuerHash: string
    extensions: {
      nid: number
      oid: string
      name: string
      critical: boolean
      value: string
    }[]
    selfSigned: boolean
  }
  validity: { status: string; daysRemaining: number }
  fingerprints: { algorithm: string; value: string }[]
  selfSigned: boolean
  rawDerBase64: string
}

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

/** 结构化字段：逐项可复制。 */
const detailRows = computed<KeyValueRow[]>(() => {
  const result = data.value
  if (!result) {
    return []
  }
  const info = result.certificate
  const rows: KeyValueRow[] = [
    { label: t('tools.cert-parse.results.subject'), value: info.subjectText },
    { label: t('tools.cert-parse.results.issuer'), value: info.issuerText },
    { label: t('tools.cert-parse.results.serial'), value: info.serial },
    {
      label: t('tools.cert-parse.results.validity'),
      value: `${result.validity.status} (${result.validity.daysRemaining}d)`,
      level: result.validity.daysRemaining < 0 ? ('danger' as const) : ('success' as const),
    },
    { label: t('tools.cert-parse.results.notBefore'), value: info.notBefore },
    { label: t('tools.cert-parse.results.notAfter'), value: info.notAfter },
    {
      label: t('tools.cert-parse.results.publicKey'),
      value: `${info.publicKey.algorithm} ${info.publicKey.bits}`,
    },
    { label: t('tools.cert-parse.results.signatureAlgorithm'), value: info.signatureAlgorithm },
    {
      label: t('tools.cert-parse.results.selfSigned'),
      value: String(result.selfSigned),
    },
  ]
  for (const fp of result.fingerprints) {
    rows.push({ label: fp.algorithm, value: fp.value })
  }
  return rows
})

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
} = useToolRun<CertParseResult>({ slug: 'cert-parse' })

const running = computed(() => state.value === 'loading')

/** 结果文本：概要 + 扩展列表。 */
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
      head.push(`  ${ext.name || ext.oid}${ext.critical ? ' (critical)' : ''}: ${ext.value}`)
    }
  }
  return head.join('\n')
})

const faq = computed<FaqItem[]>(() => tm('tools.cert-parse.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  void execute({
    input: { data: cert.value, format: certFormat.value },
    atTime: atTime.value,
    fingerprints: fingerprints.value,
  })
}

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
          <CertTextInput
            v-model="cert"
            v-model:format="certFormat"
            :label="t('tools.cert-parse.fields.cert')"
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
        :state="state"
        :text="resultText"
        :meta="meta"
        :error="error"
        :title="t('tools.cert-parse.results.info')"
        download-filename="certificate.txt"
        @retry="run"
      >
        <template
          v-if="detailRows.length > 0"
          #extra
        >
          <KeyValueTable :rows="detailRows" />
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
