<script setup lang="ts">
/**
 * 证书链验证工具页。
 *
 * 证书内容只存在于组件局部状态；离开页面即丢弃，
 * **不写**任何浏览器存储（AGENTS.md §4.11）。
 */
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import CertTextInput from '@/components/tool/CertTextInput.vue'
import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useToolRun } from '@/composables/useToolRun'
import { DEFAULT_EXAMPLE } from '@/tools/certificates/cert-chain-verify/example'

/** 后端 cert-chain-verify 工具的响应。 */
interface ChainVerifyResult {
  valid: boolean
  errors: { code: string; message: string; depth: number }[]
  details: { subject: string; issuer: string; valid: boolean; reason: string }[]
  hostnameMatch: boolean
  hostnameError: string
  implicitTrust: boolean
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const bundle = ref('')
const bundleFormat = ref('auto')
const trustedRoots = ref('')
const trustedRootsFormat = ref('auto')
const atTime = ref('')
const host = ref('')

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
  bindInput,
  useAsInput,
} = useToolRun<ChainVerifyResult>({ slug: 'cert-chain-verify' })

/** 结果区的「作为新输入」：把结果回填到主输入框（仅内存，不写存储）。 */
bindInput((text) => {
  bundle.value = text
})

const running = computed(() => state.value === 'loading')

/** 结果文本。 */
const resultText = computed(() => {
  const result = data.value
  if (!result) {
    return ''
  }
  const head = [
    `${t('tools.cert-chain-verify.results.valid')}: ${result.valid}`,
    `${t('tools.cert-chain-verify.results.implicitTrust')}: ${result.implicitTrust}`,
    `${t('tools.cert-chain-verify.results.hostnameMatch')}: ${result.hostnameMatch}`,
  ]
  if (result.hostnameError) {
    head.push(`${t('tools.cert-chain-verify.results.hostnameError')}: ${result.hostnameError}`)
  }
  if (result.errors.length > 0) {
    head.push('')
    head.push(t('tools.cert-chain-verify.results.errors'))
    for (const item of result.errors) {
      head.push(`  [${item.depth}] ${item.code}: ${item.message}`)
    }
  }
  if (result.details.length > 0) {
    head.push('')
    head.push(t('tools.cert-chain-verify.results.details'))
    for (const item of result.details) {
      head.push(`  ${item.valid ? '✓' : '✗'} ${item.subject}`)
      if (!item.valid && item.reason) {
        head.push(`      ${item.reason}`)
      }
    }
  }
  return head.join('\n')
})

const faq = computed<FaqItem[]>(() => tm('tools.cert-chain-verify.faq') as FaqItem[])

/** 结论徽章：一眼看到「过不过」。 */
const verdict = computed(() => {
  const result = data.value
  if (!result) {
    return null
  }
  if (!result.valid) {
    return { level: 'danger' as const, label: t('tool.verdict.fail') }
  }
  if (!result.hostnameMatch) {
    return { level: 'warning' as const, label: t('tool.verdict.warn') }
  }
  return { level: 'success' as const, label: t('tool.verdict.pass') }
})

/** 运行当前输入。 */
function run(): void {
  const roots: { data: string; format: string }[] = []
  if (trustedRoots.value.trim() !== '') {
    roots.push({ data: trustedRoots.value, format: trustedRootsFormat.value })
  }
  void execute({
    bundle: { data: bundle.value, format: bundleFormat.value },
    trustedRoots: roots,
    atTime: atTime.value,
    host: host.value,
  })
}

/** 重置表单与结果。 */
function reset(): void {
  bundle.value = ''
  bundleFormat.value = 'auto'
  trustedRoots.value = ''
  trustedRootsFormat.value = 'auto'
  atTime.value = ''
  host.value = ''
  resetRun()
}

/** 填入示例。 */
function fillExample(): void {
  bundle.value = DEFAULT_EXAMPLE.bundle
  bundleFormat.value = 'auto'
  trustedRoots.value = DEFAULT_EXAMPLE.roots
  trustedRootsFormat.value = 'auto'
  atTime.value = ''
  host.value = DEFAULT_EXAMPLE.host
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="cert-chain-verify"
    :name="t('tools.cert-chain-verify.name')"
    name-en="Certificate Chain Verifier"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <CertTextInput
            v-model="bundle"
            v-model:format="bundleFormat"
            :label="t('tools.cert-chain-verify.fields.bundle')"
            :rows="10"
          />

          <CertTextInput
            v-model="trustedRoots"
            v-model:format="trustedRootsFormat"
            :label="t('tools.cert-chain-verify.fields.trustedRoots')"
            :rows="6"
          />
        </template>

        <template #options>
          <label class="field">
            <span class="field__label">{{ t('tools.cert-chain-verify.fields.host') }}</span>
            <ElInput
              v-model="host"
              class="field__input"
              placeholder="leaf.example.com"
              spellcheck="false"
              autocomplete="off"
            />
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.cert-chain-verify.fields.atTime') }}</span>
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
        :verdict="verdict"
        :title="t('tools.cert-chain-verify.results.info')"
        :use-as-input="useAsInput"
        @retry="run"
      />
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.cert-chain-verify.description') }}
      </p>
      <p
        v-if="data?.implicitTrust"
        class="prose"
      >
        {{ t('tools.cert-chain-verify.hints.implicitTrust') }}
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
