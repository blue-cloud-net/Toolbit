<script setup lang="ts">
/**
 * 证书链获取工具页。
 *
 * 只读取 TLS 握手中服务端发来的证书，不做补链。
 */
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useToolRun } from '@/composables/useToolRun'

import { DEFAULT_EXAMPLE } from './example'

/** 链摘要。 */
interface ChainSummary {
  count: number
  subject: string
  issuer: string
  notBefore: string
  notAfter: string
  selfSigned: boolean
  chainValid: boolean
  chainErrors: string[]
  verified: boolean
}

/** 后端 cert-chain-fetch 工具的响应。 */
interface ChainFetchResult {
  host: string
  port: number
  chainPem: string
  chainSummary: ChainSummary
}

interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const host = ref(DEFAULT_EXAMPLE.host)
const port = ref(DEFAULT_EXAMPLE.port)
const serverName = ref(DEFAULT_EXAMPLE.serverName)

const { state, data, meta, error, run: execute, reset: resetRun } =
  useToolRun<ChainFetchResult>({ slug: 'cert-chain-fetch' })

const running = computed(() => state.value === 'loading')

const resultText = computed(() => {
  if (!data.value) {
    return ''
  }
  const summary = data.value.chainSummary
  const header = [
    `# ${data.value.host}:${data.value.port} — ${summary.count} cert(s)`,
    `# subject: ${summary.subject}`,
    `# issuer: ${summary.issuer}`,
    `# valid: ${summary.notBefore} ~ ${summary.notAfter}`,
    `# chainValid: ${String(summary.chainValid)}  verified: ${String(summary.verified)}`,
  ]
  if (summary.chainErrors.length > 0) {
    header.push(`# errors: ${summary.chainErrors.join(', ')}`)
  }
  return `${header.join('\n')}\n\n${data.value.chainPem}`
})

const faq = computed<FaqItem[]>(() => tm('tools.cert-chain-fetch.faq') as FaqItem[])

/** 结论徽章：拉到的链是否验证通过。 */
const verdict = computed(() => {
  const result = data.value
  if (!result) {
    return null
  }
  if (!result.chainSummary.chainValid) {
    return { level: 'danger' as const, label: t('tool.verdict.fail') }
  }
  if (!result.chainSummary.verified) {
    return { level: 'warning' as const, label: t('tool.verdict.warn') }
  }
  return { level: 'success' as const, label: t('tool.verdict.pass') }
})

function run(): void {
  const payload: Record<string, unknown> = { host: host.value, port: port.value }
  if (serverName.value.trim() !== '') {
    payload.serverName = serverName.value.trim()
  }
  void execute(payload)
}

function reset(): void {
  host.value = DEFAULT_EXAMPLE.host
  port.value = DEFAULT_EXAMPLE.port
  serverName.value = DEFAULT_EXAMPLE.serverName
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="cert-chain-fetch"
    name="证书链获取"
    name-en="Certificate Chain Fetcher"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.cert-chain-fetch.fields.host') }}</span>
            <input
              v-model="host"
              type="text"
              class="field__input"
              spellcheck="false"
              autocomplete="off"
            >
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.cert-chain-fetch.fields.port') }}</span>
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
            <span class="field__label">{{ t('tools.cert-chain-fetch.fields.serverName') }}</span>
            <input
              v-model="serverName"
              type="text"
              class="field__input"
              spellcheck="false"
              autocomplete="off"
            >
          </label>
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
        :state="state"
        :text="resultText"
        :meta="meta"
        :error="error"
        :verdict="verdict"
        :title="t('tools.cert-chain-fetch.results.output')"
        @retry="run"
      />
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.cert-chain-fetch.description') }}
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
