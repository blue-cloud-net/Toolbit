<script setup lang="ts">
/**
 * 域名验证工具页。
 *
 * 三种方式：DNS TXT、DNS CNAME、站点文件。
 */
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useToolRun } from '@/composables/useToolRun'

import { DEFAULT_EXAMPLE } from './example'

/** 单解析器结论。 */
interface ResolverOutcome {
  resolver: string
  values: string[]
  matched: boolean
  error: { code: string; message: string } | null
}

/** 后端 domain-verify 工具的响应。 */
interface DomainVerifyResult {
  type: string
  domain: string
  expected: string
  matched: boolean
  perResolver: ResolverOutcome[]
  httpStatus: number
  bodyPreview: string
}

interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const type = ref<'dns-txt' | 'dns-cname' | 'http-file'>(DEFAULT_EXAMPLE.type)
const domain = ref(DEFAULT_EXAMPLE.domain)
const expected = ref(DEFAULT_EXAMPLE.expected)
const filePath = ref(DEFAULT_EXAMPLE.filePath)

const { state, data, meta, error, run: execute, reset: resetRun } =
  useToolRun<DomainVerifyResult>({ slug: 'domain-verify' })

const running = computed(() => state.value === 'loading')

const resultText = computed(() => {
  if (!data.value) {
    return ''
  }
  const value = data.value
  const lines = [`matched: ${String(value.matched)}`]
  if (value.type === 'http-file') {
    lines.push(`httpStatus: ${value.httpStatus}`, '', value.bodyPreview)
  }
  for (const outcome of value.perResolver) {
    const state = outcome.error
      ? `! ${outcome.error.code}`
      : outcome.matched
        ? 'match'
        : 'no match'
    lines.push(`# ${outcome.resolver} — ${state}`)
    for (const item of outcome.values) {
      lines.push(`  ${item}`)
    }
  }
  return lines.join('\n')
})

const faq = computed<FaqItem[]>(() => tm('tools.domain-verify.faq') as FaqItem[])

/** 汇总：各解析器的成功 / 失败数。 */
const summary = computed(() => {
  const result = data.value
  if (!result || result.perResolver.length === 0) {
    return null
  }
  const succeeded = result.perResolver.filter((item) => item.error === null).length
  return {
    succeeded,
    failed: result.perResolver.length - succeeded,
    total: result.perResolver.length,
  }
})

/** 结论徽章：期望值是否匹配。 */
const verdict = computed(() => {
  const result = data.value
  if (!result) {
    return null
  }
  return result.matched
    ? { level: 'success' as const, label: t('tool.verdict.pass') }
    : { level: 'danger' as const, label: t('tool.verdict.fail') }
})

const canRun = computed(
  () => domain.value.trim() !== '' && expected.value.trim() !== '',
)

function run(): void {
  const payload: Record<string, unknown> = {
    type: type.value,
    domain: domain.value.trim(),
    expected: expected.value,
  }
  if (type.value === 'http-file') {
    payload.filePath = filePath.value.trim()
  }
  void execute(payload)
}

function reset(): void {
  type.value = DEFAULT_EXAMPLE.type
  domain.value = DEFAULT_EXAMPLE.domain
  expected.value = DEFAULT_EXAMPLE.expected
  filePath.value = DEFAULT_EXAMPLE.filePath
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="domain-verify"
    name="域名验证"
    name-en="Domain Verifier"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.domain-verify.fields.type') }}</span>
            <select
              v-model="type"
              class="field__select"
            >
              <option value="dns-txt">
                {{ t('tools.domain-verify.options.dnsTxt') }}
              </option>
              <option value="dns-cname">
                {{ t('tools.domain-verify.options.dnsCname') }}
              </option>
              <option value="http-file">
                {{ t('tools.domain-verify.options.httpFile') }}
              </option>
            </select>
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.domain-verify.fields.domain') }}</span>
            <input
              v-model="domain"
              type="text"
              class="field__input"
              spellcheck="false"
              autocomplete="off"
            >
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.domain-verify.fields.expected') }}</span>
            <input
              v-model="expected"
              type="text"
              class="field__input"
              spellcheck="false"
              autocomplete="off"
            >
          </label>
        </template>

        <template #options>
          <label
            v-if="type === 'http-file'"
            class="field"
          >
            <span class="field__label">{{ t('tools.domain-verify.fields.filePath') }}</span>
            <input
              v-model="filePath"
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
            :disabled="!canRun"
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
        :title="t('tools.domain-verify.results.output')"
        @retry="run"
      >
        <template
          v-if="summary"
          #summary
        >
          <span>{{ t('tool.result.summary.succeeded') }} {{ summary.succeeded }} / {{ summary.total }}</span>
          <span v-if="summary.failed > 0">{{ t('tool.result.summary.failed') }} {{ summary.failed }}</span>
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.domain-verify.description') }}
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
