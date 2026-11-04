<script setup lang="ts">
/**
 * Protobuf 编码器工具页。
 *
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
 */
import { ElButton } from 'element-plus/es/components/button/index'
import 'element-plus/es/components/button/style/css'
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { ElOption, ElSelect } from 'element-plus/es/components/select/index'
import 'element-plus/es/components/select/style/css'
import { ElRadioButton, ElRadioGroup } from 'element-plus/es/components/radio/index'
import 'element-plus/es/components/radio/style/css'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import CodeEditor from '@/components/tool/CodeEditor.vue'
import EncodingSelect from '@/components/tool/EncodingSelect.vue'
import ProductText from '@/components/tool/ProductText.vue'
import ResultPanel, { type ResultView } from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import StatGrid, { type StatGridItem } from '@/components/tool/StatGrid.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useToolRun } from '@/composables/useToolRun'
import { EXAMPLES } from '@/tools/encoding/protobuf-encoder/example'
import FieldTable from '@/tools/encoding/protobuf-encoder/FieldTable.vue'
import {
  cloneNodes,
  createNode,
  fromJsonText,
  toJsonText,
  toPayload,
  type ProtoFieldNode,
} from '@/tools/encoding/protobuf-encoder/fields'

/** 后端 protobuf-encoder 工具的响应。 */
interface ProtobufEncodeResult {
  output: string
  encoding: string
  byteLength: number
  fieldCount: number
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

type Mode = 'wire' | 'schema'
type EditorMode = 'table' | 'json'
type OutputEncoding = 'hex' | 'base64' | 'base64url'

const { t, tm } = useI18n()

const mode = ref<Mode>('wire')
const editorMode = ref<EditorMode>('table')
const fields = ref<ProtoFieldNode[]>([])
const jsonText = ref('[]')
const jsonError = ref('')
const schema = ref('')
const messageType = ref('')
const messageText = ref('')
const outputEncoding = ref<OutputEncoding>('base64')

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
  bindInput,
  useAsInput,
} = useToolRun<ProtobufEncodeResult>({ slug: 'protobuf-encoder' })

/** 结果区的「作为新输入」：把结果回填到主输入框（仅内存，不写存储）。 */
bindInput((text) => {
  jsonText.value = text
})

const running = computed(() => state.value === 'loading')

/** 可选的输出编码。 */
const OUTPUT_ENCODINGS: readonly OutputEncoding[] = ['base64', 'hex', 'base64url']

const faq = computed<FaqItem[]>(() => tm('tools.protobuf-encoder.faq') as FaqItem[])

/** 结果摘要。 */
const stats = computed<StatGridItem[]>(() => {
  const result = data.value
  if (!result) {
    return []
  }
  return [
    { label: t('tools.protobuf-encoder.results.byteLength'), value: `${result.byteLength} B` },
    { label: t('tools.protobuf-encoder.results.fieldCount'), value: String(result.fieldCount) },
    { label: t('tools.protobuf-encoder.results.encoding'), value: result.encoding },
  ]
})

/** 结果文本。 */
const resultText = computed(() => data.value?.output ?? '')

/** 结果页签：摘要 / 编码结果。 */
const views = computed<ResultView[]>(() => [
  { id: 'summary', label: t('tools.protobuf-encoder.views.summary') },
  { id: 'output', label: t('tools.protobuf-encoder.views.output') },
])

/** 当前页签；默认停在「摘要」。 */
const activeView = ref('summary')

/** 产物文件名：随输出编码变化，避免把十六进制文本存成二进制。 */
const outputFilename = computed(() =>
  outputEncoding.value === 'hex' ? 'protobuf.hex' : 'protobuf.b64',
)

/** 追加一个空字段。 */
function addField(): void {
  fields.value = [...fields.value, createNode()]
}

/** 切换编辑形态：两种形态之间同步字段数据。 */
function changeEditor(next: EditorMode): void {
  if (next === editorMode.value) {
    return
  }
  if (next === 'json') {
    jsonText.value = toJsonText(fields.value)
    jsonError.value = ''
    editorMode.value = 'json'
    return
  }
  try {
    fields.value = fromJsonText(jsonText.value)
    jsonError.value = ''
    editorMode.value = 'table'
  } catch (error) {
    jsonError.value = error instanceof Error ? error.message : String(error)
  }
}

/** 取当前生效的字段规格；JSON 形态非法时返回 null。 */
function currentFields(): ProtoFieldNode[] | null {
  if (editorMode.value === 'table') {
    return fields.value
  }
  try {
    const parsed = fromJsonText(jsonText.value)
    jsonError.value = ''
    fields.value = parsed
    return parsed
  } catch (error) {
    jsonError.value = error instanceof Error ? error.message : String(error)
    return null
  }
}

/** 运行当前输入。 */
function run(): void {
  if (mode.value === 'wire') {
    const nodes = currentFields()
    if (!nodes) {
      return
    }
    void execute({
      mode: 'wire',
      fields: toPayload(nodes),
      outputEncoding: outputEncoding.value,
    })
    return
  }
  void execute({
    mode: 'schema',
    schema: schema.value,
    messageType: messageType.value,
    message: messageText.value,
    outputEncoding: outputEncoding.value,
  })
}

/** 重置表单与结果。 */
function reset(): void {
  mode.value = 'wire'
  editorMode.value = 'table'
  fields.value = []
  jsonText.value = '[]'
  jsonError.value = ''
  schema.value = ''
  messageType.value = ''
  messageText.value = ''
  outputEncoding.value = 'base64'
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  mode.value = example.mode
  fields.value = cloneNodes(example.fields)
  jsonText.value = toJsonText(example.fields)
  jsonError.value = ''
  schema.value = example.schema
  messageType.value = example.messageType
  messageText.value = example.message
  outputEncoding.value = example.outputEncoding
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="protobuf-encoder"
    :name="t('tools.protobuf-encoder.name')"
    name-en="Protobuf Encoder"
    wide
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.protobuf-encoder.fields.mode') }}</span>
            <ElSelect
              v-model="mode"
              class="field__select"
              :aria-label="t('tools.protobuf-encoder.fields.mode')"
            >
              <ElOption
                :label="t('tools.protobuf-encoder.options.mode.wire')"
                value="wire"
              />
              <ElOption
                :label="t('tools.protobuf-encoder.options.mode.schema')"
                value="schema"
              />
            </ElSelect>
          </label>

          <template v-if="mode === 'wire'">
            <div class="pe__switch">
              <ElRadioGroup
                :model-value="editorMode"
                @update:model-value="(value) => changeEditor(value as EditorMode)"
              >
                <ElRadioButton value="table">
                  {{ t('tools.protobuf-encoder.options.editor.table') }}
                </ElRadioButton>
                <ElRadioButton value="json">
                  {{ t('tools.protobuf-encoder.options.editor.json') }}
                </ElRadioButton>
              </ElRadioGroup>

              <ElButton
                v-if="editorMode === 'table'"
                size="small"
                @click="addField"
              >
                {{ t('tools.protobuf-encoder.actions.addField') }}
              </ElButton>
            </div>

            <FieldTable
              v-if="editorMode === 'table'"
              v-model="fields"
            />

            <label
              v-else
              class="field"
            >
              <span class="field__label">{{ t('tools.protobuf-encoder.fields.fieldsJson') }}</span>
              <CodeEditor
                v-model="jsonText"
                :label="t('tools.protobuf-encoder.fields.fieldsJson')"
                :max-rows="12"
                :show-line-numbers="false"
              />
            </label>

            <p
              v-if="jsonError"
              class="input-error"
            >
              {{ t('tools.protobuf-encoder.errors.invalidJson') }}：{{ jsonError }}
            </p>

            <p class="field__hint">
              {{ t('tools.protobuf-encoder.hints.fields') }}
            </p>
          </template>

          <template v-else>
            <label class="field">
              <span class="field__label">{{ t('tools.protobuf-encoder.fields.schema') }}</span>
              <CodeEditor
                v-model="schema"
                :label="t('tools.protobuf-encoder.fields.schema')"
                :max-rows="10"
                :show-line-numbers="false"
              />
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.protobuf-encoder.fields.messageType') }}</span>
              <ElInput
                v-model="messageType"
                class="field__input tb-mono"
                :aria-label="t('tools.protobuf-encoder.fields.messageType')"
              />
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.protobuf-encoder.fields.message') }}</span>
              <CodeEditor
                v-model="messageText"
                :label="t('tools.protobuf-encoder.fields.message')"
                :max-rows="10"
                :show-line-numbers="false"
              />
            </label>

            <p class="field__hint">
              {{ t('tools.protobuf-encoder.hints.schema') }}
            </p>
          </template>
        </template>

        <template #options>
          <label class="field">
            <span class="field__label">{{ t('tools.protobuf-encoder.fields.outputEncoding') }}</span>
            <EncodingSelect
              v-model="outputEncoding"
              :options="OUTPUT_ENCODINGS"
            />
          </label>
        </template>

        <template #footer>
          <RunBar
            :running="running"
            :disabled="mode === 'wire' && fields.length === 0 && editorMode === 'table'"
            :examples="EXAMPLES"
            @run="run"
            @reset="reset"
            @fill="fillExample"
          />
        </template>
      </ToolPanel>
    </template>

    <template #result>
      <ResultPanel
        v-model:active-view="activeView"
        :state="state"
        :text="resultText"
        :meta="meta"
        :error="error"
        :title="t('tools.protobuf-encoder.results.output')"
        :views="views"
        :use-as-input="useAsInput"
        :download-filename="outputFilename"
        @retry="run"
      >
        <template #view="{ viewId }">
          <StatGrid
            v-if="viewId === 'summary'"
            :items="stats"
          />

          <ProductText
            v-else-if="viewId === 'output' && data"
            :label="t('tools.protobuf-encoder.views.output')"
            :content="data.output"
            :filename="outputFilename"
          />
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.protobuf-encoder.description') }}
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

<style scoped>
.pe__switch {
  display: flex;
  gap: var(--tb-space-2);
  align-items: center;
  justify-content: space-between;
}
</style>
