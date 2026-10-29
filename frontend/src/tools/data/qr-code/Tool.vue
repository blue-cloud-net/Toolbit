<script setup lang="ts">
/**
 * 二维码工具页。
 *
 * 图片只在浏览器内存中转成 base64 后提交，**不写**任何浏览器存储。
 */
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import Base64FileInput, { type LoadedFile } from '@/components/tool/Base64FileInput.vue'
import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
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

/** 待识别图片的字节上限；与服务端一致。 */
const MAX_IMAGE_BYTES = 4 * 1024 * 1024

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
const fileError = ref('')

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<QRCodeResult>({
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

const faq = computed<FaqItem[]>(() => tm('tools.qr-code.faq') as FaqItem[])

/** 解析整数输入；非法时返回 undefined。 */
function parseIntOrUndefined(raw: number): number | undefined {
  return Number.isFinite(raw) ? Math.trunc(raw) : undefined
}

/** 运行当前输入。 */
function run(): void {
  if (operation.value === 'decode') {
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

/** 处理读入的图片。 */
function onLoaded(file: LoadedFile): void {
  imageBase64.value = file.base64
  imageName.value = file.name
  fileError.value = ''
  resetRun()
}

/** 处理读入失败。 */
function onFileError(message: string): void {
  fileError.value = message
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
  imageBase64.value = ''
  imageName.value = ''
  fileError.value = ''
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
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="qr-code"
    :name="t('tools.qr-code.name')"
    name-en="QR Code Generator"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.qr-code.fields.operation') }}</span>
            <select
              v-model="operation"
              class="field__select"
            >
              <option value="generate">
                {{ t('tools.qr-code.options.operation.generate') }}
              </option>
              <option value="decode">
                {{ t('tools.qr-code.options.operation.decode') }}
              </option>
            </select>
          </label>

          <template v-if="operation === 'generate'">
            <label class="field">
              <span class="field__label">{{ t('tools.qr-code.fields.text') }}</span>
              <textarea
                v-model="text"
                class="field__textarea tb-mono"
                rows="4"
                spellcheck="false"
                placeholder="https://example.com"
                :aria-label="t('tools.qr-code.fields.text')"
              />
            </label>

            <p class="field__hint">
              {{ t('tools.qr-code.hints.text') }}
            </p>
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

            <p
              v-if="fileError"
              class="field__hint"
            >
              {{ fileError }}
            </p>
            <p
              v-if="imageName"
              class="field__hint"
            >
              {{ t('tools.qr-code.results.file') }}: {{ imageName }}
            </p>
            <p class="field__hint">
              {{ t('tools.qr-code.hints.image') }}
            </p>
          </template>
        </template>

        <template #options>
          <template v-if="operation === 'generate'">
            <div class="field-row">
              <label class="field">
                <span class="field__label">{{ t('tools.qr-code.fields.size') }}</span>
                <input
                  v-model.number="size"
                  type="number"
                  class="field__input tb-mono"
                  min="1"
                  max="40"
                  step="1"
                  :aria-label="t('tools.qr-code.fields.size')"
                >
              </label>

              <label class="field">
                <span class="field__label">{{ t('tools.qr-code.fields.margin') }}</span>
                <input
                  v-model.number="margin"
                  type="number"
                  class="field__input tb-mono"
                  min="0"
                  max="16"
                  step="1"
                  :aria-label="t('tools.qr-code.fields.margin')"
                >
              </label>

              <label class="field">
                <span class="field__label">{{ t('tools.qr-code.fields.ecLevel') }}</span>
                <select
                  v-model="ecLevel"
                  class="field__select"
                >
                  <option value="L">
                    {{ t('tools.qr-code.options.ecLevel.L') }}
                  </option>
                  <option value="M">
                    {{ t('tools.qr-code.options.ecLevel.M') }}
                  </option>
                  <option value="Q">
                    {{ t('tools.qr-code.options.ecLevel.Q') }}
                  </option>
                  <option value="H">
                    {{ t('tools.qr-code.options.ecLevel.H') }}
                  </option>
                </select>
              </label>
            </div>

            <div class="field-row">
              <label class="field">
                <span class="field__label">{{ t('tools.qr-code.fields.fg') }}</span>
                <input
                  v-model="fg"
                  type="text"
                  class="field__input tb-mono"
                  autocomplete="off"
                  placeholder="#000000"
                  :aria-label="t('tools.qr-code.fields.fg')"
                >
              </label>

              <label class="field">
                <span class="field__label">{{ t('tools.qr-code.fields.bg') }}</span>
                <input
                  v-model="bg"
                  type="text"
                  class="field__input tb-mono"
                  autocomplete="off"
                  placeholder="#FFFFFF"
                  :aria-label="t('tools.qr-code.fields.bg')"
                >
              </label>
            </div>
            <p class="field__hint">
              {{ t('tools.qr-code.hints.fg') }}
            </p>
          </template>
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
        :title="t('tools.qr-code.results.output')"
        @retry="run"
      >
        <template
          v-if="data"
          #extra
        >
          <template v-if="operation === 'generate'">
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

            <dl class="detail">
              <div class="detail__row">
                <dt>
                  {{ t('tools.qr-code.results.dimension') }}
                </dt>
                <dd class="tb-mono">
                  {{ data?.dimension }}
                </dd>
              </div>
              <div class="detail__row">
                <dt>
                  {{ t('tools.qr-code.results.moduleSize') }}
                </dt>
                <dd class="tb-mono">
                  {{ data?.moduleSize }}
                </dd>
              </div>
            </dl>
          </template>

          <dl
            v-else
            class="detail"
          >
            <div class="detail__row">
              <dt>
                {{ t('tools.qr-code.results.content') }}
              </dt>
              <dd class="tb-mono">
                {{ data?.content }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>
                {{ t('tools.qr-code.results.format') }}
              </dt>
              <dd class="tb-mono">
                {{ data?.format }}
              </dd>
            </div>
          </dl>
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
.qr {
  margin: var(--tb-space-3) 0 0;
}

.qr__image {
  display: block;
  max-width: 100%;
  padding: var(--tb-space-2);
  background: var(--tb-bg-sunken);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-control);
  image-rendering: pixelated;
}

.qr__caption {
  margin-top: var(--tb-space-1);
  font-size: var(--tb-font-xs);
  color: var(--tb-text-tertiary);
}

.detail {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-1);
  margin: var(--tb-space-3) 0 0;
}

.detail__row {
  display: flex;
  gap: var(--tb-space-2);
  font-size: var(--tb-font-sm);
}

.detail__row dt {
  min-width: 96px;
  color: var(--tb-text-tertiary);
}

.detail__row dd {
  margin: 0;
  word-break: break-all;
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
