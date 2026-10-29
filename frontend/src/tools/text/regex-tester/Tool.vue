<script setup lang="ts">
/**
 * 正则测试工具页。
 *
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
 */
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import CodeEditor from '@/components/tool/CodeEditor.vue'
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
              <label
                v-for="flag in FLAGS"
                :key="flag"
                class="field field--inline"
              >
                <input
                  type="checkbox"
                  class="field__checkbox"
                  :checked="flags.includes(flag)"
                  @change="toggleFlag(flag)"
                >
                <span>{{ t(`tools.regex-tester.options.flags.${flag}`) }}</span>
              </label>
            </div>
          </fieldset>
          <p class="field__hint">
            {{ t('tools.regex-tester.hints.flags') }}
          </p>

          <div class="field-row">
            <label class="field">
              <span class="field__label">{{ t('tools.regex-tester.fields.maxMatches') }}</span>
              <input
                v-model.number="maxMatches"
                type="number"
                class="field__input"
                min="0"
                :max="MAX_MATCH_LIMIT"
                :aria-label="t('tools.regex-tester.fields.maxMatches')"
              >
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.regex-tester.fields.groupNames') }}</span>
              <input
                v-model="groupNames"
                type="text"
                class="field__input tb-mono"
                autocomplete="off"
                :aria-label="t('tools.regex-tester.fields.groupNames')"
              >
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

          <ol
            v-if="matches.length > 0"
            class="matches"
          >
            <li
              v-for="(match, index) in matches"
              :key="index"
              class="match"
            >
              <div class="match__head">
                <span class="match__no">
                  #{{ index + 1 }}
                </span>
                <span class="match__pos">
                  {{ t('tools.regex-tester.results.position') }} {{ match.line }}:{{ match.column }}
                </span>
                <span class="match__offset">
                  {{ t('tools.regex-tester.results.offset') }} {{ match.index }}
                </span>
              </div>

              <pre class="match__text tb-mono">{{ match.text }}</pre>

              <dl
                v-if="match.groups.length > 0"
                class="match__groups"
              >
                <div
                  v-for="(group, groupIndex) in match.groups"
                  :key="groupIndex"
                  class="match__group"
                >
                  <dt>{{ groupLabel(groupIndex) }}</dt>
                  <dd class="tb-mono">
                    {{ group || '—' }}
                  </dd>
                </div>
              </dl>
            </li>
          </ol>
        </template>
      </ResultPanel>
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

.matches {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-2);
  margin: var(--tb-space-3) 0 0;
  padding: 0;
  list-style: none;
}

.match {
  padding: var(--tb-space-2) var(--tb-space-3);
  background: var(--tb-bg-sunken);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-control);
}

.match__head {
  display: flex;
  flex-wrap: wrap;
  gap: var(--tb-space-2);
  font-size: var(--tb-font-xs);
  color: var(--tb-text-tertiary);
}

.match__no {
  color: var(--tb-text-secondary);
}

.match__text {
  margin: var(--tb-space-2) 0 0;
  padding: var(--tb-space-2);
  overflow-x: auto;
  font-family: var(--tb-font-mono);
  font-size: var(--tb-font-sm);
  color: var(--tb-text);
  white-space: pre-wrap;
  word-break: break-all;
  background: var(--tb-bg-elevated);
  border-radius: var(--tb-radius-control);
}

.match__groups {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-1);
  margin: var(--tb-space-2) 0 0;
}

.match__group {
  display: flex;
  gap: var(--tb-space-2);
  font-size: var(--tb-font-xs);
}

.match__group dt {
  min-width: 88px;
  color: var(--tb-text-tertiary);
}

.match__group dd {
  margin: 0;
  word-break: break-all;
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
