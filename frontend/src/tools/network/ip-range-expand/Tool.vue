<script setup lang="ts">
/**
 * IP 段展开工具页。
 *
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
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
import { EXAMPLES } from '@/tools/network/ip-range-expand/example'

/** 后端 ip-range-expand 工具的响应。 */
interface IPRangeResult {
  /** 模式。 */
  mode: string
  /** 规范化后的 CIDR。 */
  cidr: string
  /** 起始地址。 */
  start: string
  /** 结束地址。 */
  end: string
  /** 地址总数。 */
  total: number
  /** 展开的地址列表。 */
  addresses: string[]
  /** 是否被上限截断。 */
  truncated: boolean
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const mode = ref('expand')
const cidr = ref('192.168.1.0/24')
const start = ref('192.168.1.10')
const end = ref('192.168.1.20')
const limit = ref(1024)

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<IPRangeResult>({
  slug: 'ip-range-expand',
})

const running = computed(() => state.value === 'loading')

/** 结果文本：地址列表。 */
const resultText = computed(() => data.value?.addresses.join('\n') ?? '')

const faq = computed<FaqItem[]>(() => tm('tools.ip-range-expand.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  void execute({
    mode: mode.value,
    cidr: cidr.value,
    start: start.value,
    end: end.value,
    limit: limit.value,
  })
}

// 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。
useAutoRun({
  slug: 'ip-range-expand',
  run,
  watch: [mode, cidr, start, end, limit],
  enabled: () => true,
})

/** 重置表单与结果。 */
function reset(): void {
  mode.value = 'expand'
  cidr.value = '192.168.1.0/24'
  start.value = '192.168.1.10'
  end.value = '192.168.1.20'
  limit.value = 1024
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  cidr.value = example.cidr
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="ip-range-expand"
    :name="t('tools.ip-range-expand.name')"
    name-en="IP Range Expander"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.ip-range-expand.fields.mode') }}</span>
            <ElSelect
              v-model="mode"
              class="field__select"
              :aria-label="t('tools.ip-range-expand.fields.mode')"
            >
              <ElOption
                :label="t('tools.ip-range-expand.options.mode.expand')"
                value="expand"
              />
              <ElOption
                :label="t('tools.ip-range-expand.options.mode.toCidr')"
                value="toCidr"
              />
              <ElOption
                :label="t('tools.ip-range-expand.options.mode.toRange')"
                value="toRange"
              />
            </ElSelect>
          </label>

          <label
            v-if="mode === 'expand'"
            class="field field--grow"
          >
            <span class="field__label">{{ t('tools.ip-range-expand.fields.cidr') }}</span>
            <ElInput
              v-model="cidr"
              class="field__input tb-mono"
              autocomplete="off"
              :aria-label="t('tools.ip-range-expand.fields.cidr')"
              @keydown.enter.prevent="run"
            />
          </label>

          <label
            v-if="mode !== 'expand'"
            class="field field--grow"
          >
            <span class="field__label">{{ t('tools.ip-range-expand.fields.start') }}</span>
            <ElInput
              v-model="start"
              class="field__input tb-mono"
              autocomplete="off"
              :aria-label="t('tools.ip-range-expand.fields.start')"
              @keydown.enter.prevent="run"
            />
          </label>

          <label
            v-if="mode !== 'expand'"
            class="field field--grow"
          >
            <span class="field__label">{{ t('tools.ip-range-expand.fields.end') }}</span>
            <ElInput
              v-model="end"
              class="field__input tb-mono"
              autocomplete="off"
              :aria-label="t('tools.ip-range-expand.fields.end')"
              @keydown.enter.prevent="run"
            />
          </label>
        </template>

        <template #options>
          <label class="field">
            <span class="field__label">{{ t('tools.ip-range-expand.fields.limit') }}</span>
            <ElInputNumber
              v-model="limit"
              class="field__select"
              :min="1"
              :max="65536"
              :aria-label="t('tools.ip-range-expand.fields.limit')"
            />
          </label>

          <p class="field__hint">
            {{ t('tools.ip-range-expand.hints.limit') }}
          </p>
        </template>

        <template #footer>
          <RunBar
            :running="running"
            @run="run"
            @reset="reset"
            @example="fillExample(0)"
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
        :title="t('tools.ip-range-expand.results.addresses')"
        @retry="run"
      >
        <template #extra>
          <dl class="detail">
            <div class="detail__row">
              <dt>{{ t('tools.ip-range-expand.results.cidr') }}</dt>
              <dd class="tb-mono">
                {{ data?.cidr ?? '' }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.ip-range-expand.results.start') }}</dt>
              <dd class="tb-mono">
                {{ data?.start ?? '' }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.ip-range-expand.results.end') }}</dt>
              <dd class="tb-mono">
                {{ data?.end ?? '' }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.ip-range-expand.results.total') }}</dt>
              <dd class="tb-mono">
                {{ data?.total ?? 0 }}
              </dd>
            </div>
            <div
              v-if="data?.truncated"
              class="detail__row"
            >
              <dt>!</dt>
              <dd>{{ t('tools.ip-range-expand.results.truncated') }}</dd>
            </div>
          </dl>
        </template>
      </ResultPanel>
    </template>

    <template #faq>
      <dl class="faq">
        <template
          v-for="(item, index) in faq"
          :key="index"
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
