<script setup lang="ts">
/**
 * 证书链修复工具页。
 *
 * 证书内容只存在于组件局部状态；离开页面即丢弃，
 * **不写**任何浏览器存储（AGENTS.md §4.11）。
 */
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import MaterialInput from '@/components/tool/MaterialInput.vue'
import ResultPanel, { type ResultView } from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import StatGrid, { type StatGridItem } from '@/components/tool/StatGrid.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useToolRun } from '@/composables/useToolRun'
import { DEFAULT_EXAMPLE } from '@/tools/certificates/cert-chain-repair/example'

/** 后端 cert-chain-repair 工具的响应。 */
interface ChainRepairResult {
  chainPem: string[]
  addedCount: number
  addedFroms: string[]
  reachedRoot: boolean
  note: string
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const cert = ref('')
const certFormat = ref('auto')
const intermediates = ref('')
const intermediatesFormat = ref('auto')
const trustedRoots = ref('')
const trustedRootsFormat = ref('auto')

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
} = useToolRun<ChainRepairResult>({ slug: 'cert-chain-repair' })

const running = computed(() => state.value === 'loading')

/** 结果文本。 */
const resultText = computed(() => {
  const result = data.value
  if (!result) {
    return ''
  }
  return [
    `${t('tools.cert-chain-repair.results.addedCount')}: ${result.addedCount}`,
    `${t('tools.cert-chain-repair.results.addedFroms')}: ${result.addedFroms.join(', ') || '-'}`,
    `${t('tools.cert-chain-repair.results.reachedRoot')}: ${result.reachedRoot}`,
    `${t('tools.cert-chain-repair.results.note')}: ${result.note}`,
  ].join('\n')
})

const faq = computed<FaqItem[]>(() => tm('tools.cert-chain-repair.faq') as FaqItem[])

/** 结果页签：摘要 / 补全后的链。 */
const views = computed<ResultView[]>(() => [
  { id: 'summary', label: t('tools.cert-chain-repair.views.summary') },
  { id: 'chain', label: t('tools.cert-chain-repair.views.chain') },
])

/** 当前页签。 */
const activeView = ref('summary')

/** 摘要统计卡片。 */
const stats = computed<StatGridItem[]>(() => {
  const result = data.value
  if (!result) {
    return []
  }
  return [
    { label: t('tools.cert-chain-repair.results.addedCount'), value: String(result.addedCount) },
    {
      label: t('tools.cert-chain-repair.results.addedFroms'),
      value: result.addedFroms.join(', ') || '-',
    },
    {
      label: t('tools.cert-chain-repair.results.reachedRoot'),
      value: result.reachedRoot ? t('common.states.yes') : t('common.states.no'),
      level: result.reachedRoot ? ('success' as const) : ('warning' as const),
    },
    { label: t('tools.cert-chain-repair.results.note'), value: result.note },
  ]
})

/** 运行当前输入。 */
function run(): void {
  const intermediatesInput: { data: string; format: string }[] = []
  if (intermediates.value.trim() !== '') {
    intermediatesInput.push({ data: intermediates.value, format: intermediatesFormat.value })
  }
  const rootsInput: { data: string; format: string }[] = []
  if (trustedRoots.value.trim() !== '') {
    rootsInput.push({ data: trustedRoots.value, format: trustedRootsFormat.value })
  }
  void execute({
    cert: { data: cert.value, format: certFormat.value },
    intermediates: intermediatesInput,
    trustedRoots: rootsInput,
  })
}

/** 重置表单与结果。 */
function reset(): void {
  cert.value = ''
  certFormat.value = 'auto'
  intermediates.value = ''
  intermediatesFormat.value = 'auto'
  trustedRoots.value = ''
  trustedRootsFormat.value = 'auto'
  resetRun()
}

/** 填入示例。 */
function fillExample(): void {
  cert.value = DEFAULT_EXAMPLE.cert
  certFormat.value = 'auto'
  intermediates.value = DEFAULT_EXAMPLE.intermediates
  intermediatesFormat.value = 'auto'
  trustedRoots.value = DEFAULT_EXAMPLE.roots
  trustedRootsFormat.value = 'auto'
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="cert-chain-repair"
    :name="t('tools.cert-chain-repair.name')"
    name-en="Certificate Chain Repairer"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <MaterialInput
            v-model="cert"
            v-model:format="certFormat"
            :label="t('tools.cert-chain-repair.fields.cert')"
            :rows="8"
          />

          <MaterialInput
            v-model="intermediates"
            v-model:format="intermediatesFormat"
            :label="t('tools.cert-chain-repair.fields.intermediates')"
            :rows="8"
          />

          <MaterialInput
            v-model="trustedRoots"
            v-model:format="trustedRootsFormat"
            :label="t('tools.cert-chain-repair.fields.trustedRoots')"
            :rows="6"
          />
        </template>

        <template #footer>
          <RunBar
            :running="running"
            :disabled="cert === ''"
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
        :title="t('tools.cert-chain-repair.results.chain')"
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
          ><code>{{ data.chainPem.join('\n') }}</code></pre>
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.cert-chain-repair.description') }}
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
