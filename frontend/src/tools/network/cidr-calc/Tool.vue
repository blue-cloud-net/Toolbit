<script setup lang="ts">
/**
 * CIDR 计算工具页。
 *
 * 结果用页签呈现：**概览**（统计卡片 + 地址区间条 + 前缀位图）与
 * **子网**（拆分 / 反推得到的 CIDR 列表）。全文复制仍由 `ResultPanel` 提供。
 */
import { ElButton } from 'element-plus/es/components/button/index'
import 'element-plus/es/components/button/style/css'
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { ElInputNumber } from 'element-plus/es/components/input-number/index'
import 'element-plus/es/components/input-number/style/css'
import { ElOption, ElSelect } from 'element-plus/es/components/select/index'
import 'element-plus/es/components/select/style/css'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import CopyButton from '@/components/tool/CopyButton.vue'
import ResultPanel, { type ResultView } from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import StatGrid, { type StatGridItem } from '@/components/tool/StatGrid.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useAutoRun } from '@/composables/useAutoRun'
import { useToolRun } from '@/composables/useToolRun'

import AddressRangeBar from './AddressRangeBar.vue'
import { DEFAULT_EXAMPLE } from './example'
import PrefixBitmap from './PrefixBitmap.vue'

/** 后端 cidr-calc 工具的响应。 */
interface CIDRCalcResult {
  version: number
  network: string
  prefix: number
  mask: string
  firstAddress: string
  lastAddress: string
  broadcast: string
  firstUsable: string
  lastUsable: string
  hostCount: string
  totalAddresses: string
  cidrs: string[]
  range: string[]
}

interface FaqItem {
  q: string
  a: string
}

/** 子网标签网格一次展示的上限；超出折叠。 */
const SUBNET_PREVIEW = 64

const { t, tm } = useI18n()

const mode = ref<'cidr' | 'range'>(DEFAULT_EXAMPLE.mode)
const cidr = ref(DEFAULT_EXAMPLE.cidr)
const start = ref(DEFAULT_EXAMPLE.start)
const end = ref(DEFAULT_EXAMPLE.end)
const split = ref(0)

/** 子网列表是否展开全部。 */
const showAllSubnets = ref(false)

const { state, data, meta, error, run: execute, reset: resetRun } =
  useToolRun<CIDRCalcResult>({ slug: 'cidr-calc' })

const running = computed(() => state.value === 'loading')

/** 结果页签：概览 + 子网（仅在有子网时出现）。 */
const views = computed<ResultView[]>(() => {
  const list: ResultView[] = [{ id: 'overview', label: t('tools.cidr-calc.views.overview') }]
  const subnets = subnetList.value
  if (subnets.length > 0) {
    list.push({ id: 'subnets', label: t('tools.cidr-calc.views.subnets') })
  }
  return list
})

/** 当前页签。 */
const activeView = ref('overview')

/** 子网列表：拆分结果优先，否则用范围反推结果。 */
const subnetList = computed<string[]>(() => {
  if (!data.value) {
    return []
  }
  if (data.value.cidrs.length > 0) {
    return data.value.cidrs
  }
  return data.value.range
})

/** 是否折叠了子网。 */
const subnetsTruncated = computed(
  () => subnetList.value.length > SUBNET_PREVIEW && !showAllSubnets.value,
)

/** 展示的子网。 */
const visibleSubnets = computed(() =>
  subnetsTruncated.value ? subnetList.value.slice(0, SUBNET_PREVIEW) : subnetList.value,
)

/** 结果文本：结构化摘要 + 子网列表。 */
const resultText = computed(() => {
  const value = data.value
  if (!value) {
    return ''
  }
  const lines = [
    `network: ${value.network}`,
    `mask: ${value.mask}`,
    `firstAddress: ${value.firstAddress}`,
    `lastAddress: ${value.lastAddress}`,
  ]
  if (value.version === 4) {
    lines.push(`broadcast: ${value.broadcast}`)
  }
  lines.push(
    `firstUsable: ${value.firstUsable}`,
    `lastUsable: ${value.lastUsable}`,
    `hostCount: ${value.hostCount}`,
    `totalAddresses: ${value.totalAddresses}`,
  )
  if (value.cidrs.length > 0) {
    lines.push('', ...value.cidrs)
  }
  if (value.range.length > 0) {
    lines.push('', ...value.range)
  }
  return lines.join('\n')
})

/** 概览统计卡片。 */
const stats = computed<StatGridItem[]>(() => {
  const value = data.value
  if (!value) {
    return []
  }
  const items: StatGridItem[] = [
    { label: t('tools.cidr-calc.results.network'), value: value.network },
    { label: t('tools.cidr-calc.results.mask'), value: value.mask },
    { label: t('tools.cidr-calc.results.prefix'), value: `/${value.prefix}` },
    {
      label: t('tools.cidr-calc.results.version'),
      value: `IPv${value.version}`,
    },
    { label: t('tools.cidr-calc.results.totalAddresses'), value: value.totalAddresses },
    { label: t('tools.cidr-calc.results.hostCount'), value: value.hostCount },
  ]
  if (value.version === 4) {
    items.push({ label: t('tools.cidr-calc.results.broadcast'), value: value.broadcast })
  }
  items.push({
    label: t('tools.cidr-calc.results.usableRange'),
    value: `${value.firstUsable} – ${value.lastUsable}`,
  })
  return items
})

const faq = computed<FaqItem[]>(() => tm('tools.cidr-calc.faq') as FaqItem[])

const canRun = computed(() =>
  mode.value === 'cidr'
    ? cidr.value.trim() !== ''
    : start.value.trim() !== '' && end.value.trim() !== '',
)

function run(): void {
  const payload =
    mode.value === 'cidr'
      ? { cidr: cidr.value, ...(split.value > 0 ? { split: split.value } : {}) }
      : { start: start.value, end: end.value }
  void execute(payload)
}

/** 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。 */
useAutoRun({
  slug: 'cidr-calc',
  run,
  watch: [mode, cidr, start, end, split],
})

function reset(): void {
  mode.value = DEFAULT_EXAMPLE.mode
  cidr.value = DEFAULT_EXAMPLE.cidr
  start.value = DEFAULT_EXAMPLE.start
  end.value = DEFAULT_EXAMPLE.end
  split.value = 0
  showAllSubnets.value = false
  activeView.value = 'overview'
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="cidr-calc"
    name="CIDR 计算"
    name-en="CIDR Calculator"
    wide
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.cidr-calc.fields.mode') }}</span>
            <ElSelect
              v-model="mode"
              class="field__select"
              :aria-label="t('tools.cidr-calc.fields.mode')"
            >
              <ElOption
                :label="t('tools.cidr-calc.options.modeCidr')"
                value="cidr"
              />
              <ElOption
                :label="t('tools.cidr-calc.options.modeRange')"
                value="range"
              />
            </ElSelect>
          </label>

          <label
            v-if="mode === 'cidr'"
            class="field"
          >
            <span class="field__label">{{ t('tools.cidr-calc.fields.cidr') }}</span>
            <ElInput
              v-model="cidr"
              class="field__input"
              spellcheck="false"
              autocomplete="off"
            />
          </label>

          <template v-else>
            <label class="field">
              <span class="field__label">{{ t('tools.cidr-calc.fields.start') }}</span>
              <ElInput
                v-model="start"
                class="field__input"
                spellcheck="false"
                autocomplete="off"
              />
            </label>
            <label class="field">
              <span class="field__label">{{ t('tools.cidr-calc.fields.end') }}</span>
              <ElInput
                v-model="end"
                class="field__input"
                spellcheck="false"
                autocomplete="off"
              />
            </label>
          </template>
        </template>

        <template #options>
          <label
            v-if="mode === 'cidr'"
            class="field"
          >
            <span class="field__label">{{ t('tools.cidr-calc.fields.split') }}</span>
            <ElInputNumber
              v-model="split"
              class="field__input"
              :min="0"
              :max="128"
            />
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
        v-model:active-view="activeView"
        :state="state"
        :text="resultText"
        :meta="meta"
        :error="error"
        :title="t('tools.cidr-calc.results.output')"
        :views="views"
        @retry="run"
      >
        <template #view="{ viewId }">
          <template v-if="viewId === 'overview'">
            <StatGrid :items="stats" />

            <section
              v-if="data"
              class="section"
            >
              <h3 class="section__title">
                {{ t('tools.cidr-calc.results.range') }}
              </h3>
              <AddressRangeBar
                :version="data.version"
                :network="data.network"
                :first-address="data.firstAddress"
                :last-address="data.lastAddress"
                :broadcast="data.broadcast"
                :first-usable="data.firstUsable"
                :last-usable="data.lastUsable"
              />
            </section>

            <section
              v-if="data"
              class="section"
            >
              <h3 class="section__title">
                {{ t('tools.cidr-calc.results.prefixBits') }}
              </h3>
              <PrefixBitmap
                :version="data.version"
                :prefix="data.prefix"
              />
              <p class="section__hint">
                {{ t('tools.cidr-calc.hints.bits', { prefix: data.prefix, bits: data.version === 4 ? 32 : 128 }) }}
              </p>
            </section>
          </template>

          <template v-else>
            <div class="subnets">
              <div
                v-for="item in visibleSubnets"
                :key="item"
                class="subnet"
              >
                <span class="subnet__value tb-mono">{{ item }}</span>
                <CopyButton
                  :text="item"
                  size="small"
                />
              </div>
            </div>

            <p
              v-if="subnetsTruncated"
              class="section__hint"
            >
              {{ t('tools.cidr-calc.hints.subnetTruncated', {
                shown: SUBNET_PREVIEW,
                total: subnetList.length,
              }) }}
            </p>

            <ElButton
              v-if="subnetsTruncated || (subnetList.length > SUBNET_PREVIEW && showAllSubnets)"
              size="small"
              @click="showAllSubnets = !showAllSubnets"
            >
              {{ showAllSubnets
                ? t('tools.cidr-calc.actions.collapse')
                : t('tools.cidr-calc.actions.showAll', { count: subnetList.length }) }}
            </ElButton>
          </template>
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.cidr-calc.description') }}
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
.section {
  margin-top: var(--tb-space-4);
}

.section__title {
  margin: 0 0 var(--tb-space-2);
  font-size: var(--tb-font-sm);
  font-weight: 600;
  color: var(--tb-text-secondary);
}

.section__hint {
  margin: var(--tb-space-2) 0 0;
  font-size: var(--tb-font-xs);
  color: var(--tb-text-tertiary);
}

.subnets {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(180px, 1fr));
  gap: var(--tb-space-1) var(--tb-space-2);
}

.subnet {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--tb-space-2);
  padding: var(--tb-space-1) var(--tb-space-2);
  font-size: var(--tb-font-xs);
  background: var(--tb-bg-sunken);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-control);
}

.subnet__value {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.prose {
  margin: 0;
  font-size: var(--tb-font-sm);
  color: var(--tb-text-secondary);
}

.faq {
  margin: 0;
  font-size: var(--tb-font-sm);
}

.faq__q {
  font-weight: 600;
}

.faq__a {
  margin: 0 0 var(--tb-space-2);
  color: var(--tb-text-secondary);
}
</style>
