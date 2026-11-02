<script setup lang="ts">
/**
 * 证书链解析工具页。
 *
 * 证书内容只存在于组件局部状态；离开页面即丢弃，
 * **不写**任何浏览器存储（AGENTS.md §4.11）。
 *
 * 结果区分两个页签：摘要（链梯 + 逐张详情）/ 证书链 PEM；
 * 链上每一级都能点开详情浮层，查看该证书的完整字段（与证书解析同版式）。
 */
import { ElCheckbox, ElCheckboxGroup } from 'element-plus/es/components/checkbox/index'
import 'element-plus/es/components/checkbox/style/css'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import CertChainTable from '@/components/tool/CertChainTable.vue'
import CertDetailDialog, { type CertDetailPayload } from '@/components/tool/CertDetailDialog.vue'
import CopyButton from '@/components/tool/CopyButton.vue'
import MaterialInput from '@/components/tool/MaterialInput.vue'
import DownloadButton from '@/components/tool/DownloadButton.vue'
import ResultPanel, { type ResultView, type Verdict } from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import StatGrid, { type StatGridItem } from '@/components/tool/StatGrid.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useAutoRun } from '@/composables/useAutoRun'
import { useToolRun } from '@/composables/useToolRun'
import { joinChainPem, toList } from '@/tools/certificates/certDetail'
import { DEFAULT_EXAMPLE } from '@/tools/certificates/cert-chain-parse/example'
import type { CertChainParseResult, ChainCertificate } from '@/types/certificate'

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

/** 可选的指纹算法；与证书解析页保持一致。 */
const FINGERPRINT_ALGORITHMS = ['sha256', 'sha1', 'sm3', 'sha384', 'sha512', 'md5']

const bundle = ref('')
const bundleFormat = ref('auto')
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
} = useToolRun<CertChainParseResult>({ slug: 'cert-chain-parse' })

/**
 * 结果区的「作为新输入」：把**整链 PEM**回填到主输入框（仅内存，不写存储）。
 *
 * 回填的是 PEM 而不是结果摘要：摘要不是可解析的证书材料，
 * 回填后应当能直接再次解析。
 */
bindInput(() => {
  bundle.value = chainPem.value
})

const running = computed(() => state.value === 'loading')

/** 结果页签：摘要 / 链 PEM。 */
const views = computed<ResultView[]>(() => [
  { id: 'summary', label: t('tools.cert-chain-parse.views.summary') },
  { id: 'pem', label: t('tools.cert-chain-parse.views.pem') },
])

/** 当前页签。 */
const activeView = ref('summary')

/** 详情浮层的内容。 */
const detail = ref<CertDetailPayload | null>(null)

/** 打开某张证书的详情。 */
function openDetail(cert: ChainCertificate): void {
  detail.value = {
    info: cert.info,
    validity: cert.validity,
    fingerprints: toList(cert.fingerprints),
    publicKey: cert.publicKey,
    pem: cert.pem,
  }
}

/** 关闭详情浮层。 */
function closeDetail(): void {
  detail.value = null
}

/** 结论徽章：链是否完整。 */
const verdict = computed<Verdict | null>(() => {
  const result = data.value
  if (!result) {
    return null
  }
  const broken = result.chain.some((cert) => !cert.link.ok)
  if (broken || result.missingRoot) {
    return { level: 'warning', label: t('tool.verdict.warn') }
  }
  return { level: 'success', label: t('tool.verdict.pass') }
})

/** 摘要统计卡片。 */
const stats = computed<StatGridItem[]>(() => {
  const result = data.value
  if (!result) {
    return []
  }
  const broken = result.chain.filter((cert) => !cert.link.ok).length
  const items: StatGridItem[] = [
    { label: t('tools.cert-chain-parse.results.count'), value: String(result.count) },
    {
      label: t('tools.cert-chain-parse.results.ordered'),
      value: result.ordered ? t('common.states.yes') : t('common.states.no'),
      level: result.ordered ? ('success' as const) : ('warning' as const),
    },
    {
      label: t('tools.cert-chain-parse.results.missingRoot'),
      value: result.missingRoot ? t('common.states.yes') : t('common.states.no'),
      level: result.missingRoot ? ('warning' as const) : ('success' as const),
    },
    {
      label: t('tools.cert-chain-parse.results.selfSigned'),
      value: result.selfSigned ? t('common.states.yes') : t('common.states.no'),
    },
    {
      label: t('tools.cert-chain-parse.results.brokenLinks'),
      value: String(broken),
      level: broken > 0 ? ('danger' as const) : ('success' as const),
    },
  ]
  return items
})

/** 整链 PEM：按链序拼接（RFC 7468 严格形式，块间恰好一个换行）。 */
const chainPem = computed(() =>
  joinChainPem(toList(data.value?.chain).map((cert) => cert.pem)),
)

/** 结果文本：链摘要（复制 / 下载的数据源）。 */
const resultText = computed(() => {
  const result = data.value
  if (!result) {
    return ''
  }
  const head = [
    `${t('tools.cert-chain-parse.results.count')}: ${result.count}`,
    `${t('tools.cert-chain-parse.results.ordered')}: ${result.ordered}`,
    `${t('tools.cert-chain-parse.results.missingRoot')}: ${result.missingRoot}`,
    `${t('tools.cert-chain-parse.results.selfSigned')}: ${result.selfSigned}`,
    '',
  ]
  toList(result.chain).forEach((cert, index) => {
    head.push(`[${index}] ${cert.info.subjectText}`)
    head.push(`    issuer: ${cert.info.issuerText}`)
    head.push(`    serial: ${cert.info.serial}  notAfter: ${cert.info.notAfter}`)
    head.push(`    role: ${cert.role}  link: ${cert.link.method}  ok: ${cert.link.ok}`)
    for (const fp of toList(cert.fingerprints)) {
      head.push(`    ${fp.algorithm}: ${fp.value}`)
    }
  })
  return head.join('\n')
})

const faq = computed<FaqItem[]>(() => tm('tools.cert-chain-parse.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  if (bundle.value.trim() === '') {
    return
  }
  void execute({
    bundle: { data: bundle.value, format: bundleFormat.value },
    fingerprints: fingerprints.value,
  })
}

// 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。
useAutoRun({
  slug: 'cert-chain-parse',
  run,
  watch: [bundle, bundleFormat, fingerprints],
  enabled: () => bundle.value.trim() !== '',
})

/** 重置表单与结果。 */
function reset(): void {
  bundle.value = ''
  bundleFormat.value = 'auto'
  fingerprints.value = ['sha256']
  resetRun()
}

/** 填入示例。 */
function fillExample(): void {
  bundle.value = DEFAULT_EXAMPLE.bundle
  bundleFormat.value = 'auto'
  fingerprints.value = ['sha256']
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="cert-chain-parse"
    :name="t('tools.cert-chain-parse.name')"
    name-en="Certificate Chain Parser"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <MaterialInput
            v-model="bundle"
            v-model:format="bundleFormat"
            :label="t('tools.cert-chain-parse.fields.bundle')"
            :rows="12"
            slug="cert-chain-parse"
            append
          />
        </template>

        <template #options>
          <div class="field">
            <span class="field__label">{{ t('tools.cert-chain-parse.fields.fingerprints') }}</span>
            <ElCheckboxGroup
              v-model="fingerprints"
              class="field__checks"
              :aria-label="t('tools.cert-chain-parse.fields.fingerprints')"
            >
              <ElCheckbox
                v-for="alg in FINGERPRINT_ALGORITHMS"
                :key="alg"
                :value="alg"
              >
                {{ alg.toUpperCase() }}
              </ElCheckbox>
            </ElCheckboxGroup>
            <p class="field__hint">
              {{ t('tools.cert-chain-parse.hints.fingerprints') }}
            </p>
          </div>
        </template>

        <template #footer>
          <RunBar
            :running="running"
            :disabled="bundle === ''"
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
        :title="t('tools.cert-chain-parse.results.chain')"
        :views="views"
        :use-as-input="useAsInput"
        download-filename="chain-summary.txt"
        @retry="run"
      >
        <template #view="{ viewId }">
          <div
            v-if="viewId === 'summary' && data"
            class="chain-summary"
          >
            <StatGrid :items="stats" />
            <CertChainTable
              class="chain-summary__table"
              :chain="data.chain"
              :missing-root="data.missingRoot"
              @detail="openDetail"
            />
          </div>
          <div
            v-else-if="viewId === 'pem' && data"
            class="chain-pem"
          >
            <div class="chain-pem__head">
              <span class="chain-pem__label">{{ t('tools.cert-chain-parse.results.chainPem') }}</span>
              <div class="chain-pem__actions">
                <CopyButton
                  :text="chainPem"
                  :label="t('tool.cert.actions.copyPem')"
                />
                <DownloadButton
                  filename="chain.pem"
                  :content="chainPem"
                  mime="application/x-pem-file"
                />
              </div>
            </div>
            <pre class="product"><code>{{ chainPem }}</code></pre>
          </div>
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.cert-chain-parse.description') }}
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

  <CertDetailDialog
    :detail="detail"
    :close="closeDetail"
  />
</template>

<style scoped>
.field__checks {
  display: flex;
  flex-wrap: wrap;
  gap: var(--tb-space-1) var(--tb-space-3);
}

.chain-summary {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-3);
}

.chain-summary__table {
  margin-top: var(--tb-space-1);
}

.chain-pem {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-2);
}

.chain-pem__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--tb-space-2);
}

.chain-pem__actions {
  display: flex;
  align-items: center;
  gap: var(--tb-space-2);
}

.chain-pem__label {
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
