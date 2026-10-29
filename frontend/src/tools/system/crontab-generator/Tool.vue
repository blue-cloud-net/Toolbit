<script setup lang="ts">
/**
 * Crontab 表达式生成工具页。
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
import { EXAMPLES, type CronPreset } from '@/tools/system/crontab-generator/example'

/** 一次触发时间。 */
interface CronOccurrence {
  /** 触发时刻（RFC 3339）。 */
  time: string
}

/** 后端 crontab-generator 工具的响应。 */
interface CrontabGeneratorResult {
  /** 五段式表达式。 */
  expression: string
  /** 表达式的人类可读描述。 */
  description: string
  /** 接下来的触发时间。 */
  occurrences: CronOccurrence[]
  /** 使用的时区。 */
  timezone: string
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

/** 受支持的预设；顺序与服务端一致。 */
const PRESETS: readonly CronPreset[] = [
  'every-minute',
  'every-5-minutes',
  'hourly',
  'daily',
  'weekly',
  'monthly',
  'yearly',
]

/** 预览次数上限，与服务端保持一致。 */
const MAX_COUNT = 10

const preset = ref<CronPreset | ''>('daily')
const minute = ref('0')
const hour = ref('0')
const dayOfMonth = ref('*')
const month = ref('*')
const dayOfWeek = ref('*')
const timezone = ref('UTC')
const count = ref(5)

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<CrontabGeneratorResult>(
  { slug: 'crontab-generator' },
)

const running = computed(() => state.value === 'loading')

/** 结果文本：生成的表达式。 */
const resultText = computed(() => data.value?.expression ?? '')

/** 接下来的触发时间。 */
const occurrences = computed<CronOccurrence[]>(() => data.value?.occurrences ?? [])

/** 是否使用自定义字段。 */
const custom = computed(() => preset.value === '')

/** 预览次数；空值或越界都归一到服务端允许的范围。 */
const countValue = computed(() => {
  const value = Number(count.value)
  if (!Number.isFinite(value) || value <= 0) {
    return 5
  }
  return Math.min(Math.trunc(value), MAX_COUNT)
})

const faq = computed<FaqItem[]>(() => tm('tools.crontab-generator.faq') as FaqItem[])

/** 预设项的显示文案。 */
function presetLabel(value: CronPreset | ''): string {
  const key = value === '' ? 'custom' : value
  return t(`tools.crontab-generator.options.preset.${key}`)
}

/** 运行当前输入。 */
function run(): void {
  void execute({
    preset: preset.value,
    minute: minute.value,
    hour: hour.value,
    dayOfMonth: dayOfMonth.value,
    month: month.value,
    dayOfWeek: dayOfWeek.value,
    timezone: timezone.value,
    count: countValue.value,
  })
}

/** 重置表单与结果。 */
function reset(): void {
  preset.value = 'daily'
  minute.value = '0'
  hour.value = '0'
  dayOfMonth.value = '*'
  month.value = '*'
  dayOfWeek.value = '*'
  timezone.value = 'UTC'
  count.value = 5
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  preset.value = example.preset
  minute.value = example.minute
  hour.value = example.hour
  dayOfMonth.value = example.dayOfMonth
  month.value = example.month
  dayOfWeek.value = example.dayOfWeek
  timezone.value = example.timezone
  count.value = example.count
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="crontab-generator"
    :name="t('tools.crontab-generator.name')"
    name-en="Crontab Generator"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.crontab-generator.fields.preset') }}</span>
            <select
              v-model="preset"
              class="field__select"
            >
              <option
                v-for="item in PRESETS"
                :key="item"
                :value="item"
              >
                {{ presetLabel(item) }}
              </option>
              <option value="">
                {{ presetLabel('') }}
              </option>
            </select>
          </label>

          <p class="field__hint">
            {{ t('tools.crontab-generator.hints.preset') }}
          </p>
        </template>

        <template #options>
          <fieldset
            class="field"
            :disabled="!custom"
          >
            <legend class="field__label">
              {{ t('tools.crontab-generator.fields.fields') }}
            </legend>

            <div class="field-row">
              <label class="field">
                <span class="field__label">{{ t('tools.crontab-generator.fields.minute') }}</span>
                <input
                  v-model="minute"
                  type="text"
                  class="field__input tb-mono"
                  autocomplete="off"
                  :aria-label="t('tools.crontab-generator.fields.minute')"
                >
              </label>

              <label class="field">
                <span class="field__label">{{ t('tools.crontab-generator.fields.hour') }}</span>
                <input
                  v-model="hour"
                  type="text"
                  class="field__input tb-mono"
                  autocomplete="off"
                  :aria-label="t('tools.crontab-generator.fields.hour')"
                >
              </label>

              <label class="field">
                <span class="field__label">{{ t('tools.crontab-generator.fields.dayOfMonth') }}</span>
                <input
                  v-model="dayOfMonth"
                  type="text"
                  class="field__input tb-mono"
                  autocomplete="off"
                  :aria-label="t('tools.crontab-generator.fields.dayOfMonth')"
                >
              </label>

              <label class="field">
                <span class="field__label">{{ t('tools.crontab-generator.fields.month') }}</span>
                <input
                  v-model="month"
                  type="text"
                  class="field__input tb-mono"
                  autocomplete="off"
                  :aria-label="t('tools.crontab-generator.fields.month')"
                >
              </label>

              <label class="field">
                <span class="field__label">{{ t('tools.crontab-generator.fields.dayOfWeek') }}</span>
                <input
                  v-model="dayOfWeek"
                  type="text"
                  class="field__input tb-mono"
                  autocomplete="off"
                  :aria-label="t('tools.crontab-generator.fields.dayOfWeek')"
                >
              </label>
            </div>
          </fieldset>
          <p class="field__hint">
            {{ t('tools.crontab-generator.hints.fields') }}
          </p>

          <div class="field-row">
            <label class="field">
              <span class="field__label">{{ t('tools.crontab-generator.fields.timezone') }}</span>
              <input
                v-model="timezone"
                type="text"
                class="field__input tb-mono"
                autocomplete="off"
                placeholder="UTC"
                :aria-label="t('tools.crontab-generator.fields.timezone')"
              >
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.crontab-generator.fields.count') }}</span>
              <input
                v-model.number="count"
                type="number"
                class="field__input"
                min="1"
                :max="MAX_COUNT"
                :aria-label="t('tools.crontab-generator.fields.count')"
              >
            </label>
          </div>

          <p class="field__hint">
            {{ t('tools.crontab-generator.hints.timezone') }}
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
        :title="t('tools.crontab-generator.results.expression')"
        @retry="run"
      >
        <template
          v-if="data"
          #extra
        >
          <dl class="detail">
            <div class="detail__row">
              <dt>{{ t('tools.crontab-generator.results.description') }}</dt>
              <dd>{{ data?.description }}</dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.crontab-generator.results.timezone') }}</dt>
              <dd class="tb-mono">
                {{ data?.timezone }}
              </dd>
            </div>
          </dl>

          <div
            v-if="occurrences.length > 0"
            class="occurrences"
          >
            <span class="occurrences__label">
              {{ t('tools.crontab-generator.results.occurrences') }}
            </span>
            <ol class="occurrences__list">
              <li
                v-for="(item, index) in occurrences"
                :key="index"
                class="occurrences__item tb-mono"
              >
                {{ item.time }}
              </li>
            </ol>
          </div>
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.crontab-generator.hints.output') }}
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

.occurrences {
  margin-top: var(--tb-space-3);
}

.occurrences__label {
  font-size: var(--tb-font-xs);
  color: var(--tb-text-tertiary);
}

.occurrences__list {
  margin: var(--tb-space-1) 0 0;
  padding-left: var(--tb-space-4);
  font-size: var(--tb-font-sm);
}

.occurrences__item {
  color: var(--tb-text-secondary);
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
