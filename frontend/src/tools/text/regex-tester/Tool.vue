<script setup lang="ts">
/**
 * 正则测试工具页。
 *
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
 */
import { ElCheckbox } from 'element-plus/es/components/checkbox/index'
import 'element-plus/es/components/checkbox/style/css'
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { ElInputNumber } from 'element-plus/es/components/input-number/index'
import 'element-plus/es/components/input-number/style/css'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import CodeEditor from '@/components/tool/CodeEditor.vue'
import DataTable, { type DataTableColumn, type DataTableRow } from '@/components/tool/DataTable.vue'
import QuickReference from '@/components/tool/QuickReference.vue'
import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useToolRun } from '@/composables/useToolRun'
import { EXAMPLES } from '@/tools/text/regex-tester/example'

/** 一条匹配。 */
interface RegexMatch {
  /** 匹配在原文中的起始字节偏移。 */
  index: number
  /** 匹配的文本。 */
  text: string
  /** 匹配所在行号，从 1 开始。 */
  line: number
  /** 匹配所在列号，从 1 开始。 */
  column: number
  /** 捕获组内容；未参与匹配的组为空串。 */
  groups: string[]
  /** 捕获组的命名。 */
  names?: Record<string, string>
}

/** 后端 regex-tester 工具的响应。 */
interface RegexTesterResult {
  /** 匹配列表。 */
  matches: RegexMatch[]
  /** 匹配总数。 */
  count: number
  /** 是否因上限截断。 */
  truncated: boolean
  /** 实际使用的正则。 */
  pattern: string
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

/** 受支持的标志。 */
const FLAGS = ['i', 'm', 's', 'x'] as const

/** 匹配条数上限，与服务端保持一致。 */
const MAX_MATCH_LIMIT = 1000

const pattern = ref('')
const input = ref('')
const flags = ref<string[]>([])
const maxMatches = ref(0)
const groupNames = ref('')

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
  bindInput,
  useAsInput,
} = useToolRun<RegexTesterResult>({
  slug: 'regex-tester',
})

/** 结果区的「作为新输入」：把结果回填到主输入框（仅内存，不写存储）。 */
bindInput((text) => {
  input.value = text
})

const running = computed(() => state.value === 'loading')

/** 结果文本：每行一条匹配文本。 */
const resultText = computed(() => (data.value?.matches ?? []).map((m) => m.text).join('\n'))

/** 匹配列表。 */
const matches = computed<RegexMatch[]>(() => data.value?.matches ?? [])

/** 捕获组名称，按顺序对应第 1、2 … 组。 */
const groupNameList = computed(() =>
  groupNames.value
    .split(',')
    .map((name) => name.trim())
    .filter((name) => name !== ''),
)

/** 匹配条数上限；空值或越界都归一到服务端允许的范围。 */
const maxMatchesValue = computed(() => {
  const value = Number(maxMatches.value)
  if (!Number.isFinite(value) || value <= 0) {
    return 0
  }
  return Math.min(Math.trunc(value), MAX_MATCH_LIMIT)
})

const faq = computed<FaqItem[]>(() => tm('tools.regex-tester.faq') as FaqItem[])

/** 表格列定义。 */
const matchColumns = computed<DataTableColumn[]>(() => [
  { key: 'no', label: t('tools.regex-tester.results.no'), mono: true, align: 'right' },
  { key: 'position', label: t('tools.regex-tester.results.position'), mono: true },
  { key: 'offset', label: t('tools.regex-tester.results.offset'), mono: true, align: 'right' },
  { key: 'text', label: t('tools.regex-tester.results.matchText'), mono: true },
  { key: 'groups', label: t('tools.regex-tester.results.groups'), mono: true },
])

/** 表格行；捕获组多组时按「组名=值」拼在一列。 */
const matchRows = computed<DataTableRow[]>(() =>
  matches.value.map((match, index) => ({
    key: String(index),
    cells: {
      no: String(index + 1),
      position: `${match.line}:${match.column}`,
      offset: String(match.index),
      text: match.text,
      groups: match.groups
        .map((group, groupIndex) => `${groupLabel(groupIndex)}=${group || '—'}`)
        .join('; '),
    },
  })),
)

/** 切换某个标志。 */
function toggleFlag(name: string): void {
  const index = flags.value.indexOf(name)
  if (index >= 0) {
    flags.value.splice(index, 1)
  } else {
    flags.value.push(name)
  }
}

/** 捕获组标签：优先用组名，否则用序号。 */
function groupLabel(index: number): string {
  const name = groupNameList.value[index]
  if (name !== undefined) {
    return name
  }
  return t('tools.regex-tester.results.groupIndex', { index: index + 1 })
}

/** 运行当前输入。 */
/** 快速参考条目：点击后填入输入框。 */
const quickRefItems = [
    { label: '任意字符', value: '.' },
    { label: '数字', value: '\\d' },
    { label: '非数字', value: '\\D' },
    { label: '单词字符', value: '\\w' },
    { label: '空白字符', value: '\\s' },
    { label: '行首', value: '^' },
    { label: '行尾', value: '$' },
    { label: '或', value: '|' },
    { label: '分组', value: '()' },
    { label: '量词 0+ 次', value: '+' },
    { label: '量词 0-1 次', value: '?' },
    { label: '量词 0-∞ 次', value: '*' },
  ]

/** 把速查条目填入输入框。 */
function applyQuickRef(value: string): void {
  pattern.value = value
}

function run(): void {
  void execute({
    pattern: pattern.value,
    flags: flags.value,
    input: { data: input.value },
    maxMatches: maxMatchesValue.value,
    groupNames: groupNameList.value,
  })
}

/** 重置表单与结果。 */
function reset(): void {
  pattern.value = ''
  input.value = ''
  flags.value = []
  maxMatches.value = 0
  groupNames.value = ''
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  pattern.value = example.pattern
  input.value = example.input
  flags.value = [...example.flags]
  maxMatches.value = example.maxMatches
  groupNames.value = example.groupNames.join(', ')
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="regex-tester"
    :name="t('tools.regex-tester.name')"
    name-en="Regex Tester"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.regex-tester.fields.pattern') }}</span>
            <CodeEditor
              v-model="pattern"
              :label="t('tools.regex-tester.fields.pattern')"
              :max-rows="4"
              @run="run"
            />
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.regex-tester.fields.input') }}</span>
            <CodeEditor
              v-model="input"
              :label="t('tools.regex-tester.fields.input')"
              :max-rows="12"
              @run="run"
            />
          </label>

          <p class="field__hint">
            {{ t('tools.regex-tester.hints.input') }}
          </p>
        </template>

        <template #options>
          <fieldset class="field">
            <legend class="field__label">
              {{ t('tools.regex-tester.fields.flags') }}
            </legend>
            <div class="flag-list">
              <ElCheckbox
                v-for="flag in FLAGS"
                :key="flag"
                class="field__checkbox"
                :model-value="flags.includes(flag)"
                @change="toggleFlag(flag)"
              >
                {{ t(`tools.regex-tester.options.flags.${flag}`) }}
              </ElCheckbox>
            </div>
          </fieldset>
          <p class="field__hint">
            {{ t('tools.regex-tester.hints.flags') }}
          </p>

          <div class="field-row">
            <label class="field">
              <span class="field__label">{{ t('tools.regex-tester.fields.maxMatches') }}</span>
              <ElInputNumber
                v-model="maxMatches"
                class="field__input"
                :min="0"
                :max="MAX_MATCH_LIMIT"
                :aria-label="t('tools.regex-tester.fields.maxMatches')"
              />
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.regex-tester.fields.groupNames') }}</span>
              <ElInput
                v-model="groupNames"
                class="field__input tb-mono"
                autocomplete="off"
                :aria-label="t('tools.regex-tester.fields.groupNames')"
              />
            </label>
          </div>

          <p class="field__hint">
            {{ t('tools.regex-tester.hints.maxMatches') }}
          </p>
          <p class="field__hint">
            {{ t('tools.regex-tester.hints.groupNames') }}
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
        :title="t('tools.regex-tester.results.matches')"
        :use-as-input="useAsInput"
        hide-text
        @retry="run"
      >
        <template
          v-if="data"
          #extra
        >
          <dl class="detail">
            <div class="detail__row">
              <dt>{{ t('tools.regex-tester.results.count') }}</dt>
              <dd>{{ data?.count }}</dd>
            </div>
          </dl>

          <p
            v-if="data?.truncated"
            class="notice"
          >
            {{ t('tools.regex-tester.results.truncated') }}
          </p>

          <DataTable
            class="table"
            :columns="matchColumns"
            :rows="matchRows"
            :copy-column="'text'"
            :empty="t('tools.regex-tester.results.empty')"
          />
        </template>
      </ResultPanel>
    </template>

    <template #reference>
      <QuickReference
        :title="t('tool.sections.reference')"
        :items="quickRefItems"
        :insert="applyQuickRef"
      />
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.regex-tester.hints.output') }}
      </p>
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

<style scoped>
.flag-list {
  display: flex;
  flex-wrap: wrap;
  gap: var(--tb-space-2) var(--tb-space-4);
}

.detail {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-1);
  margin: var(--tb-space-3) 0 0;
}

.detail__row {
  display: flex;
  gap: var(--tb-space-2);
  font-size: var(--tb-font-sm);
}

.detail__row dt {
  min-width: 96px;
  color: var(--tb-text-tertiary);
}

.detail__row dd {
  margin: 0;
}

.notice {
  margin: var(--tb-space-2) 0 0;
  font-size: var(--tb-font-xs);
  color: var(--tb-warning);
}

.table {
  margin-top: var(--tb-space-3);
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
  margin-top: var(--tb-space-2);
  font-weight: 600;
  color: var(--tb-text);
}

.faq__a {
  margin: var(--tb-space-1) 0 0;
  color: var(--tb-text-secondary);
}
</style>
