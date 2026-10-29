<script setup lang="ts">
/**
 * TOTP 工具页。
 *
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
 */
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { ElInputNumber } from 'element-plus/es/components/input-number/index'
import 'element-plus/es/components/input-number/style/css'
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
import { EXAMPLES } from '@/tools/crypto/totp/example'

/** 后端 totp 工具的响应。 */
interface TOTPResult {
  /** 操作。 */
  operation: string
  /** 生成的验证码。 */
  code: string
  /** 验证结果。 */
  valid: boolean
  /** 剩余有效秒数。 */
  remaining: number
  /** 用于验证的时间步。 */
  counter: number
  /** Base32 密钥（仅生成模式下返回）。 */
  secret: string
  /** 密钥对应的 otpauth URI。 */
  uri: string
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const operation = ref('generate')
const secret = ref('JBSWY3DPEHPK3PXP')
const algorithm = ref('SHA1')
const digits = ref(6)
const period = ref(30)
const timestamp = ref(0)
const code = ref('')
const window = ref(1)

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<TOTPResult>({
  slug: 'totp',
})

const running = computed(() => state.value === 'loading')

/** 结果文本：验证码。 */
const resultText = computed(() => data.value?.code ?? '')

/** 结论徽章：验证结果。 */
const verdict = computed(() => {
  if (!data.value || data.value.operation !== 'verify') {
    return null
  }
  return data.value.valid
    ? { level: 'success' as const, label: t('tools.totp.results.valid') }
    : { level: 'danger' as const, label: t('tools.totp.results.invalid') }
})

const faq = computed<FaqItem[]>(() => tm('tools.totp.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  void execute({
    operation: operation.value,
    secret: secret.value,
    algorithm: algorithm.value,
    digits: digits.value,
    period: period.value,
    timestamp: timestamp.value,
    code: code.value,
    window: window.value,
  })
}

// 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。
useAutoRun({
  slug: 'totp',
  run,
  watch: [operation, secret, algorithm, digits, period, timestamp, code, window],
  enabled: () => secret.value.trim() !== '',
})

/** 重置表单与结果。 */
function reset(): void {
  operation.value = 'generate'
  secret.value = 'JBSWY3DPEHPK3PXP'
  algorithm.value = 'SHA1'
  digits.value = 6
  period.value = 30
  timestamp.value = 0
  code.value = ''
  window.value = 1
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  secret.value = example.secret
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="totp"
    :name="t('tools.totp.name')"
    name-en="TOTP Generator"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.totp.fields.operation') }}</span>
            <ElSelect
              v-model="operation"
              class="field__select"
              :aria-label="t('tools.totp.fields.operation')"
            >
              <ElOption
                :label="t('tools.totp.options.operation.generate')"
                value="generate"
              />
              <ElOption
                :label="t('tools.totp.options.operation.verify')"
                value="verify"
              />
            </ElSelect>
          </label>

          <label class="field field--grow">
            <span class="field__label">{{ t('tools.totp.fields.secret') }}</span>
            <ElInput
              v-model="secret"
              class="field__input tb-mono"
              autocomplete="off"
              :aria-label="t('tools.totp.fields.secret')"
              @keydown.enter.prevent="run"
            />
          </label>

          <label
            v-if="operation === 'verify'"
            class="field field--grow"
          >
            <span class="field__label">{{ t('tools.totp.fields.code') }}</span>
            <ElInput
              v-model="code"
              class="field__input tb-mono"
              autocomplete="off"
              :aria-label="t('tools.totp.fields.code')"
              @keydown.enter.prevent="run"
            />
          </label>
        </template>

        <template #options>
          <label class="field">
            <span class="field__label">{{ t('tools.totp.fields.algorithm') }}</span>
            <ElSelect
              v-model="algorithm"
              class="field__select"
              :aria-label="t('tools.totp.fields.algorithm')"
            >
              <ElOption
                v-for="name in ['SHA1', 'SHA256', 'SHA512']"
                :key="name"
                :label="name"
                :value="name"
              />
            </ElSelect>
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.totp.fields.digits') }}</span>
            <ElInputNumber
              v-model="digits"
              class="field__select"
              :min="6"
              :max="8"
              :aria-label="t('tools.totp.fields.digits')"
            />
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.totp.fields.period') }}</span>
            <ElInputNumber
              v-model="period"
              class="field__select"
              :min="10"
              :max="120"
              :aria-label="t('tools.totp.fields.period')"
            />
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.totp.fields.timestamp') }}</span>
            <ElInputNumber
              v-model="timestamp"
              class="field__select"
              :min="0"
              :aria-label="t('tools.totp.fields.timestamp')"
            />
          </label>

          <label
            v-if="operation === 'verify'"
            class="field"
          >
            <span class="field__label">{{ t('tools.totp.fields.window') }}</span>
            <ElInputNumber
              v-model="window"
              class="field__select"
              :min="0"
              :max="10"
              :aria-label="t('tools.totp.fields.window')"
            />
          </label>

          <p class="field__hint">
            {{ t('tools.totp.hints.secret') }}
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
        :title="t('tools.totp.results.code')"
        @retry="run"
      >
        <template #extra>
          <dl class="detail">
            <div class="detail__row">
              <dt>{{ t('tools.totp.results.remaining') }}</dt>
              <dd class="tb-mono">
                {{ data?.remaining ?? 0 }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.totp.results.counter') }}</dt>
              <dd class="tb-mono">
                {{ data?.counter ?? 0 }}
              </dd>
            </div>
            <div
              v-if="data?.uri"
              class="detail__row"
            >
              <dt>{{ t('tools.totp.results.uri') }}</dt>
              <dd class="tb-mono">
                {{ data.uri }}
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
