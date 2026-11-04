<script setup lang="ts">
/**
 * 正则表达式校验工具页。
 *
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
 */
import { ElCheckbox } from 'element-plus/es/components/checkbox/index'
import 'element-plus/es/components/checkbox/style/css'
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { ElCheckboxGroup } from 'element-plus/es/components/checkbox/index'
import 'element-plus/es/components/checkbox/style/css'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import DataTable, { type DataTableColumn, type DataTableRow } from '@/components/tool/DataTable.vue'
import StatGrid from '@/components/tool/StatGrid.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { buildSummary } from '@/components/tool/summary'
import { useAutoRun } from '@/composables/useAutoRun'
import { useToolRun } from '@/composables/useToolRun'
import { EXAMPLES } from '@/tools/text/regex-validate/example'

/** 一条建议。 */
interface RegexSuggestion {
  /** 建议标识。 */
  code: string
  /** 建议说明。 */
  message: string
  /** 优先级。 */
  priority: string
}

/** 语法错误。 */
interface RegexError {
  /** 错误码标识。 */
  code: string
  /** 错误说明。 */
  message: string
  /** 错误偏移。 */
  offset: number
  /** 错误位置。 */
  position: string
}

/** 后端 regex-validate 工具的响应。 */
interface RegexValidateResult {
  /** 是否合法。 */
  valid: boolean
  /** 正则方言。 */
  dialect: string
  /** 生效的标志。 */
  flags: string[]
  /** 语法错误。 */
  error: RegexError | null
  /** 改进建议。 */
  suggestions: RegexSuggestion[]
  /** 捕获组数量。 */
  captureGroups: number
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const pattern = ref('^[a-z]+$')
const flags = ref<string[]>([])

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<RegexValidateResult>({
  slug: 'regex-validate',
})

const running = computed(() => state.value === 'loading')

/** 结果文本：错误说明；合法时为空串。 */
const resultText = computed(() => data.value?.error?.message ?? '')

/** 结论徽章：是否合法。 */
const verdict = computed(() => {
  if (!data.value) {
    return null
  }
  return data.value.valid
    ? { level: 'success' as const, label: t('tools.regex-validate.results.valid') }
    : { level: 'danger' as const, label: t('tools.regex-validate.results.invalid') }
})

/** 摘要卡片：3 项关键字段。 */
const stats = computed(() =>
  buildSummary(
    data.value as unknown as Record<string, unknown> | null,
    [
      { key: 'dialect', label: t('tools.regex-validate.results.dialect') },
      { key: 'captureGroups', label: t('tools.regex-validate.results.captureGroups') },
      { key: 'error.position', label: t('tools.regex-validate.results.position') },
      { key: 'error.message', label: t('tools.regex-validate.results.reason'), wide: true },
    ],
    { yes: t('common.states.yes'), no: t('common.states.no') },
  ),
)

const faq = computed<FaqItem[]>(() => tm('tools.regex-validate.faq') as FaqItem[])

/** 建议表格列定义。 */
const suggestionColumns = computed<DataTableColumn[]>(() => [
  { key: 'code', label: t('tools.regex-validate.results.suggestionCode'), mono: true },
  { key: 'priority', label: t('tools.regex-validate.results.suggestionPriority') },
  { key: 'message', label: t('tools.regex-validate.results.suggestionMessage') },
])

/** 建议表格行。 */
const suggestionRows = computed<DataTableRow[]>(() =>
  (data.value?.suggestions ?? []).map((suggestion) => ({
    key: suggestion.code,
    cells: {
      code: suggestion.code,
      priority: suggestion.priority,
      message: suggestion.message,
    },
  })),
)

/** 运行当前输入。 */
function run(): void {
  if (pattern.value.trim() === '') {
    return
  }
  void execute({ pattern: pattern.value, flags: flags.value })
}

// 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。
useAutoRun({
  slug: 'regex-validate',
  run,
  watch: [pattern, flags],
  enabled: () => pattern.value.trim() !== '',
})

/** 重置表单与结果。 */
function reset(): void {
  pattern.value = '^[a-z]+$'
  flags.value = []
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  pattern.value = example.pattern
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="regex-validate"
    :name="t('tools.regex-validate.name')"
    name-en="Regex Validator"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field field--grow">
            <span class="field__label">{{ t('tools.regex-validate.fields.pattern') }}</span>
            <ElInput
              v-model="pattern"
              class="field__input tb-mono"
              autocomplete="off"
              placeholder="^[a-z]+$"
              :aria-label="t('tools.regex-validate.fields.pattern')"
              @keydown.enter.prevent="run"
            />
          </label>

          <div class="field field--grow">
            <span class="field__label">{{ t('tools.regex-validate.fields.flags') }}</span>
            <ElCheckboxGroup
              v-model="flags"
              class="field__checkgroup"
              :aria-label="t('tools.regex-validate.fields.flags')"
            >
              <ElCheckbox value="i">
                {{ t('tools.regex-validate.options.flags.i') }}
              </ElCheckbox>
              <ElCheckbox value="m">
                {{ t('tools.regex-validate.options.flags.m') }}
              </ElCheckbox>
              <ElCheckbox value="s">
                {{ t('tools.regex-validate.options.flags.s') }}
              </ElCheckbox>
              <ElCheckbox value="x">
                {{ t('tools.regex-validate.options.flags.x') }}
              </ElCheckbox>
            </ElCheckboxGroup>
          </div>

          <p class="field__hint">
            {{ t('tools.regex-validate.hints.pattern') }}
          </p>
        </template>

        <template #footer>
          <RunBar
            :running="running"
            :examples="EXAMPLES"
            @run="run"
            @reset="reset"
            @fill="fillExample"
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
        :title="t('tools.regex-validate.results.error')"
        hide-text
        @retry="run"
      >
        <template #extra>
          <StatGrid :items="stats" />

          <DataTable
            v-if="suggestionRows.length > 0"
            class="table"
            :columns="suggestionColumns"
            :rows="suggestionRows"
            :copy-column="'message'"
          />
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
