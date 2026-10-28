<script setup lang="ts">
/**
 * SM4 文件加解密工具页。
 *
 * 输入是上传的文件；结果是一个可下载的文件。
 * 文件只存在于组件局部状态，离开页面即丢弃，
 * **不写**任何浏览器存储。
 */
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import ErrorAlert from '@/components/tool/ErrorAlert.vue'
import FileDropZone from '@/components/tool/FileDropZone.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useFileToolRun } from '@/composables/useFileToolRun'
import { DEFAULT_EXAMPLE } from '@/tools/crypto/sm4-file-crypt/example'

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const operation = ref<string>('encrypt')
const key = ref(DEFAULT_EXAMPLE.key)
const keyEncoding = ref<string>('hex')
const nonce = ref('')
const nonceEncoding = ref<string>('hex')
const chunkSize = ref<string>('')
const file = ref<File | null>(null)

const {
  state,
  result,
  meta,
  error,
  phase,
  percent,
  run: execute,
  reset: resetRun,
  download,
} = useFileToolRun({ slug: 'sm4-file-crypt' })

const running = computed(() => state.value === 'loading')

/** 结果文件名。 */
const resultName = computed(() => result.value?.filename ?? '')

/** 结果大小（可读）。 */
const resultSize = computed(() => {
  const size = result.value?.size ?? 0
  if (size < 1024) {
    return `${size} B`
  }
  if (size < 1024 * 1024) {
    return `${(size / 1024).toFixed(1)} KB`
  }
  return `${(size / (1024 * 1024)).toFixed(1)} MB`
})

/** 上传进度文案。 */
const progressText = computed(() => {
  if (phase.value === 'uploading') {
    return percent.value === null
      ? t('tools.sm4-file-crypt.results.uploading')
      : t('tools.sm4-file-crypt.results.uploadingPercent', { percent: percent.value })
  }
  if (phase.value === 'processing') {
    return t('tools.sm4-file-crypt.results.processing')
  }
  return ''
})

const faq = computed<FaqItem[]>(() => tm('tools.sm4-file-crypt.faq') as FaqItem[])

/** 构造字节输入的表单字段值。 */
function byteField(data: string, encoding: string): string {
  if (data === '') {
    return ''
  }
  return JSON.stringify({ data, encoding })
}

/** 运行当前输入。 */
function run(): void {
  if (!file.value) {
    return
  }
  const form = new FormData()
  form.append('operation', operation.value)
  form.append('mode', 'GCM')
  form.append('key', byteField(key.value, keyEncoding.value))
  const iv = byteField(nonce.value, nonceEncoding.value)
  if (iv !== '') {
    form.append('iv', iv)
  }
  if (chunkSize.value !== '') {
    form.append('chunkSize', chunkSize.value)
  }
  // ⚠️ 文件部分**必须**是最后一个 part：
  // 服务端按顺序读表单，读到文件后就无法回头取后面的字段。
  form.append('file', file.value, file.value.name)
  void execute(form)
}

/** 重置表单与结果。 */
function reset(): void {
  operation.value = 'encrypt'
  key.value = DEFAULT_EXAMPLE.key
  keyEncoding.value = 'hex'
  nonce.value = ''
  nonceEncoding.value = 'hex'
  chunkSize.value = ''
  file.value = null
  resetRun()
}

/** 填入示例。 */
function fillExample(): void {
  operation.value = DEFAULT_EXAMPLE.operation
  key.value = DEFAULT_EXAMPLE.key
  keyEncoding.value = 'hex'
  nonce.value = ''
  nonceEncoding.value = 'hex'
  chunkSize.value = ''
  file.value = null
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="sm4-file-crypt"
    :name="t('tools.sm4-file-crypt.name')"
    name-en="SM4 File Encryption"
    wide
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <FileDropZone
            v-model="file"
            :label="t('tools.sm4-file-crypt.fields.file')"
          />
        </template>

        <template #options>
          <div class="field-row">
            <label class="field">
              <span class="field__label">{{ t('tools.sm4-file-crypt.fields.operation') }}</span>
              <select
                v-model="operation"
                class="field__select"
              >
                <option value="encrypt">{{ t('tools.sm4-file-crypt.options.operation.encrypt') }}</option>
                <option value="decrypt">{{ t('tools.sm4-file-crypt.options.operation.decrypt') }}</option>
              </select>
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.sm4-file-crypt.fields.keyEncoding') }}</span>
              <select
                v-model="keyEncoding"
                class="field__select"
              >
                <option value="hex">{{ t('tool.encoding.hex') }}</option>
                <option value="base64">{{ t('tool.encoding.base64') }}</option>
              </select>
            </label>
          </div>

          <label class="field">
            <span class="field__label">{{ t('tools.sm4-file-crypt.fields.key') }}</span>
            <input
              v-model="key"
              type="text"
              class="field__input tb-mono"
              autocomplete="off"
              :aria-label="t('tools.sm4-file-crypt.fields.key')"
            >
          </label>

          <div class="field-row">
            <label class="field">
              <span class="field__label">{{ t('tools.sm4-file-crypt.fields.nonceEncoding') }}</span>
              <select
                v-model="nonceEncoding"
                class="field__select"
              >
                <option value="hex">{{ t('tool.encoding.hex') }}</option>
                <option value="base64">{{ t('tool.encoding.base64') }}</option>
              </select>
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.sm4-file-crypt.fields.chunkSize') }}</span>
              <input
                v-model="chunkSize"
                type="text"
                class="field__input tb-mono"
                inputmode="numeric"
                autocomplete="off"
                :placeholder="t('tools.sm4-file-crypt.options.chunkSize.default')"
                :aria-label="t('tools.sm4-file-crypt.fields.chunkSize')"
              >
            </label>
          </div>

          <label class="field">
            <span class="field__label">{{ t('tools.sm4-file-crypt.fields.nonce') }}</span>
            <input
              v-model="nonce"
              type="text"
              class="field__input tb-mono"
              autocomplete="off"
              :aria-label="t('tools.sm4-file-crypt.fields.nonce')"
            >
          </label>

          <p class="field__hint">
            {{ t('tools.sm4-file-crypt.hints.nonce') }}
          </p>
        </template>

        <template #footer>
          <RunBar
            :running="running"
            :disabled="!file"
            @run="run"
            @reset="reset"
            @example="fillExample"
          />
        </template>
      </ToolPanel>
    </template>

    <template #result>
      <section
        class="result"
        :aria-label="t('tool.sections.result')"
      >
        <header class="result__head">
          <h2 class="result__title">
            {{ t('tool.sections.result') }}
          </h2>
          <span
            v-if="meta?.durationMs"
            class="result__duration"
          >{{ t('tool.result.duration', { ms: meta.durationMs }) }}</span>
        </header>

        <p
          v-if="state === 'idle'"
          class="result__placeholder"
        >
          {{ t('tool.result.idle') }}
          <span class="result__hint">{{ t('tool.result.idleHint') }}</span>
        </p>

        <p
          v-else-if="state === 'loading'"
          class="result__placeholder"
        >
          {{ progressText || t('tool.result.loading') }}
        </p>

        <ErrorAlert
          v-else-if="state === 'error'"
          :error="error ?? null"
          @retry="run"
        />

        <template v-else>
          <ul
            v-if="(meta?.warnings ?? []).length > 0"
            class="result__warnings"
          >
            <li
              v-for="(w, i) in meta?.warnings ?? []"
              :key="i"
              class="result__warning"
            >
              {{ w.code }}
            </li>
          </ul>

          <div class="result__file">
            <p class="result__filename">
              {{ resultName }}
            </p>
            <p class="result__filesize">
              {{ resultSize }}
            </p>
            <button
              class="result__download"
              type="button"
              @click="download"
            >
              {{ t('tools.sm4-file-crypt.results.download') }}
            </button>
          </div>
        </template>
      </section>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.sm4-file-crypt.description') }}
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
.result {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-2);
  padding: var(--tb-space-4);
  background: var(--tb-bg-elevated);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-panel);
}

.result__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--tb-space-2);
}

.result__title {
  margin: 0;
  font-size: var(--tb-font-lg);
}

.result__duration {
  font-size: var(--tb-font-sm);
  color: var(--tb-text-secondary);
}

.result__placeholder {
  margin: 0;
  color: var(--tb-text-secondary);
}

.result__hint {
  display: block;
  font-size: var(--tb-font-sm);
  color: var(--tb-text-tertiary);
}

.result__warnings {
  margin: 0;
  padding-left: var(--tb-space-4);
  font-size: var(--tb-font-sm);
  color: var(--tb-warning);
}

.result__file {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-1);
  align-items: flex-start;
  padding: var(--tb-space-3);
  background: var(--tb-bg-sunken);
  border-radius: var(--tb-radius-card);
}

.result__filename {
  margin: 0;
  font-family: var(--tb-font-mono);
  font-size: var(--tb-font-sm);
  color: var(--tb-text);
  word-break: break-all;
}

.result__filesize {
  margin: 0;
  font-size: var(--tb-font-sm);
  color: var(--tb-text-secondary);
}

.result__download {
  padding: var(--tb-space-1) var(--tb-space-3);
  font-family: inherit;
  font-size: var(--tb-font-sm);
  color: var(--tb-primary-contrast);
  cursor: pointer;
  background: var(--tb-primary);
  border: none;
  border-radius: var(--tb-radius-control);
}

.result__download:hover {
  background: var(--tb-primary-hover);
}
</style>
