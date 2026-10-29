<script setup lang="ts">
/**
 * 文本差异对比工具页。
 *
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
 */
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { ElOption, ElSelect } from 'element-plus/es/components/select/index'
import 'element-plus/es/components/select/style/css'
import { ElCheckbox } from 'element-plus/es/components/checkbox/index'
import 'element-plus/es/components/checkbox/style/css'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useAutoRun } from '@/composables/useAutoRun'
import { useToolRun } from '@/composables/useToolRun'
import { EXAMPLES } from '@/tools/text/text-diff/example'

/** 一行差异。 */
interface DiffLine {
  /** 操作类型：equal / insert / delete。 */
  op: string
  /** 基准侧行号。 */
  leftLineNum: number
  /** 对比侧行号。 */
  rightLineNum: number
  /** 行内容。 */
  text: string
}

/** 一处结构化差异。 */
interface DiffChange {
  /** 结构化路径。 */
  path: string
  /** 操作类型。 */
  op: string
  /** 基准侧的值。 */
  left: string
  /** 对比侧的值。 */
  right: string
}

/** 差异统计。 */
interface DiffStat {
  /** 新增行数。 */
  added: number
  /** 删除行数。 */
  removed: number
  /** 未变行数。 */
  unchanged: number
  /** 结构化差异条数。 */
  changes: number
}

/** 后端 text-diff 工具的响应。 */
interface TextDiffResult {
  /** 实际使用的对比模式。 */
  mode: string
  /** 两侧是否完全一致。 */
  identical: boolean
  /** 行级差异。 */
  lines: DiffLine[]
  /** 结构化差异。 */
  changes: DiffChange[]
  /** 差异统计。 */
  stat: DiffStat
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const mode = ref('text')
const left = ref('line A\nline B\nline C')
const right = ref('line A\nline B\nline D')
const ignoreWhitespace = ref(false)

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<TextDiffResult>({
  slug: 'text-diff',
})

const running = computed(() => state.value === 'loading')

/** 结果文本：行级差异或结构化差异。 */
const resultText = computed(() => {
  if (!data.value) {
    return ''
  }
  if (data.value.mode === 'json') {
    return data.value.changes
      .map((change) => `${change.op}\t${change.path}\t${change.left} → ${change.right}`)
      .join('\n')
  }
  return data.value.lines
    .map((line) => {
      const marker = line.op === 'insert' ? '+' : line.op === 'delete' ? '-' : ' '
      return `${marker} ${line.text}`
    })
    .join('\n')
})

/** 结论徽章：两侧是否一致。 */
const verdict = computed(() => {
  if (!data.value) {
    return null
  }
  return data.value.identical
    ? { level: 'success' as const, label: t('tools.text-diff.results.identical') }
    : { level: 'warning' as const, label: t('tools.text-diff.results.added') }
})

const faq = computed<FaqItem[]>(() => tm('tools.text-diff.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  if (left.value === '' || right.value === '') {
    return
  }
  void execute({
    mode: mode.value,
    left: { data: left.value },
    right: { data: right.value },
    ignoreWhitespace: ignoreWhitespace.value,
  })
}

// 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。
useAutoRun({
  slug: 'text-diff',
  run,
  watch: [mode, left, right, ignoreWhitespace],
  enabled: () => left.value !== '' && right.value !== '',
})

/** 重置表单与结果。 */
function reset(): void {
  mode.value = 'text'
  left.value = 'line A\nline B\nline C'
  right.value = 'line A\nline B\nline D'
  ignoreWhitespace.value = false
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  mode.value = example.mode
  left.value = example.left
  right.value = example.right
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="text-diff"
    :name="t('tools.text-diff.name')"
    name-en="Text Diff"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.text-diff.fields.mode') }}</span>
            <ElSelect
              v-model="mode"
              class="field__select"
              :aria-label="t('tools.text-diff.fields.mode')"
            >
              <ElOption
                :label="t('tools.text-diff.options.mode.text')"
                value="text"
              />
              <ElOption
                :label="t('tools.text-diff.options.mode.json')"
                value="json"
              />
            </ElSelect>
          </label>

          <label class="field field--grow">
            <span class="field__label">{{ t('tools.text-diff.fields.left') }}</span>
            <ElInput
              v-model="left"
              class="field__input tb-mono"
              type="textarea"
              :rows="6"
              autocomplete="off"
              :aria-label="t('tools.text-diff.fields.left')"
            />
          </label>

          <label class="field field--grow">
            <span class="field__label">{{ t('tools.text-diff.fields.right') }}</span>
            <ElInput
              v-model="right"
              class="field__input tb-mono"
              type="textarea"
              :rows="6"
              autocomplete="off"
              :aria-label="t('tools.text-diff.fields.right')"
            />
          </label>

          <label class="field field--check">
            <ElCheckbox
              v-model="ignoreWhitespace"
              :aria-label="t('tools.text-diff.fields.ignoreWhitespace')"
            >
              {{ t('tools.text-diff.fields.ignoreWhitespace') }}
            </ElCheckbox>
          </label>

          <p class="field__hint">
            {{ t('tools.text-diff.hints.mode') }}
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
        :title="t('tools.text-diff.results.summary')"
        @retry="run"
      >
        <template #extra>
          <dl class="detail">
            <div class="detail__row">
              <dt>{{ t('tools.text-diff.results.added') }}</dt>
              <dd class="tb-mono">
                {{ data?.stat.added ?? 0 }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.text-diff.results.removed') }}</dt>
              <dd class="tb-mono">
                {{ data?.stat.removed ?? 0 }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.text-diff.results.unchanged') }}</dt>
              <dd class="tb-mono">
                {{ data?.stat.unchanged ?? 0 }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.text-diff.results.changes') }}</dt>
              <dd class="tb-mono">
                {{ data?.stat.changes ?? 0 }}
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
