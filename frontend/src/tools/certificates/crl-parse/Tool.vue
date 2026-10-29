<script setup lang="ts">
/**
 * 吊销列表解析工具页。
 *
 * 材料内容只存在于组件局部状态；离开页面即丢弃，
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
import { DEFAULT_EXAMPLE } from '@/tools/certificates/crl-parse/example'

/** 一条吊销记录。 */
interface RevokedItem {
  serial: string
  revokedAt: string
  reasonCode: number
  reasonText: string
}

/** 后端 crl-parse 工具的响应。 */
interface CRLParseResult {
  issuer: { field: string; value: string }[]
  issuerText: string
  thisUpdate: string
  nextUpdate: string
  signatureAlgorithm: string
  signatureAlgorithmOid: string
  number: number
  revoked: RevokedItem[]
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const input = ref('')
const inputFormat = ref('auto')

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
} = useToolRun<CRLParseResult>({ slug: 'crl-parse' })

const running = computed(() => state.value === 'loading')

/** 结果文本。 */
const resultText = computed(() => {
  const result = data.value
  if (!result) {
    return ''
  }
  const head = [
    `${t('tools.crl-parse.results.issuer')}: ${result.issuerText}`,
    `${t('tools.crl-parse.results.thisUpdate')}: ${result.thisUpdate}`,
    `${t('tools.crl-parse.results.nextUpdate')}: ${result.nextUpdate}`,
    `${t('tools.crl-parse.results.signatureAlgorithm')}: ${result.signatureAlgorithm} (${result.signatureAlgorithmOid})`,
    `${t('tools.crl-parse.results.number')}: ${result.number}`,
    '',
    t('tools.crl-parse.results.revoked'),
  ]
  for (const item of result.revoked) {
    head.push(`  ${item.serial || '?'}  ${item.revokedAt}  ${item.reasonText || item.reasonCode}`)
  }
  return head.join('\n')
})

const faq = computed<FaqItem[]>(() => tm('tools.crl-parse.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  void execute({
    input: { data: input.value, format: inputFormat.value },
  })
}

/** 重置表单与结果。 */
function reset(): void {
  input.value = ''
  inputFormat.value = 'auto'
  resetRun()
}

/** 填入示例。 */
function fillExample(): void {
  input.value = DEFAULT_EXAMPLE.data
  inputFormat.value = DEFAULT_EXAMPLE.format
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="crl-parse"
    :name="t('tools.crl-parse.name')"
    name-en="CRL Parser"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <CertTextInput
            v-model="input"
            v-model:format="inputFormat"
            :label="t('tools.crl-parse.fields.input')"
            :rows="8"
          />
        </template>

        <template #footer>
          <RunBar
            :running="running"
            :disabled="input === ''"
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
        :title="t('tools.crl-parse.results.info')"
        @retry="run"
      />
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.crl-parse.description') }}
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
