<script setup lang="ts">
/**
 * 日期计算工具页。
 *
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
 */
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useToolRun } from '@/composables/useToolRun'
import { EXAMPLES } from '@/tools/data/date-calc/example'
import type { DateOperation, DateUnit } from '@/tools/data/date-calc/example'

/** 后端 date-calc 工具的响应。 */
interface DateCalcResult {
  /** 结果日期（`add` / `sub`）。 */
  resultDate?: string
  /** 相差天数（`diff`）。 */
  diffDays: number
  /** 工作日数（`workdays`）。 */
  workdays: number
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const operation = ref<DateOperation>('diff')
const dateA = ref('2024-01-01')
const dateB = ref('2024-01-31')
const days = ref(10)
const unit = ref<DateUnit>('day')
const holidays = ref('')

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<DateCalcResult>({
  slug: 'date-calc',
})

const running = computed(() => state.value === 'loading')

/** 结果文本：按操作取对应字段。 */
const resultText = computed(() => {
  const result = data.value
  if (!result) {
    return ''
  }
  if (operation.value === 'add' || operation.value === 'sub') {
    return result.resultDate ?? ''
  }
  if (operation.value === 'workdays') {
    return String(result.workdays)
  }
  return String(result.diffDays)
})

/** 是否需要结束日期。 */
const needsDateB = computed(
  () => operation.value === 'diff' || operation.value === 'workdays',
)

/** 是否需要增减数量。 */
const needsDays = computed(() => operation.value === 'add' || operation.value === 'sub')

const faq = computed<FaqItem[]>(() => tm('tools.date-calc.faq') as FaqItem[])

/** 解析整数输入；非法时返回 undefined。 */
function parseIntOrUndefined(raw: number): number | undefined {
  return Number.isFinite(raw) ? Math.trunc(raw) : undefined
}

/** 解析节假日列表。 */
function parseHolidays(): string[] {
  return holidays.value
    .split('\n')
    .map((line) => line.trim())
    .filter((line) => line !== '')
}

/** 运行当前输入。 */
function run(): void {
  const payload: Record<string, unknown> = {
    operation: operation.value,
    dateA: dateA.value,
  }
  if (needsDateB.value) {
    payload.dateB = dateB.value
  }
  if (needsDays.value) {
    payload.days = parseIntOrUndefined(days.value) ?? 0
    payload.unit = unit.value
  }
  if (operation.value === 'workdays') {
    payload.holidays = parseHolidays()
  }
  void execute(payload)
}

/** 重置表单与结果。 */
function reset(): void {
  operation.value = 'diff'
  dateA.value = '2024-01-01'
  dateB.value = '2024-01-31'
  days.value = 10
  unit.value = 'day'
  holidays.value = ''
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  operation.value = example.operation
  dateA.value = example.dateA
  dateB.value = example.dateB
  days.value = example.days
  unit.value = example.unit
  holidays.value = example.holidays.join('\n')
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="date-calc"
    :name="t('tools.date-calc.name')"
    name-en="Date Calculator"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.date-calc.fields.dateA') }}</span>
            <input
              v-model="dateA"
              type="text"
              class="field__input tb-mono"
              autocomplete="off"
              placeholder="2024-01-01"
              :aria-label="t('tools.date-calc.fields.dateA')"
            >
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.date-calc.fields.dateB') }}</span>
            <input
              v-model="dateB"
              type="text"
              class="field__input tb-mono"
              autocomplete="off"
              placeholder="2024-01-31"
              :aria-label="t('tools.date-calc.fields.dateB')"
            >
          </label>

          <p class="field__hint">
            {{ t('tools.date-calc.hints.dateA') }}
          </p>
        </template>

        <template #options>
          <div class="field-row">
            <label class="field">
              <span class="field__label">{{ t('tools.date-calc.fields.operation') }}</span>
              <select
                v-model="operation"
                class="field__select"
              >
                <option value="diff">
                  {{ t('tools.date-calc.options.operation.diff') }}
                </option>
                <option value="add">
                  {{ t('tools.date-calc.options.operation.add') }}
                </option>
                <option value="sub">
                  {{ t('tools.date-calc.options.operation.sub') }}
                </option>
                <option value="workdays">
                  {{ t('tools.date-calc.options.operation.workdays') }}
                </option>
              </select>
            </label>

            <label
              v-if="needsDays"
              class="field"
            >
              <span class="field__label">{{ t('tools.date-calc.fields.days') }}</span>
              <input
                v-model.number="days"
                type="number"
                class="field__input tb-mono"
                step="1"
                :aria-label="t('tools.date-calc.fields.days')"
              >
            </label>

            <label
              v-if="needsDays"
              class="field"
            >
              <span class="field__label">{{ t('tools.date-calc.fields.unit') }}</span>
              <select
                v-model="unit"
                class="field__select"
              >
                <option value="day">
                  {{ t('tools.date-calc.options.unit.day') }}
                </option>
                <option value="week">
                  {{ t('tools.date-calc.options.unit.week') }}
                </option>
                <option value="month">
                  {{ t('tools.date-calc.options.unit.month') }}
                </option>
                <option value="year">
                  {{ t('tools.date-calc.options.unit.year') }}
                </option>
              </select>
            </label>
          </div>

          <p
            v-if="needsDays"
            class="field__hint"
          >
            {{ t('tools.date-calc.hints.unit') }}
          </p>

          <label
            v-if="operation === 'workdays'"
            class="field"
          >
            <span class="field__label">{{ t('tools.date-calc.fields.holidays') }}</span>
            <textarea
              v-model="holidays"
              class="field__textarea tb-mono"
              rows="3"
              spellcheck="false"
              placeholder="2024-01-01"
              :aria-label="t('tools.date-calc.fields.holidays')"
            />
          </label>
          <p
            v-if="operation === 'workdays'"
            class="field__hint"
          >
            {{ t('tools.date-calc.hints.holidays') }}
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
        :title="t('tools.date-calc.results.output')"
        @retry="run"
      >
        <template
          v-if="data"
          #extra
        >
          <dl class="detail">
            <div
              v-if="data?.resultDate"
              class="detail__row"
            >
              <dt>
                {{ t('tools.date-calc.results.resultDate') }}
              </dt>
              <dd class="tb-mono">
                {{ data?.resultDate }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>
                {{ t('tools.date-calc.results.diffDays') }}
              </dt>
              <dd class="tb-mono">
                {{ data?.diffDays }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>
                {{ t('tools.date-calc.results.workdays') }}
              </dt>
              <dd class="tb-mono">
                {{ data?.workdays }}
              </dd>
            </div>
          </dl>
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.date-calc.hints.output') }}
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
