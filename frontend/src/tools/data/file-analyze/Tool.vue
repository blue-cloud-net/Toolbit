<script setup lang="ts">
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
import ProductText from '@/components/tool/ProductText.vue'
import ResultPanel, { type ResultView } from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import StatGrid from '@/components/tool/StatGrid.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { buildSummary } from '@/components/tool/summary'
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

/** 摘要卡片：6 项关键字段。 */
const stats = computed(() =>
  buildSummary(
    data.value as unknown as Record<string, unknown> | null,
    [
      { key: 'mimeType', label: t('tools.file-analyze.results.mimeType') },
      { key: 'extension', label: t('tools.file-analyze.results.extension') },
      { key: 'description', label: t('tools.file-analyze.results.description'), wide: true },
      { key: 'size', label: t('tools.file-analyze.results.size') },
      { key: 'magicHex', label: t('tools.file-analyze.results.magicHex'), wide: true },
      { key: 'digestAlgorithm', label: t('tools.file-analyze.results.digestAlgorithm') },
    ],
    { yes: t('common.states.yes'), no: t('common.states.no') },
  ),
)

/** 结果页签：摘要 / 产物。 */
const views = computed<ResultView[]>(() => [
  { id: 'summary', label: t('tools.file-analyze.views.summary') },
  { id: 'digest', label: t('tools.file-analyze.views.digest') },
])

/** 当前页签；默认停在「摘要」。 */
const activeView = ref('summary')

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
        :views="views"
        :title="t('tools.file-analyze.results.digest')"
        @retry="run"
      >
        <template #view="{ viewId }">
          <template v-if="viewId === 'summary' && data">
            <StatGrid :items="stats" />
          </template>

          <ProductText
            v-else-if="viewId === 'digest' && data"
            :label="t('tools.file-analyze.views.digest')"
            :content="String(data.digest ?? '')"
          />
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
