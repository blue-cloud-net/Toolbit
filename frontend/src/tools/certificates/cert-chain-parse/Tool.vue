<script setup lang="ts">
/**
 * 证书链解析工具页。
 *
 * 证书内容只存在于组件局部状态；离开页面即丢弃，
 * **不写**任何浏览器存储（AGENTS.md §4.11）。
 */
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import CertTextInput from '@/components/tool/CertTextInput.vue'
import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useToolRun } from '@/composables/useToolRun'
import { DEFAULT_EXAMPLE } from '@/tools/certificates/cert-chain-parse/example'

/** 链上一张证书。 */
interface ChainCert {
  pem: string
  info: {
    subjectText: string
    issuerText: string
    serial: string
    notAfter: string
    isCa: boolean
  }
}

/** 后端 cert-chain-parse 工具的响应。 */
interface ChainParseResult {
  count: number
  chain: ChainCert[]
  ordered: boolean
  missingRoot: boolean
  selfSigned: boolean
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const bundle = ref('')
const bundleFormat = ref('auto')

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
} = useToolRun<ChainParseResult>({ slug: 'cert-chain-parse' })

const running = computed(() => state.value === 'loading')

/** 结果文本。 */
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
  result.chain.forEach((cert, index) => {
    head.push(`[${index}] ${cert.info.subjectText}`)
    head.push(`    issuer: ${cert.info.issuerText}`)
    head.push(`    serial: ${cert.info.serial}  notAfter: ${cert.info.notAfter}`)
  })
  return head.join('\n')
})

const faq = computed<FaqItem[]>(() => tm('tools.cert-chain-parse.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  void execute({
    bundle: { data: bundle.value, format: bundleFormat.value },
  })
}

/** 重置表单与结果。 */
function reset(): void {
  bundle.value = ''
  bundleFormat.value = 'auto'
  resetRun()
}

/** 填入示例。 */
function fillExample(): void {
  bundle.value = DEFAULT_EXAMPLE.bundle
  bundleFormat.value = 'auto'
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
          <CertTextInput
            v-model="bundle"
            v-model:format="bundleFormat"
            :label="t('tools.cert-chain-parse.fields.bundle')"
            :rows="12"
          />
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
        :state="state"
        :text="resultText"
        :meta="meta"
        :error="error"
        :title="t('tools.cert-chain-parse.results.chain')"
        @retry="run"
      />
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
</template>
