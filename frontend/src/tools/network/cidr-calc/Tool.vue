<script setup lang="ts">
/**
 * CIDR 计算工具页。
 *
 * 纯本地计算，结果由后端返回后按行渲染，便于复制。
 */
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { ElInputNumber } from 'element-plus/es/components/input-number/index'
import 'element-plus/es/components/input-number/style/css'
import { ElOption, ElSelect } from 'element-plus/es/components/select/index'
import 'element-plus/es/components/select/style/css'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useAutoRun } from '@/composables/useAutoRun'
import { useToolRun } from '@/composables/useToolRun'

import { DEFAULT_EXAMPLE } from './example'

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

const { t, tm } = useI18n()

const mode = ref<'cidr' | 'range'>(DEFAULT_EXAMPLE.mode)
const cidr = ref(DEFAULT_EXAMPLE.cidr)
const start = ref(DEFAULT_EXAMPLE.start)
const end = ref(DEFAULT_EXAMPLE.end)
const split = ref(0)

const { state, data, meta, error, run: execute, reset: resetRun } =
  useToolRun<CIDRCalcResult>({ slug: 'cidr-calc' })

const running = computed(() => state.value === 'loading')

const resultText = computed(() => {
  if (!data.value) {
    return ''
  }
  const value = data.value
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
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="cidr-calc"
    name="CIDR 计算"
    name-en="CIDR Calculator"
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
        :state="state"
        :text="resultText"
        :meta="meta"
        :error="error"
        :title="t('tools.cidr-calc.results.output')"
        @retry="run"
      />
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
