<script setup lang="ts">
/**
 * 正则表达式生成工具页。
 *
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
 */
import { ElCheckbox } from 'element-plus/es/components/checkbox/index'
import 'element-plus/es/components/checkbox/style/css'
import { ElCheckboxGroup } from 'element-plus/es/components/checkbox/index'
import 'element-plus/es/components/checkbox/style/css'
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { ElInputNumber } from 'element-plus/es/components/input-number/index'
import 'element-plus/es/components/input-number/style/css'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useAutoRun } from '@/composables/useAutoRun'
import { useToolRun } from '@/composables/useToolRun'
import { EXAMPLES } from '@/tools/text/regex-build/example'

/** 一段释义。 */
interface RegexPartExplain {
  /** 部件标识。 */
  part: string
  /** 表达式片段。 */
  fragment: string
  /** 片段含义。 */
  description: string
}

/** 后端 regex-build 工具的响应。 */
interface RegexBuildResult {
  /** 生成的表达式。 */
  pattern: string
  /** 是否合法。 */
  valid: boolean
  /** 逐段释义。 */
  explanation: RegexPartExplain[]
  /** 校验错误。 */
  error: RegexError | null
  /** 捕获组数量。 */
  captureGroups: number
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

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const parts = ref<string[]>(['digits'])
const prefix = ref('')
const suffix = ref('')
const minRepeat = ref(1)
const maxRepeat = ref(1)
const anchored = ref(false)

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<RegexBuildResult>({
  slug: 'regex-build',
})

const running = computed(() => state.value === 'loading')

/** 结果文本：生成的表达式。 */
const resultText = computed(() => data.value?.pattern ?? '')

/** 结论徽章：表达式是否合法。 */
const verdict = computed(() => {
  if (!data.value) {
    return null
  }
  return data.value.valid
    ? { level: 'success' as const, label: t('tools.regex-build.results.valid') }
    : { level: 'danger' as const, label: t('tools.regex-build.results.error') }
})

const faq = computed<FaqItem[]>(() => tm('tools.regex-build.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  void execute({
    parts: parts.value,
    prefix: prefix.value,
    suffix: suffix.value,
    quantifier: { min: minRepeat.value, max: maxRepeat.value },
    anchored: anchored.value,
  })
}

// 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。
useAutoRun({
  slug: 'regex-build',
  run,
  watch: [parts, prefix, suffix, minRepeat, maxRepeat, anchored],
  enabled: () => true,
})

/** 重置表单与结果。 */
function reset(): void {
  parts.value = ['digits']
  prefix.value = ''
  suffix.value = ''
  minRepeat.value = 1
  maxRepeat.value = 1
  anchored.value = false
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  prefix.value = example.prefix
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="regex-build"
    :name="t('tools.regex-build.name')"
    name-en="Regex Builder"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field field--grow">
            <span class="field__label">{{ t('tools.regex-build.fields.prefix') }}</span>
            <ElInput
              v-model="prefix"
              class="field__input tb-mono"
              autocomplete="off"
              :aria-label="t('tools.regex-build.fields.prefix')"
              @keydown.enter.prevent="run"
            />
          </label>

          <label class="field field--grow">
            <span class="field__label">{{ t('tools.regex-build.fields.suffix') }}</span>
            <ElInput
              v-model="suffix"
              class="field__input tb-mono"
              autocomplete="off"
              :aria-label="t('tools.regex-build.fields.suffix')"
              @keydown.enter.prevent="run"
            />
          </label>
        </template>

        <template #options>
          <div class="field field--grow">
            <span class="field__label">{{ t('tools.regex-build.fields.parts') }}</span>
            <ElCheckboxGroup
              v-model="parts"
              class="field__checkgroup"
              :aria-label="t('tools.regex-build.fields.parts')"
            >
              <ElCheckbox value="anyChar">
                {{ t('tools.regex-build.options.parts.anyChar') }}
              </ElCheckbox>
              <ElCheckbox value="digits">
                {{ t('tools.regex-build.options.parts.digits') }}
              </ElCheckbox>
              <ElCheckbox value="letters">
                {{ t('tools.regex-build.options.parts.letters') }}
              </ElCheckbox>
              <ElCheckbox value="word">
                {{ t('tools.regex-build.options.parts.word') }}
              </ElCheckbox>
              <ElCheckbox value="space">
                {{ t('tools.regex-build.options.parts.space') }}
              </ElCheckbox>
              <ElCheckbox value="upper">
                {{ t('tools.regex-build.options.parts.upper') }}
              </ElCheckbox>
              <ElCheckbox value="lower">
                {{ t('tools.regex-build.options.parts.lower') }}
              </ElCheckbox>
              <ElCheckbox value="chinese">
                {{ t('tools.regex-build.options.parts.chinese') }}
              </ElCheckbox>
              <ElCheckbox value="hyphen">
                {{ t('tools.regex-build.options.parts.hyphen') }}
              </ElCheckbox>
              <ElCheckbox value="dot">
                {{ t('tools.regex-build.options.parts.dot') }}
              </ElCheckbox>
              <ElCheckbox value="at">
                {{ t('tools.regex-build.options.parts.at') }}
              </ElCheckbox>
              <ElCheckbox value="slash">
                {{ t('tools.regex-build.options.parts.slash') }}
              </ElCheckbox>
            </ElCheckboxGroup>
          </div>

          <label class="field">
            <span class="field__label">{{ t('tools.regex-build.fields.minRepeat') }}</span>
            <ElInputNumber
              v-model="minRepeat"
              class="field__select"
              :min="0"
              :max="100"
              :aria-label="t('tools.regex-build.fields.minRepeat')"
            />
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.regex-build.fields.maxRepeat') }}</span>
            <ElInputNumber
              v-model="maxRepeat"
              class="field__select"
              :min="-1"
              :max="100"
              :aria-label="t('tools.regex-build.fields.maxRepeat')"
            />
          </label>

          <label class="field field--check">
            <ElCheckbox
              v-model="anchored"
              :aria-label="t('tools.regex-build.fields.anchored')"
            >
              {{ t('tools.regex-build.fields.anchored') }}
            </ElCheckbox>
          </label>

          <p class="field__hint">
            {{ t('tools.regex-build.hints.parts') }}
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
        :title="t('tools.regex-build.results.pattern')"
        @retry="run"
      >
        <template #extra>
          <dl class="detail">
            <div
              v-for="item in data?.explanation ?? []"
              :key="item.part"
              class="detail__row"
            >
              <dt>{{ item.description }}</dt>
              <dd class="tb-mono">
                {{ item.fragment }}
              </dd>
            </div>
            <div
              v-if="data?.error"
              class="detail__row"
            >
              <dt>{{ t('tools.regex-build.results.error') }}</dt>
              <dd class="tb-mono">
                {{ data.error.message }}
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
