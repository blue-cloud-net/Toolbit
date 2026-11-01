<script setup lang="ts">
/**
 * PKCS#12 解析工具页。
 *
 * 容器与口令只存在于组件局部状态；离开页面即丢弃，**不写**浏览器存储。
 */
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import MaterialInput from '@/components/tool/MaterialInput.vue'
import ResultPanel, { type ResultView } from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import StatGrid, { type StatGridItem } from '@/components/tool/StatGrid.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useToolRun } from '@/composables/useToolRun'
import { TEST_P12_BASE64 } from '@/tools/certificates/fixtures-v060'

/** 后端 pfx-parse 工具的响应。 */
interface PFXParseResult {
  hasPrivateKey: boolean
  certificate: { subjectText: string; issuerText: string; serial: string } | null
  privateKeyPem: string
  caCount: number
}

interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const container = ref('')
const containerFormat = ref('auto')
const passphrase = ref('')

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
} = useToolRun<PFXParseResult>({ slug: 'pfx-parse' })

const running = computed(() => state.value === 'loading')

const resultText = computed(() => {
  const result = data.value
  if (!result) {
    return ''
  }
  const head = [
    `${t('tools.pfx-parse.results.hasPrivateKey')}: ${result.hasPrivateKey}`,
    `${t('tools.pfx-parse.results.caCount')}: ${result.caCount}`,
  ]
  if (result.certificate) {
    head.push(`${t('tools.pfx-parse.results.subject')}: ${result.certificate.subjectText}`)
    head.push(`${t('tools.pfx-parse.results.issuer')}: ${result.certificate.issuerText}`)
    head.push(`${t('tools.pfx-parse.results.serial')}: ${result.certificate.serial}`)
  }
  if (result.privateKeyPem) {
    head.push('', result.privateKeyPem)
  }
  return head.join('\n')
})

const faq = computed<FaqItem[]>(() => tm('tools.pfx-parse.faq') as FaqItem[])

/** 结果页签：摘要 / 私钥。 */
const views = computed<ResultView[]>(() => {
  const list: ResultView[] = [{ id: 'summary', label: t('tools.pfx-parse.views.summary') }]
  if ((data.value?.privateKeyPem ?? '') !== '') {
    list.push({ id: 'key', label: t('tools.pfx-parse.views.key') })
  }
  return list
})

/** 当前页签。 */
const activeView = ref('summary')

/** 摘要统计卡片。 */
const stats = computed<StatGridItem[]>(() => {
  const result = data.value
  if (!result) {
    return []
  }
  const items: StatGridItem[] = [
    {
      label: t('tools.pfx-parse.results.hasPrivateKey'),
      value: result.hasPrivateKey ? t('common.states.yes') : t('common.states.no'),
      level: result.hasPrivateKey ? ('success' as const) : ('warning' as const),
    },
    { label: t('tools.pfx-parse.results.caCount'), value: String(result.caCount) },
  ]
  if (result.certificate) {
    items.push({ label: t('tools.pfx-parse.results.subject'), value: result.certificate.subjectText })
    items.push({ label: t('tools.pfx-parse.results.issuer'), value: result.certificate.issuerText })
    items.push({ label: t('tools.pfx-parse.results.serial'), value: result.certificate.serial })
  }
  return items
})

function run(): void {
  void execute({
    container: { data: container.value, format: containerFormat.value },
    passphrase: passphrase.value,
  })
}

function reset(): void {
  container.value = ''
  containerFormat.value = 'auto'
  passphrase.value = ''
  resetRun()
}

function fillExample(): void {
  reset()
  container.value = TEST_P12_BASE64
  containerFormat.value = 'base64'
  passphrase.value = 'test-pass'
}
</script>

<template>
  <ToolPageLayout
    slug="pfx-parse"
    name="PKCS#12 解析"
    name-en="PKCS#12 Parser"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <MaterialInput
            v-model="container"
            v-model:format="containerFormat"
            :label="t('tools.pfx-parse.fields.container')"
            :rows="8"
          />
        </template>

        <template #options>
          <label class="field">
            <span class="field__label">{{ t('tools.pfx-parse.fields.passphrase') }}</span>
            <ElInput
              v-model="passphrase"
              type="password"
              class="field__input"
              autocomplete="off"
            />
          </label>
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
        v-model:active-view="activeView"
        :state="state"
        :text="resultText"
        :meta="meta"
        :error="error"
        :title="t('tools.pfx-parse.results.output')"
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
          ><code>{{ data.privateKeyPem }}</code></pre>
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.pfx-parse.description') }}
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
