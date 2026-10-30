<script setup lang="ts">
/**
 * 二维码工具页。
 *
 * 图片只在浏览器内存中转成 base64 后提交，**不写**任何浏览器存储。
 */
import { ElColorPicker } from 'element-plus/es/components/color-picker/index'
import 'element-plus/es/components/color-picker/style/css'
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { ElInputNumber } from 'element-plus/es/components/input-number/index'
import 'element-plus/es/components/input-number/style/css'
import { ElOption, ElSelect } from 'element-plus/es/components/select/index'
import 'element-plus/es/components/select/style/css'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import Base64FileInput, { type LoadedFile } from '@/components/tool/Base64FileInput.vue'
import DownloadButton from '@/components/tool/DownloadButton.vue'
import ResultPanel, { type ResultView } from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import StatGrid, { type StatGridItem } from '@/components/tool/StatGrid.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import ValidationHint from '@/components/tool/ValidationHint.vue'
import { useToolRun } from '@/composables/useToolRun'
import { EXAMPLES } from '@/tools/data/qr-code/example'
import type { QRECLevel, QROperation } from '@/tools/data/qr-code/example'

/** 后端 qr-code 工具的响应。 */
interface QRCodeResult {
  /** 矢量图源码（`generate`）。 */
  svg?: string
  /** 位图 base64（`generate`）。 */
  pngBase64?: string
  /** 模块边长（`generate`）。 */
  dimension?: number
  /** 模块像素边长（`generate`）。 */
  moduleSize?: number
  /** 识别出的内容（`decode`）。 */
  content?: string
  /** 图片格式（`decode`）。 */
  format?: string
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

/** 校验提示的结论。 */
interface ValidationState {
  level: 'success' | 'error'
  message: string
}

/** 待识别图片的字节上限；与服务端一致。 */
const MAX_IMAGE_BYTES = 4 * 1024 * 1024
/** 二维码内容的字节上限；与服务端一致。 */
const MAX_TEXT_BYTES = 4096
/** 十六进制颜色格式。 */
const HEX_COLOR = /^#[0-9a-fA-F]{6}$/
/** 取色器的回落色；输入非法时显示。 */
const FALLBACK_FG = '#000000'
const FALLBACK_BG = '#ffffff'

const { t, tm } = useI18n()

const operation = ref<QROperation>('generate')
const text = ref('https://example.com')
const size = ref(6)
const margin = ref(4)
const ecLevel = ref<QRECLevel>('M')
const fg = ref('')
const bg = ref('')
const imageBase64 = ref('')
const imageName = ref('')
const imageMime = ref('')
const imageBytes = ref(0)
const fileError = ref('')

const { state, data, meta, error, run: execute, reset: resetRun, bindInput, useAsInput } =
  useToolRun<QRCodeResult>({
    slug: 'qr-code',
  })

const running = computed(() => state.value === 'loading')

/** 结果文本：生成时是 SVG 源码，识别时是解出的内容。 */
const resultText = computed(() => {
  if (operation.value === 'decode') {
    return data.value?.content ?? ''
  }
  return data.value?.svg ?? ''
})

/** 位图 data URL。 */
const pngUrl = computed(() => {
  const png = data.value?.pngBase64
  return png ? `data:image/png;base64,${png}` : ''
})

/** 待识别图片的 data URL（缩略图）。 */
const imageUrl = computed(() => {
  if (imageBase64.value === '') {
    return ''
  }
  return `data:${imageMime.value || 'image/png'};base64,${imageBase64.value}`
})

/** 内容字节数。 */
const textBytes = computed(() => new TextEncoder().encode(text.value).length)

/** 颜色的校验结论；空串表示用默认配色。 */
const colorValidation = computed<ValidationState | null>(() => {
  const invalid = [fg.value, bg.value].some(
    (value) => value !== '' && !HEX_COLOR.test(value),
  )
  return invalid ? { level: 'error', message: t('tool.validation.colorInvalid') } : null
})

/** 内容的校验结论。 */
const textValidation = computed<ValidationState | null>(() => {
  if (textBytes.value <= MAX_TEXT_BYTES) {
    return null
  }
  return { level: 'error', message: t('tools.qr-code.hints.textTooLong') }
})

/** 取色器的当前值；输入非法时回落到默认色。 */
const fgPicker = computed(() => (HEX_COLOR.test(fg.value) ? fg.value : FALLBACK_FG))
const bgPicker = computed(() => (HEX_COLOR.test(bg.value) ? bg.value : FALLBACK_BG))

/** 是否可以运行。 */
const runDisabled = computed(() => {
  if (operation.value === 'decode') {
    return imageBase64.value === ''
  }
  return (
    text.value.trim() === '' ||
    textBytes.value > MAX_TEXT_BYTES ||
    colorValidation.value !== null
  )
})

/** 结果页签：生成看预览与源码，识别看内容。 */
const views = computed<ResultView[]>(() => {
  if (operation.value === 'decode') {
    return [{ id: 'content', label: t('tools.qr-code.views.content') }]
  }
  return [
    { id: 'preview', label: t('tools.qr-code.views.preview') },
    { id: 'svg', label: t('tools.qr-code.views.svg') },
  ]
})

/** 当前页签。 */
const activeView = ref('')

/** 摘要卡片。 */
const stats = computed<StatGridItem[]>(() => {
  const result = data.value
  if (!result) {
    return []
  }

  if (operation.value === 'decode') {
    const content = result.content ?? ''
    return [
      { label: t('tools.qr-code.results.format'), value: result.format ?? '' },
      {
        label: t('tools.qr-code.results.bytes'),
        value: `${new TextEncoder().encode(content).length} B`,
      },
    ]
  }

  const dimension = result.dimension ?? 0
  const moduleSize = result.moduleSize ?? 0
  return [
    { label: t('tools.qr-code.results.dimension'), value: `${dimension}` },
    { label: t('tools.qr-code.results.moduleSize'), value: `${moduleSize} px` },
    {
      label: t('tools.qr-code.results.pixels'),
      value: `${dimension * moduleSize} px`,
      hint: t('tools.qr-code.results.pixelsHint'),
    },
    { label: t('tools.qr-code.results.ecLevel'), value: ecLevel.value },
  ]
})

const faq = computed<FaqItem[]>(() => tm('tools.qr-code.faq') as FaqItem[])

/** 解析整数输入；非法时返回 undefined。 */
function parseIntOrUndefined(raw: number): number | undefined {
  return Number.isFinite(raw) ? Math.trunc(raw) : undefined
}

/** 运行当前输入。 */
function run(): void {
  if (operation.value === 'decode') {
    if (imageBase64.value === '') {
      return
    }
    void execute({ operation: 'decode', imageBase64: imageBase64.value })
    return
  }

  const options: Record<string, unknown> = { ecLevel: ecLevel.value }
  const parsedSize = parseIntOrUndefined(size.value)
  if (parsedSize !== undefined) {
    options.size = parsedSize
  }
  const parsedMargin = parseIntOrUndefined(margin.value)
  if (parsedMargin !== undefined) {
    options.margin = parsedMargin
  }
  if (fg.value !== '') {
    options.fg = fg.value
  }
  if (bg.value !== '') {
    options.bg = bg.value
  }

  void execute({ operation: 'generate', text: text.value, options })
}

/** 切换操作：清掉上一次的结果与页签。 */
function onOperationChange(): void {
  activeView.value = ''
  resetRun()
}

/** 取色器选定前景色；清空时回到默认配色。 */
function onFgChange(value: string | null): void {
  fg.value = value ?? ''
}

/** 取色器选定背景色；清空时回到默认配色。 */
function onBgChange(value: string | null): void {
  bg.value = value ?? ''
}

/** 处理读入的图片。 */
function onLoaded(file: LoadedFile): void {
  imageBase64.value = file.base64
  imageName.value = file.name
  imageMime.value = file.mime
  imageBytes.value = file.size
  fileError.value = ''
  resetRun()
}

/** 处理读入失败。 */
function onFileError(message: string): void {
  fileError.value = message
}

/** 清除已选图片。 */
function clearImage(): void {
  imageBase64.value = ''
  imageName.value = ''
  imageMime.value = ''
  imageBytes.value = 0
  fileError.value = ''
  resetRun()
}

/** 重置表单与结果。 */
function reset(): void {
  operation.value = 'generate'
  text.value = 'https://example.com'
  size.value = 6
  margin.value = 4
  ecLevel.value = 'M'
  fg.value = ''
  bg.value = ''
  clearImage()
  activeView.value = ''
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  operation.value = example.operation
  text.value = example.text
  size.value = example.size
  margin.value = example.margin
  ecLevel.value = example.ecLevel
  fg.value = example.fg
  bg.value = example.bg
  activeView.value = ''
  resetRun()
}

// 识别结果可回填到生成输入：便于「识别后再改一改重新生成」。
bindInput((value) => {
  text.value = value
  operation.value = 'generate'
  activeView.value = ''
  resetRun()
})
</script>

<template>
  <ToolPageLayout
    slug="qr-code"
    :name="t('tools.qr-code.name')"
    name-en="QR Code Generator"
    wide
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.qr-code.fields.operation') }}</span>
            <ElSelect
              v-model="operation"
              class="field__select"
              :aria-label="t('tools.qr-code.fields.operation')"
              @change="onOperationChange"
            >
              <ElOption
                :label="t('tools.qr-code.options.operation.generate')"
                value="generate"
              />
              <ElOption
                :label="t('tools.qr-code.options.operation.decode')"
                value="decode"
              />
            </ElSelect>
          </label>

          <template v-if="operation === 'generate'">
            <label class="field">
              <span class="field__label">{{ t('tools.qr-code.fields.text') }}</span>
              <ElInput
                v-model="text"
                type="textarea"
                class="field__textarea tb-mono"
                :rows="4"
                spellcheck="false"
                placeholder="https://example.com"
                :aria-label="t('tools.qr-code.fields.text')"
              />
            </label>

            <p class="field__hint">
              {{ t('tools.qr-code.hints.text') }}
              <span class="field__counter">{{ textBytes }} / {{ MAX_TEXT_BYTES }} B</span>
            </p>
            <ValidationHint :validation="textValidation" />
          </template>

          <template v-else>
            <label class="field">
              <span class="field__label">{{ t('tools.qr-code.fields.image') }}</span>
              <Base64FileInput
                accept="image/png,image/jpeg"
                :max-bytes="MAX_IMAGE_BYTES"
                @loaded="onLoaded"
                @error="onFileError"
              />
            </label>

            <ValidationHint
              v-if="fileError"
              :validation="{ level: 'error', message: fileError }"
            />

            <figure
              v-if="imageUrl"
              class="upload"
            >
              <img
                class="upload__image"
                :src="imageUrl"
                :alt="imageName"
              >
              <figcaption class="upload__caption">
                <span class="upload__name">{{ imageName }}</span>
                <span class="upload__size">{{ Math.round(imageBytes / 1024) }} KB</span>
                <button
                  type="button"
                  class="upload__clear"
                  @click="clearImage"
                >
                  {{ t('common.actions.clear') }}
                </button>
              </figcaption>
            </figure>

            <p class="field__hint">
              {{ t('tools.qr-code.hints.image') }}
            </p>
          </template>
        </template>

        <template #options>
          <template v-if="operation === 'generate'">
            <div class="field-row">
              <div class="field">
                <span class="field__label">{{ t('tools.qr-code.fields.size') }}</span>
                <ElInputNumber
                  v-model="size"
                  class="field__input tb-mono"
                  :min="1"
                  :max="40"
                  :step="1"
                  :aria-label="t('tools.qr-code.fields.size')"
                />
              </div>

              <div class="field">
                <span class="field__label">{{ t('tools.qr-code.fields.margin') }}</span>
                <ElInputNumber
                  v-model="margin"
                  class="field__input tb-mono"
                  :min="0"
                  :max="16"
                  :step="1"
                  :aria-label="t('tools.qr-code.fields.margin')"
                />
              </div>

              <div class="field">
                <span class="field__label">{{ t('tools.qr-code.fields.ecLevel') }}</span>
                <ElSelect
                  v-model="ecLevel"
                  class="field__select"
                  :aria-label="t('tools.qr-code.fields.ecLevel')"
                >
                  <ElOption
                    :label="t('tools.qr-code.options.ecLevel.L')"
                    value="L"
                  />
                  <ElOption
                    :label="t('tools.qr-code.options.ecLevel.M')"
                    value="M"
                  />
                  <ElOption
                    :label="t('tools.qr-code.options.ecLevel.Q')"
                    value="Q"
                  />
                  <ElOption
                    :label="t('tools.qr-code.options.ecLevel.H')"
                    value="H"
                  />
                </ElSelect>
              </div>
            </div>

            <p class="field__hint">
              {{ t('tools.qr-code.hints.size') }}
            </p>
            <p class="field__hint">
              {{ t('tools.qr-code.hints.margin') }}
            </p>
            <p class="field__hint">
              {{ t('tools.qr-code.hints.ecLevel') }}
            </p>

            <div class="field-row">
              <div class="field">
                <span class="field__label">{{ t('tools.qr-code.fields.fg') }}</span>
                <div class="color">
                  <ElColorPicker
                    :model-value="fgPicker"
                    color-format="hex"
                    :aria-label="t('tools.qr-code.fields.fg')"
                    @change="onFgChange"
                  />
                  <ElInput
                    v-model="fg"
                    class="field__input tb-mono"
                    autocomplete="off"
                    placeholder="#000000"
                    :aria-label="t('tools.qr-code.fields.fg')"
                  />
                </div>
              </div>

              <div class="field">
                <span class="field__label">{{ t('tools.qr-code.fields.bg') }}</span>
                <div class="color">
                  <ElColorPicker
                    :model-value="bgPicker"
                    color-format="hex"
                    :aria-label="t('tools.qr-code.fields.bg')"
                    @change="onBgChange"
                  />
                  <ElInput
                    v-model="bg"
                    class="field__input tb-mono"
                    autocomplete="off"
                    placeholder="#FFFFFF"
                    :aria-label="t('tools.qr-code.fields.bg')"
                  />
                </div>
              </div>
            </div>

            <ValidationHint :validation="colorValidation" />
            <p class="field__hint">
              {{ t('tools.qr-code.hints.fg') }}
            </p>
          </template>
        </template>

        <template #footer>
          <RunBar
            :running="running"
            :disabled="runDisabled"
            @run="run"
            @reset="reset"
            @example="fillExample(0)"
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
        :title="t('tools.qr-code.results.output')"
        :views="views"
        :use-as-input="operation === 'decode' ? useAsInput : null"
        :download-filename="operation === 'generate' ? 'qrcode.svg' : undefined"
        :download-mime="operation === 'generate' ? 'image/svg+xml' : undefined"
        @retry="run"
      >
        <template #summary>
          <StatGrid :items="stats" />
        </template>

        <template #view="{ viewId }">
          <template v-if="viewId === 'preview'">
            <figure
              v-if="pngUrl"
              class="qr"
            >
              <img
                class="qr__image"
                :src="pngUrl"
                :alt="t('tools.qr-code.results.image')"
              >
              <figcaption class="qr__caption">
                {{ t('tools.qr-code.results.image') }}
              </figcaption>
            </figure>

            <div
              v-if="pngUrl"
              class="qr__download"
            >
              <DownloadButton
                filename="qrcode.png"
                :base64="data?.pngBase64 ?? ''"
                mime="image/png"
                :label="t('tools.qr-code.results.downloadPng')"
              />
            </div>
          </template>

          <pre
            v-else-if="viewId === 'svg' && data?.svg"
            class="product"
          ><code>{{ data.svg }}</code></pre>

          <pre
            v-else-if="viewId === 'content' && data?.content !== undefined"
            class="product"
          ><code>{{ data.content }}</code></pre>
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.qr-code.hints.output') }}
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
.color {
  display: flex;
  align-items: center;
  gap: var(--tb-space-2);
}

.color :deep(.el-color-picker__trigger) {
  width: 34px;
  height: 32px;
  padding: 0;
  border-radius: var(--tb-radius-control);
}

.upload {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-1);
  margin: 0;
}

.upload__image {
  display: block;
  max-width: 100%;
  max-height: 160px;
  padding: var(--tb-space-2);
  background: var(--tb-bg-sunken);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-control);
  object-fit: contain;
}

.upload__caption {
  display: flex;
  align-items: center;
  gap: var(--tb-space-2);
  font-size: var(--tb-font-xs);
  color: var(--tb-text-tertiary);
}

.upload__name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.upload__clear {
  padding: 0;
  font: inherit;
  font-size: var(--tb-font-xs);
  color: var(--tb-primary);
  cursor: pointer;
  background: none;
  border: 0;
}

.upload__clear:hover {
  text-decoration: underline;
}

.qr {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--tb-space-2);
  margin: var(--tb-space-2) 0 0;
}

.qr__image {
  display: block;
  width: 100%;
  max-width: 20rem;
  padding: var(--tb-space-3);
  background: #ffffff;
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-control);
  image-rendering: pixelated;
}

.qr__caption {
  font-size: var(--tb-font-xs);
  color: var(--tb-text-tertiary);
}

.qr__download {
  display: flex;
  justify-content: center;
  margin-top: var(--tb-space-2);
}

.product {
  margin: 0;
  padding: var(--tb-space-3);
  overflow-x: auto;
  font-family: var(--tb-font-mono);
  font-size: var(--tb-font-sm);
  line-height: 1.6;
  color: var(--tb-text);
  white-space: pre-wrap;
  word-break: break-all;
  background: var(--tb-bg-sunken);
  border-radius: var(--tb-radius-control);
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
