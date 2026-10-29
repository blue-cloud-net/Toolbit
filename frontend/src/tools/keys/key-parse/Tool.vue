<script setup lang="ts">
/**
 * 密钥解析工具页。
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
import { DEFAULT_EXAMPLE } from '@/tools/keys/key-parse/example'

/** 后端 key-parse 工具的响应。 */
interface KeyParseResult {
  algorithm: string
  keyType: 'private' | 'public'
  format: string
  encrypted: boolean
  bits: number
  curve: string
  fingerprintSha256: string
  params?: {
    n: string
    e: string
    d: string
    p: string
    q: string
    dp: string
    dq: string
    qi: string
  }
  normalizedPem: string
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const key = ref('')
const keyFormat = ref('auto')
const passphrase = ref('')

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
} = useToolRun<KeyParseResult>({ slug: 'key-parse' })

const running = computed(() => state.value === 'loading')

/** 结果文本：摘要信息 + 规范化 PEM。 */
const resultText = computed(() => {
  const result = data.value
  if (!result) {
    return ''
  }
  const head = [
    `${t('tools.key-parse.results.algorithm')}: ${result.algorithm}`,
    `${t('tools.key-parse.results.keyType')}: ${result.keyType}`,
    `${t('tools.key-parse.results.format')}: ${result.format}`,
  ]
  if (result.bits) {
    head.push(`${t('tools.key-parse.results.bits')}: ${result.bits}`)
  }
  if (result.curve) {
    head.push(`${t('tools.key-parse.results.curve')}: ${result.curve}`)
  }
  head.push(`${t('tools.key-parse.results.fingerprint')}: ${result.fingerprintSha256}`)
  return [...head, '', result.normalizedPem].join('\n')
})

const faq = computed<FaqItem[]>(() => tm('tools.key-parse.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  void execute({
    key: { data: key.value, format: keyFormat.value },
    passphrase: passphrase.value,
  })
}

/** 重置表单与结果。 */
function reset(): void {
  key.value = ''
  keyFormat.value = 'auto'
  passphrase.value = ''
  resetRun()
}

/** 填入示例。 */
function fillExample(): void {
  key.value = DEFAULT_EXAMPLE.key
  keyFormat.value = 'auto'
  passphrase.value = ''
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="key-parse"
    :name="t('tools.key-parse.name')"
    name-en="Key Parser"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <KeyInput
            v-model="key"
            v-model:format="keyFormat"
            :label="t('tools.key-parse.fields.key')"
            :show-passphrase="true"
          />
        </template>

        <template #footer>
          <RunBar
            :running="running"
            :disabled="key === ''"
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
        :title="t('tools.key-parse.results.info')"
        @retry="run"
      />
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.key-parse.description') }}
      </p>
      <p
        v-if="data?.params"
        class="prose"
      >
        {{ t('tools.key-parse.results.params') }}
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
