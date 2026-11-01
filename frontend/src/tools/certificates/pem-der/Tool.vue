<script setup lang="ts">
/**
 * PEM/DER 转换工具页。
 *
 * 材料内容只存在于组件局部状态；离开页面即丢弃，
 * **不写**任何浏览器存储（AGENTS.md §4.11）。
 */
import { ElOption, ElSelect } from 'element-plus/es/components/select/index'
import 'element-plus/es/components/select/style/css'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import MaterialInput from '@/components/tool/MaterialInput.vue'
import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useToolRun } from '@/composables/useToolRun'
import { DEFAULT_EXAMPLE } from '@/tools/certificates/pem-der/example'

/** 后端 pem-der 工具的响应。 */
interface PEMDerResult {
  output: { data: string; format: string; encoding: string }
  labels: string[]
  blockCount: number
}

/** 类型过滤值。 */
const TYPES = ['any', 'cert', 'key', 'csr', 'crl']

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const input = ref('')
const inputFormat = ref('auto')
const type = ref('any')
const to = ref('der')

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
  bindInput,
  useAsInput,
} = useToolRun<PEMDerResult>({ slug: 'pem-der' })

/** 结果区的「作为新输入」：把结果回填到主输入框（仅内存，不写存储）。 */
bindInput((text) => {
  input.value = text
})

const running = computed(() => state.value === 'loading')

/** 结果文本。 */
const resultText = computed(() => {
  const result = data.value
  if (!result) {
    return ''
  }
  return [
    `${t('tools.pem-der.results.format')}: ${result.output.format}`,
    `${t('tools.pem-der.results.encoding')}: ${result.output.encoding}`,
    `${t('tools.pem-der.results.blockCount')}: ${result.blockCount}`,
    `${t('tools.pem-der.results.labels')}: ${result.labels.join(', ')}`,
    '',
    result.output.data,
  ].join('\n')
})

const faq = computed<FaqItem[]>(() => tm('tools.pem-der.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  void execute({
    input: { data: input.value, format: inputFormat.value },
    type: type.value,
    to: to.value,
  })
}

/** 重置表单与结果。 */
function reset(): void {
  input.value = ''
  inputFormat.value = 'auto'
  type.value = 'any'
  to.value = 'der'
  resetRun()
}

/** 填入示例。 */
function fillExample(): void {
  input.value = DEFAULT_EXAMPLE.data
  inputFormat.value = DEFAULT_EXAMPLE.format
  type.value = 'any'
  to.value = 'der'
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="pem-der"
    :name="t('tools.pem-der.name')"
    name-en="PEM & DER Converter"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <MaterialInput
            v-model="input"
            v-model:format="inputFormat"
            :label="t('tools.pem-der.fields.input')"
          />
        </template>

        <template #options>
          <label class="field">
            <span class="field__label">{{ t('tools.pem-der.fields.type') }}</span>
            <ElSelect
              v-model="type"
              class="field__select"
              :aria-label="t('tools.pem-der.fields.type')"
            >
              <ElOption
                v-for="name in TYPES"
                :key="name"
                :label="t(`tools.pem-der.options.types.${name}`)"
                :value="name"
              />
            </ElSelect>
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.pem-der.fields.to') }}</span>
            <ElSelect
              v-model="to"
              class="field__select"
              :aria-label="t('tools.pem-der.fields.to')"
            >
              <ElOption
                :label="t('tools.pem-der.options.to.der')"
                value="der"
              />
              <ElOption
                :label="t('tools.pem-der.options.to.pem')"
                value="pem"
              />
            </ElSelect>
          </label>
        </template>

        <template #footer>
          <RunBar
            :running="running"
            :disabled="input === ''"
            @run="run"
            @reset="reset"
            @example="fillExample"
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
        :title="t('tools.pem-der.results.output')"
        :use-as-input="useAsInput"
        @retry="run"
      />
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.pem-der.description') }}
      </p>
    </template>

    <template #faq>
      <dl class="faq">
        <template
          v-for="item in faq"
          :key="item.q"
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
