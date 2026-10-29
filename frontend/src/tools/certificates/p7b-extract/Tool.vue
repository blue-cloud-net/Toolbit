<script setup lang="ts">
/**
 * PKCS#7 提取工具页。
 *
 * 容器只存在于组件局部状态；离开页面即丢弃，**不写**浏览器存储。
 */
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import CertTextInput from '@/components/tool/CertTextInput.vue'
import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useToolRun } from '@/composables/useToolRun'
import { TEST_P7B_BASE64 } from '@/tools/certificates/fixtures-v060'

/** 后端 p7b-extract 工具的响应。 */
interface P7BExtractResult {
  certificates: { subjectText: string; issuerText: string; serial: string }[]
  count: number
  pems: string[]
}

interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const container = ref('')
const containerFormat = ref('auto')

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
} = useToolRun<P7BExtractResult>({ slug: 'p7b-extract' })

const running = computed(() => state.value === 'loading')

const resultText = computed(() => {
  const result = data.value
  if (!result) {
    return ''
  }
  return [`${t('tools.p7b-extract.results.count')}: ${result.count}`, '', ...result.pems].join('\n')
})

const faq = computed<FaqItem[]>(() => tm('tools.p7b-extract.faq') as FaqItem[])

function run(): void {
  void execute({
    container: { data: container.value, format: containerFormat.value },
  })
}

function reset(): void {
  container.value = ''
  containerFormat.value = 'auto'
  resetRun()
}

function fillExample(): void {
  reset()
  container.value = TEST_P7B_BASE64
  containerFormat.value = 'base64'
}
</script>

<template>
  <ToolPageLayout
    slug="p7b-extract"
    name="PKCS#7 提取"
    name-en="PKCS#7 Extractor"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <CertTextInput
            v-model="container"
            v-model:format="containerFormat"
            :label="t('tools.p7b-extract.fields.container')"
            :rows="8"
          />
        </template>

        <template #footer>
          <RunBar
            :running="running"
            :disabled="container === ''"
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
        :title="t('tools.p7b-extract.results.output')"
        @retry="run"
      />
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.p7b-extract.description') }}
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
