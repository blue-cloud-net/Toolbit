<script setup lang="ts">
/**
 * 假数据生成工具页。
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

import DataTable, { type DataTableColumn, type DataTableRow } from '@/components/tool/DataTable.vue'
import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import StatGrid, { type StatGridItem } from '@/components/tool/StatGrid.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useAutoRun } from '@/composables/useAutoRun'
import { useToolRun } from '@/composables/useToolRun'
import { EXAMPLES } from '@/tools/data/mock-data/example'

/** 一个生成的字段。 */
interface MockField {
  /** 字段类型。 */
  type: string
  /** 生成的值。 */
  value: string
}

/** 后端 mock-data 工具的响应。 */
interface MockDataResult {
  /** 生成的类型列表。 */
  types: string[]
  /** 生成的字段值。 */
  fields: MockField[]
  /** 每类数量。 */
  count: number
  /** 语言。 */
  locale: string
  /** 随机种子。 */
  seed: string
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const types = ref<string[]>(['email', 'name', 'uuid'])
const count = ref(3)
const locale = ref('zh-CN')
const prefix = ref('')

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<MockDataResult>({
  slug: 'mock-data',
})

const running = computed(() => state.value === 'loading')

/** 结果文本：生成的字段值。 */
const resultText = computed(() => {
  if (!data.value) {
    return ''
  }
  return data.value.fields.map((field) => `${field.type}\t${field.value}`).join('\n')
})

const faq = computed<FaqItem[]>(() => tm('tools.mock-data.faq') as FaqItem[])

/** 统计卡片。 */
const stats = computed<StatGridItem[]>(() => [
  { label: t('tools.mock-data.results.count'), value: String(data.value?.count ?? 0) },
  { label: t('tools.mock-data.results.seed'), value: data.value?.seed ?? '-' },
])

/** 表格列定义。 */
const columns = computed<DataTableColumn[]>(() => [
  { key: 'type', label: t('tools.mock-data.results.type'), mono: true },
  { key: 'value', label: t('tools.mock-data.results.value'), mono: true },
])

/** 表格行。 */
const rows = computed<DataTableRow[]>(() =>
  (data.value?.fields ?? []).map((field, index) => ({
    key: `${field.type}-${index}`,
    cells: { type: field.type, value: field.value },
  })),
)

/** 运行当前输入。 */
function run(): void {
  void execute({
    types: types.value,
    count: count.value,
    locale: locale.value,
    prefix: prefix.value,
  })
}

// 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。
useAutoRun({
  slug: 'mock-data',
  run,
  watch: [types, count, locale, prefix],
  enabled: () => true,
})

/** 重置表单与结果。 */
function reset(): void {
  types.value = ['email', 'name', 'uuid']
  count.value = 3
  locale.value = 'zh-CN'
  prefix.value = ''
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  types.value = example.types.split(',')
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="mock-data"
    :name="t('tools.mock-data.name')"
    name-en="Mock Data Generator"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <div class="field field--grow">
            <span class="field__label">{{ t('tools.mock-data.fields.types') }}</span>
            <ElCheckboxGroup
              v-model="types"
              class="field__checkgroup"
              :aria-label="t('tools.mock-data.fields.types')"
            >
              <ElCheckbox value="string">
                string
              </ElCheckbox>
              <ElCheckbox value="int">
                int
              </ElCheckbox>
              <ElCheckbox value="float">
                float
              </ElCheckbox>
              <ElCheckbox value="bool">
                bool
              </ElCheckbox>
              <ElCheckbox value="email">
                email
              </ElCheckbox>
              <ElCheckbox value="phone">
                phone
              </ElCheckbox>
              <ElCheckbox value="name">
                name
              </ElCheckbox>
              <ElCheckbox value="url">
                url
              </ElCheckbox>
              <ElCheckbox value="uuid">
                uuid
              </ElCheckbox>
              <ElCheckbox value="date">
                date
              </ElCheckbox>
              <ElCheckbox value="datetime">
                datetime
              </ElCheckbox>
              <ElCheckbox value="ip">
                ip
              </ElCheckbox>
              <ElCheckbox value="color">
                color
              </ElCheckbox>
            </ElCheckboxGroup>
          </div>
        </template>

        <template #options>
          <label class="field">
            <span class="field__label">{{ t('tools.mock-data.fields.count') }}</span>
            <ElInputNumber
              v-model="count"
              class="field__select"
              :min="0"
              :max="100"
              :aria-label="t('tools.mock-data.fields.count')"
            />
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.mock-data.fields.locale') }}</span>
            <ElSelect
              v-model="locale"
              class="field__select"
              :aria-label="t('tools.mock-data.fields.locale')"
            >
              <ElOption
                :label="t('tools.mock-data.options.locale.zh-CN')"
                value="zh-CN"
              />
              <ElOption
                :label="t('tools.mock-data.options.locale.en-US')"
                value="en-US"
              />
            </ElSelect>
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.mock-data.fields.prefix') }}</span>
            <ElInput
              v-model="prefix"
              class="field__input tb-mono"
              autocomplete="off"
              :aria-label="t('tools.mock-data.fields.prefix')"
              @keydown.enter.prevent="run"
            />
          </label>

          <p class="field__hint">
            {{ t('tools.mock-data.hints.types') }}
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
        :title="t('tools.mock-data.results.fields')"
        hide-text
        @retry="run"
      >
        <template #extra>
          <StatGrid :items="stats" />

          <DataTable
            class="table"
            :columns="columns"
            :rows="rows"
            :empty="t('tools.mock-data.results.empty')"
          />
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
</style>
