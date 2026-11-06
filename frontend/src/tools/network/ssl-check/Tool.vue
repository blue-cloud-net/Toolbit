<script setup lang="ts">
/**
 * SSL 证书检测工具页。
 *
 * 走 WebSocket 流式端点：服务端按阶段推送 `progress`，结束后推 `result` 或 `error`。
 * 一次扫描包含协议矩阵、套件矩阵、客户端握手模拟与出网附加检查，
 * 因此**不能**复用 `useToolRun`。
 *
 * 证书链部分直接复用 `cert-chain-parse` 的 `<CertChainTable>`：
 * 服务端返回的 `chain` 与那个工具的响应是同一份结构。
 */
import { ElCheckbox } from 'element-plus/es/components/checkbox/index'
import 'element-plus/es/components/checkbox/style/css'
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { ElInputNumber } from 'element-plus/es/components/input-number/index'
import 'element-plus/es/components/input-number/style/css'
import { ElOption, ElSelect } from 'element-plus/es/components/select/index'
import 'element-plus/es/components/select/style/css'
import { computed, onScopeDispose, ref, shallowRef } from 'vue'
import { useI18n } from 'vue-i18n'

import { ApiError, streamURL } from '@/api/client'
import CertChainTable from '@/components/tool/CertChainTable.vue'
import DataTable, { type DataTableColumn, type DataTableRow } from '@/components/tool/DataTable.vue'
import KeyValueTable, { type KeyValueRow } from '@/components/tool/KeyValueTable.vue'
import ProductText from '@/components/tool/ProductText.vue'
import ResultPanel, { type ResultView } from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import StatGrid, { type StatGridItem } from '@/components/tool/StatGrid.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { localizeError, type LocalizedError } from '@/i18n/errors'
import type { ChainCertificate } from '@/types/certificate'
import type { Meta } from '@/types/api'

import { DEFAULT_EXAMPLE } from './example'

/** 端点信息。 */
interface ScanEndpoint {
  host: string
  port: number
  sni: string
  addr: string
  elapsedMs: number
}

/** 协商结果。 */
interface ScanNegotiated {
  version: string
  versionWire: number
  cipher: string
  cipherId: number
  keyExchangeGroup: string
  peerSignatureAlgorithm: string
  alpn: string
}

/** 叶证书摘要。 */
interface ScanCertificate {
  subject: string
  issuer: string
  serial: string
  notBefore: string
  notAfter: string
  daysRemaining: number
  keyAlgorithm: string
  keyBits: number
  curve: string
  signatureAlgorithm: string
  sanCount: number
  sanMatchesHost: boolean
  isCa: boolean
  selfSigned: boolean
  sctCount: number
  mustStaple: boolean
  chainValid: boolean
  verified: boolean
  implicitTrust: boolean
  chainErrors: string[]
}

/** 证书链（与 cert-chain-parse 同构）。 */
interface ChainInfo {
  count: number
  chain: ChainCertificate[]
  ordered: boolean
  missingRoot: boolean
  selfSigned: boolean
}

/** 协议版本探测结果。 */
interface ProtocolResult {
  version: string
  versionWire: number
  supported: boolean
  negotiatedCipher?: string
  reason?: string
}

/** 套件探测结果。 */
interface CipherResult {
  version: string
  name: string
  id: number
  family: string
  bits: number
  kex: string
  verdict: 'supported' | 'rejected' | 'unavailable'
}

/** 特性探测结果。 */
interface FeatureResult {
  id: string
  label: string
  status: 'supported' | 'not_supported' | 'unavailable'
  detail?: string
}

/** 客户端模拟结果。 */
interface ClientResult {
  id: string
  vendor: string
  os: string
  engine: string
  profileVersion: string
  supported: boolean
  reason?: string
  negotiatedVersion?: string
  cipher?: string
  alpn?: string
}

/** CAA 查询结果。 */
interface CAAResult {
  queried: boolean
  records: string[]
  permitted: boolean
  violations: string[]
  error?: string
}

/** 吊销检查条目。 */
interface RevocationEntry {
  url: string
  status: string
  thisUpdate?: string
  nextUpdate?: string
  error?: string
}

/** 吊销检查汇总。 */
interface RevocationResult {
  ocsp: RevocationEntry[]
  crl: RevocationEntry[]
}

/** HSTS 结论。 */
interface HSTSResult {
  present: boolean
  maxAge: number
  includeSubDomains: boolean
  preload: boolean
  error?: string
}

/** 评级说明。 */
interface GradeIssue {
  id: string
  impact: string
  detail?: string
}

/** 四象限得分。 */
interface GradeCategories {
  certificate: number
  protocols: number
  keyExchange: number
  cipherStrength: number
}

/** 综合评级。 */
interface GradeResult {
  letter: string
  score: number
  categories: GradeCategories
  caps: string[]
  issues: GradeIssue[]
}

/** 完整扫描结果。 */
interface ScanResult {
  endpoint: ScanEndpoint
  negotiated: ScanNegotiated
  certificate: ScanCertificate
  chain: ChainInfo | null
  protocols: ProtocolResult[]
  ciphers: CipherResult[]
  features: FeatureResult[]
  clients: ClientResult[]
  caa: CAAResult
  revocation: RevocationResult
  hsts: HSTSResult
  grade: GradeResult
}

/** 服务端帧。 */
type StreamFrame =
  | { type: 'progress'; phase: string; done: number; total: number; item?: string }
  | { type: 'result'; data: ScanResult; warnings?: unknown[] }
  | { type: 'error'; error: { code: string; message: string; params?: Record<string, unknown> } }

interface FaqItem {
  q: string
  a: string
}

/** 可选的检查项（与后端 `AllChecks` 保持一致）。 */
const CHECK_IDS = [
  'protocols',
  'ciphers',
  'features',
  'clients',
  'cert',
  'revocation',
  'caa',
  'hsts',
] as const

/** 可选的客户端画像（与后端画像表保持一致）。 */
const CLIENT_IDS = [
  'chrome',
  'firefox',
  'safari',
  'edge',
  'android',
  'ios',
  'java11',
  'java8',
  'openssl111',
  'openssl3',
  'schannel',
  'go',
] as const

const { t, tm } = useI18n()

const host = ref(DEFAULT_EXAMPLE.host)
const port = ref(DEFAULT_EXAMPLE.port)
const sni = ref(DEFAULT_EXAMPLE.sni)
const mode = ref<'quick' | 'full'>('full')
const selectedChecks = ref<string[]>([...CHECK_IDS])
const selectedClients = ref<string[]>([])

/** 结果区状态。 */
const controlState = ref<'idle' | 'loading' | 'success' | 'error'>('idle')

const running = ref(false)
const data = shallowRef<ScanResult | null>(null)
const meta = shallowRef<Meta | null>(null)
const error = shallowRef<LocalizedError | null>(null)
const progress = ref<{ phase: string; done: number; total: number; item?: string } | null>(null)
const warnings = shallowRef<unknown[]>([])

let socket: WebSocket | null = null

const faq = computed<FaqItem[]>(() => tm('tools.ssl-check.faq') as FaqItem[])

/** 结论徽章。 */
const verdict = computed(() => {
  const grade = data.value?.grade
  if (!grade?.letter) {
    return null
  }
  const label = t('tools.ssl-check.results.grade', { grade: grade.letter })
  if (grade.letter.startsWith('A')) {
    return { level: 'success' as const, label }
  }
  if (grade.letter === 'B' || grade.letter === 'C') {
    return { level: 'warning' as const, label }
  }
  return { level: 'danger' as const, label }
})

/** 结果页签。 */
const views = computed<ResultView[]>(() => [
  { id: 'summary', label: t('tools.ssl-check.views.summary') },
  { id: 'cert', label: t('tools.ssl-check.views.cert') },
  { id: 'chain', label: t('tools.ssl-check.views.chain') },
  { id: 'protocols', label: t('tools.ssl-check.views.protocols') },
  { id: 'ciphers', label: t('tools.ssl-check.views.ciphers') },
  { id: 'features', label: t('tools.ssl-check.views.features') },
  { id: 'clients', label: t('tools.ssl-check.views.clients') },
  { id: 'pem', label: t('tools.ssl-check.views.pem') },
  { id: 'json', label: t('tools.ssl-check.views.json') },
])

const activeView = ref('summary')

/** 摘要：端点与协商结果。 */
const sessionStats = computed<StatGridItem[]>(() => {
  const result = data.value
  if (!result) {
    return []
  }
  const items: StatGridItem[] = [
    { label: t('tools.ssl-check.results.endpoint'), value: `${result.endpoint.host}:${result.endpoint.port}` },
    { label: t('tools.ssl-check.results.address'), value: result.endpoint.addr, wide: true },
    { label: t('tools.ssl-check.results.version'), value: result.negotiated.version },
    { label: t('tools.ssl-check.results.cipher'), value: result.negotiated.cipher, wide: true },
    {
      label: t('tools.ssl-check.results.keyExchangeGroup'),
      value: result.negotiated.keyExchangeGroup || '-',
    },
    {
      label: t('tools.ssl-check.results.peerSignatureAlgorithm'),
      value: result.negotiated.peerSignatureAlgorithm || '-',
    },
    { label: t('tools.ssl-check.results.alpn'), value: result.negotiated.alpn || '-' },
    {
      label: t('tools.ssl-check.results.elapsed'),
      value: `${result.endpoint.elapsedMs} ms`,
    },
  ]
  return items
})

/** 摘要：四象限得分。 */
const gradeStats = computed<StatGridItem[]>(() => {
  const grade = data.value?.grade
  if (!grade) {
    return []
  }
  return [
    { label: t('tools.ssl-check.results.gradeScore'), value: String(grade.score) },
    { label: t('tools.ssl-check.results.catCertificate'), value: String(grade.categories.certificate) },
    { label: t('tools.ssl-check.results.catProtocols'), value: String(grade.categories.protocols) },
    { label: t('tools.ssl-check.results.catKeyExchange'), value: String(grade.categories.keyExchange) },
    { label: t('tools.ssl-check.results.catCipherStrength'), value: String(grade.categories.cipherStrength) },
  ]
})

/** 摘要：评级说明。 */
const issueRows = computed<KeyValueRow[]>(() =>
  (data.value?.grade.issues ?? []).map((issue) => ({
    label: t(`tools.ssl-check.grade.issues.${issue.id}`),
    value: issue.detail ?? t(`tools.ssl-check.grade.impacts.${issue.impact}`),
    level: issue.impact === 'critical' ? ('danger' as const) : issue.impact === 'unknown' ? ('warning' as const) : undefined,
  })),
)

/** 证书：摘要键值。 */
const certRows = computed<KeyValueRow[]>(() => {
  const cert = data.value?.certificate
  if (!cert) {
    return []
  }
  const yes = t('common.states.yes')
  const no = t('common.states.no')
  return [
    { label: t('tools.ssl-check.results.certSubject'), value: cert.subject },
    { label: t('tools.ssl-check.results.certIssuer'), value: cert.issuer },
    { label: t('tools.ssl-check.results.certSerial'), value: cert.serial },
    { label: t('tools.ssl-check.results.certValidFrom'), value: cert.notBefore },
    { label: t('tools.ssl-check.results.certValidTo'), value: cert.notAfter },
    { label: t('tools.ssl-check.results.certDaysRemaining'), value: cert.daysRemaining.toFixed(1) },
    { label: t('tools.ssl-check.results.certKey'), value: `${cert.keyAlgorithm}${cert.keyBits > 0 ? ` ${cert.keyBits}` : ''}${cert.curve ? ` ${cert.curve}` : ''}` },
    { label: t('tools.ssl-check.results.certSignature'), value: cert.signatureAlgorithm },
    { label: t('tools.ssl-check.results.certSanCount'), value: String(cert.sanCount) },
    {
      label: t('tools.ssl-check.results.certSanMatches'),
      value: cert.sanMatchesHost ? yes : no,
      level: cert.sanMatchesHost ? ('success' as const) : ('danger' as const),
    },
    {
      label: t('tools.ssl-check.results.chainValid'),
      value: cert.chainValid ? yes : no,
      level: cert.chainValid ? ('success' as const) : ('danger' as const),
    },
    {
      label: t('tools.ssl-check.results.verified'),
      value: cert.verified ? yes : no,
      level: cert.verified ? ('success' as const) : ('danger' as const),
    },
    { label: t('tools.ssl-check.results.mustStaple'), value: cert.mustStaple ? yes : no },
    { label: t('tools.ssl-check.results.sctCount'), value: String(cert.sctCount) },
  ]
})

/** 证书：链错误。 */
const chainErrorRows = computed<KeyValueRow[]>(() =>
  (data.value?.certificate.chainErrors ?? []).map((code) => ({
    label: t('tools.ssl-check.results.chainError'),
    value: code,
    level: 'danger' as const,
  })),
)

/** 证书链 PEM 全文。 */
const chainPem = computed(() =>
  (data.value?.chain?.chain ?? []).map((cert) => cert.pem.trimEnd()).join('\n\n'),
)

/** 协议矩阵列。 */
const protocolColumns = computed<DataTableColumn[]>(() => [
  { key: 'version', label: t('tools.ssl-check.results.protocolVersion'), mono: true },
  { key: 'status', label: t('tools.ssl-check.results.protocolStatus') },
  { key: 'cipher', label: t('tools.ssl-check.results.protocolCipher'), mono: true },
])

/** 协议矩阵行。 */
const protocolRows = computed<DataTableRow[]>(() =>
  (data.value?.protocols ?? []).map((item) => ({
    key: item.version,
    cells: {
      version: item.version,
      status: item.supported
        ? t('tools.ssl-check.results.supported')
        : `${t('tools.ssl-check.results.notSupported')}${item.reason ? ` (${item.reason})` : ''}`,
      cipher: item.negotiatedCipher ?? '-',
    },
  })),
)

/** 套件矩阵列。 */
const cipherColumns = computed<DataTableColumn[]>(() => [
  { key: 'version', label: t('tools.ssl-check.results.protocolVersion'), mono: true },
  { key: 'name', label: t('tools.ssl-check.results.cipherName'), mono: true },
  { key: 'family', label: t('tools.ssl-check.results.cipherFamily') },
  { key: 'bits', label: t('tools.ssl-check.results.cipherBits'), align: 'right' },
  { key: 'kex', label: t('tools.ssl-check.results.cipherKex') },
  { key: 'verdict', label: t('tools.ssl-check.results.cipherVerdict') },
])

/** 套件矩阵行。 */
const cipherRows = computed<DataTableRow[]>(() =>
  (data.value?.ciphers ?? []).map((item) => ({
    key: `${item.version}-${item.name}`,
    cells: {
      version: item.version,
      name: item.name,
      family: item.family,
      bits: String(item.bits),
      kex: item.kex,
      verdict: t(`tools.ssl-check.results.verdict.${item.verdict}`),
    },
  })),
)

/** 特性行。 */
const featureRows = computed<KeyValueRow[]>(() =>
  (data.value?.features ?? []).map((item) => ({
    label: t(`tools.ssl-check.features.${item.id}`),
    value: item.detail
      ? `${t(`tools.ssl-check.results.status.${item.status}`)} · ${item.detail}`
      : t(`tools.ssl-check.results.status.${item.status}`),
    level:
      item.status === 'supported'
        ? ('success' as const)
        : item.status === 'unavailable'
          ? ('warning' as const)
          : ('danger' as const),
  })),
)

/** 客户端模拟列。 */
const clientColumns = computed<DataTableColumn[]>(() => [
  { key: 'client', label: t('tools.ssl-check.results.clientName') },
  { key: 'engine', label: t('tools.ssl-check.results.clientEngine'), mono: true },
  { key: 'status', label: t('tools.ssl-check.results.clientStatus') },
  { key: 'negotiated', label: t('tools.ssl-check.results.clientNegotiated'), mono: true },
])

/** 客户端模拟行。 */
const clientRows = computed<DataTableRow[]>(() =>
  (data.value?.clients ?? []).map((item) => ({
    key: item.id,
    cells: {
      client: t(`tools.ssl-check.clients.${item.id}`),
      engine: `${item.engine} ${item.profileVersion}`,
      status: item.supported
        ? t('tools.ssl-check.results.supported')
        : `${t('tools.ssl-check.results.notSupported')} (${t(`tools.ssl-check.results.reason.${item.reason ?? 'handshakeFailure'}`)})`,
      negotiated: item.supported
        ? `${item.negotiatedVersion ?? '-'} · ${item.cipher ?? '-'}`
        : '-',
    },
  })),
)

/** 出网附加检查（吊销 / CAA / HSTS）。 */
const offsiteRows = computed<KeyValueRow[]>(() => {
  const result = data.value
  if (!result) {
    return []
  }
  const rows: KeyValueRow[] = []

  for (const entry of result.revocation.ocsp) {
    rows.push({
      label: `OCSP ${entry.url}`,
      value: entry.error ? `${entry.status} (${entry.error})` : entry.status,
      level: entry.status === 'revoked' ? ('danger' as const) : entry.error ? ('warning' as const) : ('success' as const),
    })
  }
  for (const entry of result.revocation.crl) {
    rows.push({
      label: `CRL ${entry.url}`,
      value: entry.error ? `${entry.status} (${entry.error})` : entry.status,
      level: entry.status === 'revoked' ? ('danger' as const) : entry.error ? ('warning' as const) : ('success' as const),
    })
  }
  if (result.caa.queried) {
    rows.push({
      label: t('tools.ssl-check.results.caaRecords'),
      value: result.caa.error
        ? `${result.caa.error}`
        : result.caa.records.length > 0
          ? result.caa.records.join('; ')
          : t('tools.ssl-check.results.caaNone'),
      level: result.caa.permitted ? ('success' as const) : ('danger' as const),
    })
  }
  rows.push({
    label: t('tools.ssl-check.results.hsts'),
    value: result.hsts.error
      ? result.hsts.error
      : result.hsts.present
        ? `max-age=${result.hsts.maxAge}${result.hsts.includeSubDomains ? '; includeSubDomains' : ''}${result.hsts.preload ? '; preload' : ''}`
        : t('tools.ssl-check.results.hstsAbsent'),
    level: result.hsts.present ? ('success' as const) : ('warning' as const),
  })
  return rows
})

/** JSON 视图全文。 */
const jsonText = computed(() => (data.value ? JSON.stringify(data.value, null, 2) : ''))

/** 结果纯文本。 */
const resultText = computed(() => {
  const result = data.value
  if (!result) {
    return ''
  }
  const lines = [
    `host: ${result.endpoint.host}:${result.endpoint.port} (${result.endpoint.addr})`,
    `version: ${result.negotiated.version}`,
    `cipher: ${result.negotiated.cipher}`,
    `grade: ${result.grade.letter} (${result.grade.score})`,
    `chainValid: ${String(result.certificate.chainValid)}`,
    `verified: ${String(result.certificate.verified)}`,
  ]
  for (const item of result.protocols) {
    lines.push(`protocol ${item.version}: ${item.supported ? 'supported' : 'no'}`)
  }
  lines.push(chainPem.value)
  return lines.filter((line) => line !== '').join('\n')
})

/** 进度文案。 */
const progressText = computed(() => {
  if (!progress.value) {
    return ''
  }
  const { phase, done, total, item } = progress.value
  const label = t(`tools.ssl-check.phases.${phase}`)
  const suffix = item ? ` · ${item}` : ''
  return `${label} ${done}/${total}${suffix}`
})

function cleanup(): void {
  if (socket) {
    socket.onmessage = null
    socket.onerror = null
    socket.onclose = null
    socket.close()
    socket = null
  }
}

onScopeDispose(cleanup)

/** 把服务端错误帧转成可本地化的错误对象。 */
function frameError(frame: {
  code: string
  message: string
  params?: Record<string, unknown>
}): LocalizedError {
  return localizeError(new ApiError(0, frame.code, frame.message, frame.params ?? {}))
}

function run(): void {
  cleanup()
  running.value = true
  controlState.value = 'loading'
  data.value = null
  error.value = null
  meta.value = null
  warnings.value = []
  progress.value = null

  let opened: WebSocket
  try {
    opened = new WebSocket(streamURL('ssl-check'))
  } catch {
    running.value = false
    controlState.value = 'error'
    error.value = frameError({ code: 'NETWORK_DISABLED', message: '' })
    return
  }
  socket = opened

  opened.onopen = () => {
    opened.send(
      JSON.stringify({
        host: host.value,
        port: port.value,
        ...(sni.value.trim() === '' ? {} : { sni: sni.value.trim() }),
        mode: mode.value,
        checks: selectedChecks.value,
        ...(selectedClients.value.length === 0 ? {} : { clients: selectedClients.value }),
      }),
    )
  }

  opened.onmessage = (event: MessageEvent<string>) => {
    let frame: StreamFrame
    try {
      frame = JSON.parse(event.data) as StreamFrame
    } catch {
      return
    }
    if (frame.type === 'progress') {
      progress.value = { phase: frame.phase, done: frame.done, total: frame.total, item: frame.item }
      return
    }
    if (frame.type === 'error') {
      error.value = frameError(frame.error)
      controlState.value = 'error'
      running.value = false
      return
    }
    data.value = frame.data
    warnings.value = frame.warnings ?? []
    controlState.value = 'success'
    running.value = false
  }

  opened.onerror = () => {
    if (controlState.value === 'loading') {
      error.value = frameError({ code: 'UPSTREAM_ERROR', message: '' })
      controlState.value = 'error'
    }
    running.value = false
  }

  opened.onclose = () => {
    if (controlState.value === 'loading') {
      error.value = frameError({ code: 'UPSTREAM_ERROR', message: '' })
      controlState.value = 'error'
    }
    running.value = false
    socket = null
  }
}

function reset(): void {
  cleanup()
  host.value = DEFAULT_EXAMPLE.host
  port.value = DEFAULT_EXAMPLE.port
  sni.value = DEFAULT_EXAMPLE.sni
  mode.value = 'full'
  selectedChecks.value = [...CHECK_IDS]
  selectedClients.value = []
  running.value = false
  controlState.value = 'idle'
  data.value = null
  error.value = null
  meta.value = null
  warnings.value = []
  progress.value = null
}
</script>

<template>
  <ToolPageLayout
    slug="ssl-check"
    name="SSL 证书检测"
    name-en="SSL Checker"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.ssl-check.fields.host') }}</span>
            <ElInput
              v-model="host"
              class="field__input"
              spellcheck="false"
              autocomplete="off"
            />
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.ssl-check.fields.port') }}</span>
            <ElInputNumber
              v-model="port"
              class="field__number"
              :min="1"
              :max="65535"
            />
          </label>
        </template>

        <template #options>
          <label class="field">
            <span class="field__label">{{ t('tools.ssl-check.fields.sni') }}</span>
            <ElInput
              v-model="sni"
              class="field__input"
              spellcheck="false"
              autocomplete="off"
            />
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.ssl-check.fields.mode') }}</span>
            <ElSelect
              v-model="mode"
              class="field__input"
            >
              <ElOption
                value="quick"
                :label="t('tools.ssl-check.options.modeQuick')"
              />
              <ElOption
                value="full"
                :label="t('tools.ssl-check.options.modeFull')"
              />
            </ElSelect>
          </label>

          <div class="field">
            <span class="field__label">{{ t('tools.ssl-check.fields.checks') }}</span>
            <div class="checks">
              <ElCheckbox
                v-for="id in CHECK_IDS"
                :key="id"
                :model-value="selectedChecks.includes(id)"
                :label="t(`tools.ssl-check.checks.${id}`)"
                @change="selectedChecks = selectedChecks.includes(id) ? selectedChecks.filter((item) => item !== id) : [...selectedChecks, id]"
              />
            </div>
          </div>

          <label class="field">
            <span class="field__label">{{ t('tools.ssl-check.fields.clients') }}</span>
            <ElSelect
              v-model="selectedClients"
              class="field__input"
              multiple
              collapse-tags
              :placeholder="t('tools.ssl-check.options.clientsAll')"
            >
              <ElOption
                v-for="id in CLIENT_IDS"
                :key="id"
                :value="id"
                :label="t(`tools.ssl-check.clients.${id}`)"
              />
            </ElSelect>
          </label>

          <p
            v-if="progressText !== ''"
            class="prose"
          >
            {{ t('tools.ssl-check.results.progress') }}: {{ progressText }}
          </p>
        </template>

        <template #footer>
          <RunBar
            :running="running"
            :disabled="host.trim() === ''"
            @run="run"
            @reset="reset"
            @example="reset"
          />
        </template>
      </ToolPanel>
    </template>

    <template #result>
      <ResultPanel
        v-model:active-view="activeView"
        :state="controlState"
        :text="resultText"
        :meta="meta"
        :error="error"
        :verdict="verdict"
        :title="t('tools.ssl-check.results.output')"
        :views="views"
        download-filename="ssl-scan.txt"
        @retry="run"
      >
        <template #view="{ viewId }">
          <template v-if="viewId === 'summary' && data">
            <StatGrid :items="sessionStats" />
            <StatGrid
              v-if="gradeStats.length > 0"
              class="stack"
              :items="gradeStats"
            />
            <KeyValueTable
              v-if="issueRows.length > 0"
              :rows="issueRows"
            />
            <KeyValueTable
              v-if="offsiteRows.length > 0"
              class="stack"
              :rows="offsiteRows"
            />
          </template>

          <template v-else-if="viewId === 'cert' && data">
            <KeyValueTable :rows="certRows" />
            <KeyValueTable
              v-if="chainErrorRows.length > 0"
              class="stack"
              :rows="chainErrorRows"
            />
          </template>

          <CertChainTable
            v-else-if="viewId === 'chain' && data && data.chain"
            :chain="data.chain.chain"
            :missing-root="data.chain.missingRoot"
          />

          <DataTable
            v-else-if="viewId === 'protocols' && data"
            class="table"
            :columns="protocolColumns"
            :rows="protocolRows"
            :empty="t('tools.ssl-check.results.empty')"
          />

          <DataTable
            v-else-if="viewId === 'ciphers' && data"
            class="table"
            :columns="cipherColumns"
            :rows="cipherRows"
            :empty="t('tools.ssl-check.results.empty')"
          />

          <KeyValueTable
            v-else-if="viewId === 'features' && data"
            :rows="featureRows"
          />

          <DataTable
            v-else-if="viewId === 'clients' && data"
            class="table"
            :columns="clientColumns"
            :rows="clientRows"
            :empty="t('tools.ssl-check.results.empty')"
          />

          <ProductText
            v-else-if="viewId === 'pem' && data"
            :label="t('tools.ssl-check.views.pem')"
            :content="chainPem"
            filename="ssl-chain.pem"
            mime="application/x-pem-file"
          />

          <ProductText
            v-else-if="viewId === 'json' && data"
            :label="t('tools.ssl-check.views.json')"
            :content="jsonText"
            filename="ssl-scan.json"
            mime="application/json"
          />
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.ssl-check.description') }}
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
.checks {
  display: flex;
  flex-wrap: wrap;
  gap: 4px 12px;
}

.stack {
  margin-top: 12px;
}

.table {
  width: 100%;
}
</style>
