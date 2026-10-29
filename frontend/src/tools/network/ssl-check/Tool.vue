<script setup lang="ts">
/**
 * SSL 证书检测工具页。
 *
 * 走 WebSocket 流式端点：服务端按阶段推送 `progress`，
 * 结束后推 `result` 或 `error`。因此**不能**复用 `useToolRun`。
 */
import { computed, onScopeDispose, ref, shallowRef } from 'vue'
import { useI18n } from 'vue-i18n'

import { ApiError, streamURL } from '@/api/client'
import KeyValueTable, { type KeyValueRow } from '@/components/tool/KeyValueTable.vue'
import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
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
            <input
              v-model="host"
              type="text"
              class="field__input"
              spellcheck="false"
              autocomplete="off"
            >
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.ssl-check.fields.port') }}</span>
            <input
              v-model.number="port"
              type="number"
              class="field__input"
              min="1"
              max="65535"
            >
          </label>
        </template>

        <template #options>
          <label class="field">
            <span class="field__label">{{ t('tools.ssl-check.fields.sni') }}</span>
            <input
              v-model="sni"
              type="text"
              class="field__input"
              spellcheck="false"
              autocomplete="off"
            >
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
        :state="controlState"
        :text="resultText"
        :meta="meta"
        :error="error"
        :verdict="verdict"
        :title="t('tools.ssl-check.results.output')"
        @retry="run"
      >
        <template
          v-if="checkRows.length > 0"
          #extra
        >
          <KeyValueTable :rows="checkRows" />
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
