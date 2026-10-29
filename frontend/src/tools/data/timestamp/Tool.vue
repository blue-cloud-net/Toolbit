<script setup lang="ts">
/**
 * 时间戳转换工具页。
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
import { EXAMPLES } from '@/tools/data/timestamp/example'
import type { TimestampOperation, TimestampUnit } from '@/tools/data/timestamp/example'

/** 后端 timestamp 工具的响应。 */
interface TimestampResult {
  /** 秒与毫秒时间戳。 */
  timestamp: { s: number; ms: number }
  /** ISO 8601 格式的本地时间。 */
  iso8601: string
  /** RFC 3339 格式的本地时间。 */
  rfc3339: string
  /** 指定时区的本地时间。 */
  local: string
  /** UTC 时间。 */
  utc: string
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const operation = ref<TimestampOperation>('toDate')
const value = ref('1700000000')
const unit = ref<TimestampUnit | ''>('s')
const timezone = ref('')

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<TimestampResult>({
  slug: 'timestamp',
})

const running = computed(() => state.value === 'loading')

/** 结果文本：ISO 8601。 */
const resultText = computed(() => data.value?.iso8601 ?? '')

const faq = computed<FaqItem[]>(() => tm('tools.timestamp.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  void execute({
    operation: operation.value,
    value: value.value,
    unit: unit.value,
    timezone: timezone.value,
  })
}

/** 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。 */
useAutoRun({
  slug: 'timestamp',
  run,
  watch: [operation, value, unit, timezone],
})

/** 重置表单与结果。 */
function reset(): void {
  operation.value = 'toDate'
  value.value = '1700000000'
  unit.value = 's'
  timezone.value = ''
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  operation.value = example.operation
  value.value = example.value
  unit.value = example.unit
  timezone.value = example.timezone
  resetRun()
}

/** Ctrl / ⌘ + Enter 触发运行。 */
function onKeydown(event: KeyboardEvent): void {
  if ((event.ctrlKey || event.metaKey) && event.key === 'Enter') {
    event.preventDefault()
    run()
  }
}
</script>

<template>
  <ToolPageLayout
    slug="timestamp"
    :name="t('tools.timestamp.name')"
    name-en="Timestamp Converter"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.timestamp.fields.value') }}</span>
            <ElInput
              v-model="value"
              class="field__input tb-mono"
              autocomplete="off"
              placeholder="1700000000"
              :aria-label="t('tools.timestamp.fields.value')"
              @keydown="(event) => onKeydown(event as KeyboardEvent)"
            />
          </label>

          <p class="field__hint">
            {{ t('tools.timestamp.hints.value') }}
          </p>
        </template>

        <template #options>
          <div class="field-row">
            <label class="field">
              <span class="field__label">{{ t('tools.timestamp.fields.operation') }}</span>
              <ElSelect
                v-model="operation"
                class="field__select"
                :aria-label="t('tools.timestamp.fields.operation')"
              >
                <ElOption
                  :label="t('tools.timestamp.options.operation.toDate')"
                  value="toDate"
                />
                <ElOption
                  :label="t('tools.timestamp.options.operation.toTimestamp')"
                  value="toTimestamp"
                />
              </ElSelect>
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.timestamp.fields.unit') }}</span>
              <ElSelect
                v-model="unit"
                class="field__select"
                :aria-label="t('tools.timestamp.fields.unit')"
              >
                <ElOption
                  :label="t('tools.timestamp.options.unit.auto')"
                  value=""
                />
                <ElOption
                  :label="t('tools.timestamp.options.unit.s')"
                  value="s"
                />
                <ElOption
                  :label="t('tools.timestamp.options.unit.ms')"
                  value="ms"
                />
                <ElOption
                  :label="t('tools.timestamp.options.unit.us')"
                  value="us"
                />
                <ElOption
                  :label="t('tools.timestamp.options.unit.ns')"
                  value="ns"
                />
              </ElSelect>
            </label>
          </div>

          <label class="field">
            <span class="field__label">{{ t('tools.timestamp.fields.timezone') }}</span>
            <ElInput
              v-model="timezone"
              class="field__input tb-mono"
              autocomplete="off"
              placeholder="Asia/Shanghai"
              :aria-label="t('tools.timestamp.fields.timezone')"
            />
          </label>
          <p class="field__hint">
            {{ t('tools.timestamp.hints.timezone') }}
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
        :title="t('tools.timestamp.results.output')"
        @retry="run"
      >
        <template
          v-if="data"
          #extra
        >
          <dl class="detail">
            <div class="detail__row">
              <dt>
                {{ t('tools.timestamp.results.seconds') }}
              </dt>
              <dd class="tb-mono">
                {{ data?.timestamp.s }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>
                {{ t('tools.timestamp.results.milliseconds') }}
              </dt>
              <dd class="tb-mono">
                {{ data?.timestamp.ms }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>
                {{ t('tools.timestamp.results.iso8601') }}
              </dt>
              <dd class="tb-mono">
                {{ data?.iso8601 }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>
                {{ t('tools.timestamp.results.rfc3339') }}
              </dt>
              <dd class="tb-mono">
                {{ data?.rfc3339 }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>
                {{ t('tools.timestamp.results.local') }}
              </dt>
              <dd class="tb-mono">
                {{ data?.local }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>
                {{ t('tools.timestamp.results.utc') }}
              </dt>
              <dd class="tb-mono">
                {{ data?.utc }}
              </dd>
            </div>
          </dl>
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.timestamp.hints.output') }}
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
