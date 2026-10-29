<script setup lang="ts">
/**
 * 速查表查询工具页。
 *
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
 */
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
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
import { EXAMPLES } from '@/tools/system/lookup-table/example'

/** 一个速查条目。 */
interface LookupEntry {
  /** 类别。 */
  kind: string
  /** 查询键。 */
  key: string
  /** 简短名称。 */
  label: string
  /** 中文说明。 */
  description: string
  /** 附加信息。 */
  extra: string[]
}

/** 后端 lookup-table 工具的响应。 */
interface LookupTableResult {
  /** 类别。 */
  kind: string
  /** 查询词。 */
  query: string
  /** 是否精确命中。 */
  matched: boolean
  /** 命中的条目。 */
  entry: LookupEntry | null
  /** 模糊检索的候选。 */
  candidates: LookupEntry[]
  /** 数据集版本。 */
  dataVersion: string
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const kind = ref('port')
const key = ref('443')

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<LookupTableResult>({
  slug: 'lookup-table',
})

const running = computed(() => state.value === 'loading')

/** 结果文本：命中条目或候选列表。 */
const resultText = computed(() => {
  if (!data.value) {
    return ''
  }
  if (data.value.entry) {
    return `${data.value.entry.label}\t${data.value.entry.description}`
  }
  return data.value.candidates
    .map((candidate) => `${candidate.label}\t${candidate.description}`)
    .join('\n')
})

/** 结论徽章：是否精确命中。 */
const verdict = computed(() => {
  if (!data.value) {
    return null
  }
  return data.value.matched
    ? { level: 'success' as const, label: t('tools.lookup-table.results.matched') }
    : { level: 'info' as const, label: t('tools.lookup-table.results.notMatched') }
})

const faq = computed<FaqItem[]>(() => tm('tools.lookup-table.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  if (key.value.trim() === '') {
    return
  }
  void execute({ kind: kind.value, key: key.value })
}

// 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。
useAutoRun({
  slug: 'lookup-table',
  run,
  watch: [kind, key],
  enabled: () => key.value.trim() !== '',
})

/** 重置表单与结果。 */
function reset(): void {
  kind.value = 'port'
  key.value = '443'
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  kind.value = example.kind
  key.value = example.key
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="lookup-table"
    :name="t('tools.lookup-table.name')"
    name-en="Lookup Table"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.lookup-table.fields.kind') }}</span>
            <ElSelect
              v-model="kind"
              class="field__select"
              :aria-label="t('tools.lookup-table.fields.kind')"
            >
              <ElOption
                :label="t('tools.lookup-table.options.kind.port')"
                value="port"
              />
              <ElOption
                :label="t('tools.lookup-table.options.kind.exitcode')"
                value="exitcode"
              />
              <ElOption
                :label="t('tools.lookup-table.options.kind.timestamp')"
                value="timestamp"
              />
            </ElSelect>
          </label>

          <label class="field field--grow">
            <span class="field__label">{{ t('tools.lookup-table.fields.key') }}</span>
            <ElInput
              v-model="key"
              class="field__input tb-mono"
              autocomplete="off"
              :aria-label="t('tools.lookup-table.fields.key')"
              @keydown.enter.prevent="run"
            />
          </label>

          <p class="field__hint">
            {{ t('tools.lookup-table.hints.key') }}
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
        :verdict="verdict"
        :title="t('tools.lookup-table.results.entry')"
        @retry="run"
      >
        <template #extra>
          <dl class="detail">
            <div
              v-if="data?.entry"
              class="detail__row"
            >
              <dt>{{ data.entry.label }}</dt>
              <dd class="tb-mono">
                {{ data.entry.extra.join(' · ') }}
              </dd>
            </div>
            <div
              v-for="candidate in data?.candidates ?? []"
              :key="candidate.key"
              class="detail__row"
            >
              <dt class="tb-mono">
                {{ candidate.key }}
              </dt>
              <dd>{{ candidate.label }}</dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.lookup-table.results.dataVersion') }}</dt>
              <dd class="tb-mono">
                {{ data?.dataVersion ?? '' }}
              </dd>
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
