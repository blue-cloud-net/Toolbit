<script setup lang="ts">
/**
 * SSL 证书检测工具页。
 *
 * 走 WebSocket 流式端点：服务端按阶段推送 `progress`，
 * 结束后推 `result` 或 `error`。因此**不能**复用 `useToolRun`。
 */
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { ElInputNumber } from 'element-plus/es/components/input-number/index'
import 'element-plus/es/components/input-number/style/css'
import { computed, onScopeDispose, ref, shallowRef } from 'vue'
import { useI18n } from 'vue-i18n'

import { ApiError, streamURL } from '@/api/client'
import DataTable, { type DataTableColumn, type DataTableRow } from '@/components/tool/DataTable.vue'
import KeyValueTable, { type KeyValueRow } from '@/components/tool/KeyValueTable.vue'
import ProductText from '@/components/tool/ProductText.vue'
import ResultPanel, { type ResultView } from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import StatGrid, { type StatGridItem } from '@/components/tool/StatGrid.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { localizeError, type LocalizedError } from '@/i18n/errors'
import type { Meta } from '@/types/api'

import { DEFAULT_EXAMPLE } from './example'

/** 链上的一张证书。 */
interface ChainCert {
  pem: string
  info: {
    subjectText: string
    issuerText: string
    notBefore: string
    notAfter: string
    selfSigned: boolean
  } | null
}

/** 探测结果。 */
interface SSLProbeResult {
  host: string
  port: number
  addr: string
  protocol: string
  negotiatedVersion: string
  cipherSuite: string
  certificates: ChainCert[]
  chainValid: boolean
  chainErrors: string[]
  verified: boolean
  implicitTrust: boolean
  ocspStapled: boolean
  grade: string
  checks: { id: string; label: string; passed: boolean; detail?: string }[]
}

/** 服务端帧。 */
type StreamFrame =
  | { type: 'progress'; phase: string; done: number; total: number }
  | { type: 'result'; data: SSLProbeResult }
  | { type: 'error'; error: { code: string; message: string; params?: Record<string, unknown> } }

interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const host = ref(DEFAULT_EXAMPLE.host)
const port = ref(DEFAULT_EXAMPLE.port)
const sni = ref(DEFAULT_EXAMPLE.sni)

/** 结果区状态；与 `useToolRun` 的状态机保持一致的四态。 */
const controlState = ref<'idle' | 'loading' | 'success' | 'error'>('idle')

const running = ref(false)
const data = shallowRef<SSLProbeResult | null>(null)
const meta = shallowRef<Meta | null>(null)
const error = shallowRef<LocalizedError | null>(null)
const progress = ref<{ phase: string; done: number; total: number } | null>(null)

let socket: WebSocket | null = null

const faq = computed<FaqItem[]>(() => tm('tools.ssl-check.faq') as FaqItem[])

/** 结论徽章：按综合评级给出。 */
const verdict = computed(() => {
  const result = data.value
  if (!result || !result.grade) {
    return null
  }
  const label = t('tools.ssl-check.results.grade', { grade: result.grade })
  if (result.grade === 'A+' || result.grade === 'A') {
    return { level: 'success' as const, label }
  }
  if (result.grade === 'B') {
    return { level: 'warning' as const, label }
  }
  return { level: 'danger' as const, label }
})

/** 逐项检查结论。 */
const checkRows = computed<KeyValueRow[]>(() => {
  const result = data.value
  if (!result || result.checks.length === 0) {
    return []
  }
  return result.checks.map((check) => ({
    label: check.label,
    value: check.detail || (check.passed ? t('tools.ssl-check.results.checkPass') : t('tools.ssl-check.results.checkFail')),
    level: check.passed ? ('success' as const) : ('danger' as const),
  }))
})

/** 结果页签：摘要 / 证书链 / PEM。 */
const views = computed<ResultView[]>(() => [
  { id: 'summary', label: t('tools.ssl-check.views.summary') },
  { id: 'chain', label: t('tools.ssl-check.views.chain') },
  { id: 'pem', label: t('tools.ssl-check.views.pem') },
])

/** 当前页签；默认停在「摘要」。 */
const activeView = ref('summary')

/** 握手参数卡片。 */
const sessionStats = computed<StatGridItem[]>(() => {
  const result = data.value
  if (!result) {
    return []
  }
  const items: StatGridItem[] = [
    { label: t('tools.ssl-check.results.endpoint'), value: `${result.host}:${result.port}` },
    { label: t('tools.ssl-check.results.address'), value: result.addr, wide: true },
    { label: t('tools.ssl-check.results.version'), value: result.negotiatedVersion },
    { label: t('tools.ssl-check.results.cipher'), value: result.cipherSuite, wide: true },
    {
      label: t('tools.ssl-check.results.chainValid'),
      value: result.chainValid ? t('common.states.yes') : t('common.states.no'),
      level: result.chainValid ? ('success' as const) : ('danger' as const),
    },
    {
      label: t('tools.ssl-check.results.verified'),
      value: result.verified ? t('common.states.yes') : t('common.states.no'),
      level: result.verified ? ('success' as const) : ('danger' as const),
    },
    {
      label: t('tools.ssl-check.results.implicitTrust'),
      value: result.implicitTrust ? t('common.states.yes') : t('common.states.no'),
      level: result.implicitTrust ? ('success' as const) : ('warning' as const),
    },
    {
      label: t('tools.ssl-check.results.ocspStapled'),
      value: result.ocspStapled ? t('common.states.yes') : t('common.states.no'),
    },
    { label: t('tools.ssl-check.results.certCount'), value: String(result.certificates.length) },
  ]
  if (result.chainErrors.length > 0) {
    items.push({
      label: t('tools.ssl-check.results.chainErrors'),
      value: result.chainErrors.join(', '),
      level: 'danger',
      wide: true,
    })
  }
  return items
})

/** 证书链表格列定义。 */
const chainColumns = computed<DataTableColumn[]>(() => [
  { key: 'index', label: t('tools.ssl-check.results.certIndex'), align: 'right' },
  { key: 'subject', label: t('tools.ssl-check.results.certSubject'), mono: true },
  { key: 'issuer', label: t('tools.ssl-check.results.certIssuer'), mono: true },
  { key: 'validity', label: t('tools.ssl-check.results.certValidity'), mono: true },
])

/** 证书链表格行。 */
const chainRows = computed<DataTableRow[]>(() =>
  (data.value?.certificates ?? []).map((cert, index) => ({
    key: `${index}-${cert.info?.subjectText ?? 'unparsed'}`,
    cells: {
      index: String(index),
      subject: cert.info?.subjectText ?? '-',
      issuer: cert.info?.issuerText ?? '-',
      validity: cert.info ? `${cert.info.notBefore} ~ ${cert.info.notAfter}` : '-',
    },
  })),
)

/** 证书链 PEM 全文。 */
const chainPem = computed(() =>
  (data.value?.certificates ?? []).map((cert) => cert.pem.trimEnd()).join('\n\n'),
)

const progressText = computed(() => {
  if (!progress.value) {
    return ''
  }
  const { phase, done, total } = progress.value
  return `${phase} ${done}/${total}`
})

const resultText = computed(() => {
  if (!data.value) {
    return ''
  }
  const value = data.value
  const lines = [
    `host: ${value.host}:${value.port} (${value.addr})`,
    `version: ${value.negotiatedVersion}`,
    `cipherSuite: ${value.cipherSuite}`,
    `chainValid: ${String(value.chainValid)}`,
    `verified: ${String(value.verified)}`,
    `implicitTrust: ${String(value.implicitTrust)}`,
    `ocspStapled: ${String(value.ocspStapled)}`,
    `certificates: ${value.certificates.length}`,
  ]
  if (value.chainErrors.length > 0) {
    lines.push(`chainErrors: ${value.chainErrors.join(', ')}`)
  }
  lines.push('')
  for (const [index, cert] of value.certificates.entries()) {
    lines.push(`# [${index}] ${cert.info?.subjectText ?? '(unparsed)'}`)
    if (cert.info) {
      lines.push(`  issuer: ${cert.info.issuerText}`)
      lines.push(`  valid: ${cert.info.notBefore} ~ ${cert.info.notAfter}`)
    }
    lines.push(cert.pem.trimEnd(), '')
  }
  return lines.join('\n').trimEnd()
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
      progress.value = { phase: frame.phase, done: frame.done, total: frame.total }
      return
    }
    if (frame.type === 'error') {
      error.value = frameError(frame.error)
      controlState.value = 'error'
      running.value = false
      return
    }
    data.value = frame.data
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
  running.value = false
  controlState.value = 'idle'
  data.value = null
  error.value = null
  meta.value = null
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
        download-filename="ssl-chain.pem"
        download-mime="application/x-pem-file"
        @retry="run"
      >
        <template #view="{ viewId }">
          <template v-if="viewId === 'summary' && data">
            <StatGrid :items="sessionStats" />

            <KeyValueTable
              v-if="checkRows.length > 0"
              :rows="checkRows"
            />
          </template>

          <DataTable
            v-else-if="viewId === 'chain' && data"
            class="table"
            :columns="chainColumns"
            :rows="chainRows"
            :empty="t('tools.ssl-check.results.noCerts')"
          />

          <ProductText
            v-else-if="viewId === 'pem' && data"
            :label="t('tools.ssl-check.views.pem')"
            :content="chainPem"
            filename="ssl-chain.pem"
            mime="application/x-pem-file"
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
