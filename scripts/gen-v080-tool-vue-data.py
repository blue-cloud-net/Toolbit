#!/usr/bin/env python3
"""生成 v0.8.0 批次新增工具的 Tool.vue（数据类）。

每个工具的组件代码在这里完整给出，不做抽象：
交互形态差异大，抽象反而会产出千篇一律的界面。
"""
from __future__ import annotations

import pathlib

ROOT = pathlib.Path(__file__).resolve().parents[1]
TOOLS = ROOT / "frontend" / "src" / "tools" / "data"

COMPONENTS: dict[str, str] = {}

COMPONENTS["id-validator"] = '''<script setup lang="ts">
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
'''

COMPONENTS["csv-convert"] = '''<script setup lang="ts">
/**
 * CSV 转换工具页。
 *
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
 */
import { ElCheckbox } from 'element-plus/es/components/checkbox/index'
import 'element-plus/es/components/checkbox/style/css'
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
import { EXAMPLES } from '@/tools/data/csv-convert/example'

/** 后端 csv-convert 工具的响应。 */
interface CSVConvertResult {
  /** 转换方向。 */
  operation: string
  /** 转换结果。 */
  output: string
  /** 数据行数。 */
  rowCount: number
  /** 列数。 */
  columnCount: number
  /** 列名。 */
  columns: string[]
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const operation = ref('csvToJson')
const input = ref('name,age\\nAlice,30\\nBob,25')
const delimiter = ref(',')
const hasHeader = ref(true)
const pretty = ref(false)

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<CSVConvertResult>({
  slug: 'csv-convert',
})

const running = computed(() => state.value === 'loading')

/** 结果文本：转换结果。 */
const resultText = computed(() => data.value?.output ?? '')

const faq = computed<FaqItem[]>(() => tm('tools.csv-convert.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  if (input.value.trim() === '') {
    return
  }
  void execute({
    operation: operation.value,
    input: input.value,
    delimiter: delimiter.value,
    hasHeader: hasHeader.value,
    pretty: pretty.value,
  })
}

// 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。
useAutoRun({
  slug: 'csv-convert',
  run,
  watch: [operation, input, delimiter, hasHeader, pretty],
  enabled: () => input.value.trim() !== '',
})

/** 重置表单与结果。 */
function reset(): void {
  operation.value = 'csvToJson'
  input.value = 'name,age\\nAlice,30\\nBob,25'
  delimiter.value = ','
  hasHeader.value = true
  pretty.value = false
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  input.value = example.input
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="csv-convert"
    :name="t('tools.csv-convert.name')"
    name-en="CSV Converter"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.csv-convert.fields.operation') }}</span>
            <ElSelect
              v-model="operation"
              class="field__select"
              :aria-label="t('tools.csv-convert.fields.operation')"
            >
              <ElOption
                :label="t('tools.csv-convert.options.operation.csvToJson')"
                value="csvToJson"
              />
              <ElOption
                :label="t('tools.csv-convert.options.operation.jsonToCsv')"
                value="jsonToCsv"
              />
            </ElSelect>
          </label>

          <label class="field field--grow">
            <span class="field__label">{{ t('tools.csv-convert.fields.input') }}</span>
            <ElInput
              v-model="input"
              class="field__input tb-mono"
              type="textarea"
              :rows="8"
              autocomplete="off"
              :aria-label="t('tools.csv-convert.fields.input')"
            />
          </label>
        </template>

        <template #options>
          <label class="field">
            <span class="field__label">{{ t('tools.csv-convert.fields.delimiter') }}</span>
            <ElInput
              v-model="delimiter"
              class="field__input tb-mono"
              maxlength="1"
              autocomplete="off"
              :aria-label="t('tools.csv-convert.fields.delimiter')"
            />
          </label>

          <label class="field field--check">
            <ElCheckbox
              v-model="hasHeader"
              :aria-label="t('tools.csv-convert.fields.hasHeader')"
            >
              {{ t('tools.csv-convert.fields.hasHeader') }}
            </ElCheckbox>
          </label>

          <label class="field field--check">
            <ElCheckbox
              v-model="pretty"
              :aria-label="t('tools.csv-convert.fields.pretty')"
            >
              {{ t('tools.csv-convert.fields.pretty') }}
            </ElCheckbox>
          </label>

          <p class="field__hint">
            {{ t('tools.csv-convert.hints.input') }}
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
        :title="t('tools.csv-convert.results.output')"
        @retry="run"
      >
        <template #extra>
          <dl class="detail">
            <div class="detail__row">
              <dt>{{ t('tools.csv-convert.results.rowCount') }}</dt>
              <dd class="tb-mono">
                {{ data?.rowCount ?? 0 }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.csv-convert.results.columnCount') }}</dt>
              <dd class="tb-mono">
                {{ data?.columnCount ?? 0 }}
              </dd>
            </div>
            <div
              v-if="data && data.columns.length > 0"
              class="detail__row"
            >
              <dt>{{ t('tools.csv-convert.results.columns') }}</dt>
              <dd class="tb-mono">
                {{ data.columns.join(', ') }}
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

COMPONENTS["json-to-types"] = '''<script setup lang="ts">
/**
 * JSON 转类型定义工具页。
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
import { EXAMPLES } from '@/tools/data/json-to-types/example'

/** 后端 json-to-types 工具的响应。 */
interface JSONToTypesResult {
  /** 目标语言。 */
  language: string
  /** 生成的类型定义。 */
  output: string
  /** 类型数量。 */
  typeCount: number
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const input = ref('{\\n  "id": 1,\\n  "name": "Alice",\\n  "tags": ["a", "b"]\\n}')
const language = ref('typescript')
const rootName = ref('Root')

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<JSONToTypesResult>({
  slug: 'json-to-types',
})

const running = computed(() => state.value === 'loading')

/** 结果文本：生成的类型定义。 */
const resultText = computed(() => data.value?.output ?? '')

const faq = computed<FaqItem[]>(() => tm('tools.json-to-types.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  if (input.value.trim() === '') {
    return
  }
  void execute({
    input: input.value,
    language: language.value,
    rootName: rootName.value,
  })
}

// 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。
useAutoRun({
  slug: 'json-to-types',
  run,
  watch: [input, language, rootName],
  enabled: () => input.value.trim() !== '',
})

/** 重置表单与结果。 */
function reset(): void {
  input.value = '{\\n  "id": 1,\\n  "name": "Alice",\\n  "tags": ["a", "b"]\\n}'
  language.value = 'typescript'
  rootName.value = 'Root'
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  input.value = example.input
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="json-to-types"
    :name="t('tools.json-to-types.name')"
    name-en="JSON to Types"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field field--grow">
            <span class="field__label">{{ t('tools.json-to-types.fields.input') }}</span>
            <ElInput
              v-model="input"
              class="field__input tb-mono"
              type="textarea"
              :rows="10"
              autocomplete="off"
              :aria-label="t('tools.json-to-types.fields.input')"
            />
          </label>
        </template>

        <template #options>
          <label class="field">
            <span class="field__label">{{ t('tools.json-to-types.fields.language') }}</span>
            <ElSelect
              v-model="language"
              class="field__select"
              :aria-label="t('tools.json-to-types.fields.language')"
            >
              <ElOption
                :label="t('tools.json-to-types.options.language.typescript')"
                value="typescript"
              />
              <ElOption
                :label="t('tools.json-to-types.options.language.go')"
                value="go"
              />
              <ElOption
                :label="t('tools.json-to-types.options.language.rust')"
                value="rust"
              />
              <ElOption
                :label="t('tools.json-to-types.options.language.java')"
                value="java"
              />
            </ElSelect>
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.json-to-types.fields.rootName') }}</span>
            <ElInput
              v-model="rootName"
              class="field__input tb-mono"
              autocomplete="off"
              :aria-label="t('tools.json-to-types.fields.rootName')"
              @keydown.enter.prevent="run"
            />
          </label>

          <p class="field__hint">
            {{ t('tools.json-to-types.hints.input') }}
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
        :title="t('tools.json-to-types.results.output')"
        @retry="run"
      >
        <template #extra>
          <dl class="detail">
            <div class="detail__row">
              <dt>{{ t('tools.json-to-types.results.typeCount') }}</dt>
              <dd class="tb-mono">
                {{ data?.typeCount ?? 0 }}
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

COMPONENTS["mock-data"] = '''<script setup lang="ts">
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

import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
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
  return data.value.fields.map((field) => `${field.type}\\t${field.value}`).join('\\n')
})

const faq = computed<FaqItem[]>(() => tm('tools.mock-data.faq') as FaqItem[])

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
        @retry="run"
      >
        <template #extra>
          <dl class="detail">
            <div class="detail__row">
              <dt>{{ t('tools.mock-data.results.count') }}</dt>
              <dd class="tb-mono">
                {{ data?.count ?? 0 }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.mock-data.results.seed') }}</dt>
              <dd class="tb-mono">
                {{ data?.seed ?? '' }}
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

COMPONENTS["id-generator"] = '''<script setup lang="ts">
/**
 * ID 生成工具页。
 *
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
 */
import { ElCheckbox } from 'element-plus/es/components/checkbox/index'
import 'element-plus/es/components/checkbox/style/css'
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
import { EXAMPLES } from '@/tools/data/id-generator/example'

/** 后端 id-generator 工具的响应。 */
interface IDGeneratorResult {
  /** ID 格式。 */
  format: string
  /** 生成的 ID 列表。 */
  ids: string[]
  /** 生成数量。 */
  count: number
  /** ID 长度。 */
  length: number
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const format = ref('uuid')
const count = ref(5)
const length = ref(21)
const nodeId = ref(0)
const uppercase = ref(false)

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<IDGeneratorResult>({
  slug: 'id-generator',
})

const running = computed(() => state.value === 'loading')

/** 结果文本：生成的 ID 列表。 */
const resultText = computed(() => data.value?.ids.join('\\n') ?? '')

const faq = computed<FaqItem[]>(() => tm('tools.id-generator.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  void execute({
    format: format.value,
    count: count.value,
    length: length.value,
    nodeId: nodeId.value,
    uppercase: uppercase.value,
  })
}

// 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。
useAutoRun({
  slug: 'id-generator',
  run,
  watch: [format, count, length, nodeId, uppercase],
  enabled: () => true,
})

/** 重置表单与结果。 */
function reset(): void {
  format.value = 'uuid'
  count.value = 5
  length.value = 21
  nodeId.value = 0
  uppercase.value = false
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  format.value = example.format
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="id-generator"
    :name="t('tools.id-generator.name')"
    name-en="ID Generator"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.id-generator.fields.format') }}</span>
            <ElSelect
              v-model="format"
              class="field__select"
              :aria-label="t('tools.id-generator.fields.format')"
            >
              <ElOption
                :label="t('tools.id-generator.options.format.uuid')"
                value="uuid"
              />
              <ElOption
                :label="t('tools.id-generator.options.format.ulid')"
                value="ulid"
              />
              <ElOption
                :label="t('tools.id-generator.options.format.snowflake')"
                value="snowflake"
              />
              <ElOption
                :label="t('tools.id-generator.options.format.nanoid')"
                value="nanoid"
              />
            </ElSelect>
          </label>
        </template>

        <template #options>
          <label class="field">
            <span class="field__label">{{ t('tools.id-generator.fields.count') }}</span>
            <ElInputNumber
              v-model="count"
              class="field__select"
              :min="0"
              :max="1000"
              :aria-label="t('tools.id-generator.fields.count')"
            />
          </label>

          <label
            v-if="format === 'nanoid'"
            class="field"
          >
            <span class="field__label">{{ t('tools.id-generator.fields.length') }}</span>
            <ElInputNumber
              v-model="length"
              class="field__select"
              :min="0"
              :max="64"
              :aria-label="t('tools.id-generator.fields.length')"
            />
          </label>

          <label
            v-if="format === 'snowflake'"
            class="field"
          >
            <span class="field__label">{{ t('tools.id-generator.fields.nodeId') }}</span>
            <ElInputNumber
              v-model="nodeId"
              class="field__select"
              :min="0"
              :max="1023"
              :aria-label="t('tools.id-generator.fields.nodeId')"
            />
          </label>

          <label class="field field--check">
            <ElCheckbox
              v-model="uppercase"
              :aria-label="t('tools.id-generator.fields.uppercase')"
            >
              {{ t('tools.id-generator.fields.uppercase') }}
            </ElCheckbox>
          </label>

          <p class="field__hint">
            {{ t('tools.id-generator.hints.format') }}
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
        :title="t('tools.id-generator.results.ids')"
        @retry="run"
      >
        <template #extra>
          <dl class="detail">
            <div class="detail__row">
              <dt>{{ t('tools.id-generator.results.count') }}</dt>
              <dd class="tb-mono">
                {{ data?.count ?? 0 }}
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
    for slug, content in COMPONENTS.items():
        target = TOOLS / slug / "Tool.vue"
        target.write_text(content, encoding="utf-8")
        print(f"写入 {slug}/Tool.vue")


if __name__ == "__main__":
    main()
