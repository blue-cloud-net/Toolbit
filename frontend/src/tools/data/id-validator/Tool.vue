<script setup lang="ts">
/**
 * 证件号码校验工具页。
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
import { EXAMPLES } from '@/tools/data/id-validator/example'

/** 后端 id-validator 工具的响应。 */
interface IDValidateResult {
  /** 校验类型。 */
  type: string
  /** 规范化后的值。 */
  value: string
  /** 是否校验通过。 */
  valid: boolean
  /** 规范化结果。 */
  normalized: string
  /** 校验位。 */
  checkDigit: string
  /** 归属地区。 */
  region: string
  /** 出生日期。 */
  birthday: string
  /** 性别。 */
  gender: string
  /** 失败原因。 */
  issue: string
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const type = ref('idCard')
const value = ref('110101199003072516')

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<IDValidateResult>({
  slug: 'id-validator',
})

const running = computed(() => state.value === 'loading')

/** 结果文本：规范化后的值或失败原因。 */
const resultText = computed(() => {
  if (!data.value) {
    return ''
  }
  return data.value.valid ? data.value.normalized : data.value.issue
})

/** 结论徽章：校验是否通过。 */
const verdict = computed(() => {
  if (!data.value) {
    return null
  }
  return data.value.valid
    ? { level: 'success' as const, label: t('tools.id-validator.results.valid') }
    : { level: 'danger' as const, label: t('tools.id-validator.results.invalid') }
})

const faq = computed<FaqItem[]>(() => tm('tools.id-validator.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  if (value.value.trim() === '') {
    return
  }
  void execute({ type: type.value, value: value.value })
}

// 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。
useAutoRun({
  slug: 'id-validator',
  run,
  watch: [type, value],
  enabled: () => value.value.trim() !== '',
})

/** 重置表单与结果。 */
function reset(): void {
  type.value = 'idCard'
  value.value = '110101199003072516'
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  type.value = example.type
  value.value = example.value
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="id-validator"
    :name="t('tools.id-validator.name')"
    name-en="ID Validator"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.id-validator.fields.type') }}</span>
            <ElSelect
              v-model="type"
              class="field__select"
              :aria-label="t('tools.id-validator.fields.type')"
            >
              <ElOption
                :label="t('tools.id-validator.options.type.idCard')"
                value="idCard"
              />
              <ElOption
                :label="t('tools.id-validator.options.type.uscc')"
                value="uscc"
              />
              <ElOption
                :label="t('tools.id-validator.options.type.phone')"
                value="phone"
              />
              <ElOption
                :label="t('tools.id-validator.options.type.bankCard')"
                value="bankCard"
              />
              <ElOption
                :label="t('tools.id-validator.options.type.postcode')"
                value="postcode"
              />
            </ElSelect>
          </label>

          <label class="field field--grow">
            <span class="field__label">{{ t('tools.id-validator.fields.value') }}</span>
            <ElInput
              v-model="value"
              class="field__input tb-mono"
              autocomplete="off"
              :aria-label="t('tools.id-validator.fields.value')"
              @keydown.enter.prevent="run"
            />
          </label>

          <p class="field__hint">
            {{ t('tools.id-validator.hints.value') }}
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
        :title="t('tools.id-validator.results.normalized')"
        @retry="run"
      >
        <template #extra>
          <dl class="detail">
            <div
              v-if="data?.region"
              class="detail__row"
            >
              <dt>{{ t('tools.id-validator.results.region') }}</dt>
              <dd class="tb-mono">
                {{ data.region }}
              </dd>
            </div>
            <div
              v-if="data?.birthday"
              class="detail__row"
            >
              <dt>{{ t('tools.id-validator.results.birthday') }}</dt>
              <dd class="tb-mono">
                {{ data.birthday }}
              </dd>
            </div>
            <div
              v-if="data?.gender"
              class="detail__row"
            >
              <dt>{{ t('tools.id-validator.results.gender') }}</dt>
              <dd class="tb-mono">
                {{ data.gender }}
              </dd>
            </div>
            <div
              v-if="data?.checkDigit"
              class="detail__row"
            >
              <dt>{{ t('tools.id-validator.results.checkDigit') }}</dt>
              <dd class="tb-mono">
                {{ data.checkDigit }}
              </dd>
            </div>
            <div
              v-if="data && !data.valid"
              class="detail__row"
            >
              <dt>{{ t('tools.id-validator.results.issue') }}</dt>
              <dd class="tb-mono">
                {{ data.issue }}
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
