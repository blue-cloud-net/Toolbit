<script setup lang="ts">
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
const resultText = computed(() => data.value?.ids.join('\n') ?? '')

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
