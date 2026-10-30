<script setup lang="ts">
/**
 * EML 邮件解析工具页。
 *
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
 */
import { ElInputNumber } from 'element-plus/es/components/input-number/index'
import 'element-plus/es/components/input-number/style/css'
import { ElOption, ElSelect } from 'element-plus/es/components/select/index'
import 'element-plus/es/components/select/style/css'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import CodeEditor from '@/components/tool/CodeEditor.vue'
import Base64FileInput, { type LoadedFile } from '@/components/tool/Base64FileInput.vue'
import DataTable, { type DataTableColumn, type DataTableRow } from '@/components/tool/DataTable.vue'
import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useToolRun } from '@/composables/useToolRun'
import { EXAMPLES } from '@/tools/documents/eml-parser/example'

/** 一个头字段。 */
interface EMLHeader {
  /** 字段名。 */
  name: string
  /** 字段值。 */
  value: string
}

/** 一个正文分部。 */
interface EMLBody {
  /** 内容类型。 */
  contentType: string
  /** 字符编码。 */
  charset: string
  /** 正文文本。 */
  text: string
  /** HTML 正文（仅 `text/html` 分部）。 */
  html?: string
}

/** 一个附件。 */
interface EMLAttachment {
  /** 文件名。 */
  filename: string
  /** 媒体类型。 */
  mime: string
  /** 字节数。 */
  size: number
  /** 附件内容（base64）；超限时不返回。 */
  rawBase64?: string
  /** 内容因超限被丢弃。 */
  truncated: boolean
}

/** 后端 eml-parser 工具的响应。 */
interface EMLParserResult {
  /** 原始头字段。 */
  headers: EMLHeader[]
  /** 头字段映射（小写键）。 */
  headerMap: Record<string, string>
  /** 正文分部。 */
  bodies: EMLBody[]
  /** 附件。 */
  attachments: EMLAttachment[]
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

/** 待读入邮件源文的字节上限。 */
const MAX_SOURCE_BYTES = 16 * 1024 * 1024

const { t, tm } = useI18n()

const input = ref(EXAMPLES[0]?.text ?? '')
const encoding = ref('utf-8')
const maxAttachmentBytes = ref(0)
const fileError = ref('')

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<EMLParserResult>({
  slug: 'eml-parser',
})

const running = computed(() => state.value === 'loading')

/** 结果文本：按出现顺序列出头字段。 */
const resultText = computed(() =>
  (data.value?.headers ?? []).map((header) => `${header.name}: ${header.value}`).join('\n'),
)

const bodies = computed<EMLBody[]>(() => data.value?.bodies ?? [])
const attachments = computed<EMLAttachment[]>(() => data.value?.attachments ?? [])

/** 头字段表格列定义。 */
const headerColumns = computed<DataTableColumn[]>(() => [
  { key: 'name', label: t('tools.eml-parser.results.name'), mono: true },
  { key: 'value', label: t('tools.eml-parser.results.value') },
])

/** 头字段表格行；保持邮件中的原始顺序。 */
const headerRows = computed<DataTableRow[]>(() =>
  (data.value?.headers ?? []).map((header, index) => ({
    key: `${header.name}-${index}`,
    cells: { name: header.name, value: header.value },
  })),
)

const faq = computed<FaqItem[]>(() => tm('tools.eml-parser.faq') as FaqItem[])

/** 处理读入的邮件文件。 */
function onLoaded(file: LoadedFile): void {
  try {
    const binary = atob(file.base64)
    const bytes = Uint8Array.from(binary, (char) => char.charCodeAt(0))
    input.value = new TextDecoder('utf-8').decode(bytes)
    fileError.value = ''
  } catch {
    fileError.value = t('tool.dropzone.readFailed')
  }
  resetRun()
}

/** 处理读入失败。 */
function onFileError(message: string): void {
  fileError.value = message
}

/** 解析整数输入；非法时返回 undefined。 */
function parseIntOrUndefined(raw: number): number | undefined {
  return Number.isFinite(raw) ? Math.trunc(raw) : undefined
}

/** 运行当前输入。 */
function run(): void {
  void execute({
    input: { text: input.value, encoding: encoding.value },
    maxAttachmentBytes: parseIntOrUndefined(maxAttachmentBytes.value) ?? 0,
  })
}

/** 重置表单与结果。 */
function reset(): void {
  input.value = EXAMPLES[0]?.text ?? ''
  encoding.value = 'utf-8'
  maxAttachmentBytes.value = 0
  fileError.value = ''
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  input.value = example.text
  encoding.value = example.encoding
  maxAttachmentBytes.value = example.maxAttachmentBytes
  fileError.value = ''
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="eml-parser"
    :name="t('tools.eml-parser.name')"
    name-en="EML Parser"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.eml-parser.fields.input') }}</span>
            <CodeEditor
              v-model="input"
              :label="t('tools.eml-parser.fields.input')"
              :max-rows="18"
              @run="run"
            />
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.eml-parser.fields.file') }}</span>
            <Base64FileInput
              accept=".eml,message/rfc822,text/plain"
              :max-bytes="MAX_SOURCE_BYTES"
              @loaded="onLoaded"
              @error="onFileError"
            />
          </label>

          <p
            v-if="fileError"
            class="field__hint"
          >
            {{ fileError }}
          </p>
          <p class="field__hint">
            {{ t('tools.eml-parser.hints.input') }}
          </p>
        </template>

        <template #options>
          <div class="field-row">
            <label class="field">
              <span class="field__label">{{ t('tools.eml-parser.fields.encoding') }}</span>
              <ElSelect
                v-model="encoding"
                class="field__select"
                :aria-label="t('tools.eml-parser.fields.encoding')"
              >
                <ElOption
                  :label="t('tools.eml-parser.options.encoding.utf-8')"
                  value="utf-8"
                />
                <ElOption
                  :label="t('tools.eml-parser.options.encoding.gbk')"
                  value="gbk"
                />
                <ElOption
                  :label="t('tools.eml-parser.options.encoding.gb2312')"
                  value="gb2312"
                />
                <ElOption
                  :label="t('tools.eml-parser.options.encoding.big5')"
                  value="big5"
                />
                <ElOption
                  :label="t('tools.eml-parser.options.encoding.iso-8859-1')"
                  value="iso-8859-1"
                />
              </ElSelect>
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.eml-parser.fields.maxAttachmentBytes') }}</span>
              <ElInputNumber
                v-model="maxAttachmentBytes"
                class="field__input tb-mono"
                :min="0"
                :max="52428800"
                :step="1"
                :aria-label="t('tools.eml-parser.fields.maxAttachmentBytes')"
              />
            </label>
          </div>

          <p class="field__hint">
            {{ t('tools.eml-parser.hints.maxAttachmentBytes') }}
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
        :title="t('tools.eml-parser.results.output')"
        hide-text
        @retry="run"
      >
        <template
          v-if="data"
          #extra
        >
          <DataTable
            class="table"
            :columns="headerColumns"
            :rows="headerRows"
            :copy-column="'value'"
            :empty="t('tools.eml-parser.results.headersEmpty')"
          />

          <section
            v-if="bodies.length > 0"
            class="group"
          >
            <h3 class="group__title">
              {{ t('tools.eml-parser.results.bodies') }}
            </h3>
            <article
              v-for="(body, index) in bodies"
              :key="index"
              class="part"
            >
              <div class="part__head">
                <span class="part__type tb-mono">{{ body.contentType }}</span>
                <span class="part__charset tb-mono">{{ body.charset || '—' }}</span>
              </div>
              <pre class="part__body tb-mono">{{ body.text }}</pre>
            </article>
          </section>

          <section
            v-if="attachments.length > 0"
            class="group"
          >
            <h3 class="group__title">
              {{ t('tools.eml-parser.results.attachments') }}
            </h3>
            <ul class="attachments">
              <li
                v-for="(attachment, index) in attachments"
                :key="index"
                class="attachment"
              >
                <span class="attachment__name">{{ attachment.filename || '—' }}</span>
                <span class="attachment__mime tb-mono">{{ attachment.mime }}</span>
                <span class="attachment__size tb-mono">{{ attachment.size }} B</span>
                <span
                  v-if="attachment.truncated"
                  class="attachment__flag"
                >
                  {{ t('tools.eml-parser.results.truncated') }}
                </span>
              </li>
            </ul>
          </section>

          <p
            v-if="bodies.length === 0 && attachments.length === 0"
            class="empty"
          >
            {{ t('tools.eml-parser.results.empty') }}
          </p>
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.eml-parser.hints.output') }}
      </p>
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

.group {
  margin-top: var(--tb-space-3);
}

.group__title {
  margin: 0 0 var(--tb-space-2);
  font-size: var(--tb-font-sm);
  font-weight: 600;
  color: var(--tb-text-secondary);
}

.part {
  margin-bottom: var(--tb-space-2);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-control);
  overflow: hidden;
}

.part__head {
  display: flex;
  flex-wrap: wrap;
  gap: var(--tb-space-2);
  padding: var(--tb-space-1) var(--tb-space-3);
  font-size: var(--tb-font-xs);
  color: var(--tb-text-tertiary);
  background: var(--tb-bg-sunken);
}

.part__body {
  margin: 0;
  padding: var(--tb-space-3);
  overflow-x: auto;
  font-size: var(--tb-font-sm);
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-all;
}

.attachments {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-1);
  margin: 0;
  padding: 0;
  list-style: none;
  font-size: var(--tb-font-sm);
}

.attachment {
  display: flex;
  flex-wrap: wrap;
  gap: var(--tb-space-2);
  padding: var(--tb-space-1) var(--tb-space-3);
  background: var(--tb-bg-sunken);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-control);
}

.attachment__name {
  color: var(--tb-text);
}

.attachment__mime,
.attachment__size {
  color: var(--tb-text-tertiary);
}

.attachment__flag {
  color: var(--tb-warning);
}

.empty {
  margin: var(--tb-space-3) 0 0;
  font-size: var(--tb-font-sm);
  color: var(--tb-text-tertiary);
}

.prose {
  margin: 0;
  font-size: var(--tb-font-sm);
  color: var(--tb-text-secondary);
}

.faq {
  margin: 0;
  font-size: var(--tb-font-sm);
}

.faq__q {
  margin-top: var(--tb-space-2);
  font-weight: 600;
  color: var(--tb-text);
}

.faq__a {
  margin: var(--tb-space-1) 0 0;
  color: var(--tb-text-secondary);
}
</style>
