#!/usr/bin/env python3
"""生成 v0.8.0 批次新增工具的 Tool.vue（文件类）。

文件类工具统一用 FileDropZone 读文件、在前端转 Base64，
再以 JSON 文本提交给后端；前端**不**做任何密码学运算。
"""
from __future__ import annotations

import pathlib

ROOT = pathlib.Path(__file__).resolve().parents[1]
TOOLS = ROOT / "frontend" / "src" / "tools"

COMPONENTS: dict[str, str] = {}

COMPONENTS["data/archive-inspect"] = '''<script setup lang="ts">
/**
 * 归档内容查看工具页。
 *
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
 */
import { ElCheckbox } from 'element-plus/es/components/checkbox/index'
import 'element-plus/es/components/checkbox/style/css'
import { ElOption, ElSelect } from 'element-plus/es/components/select/index'
import 'element-plus/es/components/select/style/css'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import FileDropZone from '@/components/tool/FileDropZone.vue'
import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useToolRun } from '@/composables/useToolRun'
import { EXAMPLES } from '@/tools/data/archive-inspect/example'

/** 一个归档条目。 */
interface ArchiveEntry {
  /** 条目名。 */
  name: string
  /** 未压缩大小。 */
  size: number
  /** 压缩后大小。 */
  compressedSize: number
  /** 是否是目录。 */
  isDir: boolean
  /** 修改时间。 */
  modified: string
  /** 内容预览。 */
  contentPreview: string
}

/** 后端 archive-inspect 工具的响应。 */
interface ArchiveInspectResult {
  /** 归档格式。 */
  format: string
  /** 条目数。 */
  entryCount: number
  /** 解压总大小。 */
  totalSize: number
  /** 归档大小。 */
  compressedSize: number
  /** 条目清单。 */
  entries: ArchiveEntry[]
  /** 是否已截断。 */
  truncated: boolean
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const file = ref<File | null>(null)
const format = ref('')
const includeContent = ref(false)

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<ArchiveInspectResult>({
  slug: 'archive-inspect',
})

const running = computed(() => state.value === 'loading')

/** 结果文本：条目清单。 */
const resultText = computed(() => {
  if (!data.value) {
    return ''
  }
  return data.value.entries
    .map((entry) => `${entry.isDir ? 'd' : '-'} ${entry.size}\\t${entry.name}`)
    .join('\\n')
})

const faq = computed<FaqItem[]>(() => tm('tools.archive-inspect.faq') as FaqItem[])

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
      // 去掉 data URL 前缀，只保留 Base64 正文。
      const comma = result.indexOf(',')
      resolve(comma >= 0 ? result.slice(comma + 1) : result)
    }
    reader.readAsDataURL(target)
  })
}

/** 运行当前输入。 */
async function run(): Promise<void> {
  if (!file.value) {
    return
  }
  const encoded = await readAsBase64(file.value)
  await execute({
    data: encoded,
    format: format.value,
    includeContent: includeContent.value,
  })
}

/** 重置表单与结果。 */
function reset(): void {
  file.value = null
  format.value = ''
  includeContent.value = false
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  format.value = example.data
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="archive-inspect"
    :name="t('tools.archive-inspect.name')"
    name-en="Archive Inspector"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <FileDropZone
            v-model="file"
            :label="t('tools.archive-inspect.fields.data')"
          />
        </template>

        <template #options>
          <label class="field">
            <span class="field__label">{{ t('tools.archive-inspect.fields.format') }}</span>
            <ElSelect
              v-model="format"
              class="field__select"
              clearable
              :aria-label="t('tools.archive-inspect.fields.format')"
            >
              <ElOption
                :label="t('tools.archive-inspect.options.format.zip')"
                value="zip"
              />
              <ElOption
                :label="t('tools.archive-inspect.options.format.tar')"
                value="tar"
              />
              <ElOption
                :label="t('tools.archive-inspect.options.format.tarGz')"
                value="tarGz"
              />
              <ElOption
                :label="t('tools.archive-inspect.options.format.gzip')"
                value="gzip"
              />
            </ElSelect>
          </label>

          <label class="field field--check">
            <ElCheckbox
              v-model="includeContent"
              :aria-label="t('tools.archive-inspect.fields.includeContent')"
            >
              {{ t('tools.archive-inspect.fields.includeContent') }}
            </ElCheckbox>
          </label>

          <p class="field__hint">
            {{ t('tools.archive-inspect.hints.data') }}
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
        :title="t('tools.archive-inspect.results.entries')"
        @retry="run"
      >
        <template #extra>
          <dl class="detail">
            <div class="detail__row">
              <dt>{{ t('tools.archive-inspect.results.entryCount') }}</dt>
              <dd class="tb-mono">
                {{ data?.entryCount ?? 0 }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.archive-inspect.results.totalSize') }}</dt>
              <dd class="tb-mono">
                {{ data?.totalSize ?? 0 }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.archive-inspect.results.compressedSize') }}</dt>
              <dd class="tb-mono">
                {{ data?.compressedSize ?? 0 }}
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
'''

COMPONENTS["data/file-analyze"] = '''<script setup lang="ts">
/**
 * 文件类型识别工具页。
 *
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
 */
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import FileDropZone from '@/components/tool/FileDropZone.vue'
import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useToolRun } from '@/composables/useToolRun'
import { EXAMPLES } from '@/tools/data/file-analyze/example'

/** 后端 file-analyze 工具的响应。 */
interface FileAnalyzeResult {
  /** 文件名。 */
  fileName: string
  /** 字节数。 */
  size: number
  /** 媒体类型。 */
  mimeType: string
  /** 扩展名。 */
  extension: string
  /** 类型说明。 */
  description: string
  /** 是否文本类。 */
  isTextual: boolean
  /** 文件头十六进制。 */
  magicHex: string
  /** 摘要。 */
  digest: string
  /** 摘要算法。 */
  digestAlgorithm: string
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const file = ref<File | null>(null)
const fileName = ref('')
const digestAlgorithm = ref('SHA256')

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<FileAnalyzeResult>({
  slug: 'file-analyze',
})

const running = computed(() => state.value === 'loading')

/** 结果文本：摘要。 */
const resultText = computed(() => data.value?.digest ?? '')

const faq = computed<FaqItem[]>(() => tm('tools.file-analyze.faq') as FaqItem[])

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

/** 运行当前输入。 */
async function run(): Promise<void> {
  if (!file.value) {
    return
  }
  const encoded = await readAsBase64(file.value)
  await execute({
    data: encoded,
    fileName: fileName.value || file.value.name,
    digestAlgorithm: digestAlgorithm.value,
  })
}

/** 重置表单与结果。 */
function reset(): void {
  file.value = null
  fileName.value = ''
  digestAlgorithm.value = 'SHA256'
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  fileName.value = example.data
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="file-analyze"
    :name="t('tools.file-analyze.name')"
    name-en="File Analyzer"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <FileDropZone
            v-model="file"
            :label="t('tools.file-analyze.fields.data')"
          />
        </template>

        <template #options>
          <label class="field field--grow">
            <span class="field__label">{{ t('tools.file-analyze.fields.fileName') }}</span>
            <ElInput
              v-model="fileName"
              class="field__input tb-mono"
              autocomplete="off"
              :aria-label="t('tools.file-analyze.fields.fileName')"
              @keydown.enter.prevent="run"
            />
          </label>

          <p class="field__hint">
            {{ t('tools.file-analyze.hints.data') }}
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
        :title="t('tools.file-analyze.results.digest')"
        @retry="run"
      >
        <template #extra>
          <dl class="detail">
            <div class="detail__row">
              <dt>{{ t('tools.file-analyze.results.mimeType') }}</dt>
              <dd class="tb-mono">
                {{ data?.mimeType ?? '' }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.file-analyze.results.extension') }}</dt>
              <dd class="tb-mono">
                {{ data?.extension ?? '' }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.file-analyze.results.description') }}</dt>
              <dd class="tb-mono">
                {{ data?.description ?? '' }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.file-analyze.results.size') }}</dt>
              <dd class="tb-mono">
                {{ data?.size ?? 0 }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.file-analyze.results.magicHex') }}</dt>
              <dd class="tb-mono">
                {{ data?.magicHex ?? '' }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.file-analyze.results.digestAlgorithm') }}</dt>
              <dd class="tb-mono">
                {{ data?.digestAlgorithm ?? '' }}
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
'''

COMPONENTS["data/wifi-qr"] = '''<script setup lang="ts">
/**
 * 访客 WiFi 二维码工具页。
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
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useAutoRun } from '@/composables/useAutoRun'
import { useToolRun } from '@/composables/useToolRun'
import { EXAMPLES } from '@/tools/data/wifi-qr/example'

/** 后端 wifi-qr 工具的响应。 */
interface WiFiQRResult {
  /** 二维码承载的文本。 */
  payload: string
  /** PNG 图片的 Base64。 */
  png: string
  /** SVG 图片内容。 */
  svg: string
  /** 图片边长。 */
  size: number
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const ssid = ref('Guest-Network')
const password = ref('guest-password')
const security = ref('WPA')
const hidden = ref(false)
const size = ref(256)

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<WiFiQRResult>({
  slug: 'wifi-qr',
})

const running = computed(() => state.value === 'loading')

/** 结果文本：二维码承载的文本。 */
const resultText = computed(() => data.value?.payload ?? '')

const faq = computed<FaqItem[]>(() => tm('tools.wifi-qr.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  if (ssid.value.trim() === '') {
    return
  }
  void execute({
    ssid: ssid.value,
    password: password.value,
    security: security.value,
    hidden: hidden.value,
    size: size.value,
  })
}

// 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。
useAutoRun({
  slug: 'wifi-qr',
  run,
  watch: [ssid, password, security, hidden, size],
  enabled: () => ssid.value.trim() !== '',
})

/** 重置表单与结果。 */
function reset(): void {
  ssid.value = 'Guest-Network'
  password.value = 'guest-password'
  security.value = 'WPA'
  hidden.value = false
  size.value = 256
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  ssid.value = example.ssid
  password.value = example.password
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="wifi-qr"
    :name="t('tools.wifi-qr.name')"
    name-en="Guest WiFi QR Code"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field field--grow">
            <span class="field__label">{{ t('tools.wifi-qr.fields.ssid') }}</span>
            <ElInput
              v-model="ssid"
              class="field__input tb-mono"
              autocomplete="off"
              :aria-label="t('tools.wifi-qr.fields.ssid')"
              @keydown.enter.prevent="run"
            />
          </label>

          <label class="field field--grow">
            <span class="field__label">{{ t('tools.wifi-qr.fields.password') }}</span>
            <ElInput
              v-model="password"
              class="field__input tb-mono"
              type="password"
              show-password
              autocomplete="new-password"
              :aria-label="t('tools.wifi-qr.fields.password')"
              @keydown.enter.prevent="run"
            />
          </label>
        </template>

        <template #options>
          <label class="field">
            <span class="field__label">{{ t('tools.wifi-qr.fields.security') }}</span>
            <ElSelect
              v-model="security"
              class="field__select"
              :aria-label="t('tools.wifi-qr.fields.security')"
            >
              <ElOption
                :label="t('tools.wifi-qr.options.security.WPA')"
                value="WPA"
              />
              <ElOption
                :label="t('tools.wifi-qr.options.security.WEP')"
                value="WEP"
              />
              <ElOption
                :label="t('tools.wifi-qr.options.security.nopass')"
                value="nopass"
              />
            </ElSelect>
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.wifi-qr.fields.size') }}</span>
            <ElInputNumber
              v-model="size"
              class="field__select"
              :min="64"
              :max="1024"
              :step="64"
              :aria-label="t('tools.wifi-qr.fields.size')"
            />
          </label>

          <label class="field field--check">
            <ElCheckbox
              v-model="hidden"
              :aria-label="t('tools.wifi-qr.fields.hidden')"
            >
              {{ t('tools.wifi-qr.fields.hidden') }}
            </ElCheckbox>
          </label>

          <p class="field__hint">
            {{ t('tools.wifi-qr.hints.ssid') }}
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
        :title="t('tools.wifi-qr.results.png')"
        @retry="run"
      >
        <template #extra>
          <div class="detail__preview">
            <img
              v-if="data?.png"
              class="detail__image"
              :src="`data:image/png;base64,${data.png}`"
              :alt="t('tools.wifi-qr.results.png')"
            >
          </div>
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
'''


def main() -> None:
    for relative, content in COMPONENTS.items():
        target = ROOT / "frontend" / "src" / "tools" / relative / "Tool.vue"
        target.write_text(content, encoding="utf-8")
        print(f"写入 {relative}/Tool.vue")


if __name__ == "__main__":
    main()
