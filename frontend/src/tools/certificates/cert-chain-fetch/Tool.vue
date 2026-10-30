<script setup lang="ts">
/**
 * 证书链获取工具页。
 *
 * 只读取 TLS 握手中服务端发来的证书，不做补链。
 */
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { ElInputNumber } from 'element-plus/es/components/input-number/index'
import 'element-plus/es/components/input-number/style/css'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import ResultPanel, { type ResultView } from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import StatGrid, { type StatGridItem } from '@/components/tool/StatGrid.vue'
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

/** 结果页签：摘要 / 证书链 PEM。 */
const views = computed<ResultView[]>(() => [
  { id: 'summary', label: t('tools.cert-chain-fetch.views.summary') },
  { id: 'pem', label: t('tools.cert-chain-fetch.views.pem') },
])

/** 当前页签。 */
const activeView = ref('summary')

/** 摘要统计卡片。 */
const stats = computed<StatGridItem[]>(() => {
  const result = data.value
  if (!result) {
    return []
  }
  const summary = result.chainSummary
  const items: StatGridItem[] = [
    { label: t('tools.cert-chain-fetch.results.target'), value: `${result.host}:${result.port}` },
    { label: t('tools.cert-chain-fetch.results.count'), value: String(summary.count) },
    { label: t('tools.cert-chain-fetch.results.subject'), value: summary.subject },
    { label: t('tools.cert-chain-fetch.results.issuer'), value: summary.issuer },
    {
      label: t('tools.cert-chain-fetch.results.validity'),
      value: `${summary.notBefore} → ${summary.notAfter}`,
    },
    {
      label: t('tools.cert-chain-fetch.results.chainValid'),
      value: summary.chainValid ? t('tool.verdict.pass') : t('tool.verdict.fail'),
      level: summary.chainValid ? ('success' as const) : ('danger' as const),
    },
    {
      label: t('tools.cert-chain-fetch.results.verified'),
      value: summary.verified ? t('common.states.yes') : t('common.states.no'),
      level: summary.verified ? ('success' as const) : ('warning' as const),
    },
  ]
  if (summary.chainErrors.length > 0) {
    items.push({
      label: t('tools.cert-chain-fetch.results.chainErrors'),
      value: summary.chainErrors.join(', '),
      level: 'danger',
    })
  }
  return items
})

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
            <ElInput
              v-model="host"
              class="field__input"
              spellcheck="false"
              autocomplete="off"
            />
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.cert-chain-fetch.fields.port') }}</span>
            <ElInputNumber
              v-model="port"
              class="field__input"
              :min="1"
              :max="65535"
            />
          </label>
        </template>

        <template #options>
          <label class="field">
            <span class="field__label">{{ t('tools.cert-chain-fetch.fields.serverName') }}</span>
            <ElInput
              v-model="serverName"
              class="field__input"
              spellcheck="false"
              autocomplete="off"
            />
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
        v-model:active-view="activeView"
        :state="state"
        :text="resultText"
        :meta="meta"
        :error="error"
        :verdict="verdict"
        :title="t('tools.cert-chain-fetch.results.output')"
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
          ><code>{{ data.chainPem }}</code></pre>
        </template>
      </ResultPanel>
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
