#!/usr/bin/env python3
"""生成 v0.8.0 批次新增工具的 Tool.vue（加密 / 证书 / 密钥类）。"""
from __future__ import annotations

import pathlib

ROOT = pathlib.Path(__file__).resolve().parents[1]
BASE = ROOT / "frontend" / "src" / "tools"

COMPONENTS: dict[str, str] = {}

COMPONENTS["crypto/password-policy"] = '''<script setup lang="ts">
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

import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useAutoRun } from '@/composables/useAutoRun'
import { useToolRun } from '@/composables/useToolRun'
import { EXAMPLES } from '@/tools/crypto/password-policy/example'

/** 一条检查项。 */
interface PolicyCheck {
  /** 检查项标识。 */
  check: string
  /** 是否通过。 */
  passed: boolean
  /** 说明。 */
  message: string
}

/** 后端 password-policy 工具的响应。 */
interface PasswordPolicyResult {
  /** 生效的预设。 */
  preset: string
  /** 是否满足整体策略。 */
  compliant: boolean
  /** 强度评分（0-100）。 */
  score: number
  /** 强度等级。 */
  strength: string
  /** 明细检查项。 */
  checks: PolicyCheck[]
  /** 改进建议。 */
  suggestions: string[]
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
    .map((check) => `${check.passed ? 'OK' : 'NG'}\\t${check.message}`)
    .join('\\n')
})

/** 结论徽章。 */
const verdict = computed(() => {
  if (!data.value) {
    return null
  }
  if (data.value.compliant) {
    return { level: 'success' as const, label: t('tools.password-policy.results.passed') }
  }
  return { level: 'danger' as const, label: t('tools.password-policy.results.failed') }
})

const faq = computed<FaqItem[]>(() => tm('tools.password-policy.faq') as FaqItem[])

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
        @retry="run"
      >
        <template #extra>
          <dl class="detail">
            <div class="detail__row">
              <dt>{{ t('tools.password-policy.results.score') }}</dt>
              <dd class="tb-mono">
                {{ data?.score ?? 0 }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.password-policy.results.strength') }}</dt>
              <dd class="tb-mono">
                {{ data?.strength ?? '' }}
              </dd>
            </div>
            <div
              v-for="suggestion in data?.suggestions ?? []"
              :key="suggestion"
              class="detail__row"
            >
              <dt>!</dt>
              <dd>{{ suggestion }}</dd>
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
'''

COMPONENTS["crypto/totp"] = '''<script setup lang="ts">
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
  /** 验证结果；生成模式下为 null。 */
  valid: boolean | null
  /** 摘要算法。 */
  algorithm: string
  /** 验证码位数。 */
  digits: number
  /** 时间步长（秒）。 */
  period: number
  /** 时间步计数。 */
  counter: number
  /** 生效起点。 */
  validFrom: string
  /** 失效点。 */
  validUntil: string
  /** 剩余有效秒数。 */
  remainingSeconds: number
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
  if (!data.value || data.value.valid === null) {
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
                {{ data?.remainingSeconds ?? 0 }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.totp.results.counter') }}</dt>
              <dd class="tb-mono">
                {{ data?.counter ?? 0 }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.totp.results.validFrom') }}</dt>
              <dd class="tb-mono">
                {{ data?.validFrom ?? '' }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.totp.results.validUntil') }}</dt>
              <dd class="tb-mono">
                {{ data?.validUntil ?? '' }}
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
'''

COMPONENTS["certificates/cert-lifecycle"] = '''<script setup lang="ts">
/**
 * 证书生命周期工具页。
 *
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
 */
import { ElCheckbox } from 'element-plus/es/components/checkbox/index'
import 'element-plus/es/components/checkbox/style/css'
import { ElInputNumber } from 'element-plus/es/components/input-number/index'
import 'element-plus/es/components/input-number/style/css'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import CertTextInput from '@/components/tool/CertTextInput.vue'
import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useAutoRun } from '@/composables/useAutoRun'
import { useToolRun } from '@/composables/useToolRun'
import { EXAMPLES } from '@/tools/certificates/cert-lifecycle/example'

/** 后端 cert-lifecycle 工具的响应。 */
interface CertLifecycleResult {
  /** 生效时间。 */
  notBefore: string
  /** 失效时间。 */
  notAfter: string
  /** 有效期总天数。 */
  validDays: number
  /** 剩余天数。 */
  daysRemaining: number
  /** 剩余小时数。 */
  hoursRemaining: number
  /** 状态。 */
  status: string
  /** 建议。 */
  recommendations: string[]
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const certificate = ref('')
const inputFormat = ref('auto')
const isPEM = ref(true)
const warningDays = ref(30)

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<CertLifecycleResult>({
  slug: 'cert-lifecycle',
})

const running = computed(() => state.value === 'loading')

/** 结果文本：状态摘要。 */
const resultText = computed(() => {
  if (!data.value) {
    return ''
  }
  return `${data.value.notBefore} ~ ${data.value.notAfter}`
})

/** 结论徽章。 */
const verdict = computed(() => {
  if (!data.value) {
    return null
  }
  if (data.value.daysRemaining < 0) {
    return { level: 'danger' as const, label: t('tools.cert-lifecycle.results.expired') }
  }
  if (data.value.daysRemaining <= warningDays.value) {
    return { level: 'warning' as const, label: t('tools.cert-lifecycle.results.expiring') }
  }
  return { level: 'success' as const, label: t('tools.cert-lifecycle.results.valid') }
})

const faq = computed<FaqItem[]>(() => tm('tools.cert-lifecycle.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  if (certificate.value.trim() === '') {
    return
  }
  void execute({
    certificate: certificate.value,
    isPEM: isPEM.value,
    warningDays: warningDays.value,
  })
}

// 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。
useAutoRun({
  slug: 'cert-lifecycle',
  run,
  watch: [certificate, isPEM, warningDays],
  enabled: () => certificate.value.trim() !== '',
})

/** 重置表单与结果。 */
function reset(): void {
  certificate.value = ''
  inputFormat.value = 'auto'
  isPEM.value = true
  warningDays.value = 30
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  certificate.value = example.certificate
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="cert-lifecycle"
    :name="t('tools.cert-lifecycle.name')"
    name-en="Certificate Lifecycle"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <CertTextInput
            v-model="certificate"
            v-model:format="inputFormat"
            :label="t('tools.cert-lifecycle.fields.certificate')"
          />
        </template>

        <template #options>
          <label class="field">
            <span class="field__label">{{ t('tools.cert-lifecycle.fields.warningDays') }}</span>
            <ElInputNumber
              v-model="warningDays"
              class="field__select"
              :min="1"
              :max="365"
              :aria-label="t('tools.cert-lifecycle.fields.warningDays')"
            />
          </label>

          <label class="field field--check">
            <ElCheckbox
              v-model="isPEM"
              :aria-label="t('tools.cert-lifecycle.fields.isPEM')"
            >
              {{ t('tools.cert-lifecycle.fields.isPEM') }}
            </ElCheckbox>
          </label>

          <p class="field__hint">
            {{ t('tools.cert-lifecycle.hints.certificate') }}
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
        :title="t('tools.cert-lifecycle.results.validity')"
        @retry="run"
      >
        <template #extra>
          <dl class="detail">
            <div class="detail__row">
              <dt>{{ t('tools.cert-lifecycle.results.notBefore') }}</dt>
              <dd class="tb-mono">
                {{ data?.notBefore ?? '' }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.cert-lifecycle.results.notAfter') }}</dt>
              <dd class="tb-mono">
                {{ data?.notAfter ?? '' }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.cert-lifecycle.results.validDays') }}</dt>
              <dd class="tb-mono">
                {{ data?.validDays ?? 0 }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.cert-lifecycle.results.daysRemaining') }}</dt>
              <dd class="tb-mono">
                {{ data?.daysRemaining ?? 0 }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.cert-lifecycle.results.hoursRemaining') }}</dt>
              <dd class="tb-mono">
                {{ data?.hoursRemaining ?? 0 }}
              </dd>
            </div>
            <div
              v-for="recommendation in data?.recommendations ?? []"
              :key="recommendation"
              class="detail__row"
            >
              <dt>!</dt>
              <dd>{{ recommendation }}</dd>
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
'''

COMPONENTS["keys/ssh-keygen"] = '''<script setup lang="ts">
/**
 * SSH 密钥生成工具页。
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
import { EXAMPLES } from '@/tools/keys/ssh-keygen/example'

/** 后端 ssh-keygen 工具的响应。 */
interface SSHKeygenResult {
  /** 算法。 */
  algorithm: string
  /** 私钥（OpenSSH 格式）。 */
  privateKey: string
  /** 公钥（OpenSSH 格式）。 */
  publicKey: string
  /** 指纹。 */
  fingerprint: string
  /** 密钥位数。 */
  bits: number
  /** 注释。 */
  comment: string
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const algorithm = ref('Ed25519')
const bits = ref(3072)
const comment = ref('toolbit@example.com')

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<SSHKeygenResult>({
  slug: 'ssh-keygen',
})

const running = computed(() => state.value === 'loading')

/** 结果文本：私钥。 */
const resultText = computed(() => data.value?.privateKey ?? '')

const faq = computed<FaqItem[]>(() => tm('tools.ssh-keygen.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  void execute({
    algorithm: algorithm.value,
    bits: bits.value,
    comment: comment.value,
  })
}

// 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。
useAutoRun({
  slug: 'ssh-keygen',
  run,
  watch: [algorithm, bits, comment],
  enabled: () => true,
})

/** 重置表单与结果。 */
function reset(): void {
  algorithm.value = 'Ed25519'
  bits.value = 3072
  comment.value = 'toolbit@example.com'
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  algorithm.value = example.algorithm
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="ssh-keygen"
    :name="t('tools.ssh-keygen.name')"
    name-en="SSH Key Generator"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.ssh-keygen.fields.algorithm') }}</span>
            <ElSelect
              v-model="algorithm"
              class="field__select"
              :aria-label="t('tools.ssh-keygen.fields.algorithm')"
            >
              <ElOption
                :label="t('tools.ssh-keygen.options.algorithm.Ed25519')"
                value="Ed25519"
              />
              <ElOption
                :label="t('tools.ssh-keygen.options.algorithm.RSA')"
                value="RSA"
              />
            </ElSelect>
          </label>

          <label
            v-if="algorithm === 'RSA'"
            class="field"
          >
            <span class="field__label">{{ t('tools.ssh-keygen.fields.bits') }}</span>
            <ElInputNumber
              v-model="bits"
              class="field__select"
              :min="2048"
              :max="8192"
              :step="1024"
              :aria-label="t('tools.ssh-keygen.fields.bits')"
            />
          </label>

          <label class="field field--grow">
            <span class="field__label">{{ t('tools.ssh-keygen.fields.comment') }}</span>
            <ElInput
              v-model="comment"
              class="field__input tb-mono"
              autocomplete="off"
              :aria-label="t('tools.ssh-keygen.fields.comment')"
              @keydown.enter.prevent="run"
            />
          </label>

          <p class="field__hint">
            {{ t('tools.ssh-keygen.hints.privateKey') }}
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
        :title="t('tools.ssh-keygen.results.privateKey')"
        @retry="run"
      >
        <template #extra>
          <dl class="detail">
            <div class="detail__row">
              <dt>{{ t('tools.ssh-keygen.results.fingerprint') }}</dt>
              <dd class="tb-mono">
                {{ data?.fingerprint ?? '' }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.ssh-keygen.results.publicKey') }}</dt>
              <dd class="tb-mono">
                {{ data?.publicKey ?? '' }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.ssh-keygen.results.comment') }}</dt>
              <dd class="tb-mono">
                {{ data?.comment ?? '' }}
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
'''


def main() -> None:
    for relative, content in COMPONENTS.items():
        target = BASE / relative / "Tool.vue"
        target.write_text(content, encoding="utf-8")
        print(f"写入 {relative}/Tool.vue")


if __name__ == "__main__":
    main()
