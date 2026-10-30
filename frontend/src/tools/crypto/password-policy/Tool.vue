<script setup lang="ts">
/**
 * 口令强度检查工具页。
 *
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
 */
import { ElCheckbox } from 'element-plus/es/components/checkbox/index'
import 'element-plus/es/components/checkbox/style/css'
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { ElInputNumber } from 'element-plus/es/components/input-number/index'
import 'element-plus/es/components/input-number/style/css'
import { ElOption, ElSelect } from 'element-plus/es/components/select/index'
import 'element-plus/es/components/select/style/css'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import DataTable, { type DataTableColumn, type DataTableRow } from '@/components/tool/DataTable.vue'
import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import StatGrid, { type StatGridItem } from '@/components/tool/StatGrid.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useAutoRun } from '@/composables/useAutoRun'
import { useToolRun } from '@/composables/useToolRun'
import { EXAMPLES } from '@/tools/crypto/password-policy/example'

/** 一条检查项。 */
interface PolicyCheck {
  /** 检查项代码。 */
  code: string
  /** 是否通过。 */
  passed: boolean
  /** 说明。 */
  message: string
}

/** 后端 password-policy 工具的响应。 */
interface PasswordPolicyResult {
  /** 是否通过整体策略。 */
  passed: boolean
  /** 强度评分（0-100）。 */
  score: number
  /** 强度等级。 */
  strength: string
  /** 明细检查项。 */
  checks: PolicyCheck[]
  /** 改进建议。 */
  suggestions: string[]
  /** 生效的策略。 */
  policy: {
    minLength: number
    maxLength: number
    requireUpper: boolean
    requireLower: boolean
    requireDigit: boolean
    requireSymbol: boolean
    maxRepeatRun: number
  }
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const password = ref('Example123')
const preset = ref('medium')
const minLength = ref(12)
const maxLength = ref(128)
const requireUpper = ref(true)
const requireLower = ref(true)
const requireDigit = ref(true)
const requireSymbol = ref(true)
const maxRepeatRun = ref(3)

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<PasswordPolicyResult>({
  slug: 'password-policy',
})

const running = computed(() => state.value === 'loading')

/** 结果文本：逐项检查。 */
const resultText = computed(() => {
  if (!data.value) {
    return ''
  }
  return data.value.checks
    .map((check) => `${check.passed ? 'OK' : 'NG'}\t${check.message}`)
    .join('\n')
})

/** 结论徽章。 */
const verdict = computed(() => {
  if (!data.value) {
    return null
  }
  if (data.value.passed) {
    return { level: 'success' as const, label: t('tools.password-policy.results.passed') }
  }
  return { level: 'danger' as const, label: t('tools.password-policy.results.failed') }
})

const faq = computed<FaqItem[]>(() => tm('tools.password-policy.faq') as FaqItem[])

/** 统计卡片。 */
const stats = computed<StatGridItem[]>(() => [
  { label: t('tools.password-policy.results.score'), value: String(data.value?.score ?? 0) },
  { label: t('tools.password-policy.results.strength'), value: data.value?.strength ?? '-' },
])

/** 表格列定义。 */
const columns = computed<DataTableColumn[]>(() => [
  { key: 'result', label: t('tools.password-policy.results.checkResult') },
  { key: 'item', label: t('tools.password-policy.results.checks') },
])

/** 表格行；通过与否用彩色标签表达。 */
const rows = computed<DataTableRow[]>(() =>
  (data.value?.checks ?? []).map((check) => ({
    key: check.code,
    cells: {
      result: check.passed
        ? t('tools.password-policy.results.passed')
        : t('tools.password-policy.results.failed'),
      item: check.message,
    },
    badges: { result: check.passed ? 'success' : 'danger' },
  })),
)

/** 运行当前输入。 */
function run(): void {
  void execute({
    password: password.value,
    preset: preset.value,
    minLength: minLength.value,
    maxLength: maxLength.value,
    requireUpper: requireUpper.value,
    requireLower: requireLower.value,
    requireDigit: requireDigit.value,
    requireSymbol: requireSymbol.value,
    maxRepeatRun: maxRepeatRun.value,
  })
}

// 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。
useAutoRun({
  slug: 'password-policy',
  run,
  watch: [
    password, preset, minLength, maxLength,
    requireUpper, requireLower, requireDigit, requireSymbol, maxRepeatRun,
  ],
  enabled: () => password.value !== '',
})

/** 重置表单与结果。 */
function reset(): void {
  password.value = 'Example123'
  preset.value = 'medium'
  minLength.value = 12
  maxLength.value = 128
  requireUpper.value = true
  requireLower.value = true
  requireDigit.value = true
  requireSymbol.value = true
  maxRepeatRun.value = 3
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  password.value = example.password
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="password-policy"
    :name="t('tools.password-policy.name')"
    name-en="Password Policy Checker"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field field--grow">
            <span class="field__label">{{ t('tools.password-policy.fields.password') }}</span>
            <ElInput
              v-model="password"
              class="field__input tb-mono"
              type="password"
              show-password
              autocomplete="new-password"
              :aria-label="t('tools.password-policy.fields.password')"
              @keydown.enter.prevent="run"
            />
          </label>
        </template>

        <template #options>
          <label class="field">
            <span class="field__label">{{ t('tools.password-policy.fields.preset') }}</span>
            <ElSelect
              v-model="preset"
              class="field__select"
              :aria-label="t('tools.password-policy.fields.preset')"
            >
              <ElOption
                :label="t('tools.password-policy.options.preset.basic')"
                value="basic"
              />
              <ElOption
                :label="t('tools.password-policy.options.preset.medium')"
                value="medium"
              />
              <ElOption
                :label="t('tools.password-policy.options.preset.strict')"
                value="strict"
              />
            </ElSelect>
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.password-policy.fields.minLength') }}</span>
            <ElInputNumber
              v-model="minLength"
              class="field__select"
              :min="1"
              :max="128"
              :aria-label="t('tools.password-policy.fields.minLength')"
            />
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.password-policy.fields.maxLength') }}</span>
            <ElInputNumber
              v-model="maxLength"
              class="field__select"
              :min="1"
              :max="1024"
              :aria-label="t('tools.password-policy.fields.maxLength')"
            />
          </label>

          <label class="field field--check">
            <ElCheckbox
              v-model="requireUpper"
              :aria-label="t('tools.password-policy.fields.requireUpper')"
            >
              {{ t('tools.password-policy.fields.requireUpper') }}
            </ElCheckbox>
          </label>

          <label class="field field--check">
            <ElCheckbox
              v-model="requireLower"
              :aria-label="t('tools.password-policy.fields.requireLower')"
            >
              {{ t('tools.password-policy.fields.requireLower') }}
            </ElCheckbox>
          </label>

          <label class="field field--check">
            <ElCheckbox
              v-model="requireDigit"
              :aria-label="t('tools.password-policy.fields.requireDigit')"
            >
              {{ t('tools.password-policy.fields.requireDigit') }}
            </ElCheckbox>
          </label>

          <label class="field field--check">
            <ElCheckbox
              v-model="requireSymbol"
              :aria-label="t('tools.password-policy.fields.requireSymbol')"
            >
              {{ t('tools.password-policy.fields.requireSymbol') }}
            </ElCheckbox>
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.password-policy.fields.maxRepeatRun') }}</span>
            <ElInputNumber
              v-model="maxRepeatRun"
              class="field__select"
              :min="1"
              :max="16"
              :aria-label="t('tools.password-policy.fields.maxRepeatRun')"
            />
          </label>

          <p class="field__hint">
            {{ t('tools.password-policy.hints.preset') }}
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
        :title="t('tools.password-policy.results.checks')"
        hide-text
        @retry="run"
      >
        <template #extra>
          <StatGrid :items="stats" />

          <DataTable
            class="table"
            :columns="columns"
            :rows="rows"
            :copy-column="'item'"
            :empty="t('tools.password-policy.results.empty')"
          />

          <ul
            v-if="(data?.suggestions ?? []).length > 0"
            class="suggestions"
          >
            <li
              v-for="suggestion in data?.suggestions ?? []"
              :key="suggestion"
            >
              {{ suggestion }}
            </li>
          </ul>
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

<style scoped>
.table {
  margin-top: var(--tb-space-3);
}

.suggestions {
  margin: var(--tb-space-3) 0 0;
  padding-left: var(--tb-space-4);
  font-size: var(--tb-font-sm);
  color: var(--tb-text-secondary);
}
</style>
