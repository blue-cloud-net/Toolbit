<script setup lang="ts">
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
              :aria-label="t('tools.cert-lifecycle.fields.isPem')"
            >
              {{ t('tools.cert-lifecycle.fields.isPem') }}
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
