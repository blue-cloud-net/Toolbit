<script setup lang="ts">
/**
 * Protobuf 解码器工具页。
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
import { computed, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'

import CodeEditor from '@/components/tool/CodeEditor.vue'
import CodeViewer from '@/components/tool/CodeViewer.vue'
import DataTable, { type DataTableColumn, type DataTableRow } from '@/components/tool/DataTable.vue'
import EncodingSelect from '@/components/tool/EncodingSelect.vue'
import FileDropZone from '@/components/tool/FileDropZone.vue'
import JsonViewer, { type JsonNode } from '@/components/tool/JsonViewer.vue'
import ResultPanel, { type ResultView } from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import StatGrid, { type StatGridItem } from '@/components/tool/StatGrid.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useToolRun } from '@/composables/useToolRun'
import { EXAMPLES } from '@/tools/encoding/protobuf-decoder/example'

/** 后端 protobuf-decoder 的字段结构。 */
interface ProtoField {
  path: string
  offset: number
  fieldNumber: number
  name?: string
  wireType: string
  declaredType?: string
  value: string
  length?: number
  text: boolean
  message: boolean
  hex?: string
  float?: number
  children?: ProtoField[]
}

/** 后端 protobuf-decoder 工具的响应。 */
interface ProtobufDecodeResult {
  mode: string
  messageType: string
  byteLength: number
  fieldCount: number
  depth: number
  maxDepth: number
  fields: ProtoField[]
  json: string
  unknownFieldCount: number
  hexDump: string
  truncated: boolean
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

type Mode = 'wire' | 'schema'
type TransferEncoding = 'hex' | 'base64' | 'base64url'

const { t, tm } = useI18n()

const mode = ref<Mode>('wire')
const input = ref('')
const inputEncoding = ref<TransferEncoding>('hex')
const schema = ref('')
const messageType = ref('')
const maxDepth = ref(32)
const maxFields = ref(10000)
const inferStrings = ref(true)
const includeHexDump = ref(false)
const file = ref<File | null>(null)

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
} = useToolRun<ProtobufDecodeResult>({ slug: 'protobuf-decoder' })

const running = computed(() => state.value === 'loading')

/** 可选的传输编码。 */
const TRANSFER_ENCODINGS: readonly TransferEncoding[] = ['hex', 'base64', 'base64url']

const faq = computed<FaqItem[]>(() => tm('tools.protobuf-decoder.faq') as FaqItem[])

/** 摘要统计卡片。 */
const stats = computed<StatGridItem[]>(() => {
  const result = data.value
  if (!result) {
    return []
  }
  const items: StatGridItem[] = [
    { label: t('tools.protobuf-decoder.results.byteLength'), value: `${result.byteLength} B` },
    { label: t('tools.protobuf-decoder.results.fieldCount'), value: String(result.fieldCount) },
    { label: t('tools.protobuf-decoder.results.depth'), value: String(result.depth) },
  ]
  if (result.mode === 'schema') {
    items.push({
      label: t('tools.protobuf-decoder.results.unknownFieldCount'),
      value: String(result.unknownFieldCount),
    })
  }
  if (result.truncated) {
    items.push({
      label: t('tools.protobuf-decoder.results.truncated'),
      value: t('tools.protobuf-decoder.results.truncatedYes'),
      level: 'warning',
    })
  }
  return items
})

/** 结果页签。 */
const views = computed<ResultView[]>(() => [
  { id: 'summary', label: t('tools.protobuf-decoder.views.summary') },
  { id: 'fields', label: t('tools.protobuf-decoder.views.fields') },
  { id: 'json', label: t('tools.protobuf-decoder.views.json') },
  { id: 'hex', label: t('tools.protobuf-decoder.views.hex') },
])

/** 当前页签。 */
const activeView = ref('summary')

/** 字段表的列定义。 */
const fieldColumns = computed<DataTableColumn[]>(() => {
  const columns: DataTableColumn[] = [
    { key: 'path', label: t('tools.protobuf-decoder.results.path') },
    { key: 'fieldNumber', label: t('tools.protobuf-decoder.results.fieldNumber') },
  ]
  if (data.value?.mode === 'schema') {
    columns.push({ key: 'name', label: t('tools.protobuf-decoder.results.name') })
    columns.push({ key: 'declaredType', label: t('tools.protobuf-decoder.results.declaredType') })
  }
  columns.push({ key: 'wireType', label: t('tools.protobuf-decoder.results.wireType') })
  columns.push({ key: 'value', label: t('tools.protobuf-decoder.results.value') })
  return columns
})

/** 字段表的行数据；嵌套字段按层级缩进。 */
const fieldRows = computed<DataTableRow[]>(() => {
  const rows: DataTableRow[] = []
  const walk = (fields: ProtoField[], depth: number): void => {
    for (const field of fields) {
      rows.push({
        key: field.path,
        cells: {
          path: `${'  '.repeat(depth)}${field.path}`,
          fieldNumber: String(field.fieldNumber),
          name: field.name ?? '',
          declaredType: field.declaredType ?? '',
          wireType: field.wireType,
          value: protoFieldText(field),
        },
      })
      if (field.children && field.children.length > 0) {
        walk(field.children, depth + 1)
      }
    }
  }
  walk(data.value?.fields ?? [], 0)
  return rows
})

/** 字段值的展示文本。 */
function protoFieldText(field: ProtoField): string {
  if (field.message) {
    return t('tools.protobuf-decoder.results.nestedMessage')
  }
  if (field.hex) {
    return field.hex
  }
  if (field.float !== undefined) {
    return `${field.value} (${field.float})`
  }
  return field.value
}

/** 结构化 JSON 节点；解析失败时为空。 */
const jsonNodes = computed<JsonNode[]>(() => {
  const text = data.value?.json
  if (!text) {
    return []
  }
  try {
    return [{ key: 'root', value: JSON.parse(text) }]
  } catch {
    return []
  }
})

/** 读取文件并转成 Base64。 */
function readAsBase64(target: File): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader()
    reader.onerror = () => reject(new Error('cannot read file'))
    reader.onload = () => {
      const result = reader.result
      if (typeof result !== 'string') {
        reject(new Error('cannot read file'))
        return
      }
      const comma = result.indexOf(',')
      resolve(comma >= 0 ? result.slice(comma + 1) : result)
    }
    reader.readAsDataURL(target)
  })
}

/** 导入文件后把内容填进输入框。 */
watch(file, (target) => {
  if (!target) {
    return
  }
  void readAsBase64(target)
    .then((encoded) => {
      input.value = encoded
      inputEncoding.value = 'base64'
    })
    .catch(() => {
      input.value = ''
    })
})

/** 运行当前输入。 */
function run(): void {
  void execute({
    mode: mode.value,
    input: { data: input.value, encoding: inputEncoding.value },
    schema: schema.value,
    messageType: messageType.value,
    maxDepth: maxDepth.value,
    maxFields: maxFields.value,
    inferStrings: inferStrings.value,
    includeHexDump: includeHexDump.value,
  })
}

/** 重置表单与结果。 */
function reset(): void {
  mode.value = 'wire'
  input.value = ''
  inputEncoding.value = 'hex'
  schema.value = ''
  messageType.value = ''
  maxDepth.value = 32
  maxFields.value = 10000
  inferStrings.value = true
  includeHexDump.value = false
  file.value = null
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  mode.value = example.mode
  input.value = example.input
  inputEncoding.value = example.inputEncoding
  schema.value = example.schema
  messageType.value = example.messageType
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="protobuf-decoder"
    :name="t('tools.protobuf-decoder.name')"
    name-en="Protobuf Decoder"
    wide
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.protobuf-decoder.fields.input') }}</span>
            <ElInput
              v-model="input"
              type="textarea"
              class="field__textarea tb-mono"
              :rows="6"
              spellcheck="false"
              :aria-label="t('tools.protobuf-decoder.fields.input')"
            />
          </label>

          <div class="field-row">
            <label class="field">
              <span class="field__label">{{ t('tools.protobuf-decoder.fields.inputEncoding') }}</span>
              <EncodingSelect
                v-model="inputEncoding"
                :options="TRANSFER_ENCODINGS"
              />
            </label>
          </div>

          <FileDropZone
            v-model="file"
            :label="t('tools.protobuf-decoder.fields.file')"
          />

          <p class="field__hint">
            {{ t('tools.protobuf-decoder.hints.input') }}
          </p>
        </template>

        <template #options>
          <label class="field">
            <span class="field__label">{{ t('tools.protobuf-decoder.fields.mode') }}</span>
            <ElSelect
              v-model="mode"
              class="field__select"
              :aria-label="t('tools.protobuf-decoder.fields.mode')"
            >
              <ElOption
                :label="t('tools.protobuf-decoder.options.mode.wire')"
                value="wire"
              />
              <ElOption
                :label="t('tools.protobuf-decoder.options.mode.schema')"
                value="schema"
              />
            </ElSelect>
          </label>

          <label
            v-if="mode === 'schema'"
            class="field"
          >
            <span class="field__label">{{ t('tools.protobuf-decoder.fields.schema') }}</span>
            <CodeEditor
              v-model="schema"
              :label="t('tools.protobuf-decoder.fields.schema')"
              :max-rows="10"
              :show-line-numbers="false"
            />
          </label>

          <label
            v-if="mode === 'schema'"
            class="field"
          >
            <span class="field__label">{{ t('tools.protobuf-decoder.fields.messageType') }}</span>
            <ElInput
              v-model="messageType"
              class="field__input tb-mono"
              :aria-label="t('tools.protobuf-decoder.fields.messageType')"
            />
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.protobuf-decoder.fields.maxDepth') }}</span>
            <ElInputNumber
              v-model="maxDepth"
              class="field__number"
              :min="0"
              :max="128"
            />
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.protobuf-decoder.fields.maxFields') }}</span>
            <ElInputNumber
              v-model="maxFields"
              class="field__number"
              :min="1"
              :max="100000"
            />
          </label>

          <label class="field field--check">
            <ElCheckbox
              v-model="inferStrings"
              :aria-label="t('tools.protobuf-decoder.fields.inferStrings')"
            >
              {{ t('tools.protobuf-decoder.fields.inferStrings') }}
            </ElCheckbox>
          </label>

          <label class="field field--check">
            <ElCheckbox
              v-model="includeHexDump"
              :aria-label="t('tools.protobuf-decoder.fields.includeHexDump')"
            >
              {{ t('tools.protobuf-decoder.fields.includeHexDump') }}
            </ElCheckbox>
          </label>

          <p class="field__hint">
            {{ t('tools.protobuf-decoder.hints.mode') }}
          </p>
        </template>

        <template #footer>
          <RunBar
            :running="running"
            :disabled="input === ''"
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
        :meta="meta"
        :error="error"
        :title="t('tools.protobuf-decoder.results.title')"
        :views="views"
        @retry="run"
      >
        <template #view="{ viewId }">
          <StatGrid
            v-if="viewId === 'summary'"
            :items="stats"
          />
          <DataTable
            v-else-if="viewId === 'fields'"
            :columns="fieldColumns"
            :rows="fieldRows"
            :copy-columns="['value']"
            :empty="t('tools.protobuf-decoder.results.empty')"
          />
          <JsonViewer
            v-else-if="viewId === 'json' && jsonNodes.length > 0"
            :nodes="jsonNodes"
          />
          <pre
            v-else-if="viewId === 'json'"
            class="result__output"
          ><code>{{ data?.json ?? '' }}</code></pre>
          <CodeViewer
            v-else
            :text="data?.hexDump ?? ''"
            :empty="t('tools.protobuf-decoder.results.emptyHex')"
          />
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.protobuf-decoder.description') }}
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
