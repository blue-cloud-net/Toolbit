<script setup lang="ts">
/**
 * PKCS#7 证书袋打包工具页。
 *
 * 证书只存在于组件局部状态；离开页面即丢弃，**不写**浏览器存储。
 */
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import MaterialInput from '@/components/tool/MaterialInput.vue'
import ProductText from '@/components/tool/ProductText.vue'
import ResultPanel, { type ResultView } from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import StatGrid, { type StatGridItem } from '@/components/tool/StatGrid.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useToolRun } from '@/composables/useToolRun'
import {
  TEST_CA_CERT_PEM,
  TEST_LEAF_CERT_PEM,
} from '@/tools/certificates/fixtures-v060'

/** 后端 p7b-pack 工具的响应。 */
interface P7BOutput {
  pem: string
  derBase64: string
  length: number
  certCount: number
}

interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const certs = ref('')
const certsFormat = ref('auto')

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
} = useToolRun<P7BOutput>({ slug: 'p7b-pack' })

const running = computed(() => state.value === 'loading')

const resultText = computed(() => (data.value ? data.value.pem : ''))

/** 结果摘要：证书张数与容器字节数。 */
const stats = computed<StatGridItem[]>(() => {
  const result = data.value
  if (!result) {
    return []
  }
  return [
    { label: t('tools.p7b-pack.results.certCount'), value: String(result.certCount) },
    { label: t('tools.p7b-pack.results.length'), value: `${result.length} B` },
  ]
})

/** 结果页签：摘要 / PEM / DER。 */
const views: ResultView[] = [
  { id: 'summary', label: t('tools.p7b-pack.views.summary') },
  { id: 'pem', label: t('tools.p7b-pack.views.pem') },
  { id: 'der', label: t('tools.p7b-pack.views.der') },
]

/** 当前页签；默认停在「摘要」。 */
const activeView = ref('summary')

const faq = computed<FaqItem[]>(() => tm('tools.p7b-pack.faq') as FaqItem[])

function run(): void {
  void execute({
    certs: [{ data: certs.value, format: certsFormat.value }],
  })
}

function reset(): void {
  certs.value = ''
  certsFormat.value = 'auto'
  resetRun()
}

function fillExample(): void {
  reset()
  certs.value = `${TEST_LEAF_CERT_PEM}${TEST_CA_CERT_PEM}`
}
</script>

<template>
  <ToolPageLayout
    slug="p7b-pack"
    name="PKCS#7 打包"
    name-en="PKCS#7 Packer"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <MaterialInput
            v-model="certs"
            v-model:format="certsFormat"
            :label="t('tools.p7b-pack.fields.certs')"
            :rows="12"
          />
        </template>

        <template #footer>
          <RunBar
            :running="running"
            :disabled="certs === ''"
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
        :title="t('tools.p7b-pack.results.output')"
        :views="views"
        hide-text
        @retry="run"
      >
        <template #view="{ viewId }">
          <StatGrid
            v-if="viewId === 'summary' && data"
            :items="stats"
          />

          <ProductText
            v-else-if="viewId === 'pem' && data"
            :label="t('tools.p7b-pack.views.pem')"
            :content="data.pem"
            filename="bundle.p7b.pem"
            mime="application/x-pem-file"
          />

          <ProductText
            v-else-if="viewId === 'der' && data"
            :label="t('tools.p7b-pack.views.der')"
            :content="data.derBase64"
            filename="bundle.p7b.der.b64"
            mime="text/plain"
          />
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.p7b-pack.description') }}
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
