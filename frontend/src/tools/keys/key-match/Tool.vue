<script setup lang="ts">
/**
 * 密钥配对检测工具页。
 *
 * 密钥内容只存在于组件局部状态；离开页面即丢弃，
 * **不写**任何浏览器存储（AGENTS.md §4.11）。
 */
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import KeyInput from '@/components/tool/KeyInput.vue'
import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useToolRun } from '@/composables/useToolRun'
import { DEFAULT_EXAMPLE } from '@/tools/keys/key-match/example'

/** 后端 key-match 工具的响应。 */
interface KeyMatchResult {
  match: boolean
  reason: 'matched' | 'identical' | 'public_key_mismatch' | 'not_a_key_pair'
  aFingerprint: string
  bFingerprint: string
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const a = ref('')
const aFormat = ref('auto')
const b = ref('')
const bFormat = ref('auto')

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
} = useToolRun<KeyMatchResult>({ slug: 'key-match' })

const running = computed(() => state.value === 'loading')

/** 结果文本。 */
const resultText = computed(() => {
  const result = data.value
  if (!result) {
    return ''
  }
  return [
    `${t('tools.key-match.results.match')}: ${result.match}`,
    `${t('tools.key-match.results.reason')}: ${t(`tools.key-match.options.reasons.${result.reason}`)}`,
    `${t('tools.key-match.results.aFingerprint')}: ${result.aFingerprint}`,
    `${t('tools.key-match.results.bFingerprint')}: ${result.bFingerprint}`,
  ].join('\n')
})

const faq = computed<FaqItem[]>(() => tm('tools.key-match.faq') as FaqItem[])

/** 结论徽章：一眼看到两把密钥是否匹配。 */
const verdict = computed(() => {
  const result = data.value
  if (!result) {
    return null
  }
  return result.match
    ? { level: 'success' as const, label: t('tool.verdict.pass') }
    : { level: 'danger' as const, label: t('tool.verdict.fail') }
})

/** 运行当前输入。 */
function run(): void {
  void execute({
    a: { data: a.value, format: aFormat.value },
    b: { data: b.value, format: bFormat.value },
  })
}

/** 重置表单与结果。 */
function reset(): void {
  a.value = ''
  aFormat.value = 'auto'
  b.value = ''
  bFormat.value = 'auto'
  resetRun()
}

/** 填入示例。 */
function fillExample(): void {
  a.value = DEFAULT_EXAMPLE.a
  aFormat.value = 'auto'
  b.value = DEFAULT_EXAMPLE.b
  bFormat.value = 'auto'
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="key-match"
    :name="t('tools.key-match.name')"
    name-en="Key Pair Matcher"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <KeyInput
            v-model="a"
            v-model:format="aFormat"
            :label="t('tools.key-match.fields.a')"
          />
          <KeyInput
            v-model="b"
            v-model:format="bFormat"
            :label="t('tools.key-match.fields.b')"
          />
        </template>

        <template #footer>
          <RunBar
            :running="running"
            :disabled="a === '' || b === ''"
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
        :title="t('tools.key-match.results.result')"
        @retry="run"
      />
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.key-match.description') }}
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
