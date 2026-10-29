<script setup lang="ts">
/**
 * 敏感数据脱敏工具页。
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
import { EXAMPLES } from '@/tools/text/data-mask/example'

/** 一处脱敏记录。 */
interface MaskMatch {
  /** 脱敏类型。 */
  type: string
  /** 命中次数。 */
  count: number
  /** 脱敏样例。 */
  sample: string
}

/** 后端 data-mask 工具的响应。 */
interface DataMaskResult {
  /** 脱敏后的文本。 */
  masked: string
  /** 各类脱敏统计。 */
  matches: MaskMatch[]
  /** 脱敏总数。 */
  total: number
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const text = ref('手机号 13800138000，邮箱 user@example.com，身份证 110101199003072516')
const types = ref<string[]>([])
const position = ref('keepHead')
const keepCount = ref(3)

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<DataMaskResult>({
  slug: 'data-mask',
})

const running = computed(() => state.value === 'loading')

/** 结果文本：脱敏后的内容。 */
const resultText = computed(() => data.value?.masked ?? '')

/** 结论徽章：命中数量。 */
const verdict = computed(() => {
  if (!data.value) {
    return null
  }
  return data.value.total > 0
    ? { level: 'warning' as const, label: `${data.value.total}` }
    : { level: 'success' as const, label: t('tools.data-mask.results.total') }
})

const faq = computed<FaqItem[]>(() => tm('tools.data-mask.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  if (text.value.trim() === '') {
    return
  }
  void execute({
    text: text.value,
    types: types.value,
    position: position.value,
    keepCount: keepCount.value,
  })
}

// 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。
useAutoRun({
  slug: 'data-mask',
  run,
  watch: [text, types, position, keepCount],
  enabled: () => text.value.trim() !== '',
})

/** 重置表单与结果。 */
function reset(): void {
  text.value = '手机号 13800138000，邮箱 user@example.com，身份证 110101199003072516'
  types.value = []
  position.value = 'keepHead'
  keepCount.value = 3
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  text.value = example.text
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="data-mask"
    :name="t('tools.data-mask.name')"
    name-en="Data Masker"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field field--grow">
            <span class="field__label">{{ t('tools.data-mask.fields.text') }}</span>
            <ElInput
              v-model="text"
              class="field__input tb-mono"
              type="textarea"
              :rows="6"
              autocomplete="off"
              :aria-label="t('tools.data-mask.fields.text')"
            />
          </label>

          <p class="field__hint">
            {{ t('tools.data-mask.hints.text') }}
          </p>
        </template>

        <template #options>
          <div class="field field--grow">
            <span class="field__label">{{ t('tools.data-mask.fields.types') }}</span>
            <ElCheckboxGroup
              v-model="types"
              class="field__checkgroup"
              :aria-label="t('tools.data-mask.fields.types')"
            >
              <ElCheckbox value="phone">
                phone
              </ElCheckbox>
              <ElCheckbox value="idCard">
                idCard
              </ElCheckbox>
              <ElCheckbox value="bankCard">
                bankCard
              </ElCheckbox>
              <ElCheckbox value="email">
                email
              </ElCheckbox>
              <ElCheckbox value="name">
                name
              </ElCheckbox>
              <ElCheckbox value="ip">
                ip
              </ElCheckbox>
            </ElCheckboxGroup>
          </div>

          <label class="field">
            <span class="field__label">{{ t('tools.data-mask.fields.position') }}</span>
            <ElSelect
              v-model="position"
              class="field__select"
              :aria-label="t('tools.data-mask.fields.position')"
            >
              <ElOption
                :label="t('tools.data-mask.options.position.keepHead')"
                value="keepHead"
              />
              <ElOption
                :label="t('tools.data-mask.options.position.keepTail')"
                value="keepTail"
              />
              <ElOption
                :label="t('tools.data-mask.options.position.keepBoth')"
                value="keepBoth"
              />
              <ElOption
                :label="t('tools.data-mask.options.position.hash')"
                value="hash"
              />
            </ElSelect>
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.data-mask.fields.keepCount') }}</span>
            <ElInputNumber
              v-model="keepCount"
              class="field__select"
              :min="0"
              :max="32"
              :aria-label="t('tools.data-mask.fields.keepCount')"
            />
          </label>
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
        :title="t('tools.data-mask.results.masked')"
        @retry="run"
      >
        <template #extra>
          <dl class="detail">
            <div class="detail__row">
              <dt>{{ t('tools.data-mask.results.total') }}</dt>
              <dd class="tb-mono">
                {{ data?.total ?? 0 }}
              </dd>
            </div>
            <div
              v-for="item in data?.matches ?? []"
              :key="item.type"
              class="detail__row"
            >
              <dt class="tb-mono">
                {{ item.type }}
              </dt>
              <dd class="tb-mono">
                {{ item.count }} · {{ item.sample }}
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
