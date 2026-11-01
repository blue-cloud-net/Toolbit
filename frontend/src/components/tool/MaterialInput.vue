<script setup lang="ts">
/**
 * 材料输入框（通用）。
 *
 * 供**所有以 PEM 文本承载的材料**共用：X.509 证书、CSR、CRL、P7B、PFX……
 * 它们的输入形态完全一样——一段可能含多个 PEM 块的文本 + 一个格式标记，
 * 因此共用本组件，**不要**各工具页另写一份。
 *
 * 输入区有**两个互斥入口**，一屏只渲染一个面板：
 *
 *   - **上传文件**：一个大拖放区（拖入或点「选择文件」）；
 *   - **粘贴文本**：一个纯文本框，直接粘贴。
 *
 * 入口用分段控件（`ElRadioGroup` + `ElRadioButton`）切换，右侧是「格式」下拉。
 * 面板之间**没有叠加、没有浮层、没有隐藏开关**——这是本组件的核心约束：
 * 叠加式布局必须靠「点哪儿都收起」这类隐式规则维持，容易出现状态不闭环。
 *
 * 导入时会**嗅探**文件形态（PEM / DER / base64 / 十六进制）并自动切换「格式」，
 * 免去用户手工选格式。导入成功后自动切到「粘贴文本」面板，让用户看得见内容。
 *
 * 材料内容只存在于组件局部状态；离开页面即丢弃，
 * **不写**任何浏览器存储（AGENTS.md §4.11）。
 */
import { ElButton } from 'element-plus/es/components/button/index'
import 'element-plus/es/components/button/style/css'
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { ElOption, ElSelect } from 'element-plus/es/components/select/index'
import 'element-plus/es/components/select/style/css'
import { ElRadioButton, ElRadioGroup } from 'element-plus/es/components/radio/index'
import 'element-plus/es/components/radio/style/css'
import { ElUpload, type UploadFile } from 'element-plus/es/components/upload/index'
import 'element-plus/es/components/upload/style/css'
import { computed, nextTick, ref, watch } from 'vue'
import { useI18n } from 'vue-i18n'

import AppIcon from '@/components/common/AppIcon.vue'
import { useCatalogStore } from '@/stores/catalog'

import { decodeUtf8, detectMaterialFile } from './materialFile'

const value = defineModel<string>({ required: true })
const format = defineModel<string>('format', { required: true })

/** 输入面板。 */
type Panel = 'upload' | 'text'

/** 已导入的文件。 */
interface ImportedFile {
  /** 文件名。 */
  name: string
  /** 字节数。 */
  size: number
  /** 转换后的文本内容。 */
  content: string
  /** 嗅探出的格式（pem / der / base64 / hex）。 */
  detected: string
}

/** 默认的请求体上限（字节）：4 MB。 */
const DEFAULT_MAX_BYTES = 4 * 1024 * 1024

const props = withDefaults(
  defineProps<{
    /** 可选的格式集合。 */
    formats?: readonly string[]
    /** 无障碍标签。 */
    label?: string
    /** 是否禁用。 */
    disabled?: boolean
    /** 文本框行数。 */
    rows?: number
    /** 是否允许从文件导入。 */
    allowFiles?: boolean
    /** 导入时追加到已有内容（false = 替换）。 */
    append?: boolean
    /** 请求体上限（字节）；与工具目录里登记的限额取较小者。 */
    maxBytes?: number
    /** 工具 slug：用于取该工具在目录里登记的请求体上限。 */
    slug?: string
    /** 文本框占位文案。 */
    placeholder?: string
    /** 文件选择框的 accept 属性；缺省不限类型。 */
    accept?: string
  }>(),
  {
    formats: () => ['auto', 'pem', 'der', 'base64', 'hex'],
    label: undefined,
    disabled: false,
    rows: 8,
    allowFiles: true,
    append: false,
    maxBytes: DEFAULT_MAX_BYTES,
    slug: undefined,
    placeholder: undefined,
    accept: undefined,
  },
)

const { t } = useI18n()
const catalog = useCatalogStore()

/** 当前输入面板。 */
const panel = ref<Panel>('upload')

/** 导入错误；就地展示，不用弹窗（与 FileTextInput 一致）。 */
const error = ref('')

/** 已导入的文件清单。 */
const imported = ref<ImportedFile[]>([])

/** 文本框是否获得焦点（用于区分「用户删空」与「外部清空」）。 */
const focused = ref(false)

/** 是否正有文件被拖到输入区上。 */
const dragActive = ref(false)

/** 文本框实例；切到「粘贴文本」时聚焦它。 */
const textInput = ref<InstanceType<typeof ElInput> | null>(null)

const FORMAT_LABELS: Record<string, string> = {
  auto: 'tool.key.format.auto',
  pem: 'tool.key.format.pem',
  der: 'tool.key.format.der',
  base64: 'tool.key.format.base64',
  hex: 'tool.key.format.hex',
}

/** 格式名优先取词条，缺失时显示原值。 */
function formatLabel(value: string): string {
  const key = FORMAT_LABELS[value]
  if (!key) {
    return value
  }
  const translated = t(key)
  return translated === key ? value : translated
}

/** 输入里的 PEM 块数；非 PEM 输入返回 0。 */
const blockCount = computed(() => {
  if (!value.value.includes('-----BEGIN ')) {
    return 0
  }
  return value.value.split('-----BEGIN ').length - 1
})

/** 块数提示；没有块时给出格式识别提示。 */
const hint = computed(() => {
  if (blockCount.value > 0) {
    return t('tool.material.blocks', { count: blockCount.value })
  }
  return t('tool.material.hint')
})

/**
 * 上传区标题。
 *
 * 字段名就是材料名（如「证书」「证书签名请求」），所以直接用它，
 * 让提示文案随工具页走，调用方不必再传一个 material。
 */
const dropTitle = computed(() => {
  if (dragActive.value) {
    return t('tool.material.import.dropActive')
  }
  return props.label
    ? t('tool.material.import.dropNamed', { material: props.label })
    : t('tool.material.import.drop')
})

/**
 * 生效的体积上限。
 *
 * 取「显式传入」与「工具目录登记值」的**较小者**：
 * 导入后的文本要作为请求体发出去，超过服务端限额会被 413 拒绝，
 * 因此前端必须先拦一道。
 */
const effectiveMaxBytes = computed(() => {
  const limit = props.slug ? catalog.getTool(props.slug)?.maxBodyBytes : undefined
  if (typeof limit === 'number' && limit > 0) {
    return Math.min(limit, props.maxBytes)
  }
  return props.maxBytes
})

/** 可读的文件大小。 */
function formatSize(bytes: number): string {
  if (bytes < 1024) {
    return `${bytes} B`
  }
  if (bytes < 1024 * 1024) {
    return `${(bytes / 1024).toFixed(1)} KB`
  }
  return `${(bytes / 1024 / 1024).toFixed(1)} MB`
}

/**
 * 嗅探文件内容并转成可提交的文本。
 *
 * 判定逻辑在 `materialFile.ts`（纯函数，有单测覆盖）；
 * 这里只负责把 File 读成字节。
 */
async function sniff(file: File): Promise<ImportedFile> {
  const buffer = await file.arrayBuffer()
  const bytes = new Uint8Array(buffer)
  const detected = detectMaterialFile(bytes, decodeUtf8(bytes))
  if (!detected) {
    throw new Error('not-material')
  }
  return {
    name: file.name,
    size: file.size,
    content: detected.content,
    detected: detected.detected,
  }
}

/** 把导入的文件写进输入框与格式选择。 */
function applyImported(files: ImportedFile[]): void {
  imported.value = files
  value.value = files.map((item) => item.content).join('\n')
  const last = files[files.length - 1]
  if (last) {
    // 嗅探出的格式可能不在该工具的格式集合里（如只收 pem/base64/hex 的工具）；
    // 这时回落到「自动识别」：后端会自己按 PEM → base64 → hex 试。
    format.value = props.formats.includes(last.detected) ? last.detected : 'auto'
  }
}

/**
 * 面板切换。
 *
 * - 切到「上传文件」：**清空**文本、文件标签与格式，回到干净的初始态；
 * - 切到「粘贴文本」：把光标放进文本框。
 */
function onPanelChange(raw: string | number | boolean | undefined): void {
  if (props.disabled) {
    return
  }
  const next: Panel = raw === 'text' ? 'text' : 'upload'
  panel.value = next
  if (next === 'upload') {
    error.value = ''
    imported.value = []
    value.value = ''
    format.value = 'auto'
    return
  }
  void nextTick(() => textInput.value?.focus())
}

/**
 * 输入变化时的维护。
 *
 * 内容被外部改写（工具页「重置」或手工编辑）→ 文件清单已不对应内容，丢掉。
 * 「清空后回上传面板」不放在这里：value 不变时 watch 不会触发，
 * 改由文本框的 `focusout` 兜底（见 `onTextBlur`）。
 */
watch(value, (next) => {
  if (imported.value.length === 0) {
    return
  }
  const joined = imported.value.map((item) => item.content).join('\n')
  if (next !== joined) {
    imported.value = []
  }
})

/**
 * 文本框失焦。
 *
 * 焦点不在文本框、且里面没有文本 → 回到「上传文件」面板。
 * 用户在文本框里删空时仍有焦点，**不**跳走，避免输入到一半被抢走。
 */
function onTextBlur(event: FocusEvent): void {
  const next = event.relatedTarget as HTMLElement | null
  if (next && event.currentTarget instanceof HTMLElement && event.currentTarget.contains(next)) {
    return
  }
  if (value.value === '' && imported.value.length === 0) {
    error.value = ''
    panel.value = 'upload'
  }
}

/** 读取并导入一组文件。 */
async function importFiles(files: File[]): Promise<void> {
  error.value = ''
  // 非追加模式只取第一个文件（与文件选择器的单选行为一致）。
  const picked = props.append ? files : files.slice(0, 1)
  const sniffed: ImportedFile[] = []
  for (const file of picked) {
    try {
      sniffed.push(await sniff(file))
    } catch {
      error.value = t('tool.material.import.notMaterial')
      return
    }
  }
  if (sniffed.length === 0) {
    return
  }
  const next = props.append ? [...imported.value, ...sniffed] : sniffed
  const bodyBytes = next.reduce((sum, item) => sum + item.content.length, 0)
  if (bodyBytes > effectiveMaxBytes.value) {
    error.value = t('tool.material.import.tooLarge', {
      max: Math.round(effectiveMaxBytes.value / 1024),
    })
    return
  }
  applyImported(next)
  // 导入成功 → 切到「粘贴文本」面板，让用户看得见内容。
  panel.value = 'text'
}

/** 导入队列：串行执行，保证多文件时的顺序与选择顺序一致。 */
let importChain: Promise<void> = Promise.resolve()

function enqueue(files: File[]): void {
  importChain = importChain.then(() => importFiles(files))
}

/** `ElUpload` 选中文件后读取。 */
function onChange(upload: UploadFile): void {
  if (props.disabled || !upload.raw) {
    return
  }
  enqueue([upload.raw])
}

/** 判断事件是否来自文件拖放（而不是拖选文本）。 */
function isFileDrag(event: DragEvent): boolean {
  return event.dataTransfer?.types.includes('Files') === true
}

/** 拖放深度计数器：dragleave 会在子元素之间冒泡，用计数避免高亮闪烁。 */
let dragDepth = 0

function onDragEnter(event: DragEvent): void {
  if (props.disabled || !props.allowFiles || !isFileDrag(event)) {
    return
  }
  dragDepth += 1
  dragActive.value = true
}

function onDragOver(event: DragEvent): void {
  if (props.disabled || !props.allowFiles || !isFileDrag(event)) {
    return
  }
  // 必须 preventDefault，否则浏览器不会派发 drop。
  event.preventDefault()
  if (event.dataTransfer) {
    event.dataTransfer.dropEffect = 'copy'
  }
  dragActive.value = true
}

function onDragLeave(event: DragEvent): void {
  if (props.disabled || !props.allowFiles || !isFileDrag(event)) {
    return
  }
  dragDepth = Math.max(0, dragDepth - 1)
  if (dragDepth === 0) {
    dragActive.value = false
  }
}

function onDrop(event: DragEvent): void {
  if (props.disabled || !props.allowFiles) {
    return
  }
  const files = event.dataTransfer?.files
  if (!files || files.length === 0) {
    // 拖的是文本：交回浏览器默认行为（插入文本框）。
    return
  }
  event.preventDefault()
  dragDepth = 0
  dragActive.value = false
  enqueue(Array.from(files))
}

/** 移除某个已导入的文件，并重排输入内容。 */
function removeFile(index: number): void {
  const next = imported.value.filter((_, i) => i !== index)
  if (next.length === 0) {
    clearFiles()
    return
  }
  applyImported(next)
}

/** 清除全部已导入文件与输入内容，回到「上传文件」面板。 */
function clearFiles(): void {
  error.value = ''
  imported.value = []
  value.value = ''
  format.value = 'auto'
  panel.value = 'upload'
}
</script>

<template>
  <div
    class="material-input"
    :class="{ 'material-input--drag': dragActive }"
    @dragenter="onDragEnter"
    @dragover="onDragOver"
    @dragleave="onDragLeave"
    @drop="onDrop"
  >
    <div class="material-input__head">
      <span class="field__label">{{ label ?? t('tool.material.fields.material') }}</span>
      <ElRadioGroup
        :model-value="panel"
        size="small"
        class="material-input__switch"
        :disabled="props.disabled"
        @update:model-value="onPanelChange"
      >
        <ElRadioButton value="upload">
          {{ t('tool.material.panel.upload') }}
        </ElRadioButton>
        <ElRadioButton value="text">
          {{ t('tool.material.panel.text') }}
        </ElRadioButton>
      </ElRadioGroup>
      <ElSelect
        v-model="format"
        class="material-input__format"
        :aria-label="t('tool.material.fields.format')"
        :disabled="props.disabled"
      >
        <ElOption
          v-for="f in props.formats"
          :key="f"
          :label="formatLabel(f)"
          :value="f"
        />
      </ElSelect>
    </div>

    <!-- 上传面板 -->
    <div
      v-if="panel === 'upload'"
      class="material-input__upload"
    >
      <ElUpload
        class="material-input__drop"
        drag
        action="#"
        :auto-upload="false"
        :show-file-list="false"
        :multiple="props.append"
        :accept="props.accept"
        :disabled="props.disabled"
        :on-change="onChange"
      >
        <span class="material-input__badge">
          <AppIcon
            name="upload"
            :size="22"
          />
        </span>
        <p class="material-input__drop-title">
          {{ dropTitle }}
        </p>
        <p class="material-input__drop-sub">
          {{ props.append ? t('tool.material.import.acceptMulti') : t('tool.material.import.accept') }}
        </p>
        <ElButton
          size="small"
          type="primary"
          plain
          :disabled="props.disabled"
        >
          {{ t('tool.file.choose') }}
        </ElButton>
      </ElUpload>
    </div>

    <!-- 文本面板 -->
    <div
      v-else
      class="material-input__text"
      @focusout="onTextBlur"
    >
      <ul
        v-if="imported.length > 0"
        class="material-input__files"
      >
        <li
          v-for="(item, index) in imported"
          :key="`${item.name}-${index}`"
          class="material-input__file"
        >
          <AppIcon
            name="document"
            :size="12"
          />
          <span class="material-input__file-name">{{ item.name }}</span>
          <span class="material-input__file-meta">{{ formatSize(item.size) }}</span>
          <button
            type="button"
            class="material-input__file-remove"
            :aria-label="`${t('tool.material.import.remove')} ${item.name}`"
            :disabled="props.disabled"
            @click="removeFile(index)"
          >
            <AppIcon
              name="close"
              :size="11"
            />
          </button>
        </li>
      </ul>

      <ElInput
        ref="textInput"
        v-model="value"
        type="textarea"
        class="tb-mono"
        :rows="props.rows"
        :placeholder="props.placeholder ?? t('tool.material.import.placeholder')"
        spellcheck="false"
        autocomplete="off"
        :aria-label="label ?? t('tool.material.fields.material')"
        :disabled="props.disabled"
        @focus="focused = true"
        @blur="focused = false"
      />
    </div>

    <div class="material-input__foot">
      <p class="field__hint">
        {{ hint }}
      </p>
      <button
        v-if="value !== ''"
        type="button"
        class="material-input__toggle"
        :disabled="props.disabled"
        @click="clearFiles"
      >
        {{ t('tool.material.import.clear') }}
      </button>
    </div>

    <p
      v-if="error"
      class="material-input__error"
      role="alert"
    >
      {{ error }}
    </p>
  </div>
</template>

<style scoped>
.material-input {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-2);
}

/* 标题行：字段名 + 入口分段控件 + 格式。 */
.material-input__head {
  display: flex;
  align-items: center;
  gap: var(--tb-space-3);
}

.material-input__switch {
  flex: 0 0 auto;
}

.material-input__format {
  width: 150px;
  margin-left: auto;
}

/* 上传面板：一个虚线大拖放区。 */
.material-input__upload {
  border: 1px dashed var(--tb-border-strong);
  border-radius: var(--tb-radius-card);
  background: var(--tb-bg-elevated);
  transition: border-color var(--tb-transition);
}

.material-input--drag .material-input__upload {
  border-color: var(--tb-primary);
  background: var(--tb-primary-soft);
}

.material-input__drop {
  display: block;
  width: 100%;
}

.material-input__drop :deep(.el-upload-dragger) {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: var(--tb-space-1);
  min-height: 168px;
  padding: var(--tb-space-6) var(--tb-space-4);
  border: 0;
  background: transparent;
}

.material-input__drop :deep(.el-upload-dragger:hover) {
  background: var(--tb-bg-hover);
}

.material-input--drag .material-input__drop :deep(.el-upload-dragger) {
  background: transparent;
}

/* 图标底座：浅色圆底 + 主色图标。 */
.material-input__badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 44px;
  height: 44px;
  margin-bottom: var(--tb-space-2);
  border-radius: var(--tb-radius-pill);
  background: var(--tb-primary-soft);
  color: var(--tb-primary);
}

.material-input__drop-title {
  margin: 0;
  font-size: var(--tb-font-sm);
  font-weight: 500;
  color: var(--tb-text);
}

.material-input__drop-sub {
  margin: 0;
  font-size: var(--tb-font-2xs);
  color: var(--tb-text-tertiary);
}

.material-input__drop :deep(.el-button) {
  margin-top: var(--tb-space-2);
}

/* 文本面板。 */
.material-input__text {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-2);
}

/* 已导入文件：胶囊标签。 */
.material-input__files {
  display: flex;
  flex-wrap: wrap;
  gap: var(--tb-space-1);
  margin: 0;
  padding: 0;
  list-style: none;
}

.material-input__file {
  display: inline-flex;
  align-items: center;
  gap: var(--tb-space-1);
  padding: 2px var(--tb-space-1) 2px var(--tb-space-2);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-pill);
  background: var(--tb-bg-sunken);
  color: var(--tb-text-tertiary);
  font-size: var(--tb-font-2xs);
}

.material-input__file-name {
  max-width: 200px;
  overflow: hidden;
  font-family: var(--tb-font-mono);
  color: var(--tb-text);
  text-overflow: ellipsis;
  white-space: nowrap;
}

.material-input__file-meta {
  white-space: nowrap;
}

.material-input__file-remove {
  display: inline-flex;
  padding: 0;
  border: 0;
  background: none;
  color: var(--tb-text-tertiary);
  cursor: pointer;
}

.material-input__file-remove:hover {
  color: var(--tb-danger);
}

/* 文字按钮：清除内容。 */
.material-input__toggle {
  padding: 0;
  border: 0;
  background: none;
  color: var(--tb-primary);
  font: inherit;
  font-size: var(--tb-font-xs);
  white-space: nowrap;
  cursor: pointer;
}

.material-input__toggle:hover {
  text-decoration: underline;
}

.material-input__toggle:disabled {
  color: var(--tb-text-tertiary);
  cursor: not-allowed;
  text-decoration: none;
}

.material-input__foot {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--tb-space-2);
}

.material-input__error {
  margin: 0;
  font-size: var(--tb-font-2xs);
  color: var(--tb-danger);
}
</style>
