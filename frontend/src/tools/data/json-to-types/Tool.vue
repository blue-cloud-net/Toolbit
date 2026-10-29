<script setup lang="ts">
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

const input = ref('{\n  "id": 1,\n  "name": "Alice",\n  "tags": ["a", "b"]\n}')
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
  input.value = '{\n  "id": 1,\n  "name": "Alice",\n  "tags": ["a", "b"]\n}'
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
