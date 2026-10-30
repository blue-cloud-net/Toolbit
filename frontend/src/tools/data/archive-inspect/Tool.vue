<script setup lang="ts">
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

import DataTable, { type DataTableColumn, type DataTableRow } from '@/components/tool/DataTable.vue'
import FileDropZone from '@/components/tool/FileDropZone.vue'
import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import StatGrid, { type StatGridItem } from '@/components/tool/StatGrid.vue'
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
    .map((entry) => `${entry.isDir ? 'd' : '-'} ${entry.size}\t${entry.name}`)
    .join('\n')
})

const faq = computed<FaqItem[]>(() => tm('tools.archive-inspect.faq') as FaqItem[])

/** 统计卡片。 */
const stats = computed<StatGridItem[]>(() => [
  { label: t('tools.archive-inspect.results.entryCount'), value: String(data.value?.entryCount ?? 0) },
  { label: t('tools.archive-inspect.results.totalSize'), value: String(data.value?.totalSize ?? 0) },
  {
    label: t('tools.archive-inspect.results.compressedSize'),
    value: String(data.value?.compressedSize ?? 0),
  },
])

/** 表格列定义。 */
const columns = computed<DataTableColumn[]>(() => [
  { key: 'name', label: t('tools.archive-inspect.results.name'), mono: true },
  { key: 'type', label: t('tools.archive-inspect.results.entryType') },
  { key: 'size', label: t('tools.archive-inspect.results.entrySize'), mono: true, align: 'right' },
  { key: 'modified', label: t('tools.archive-inspect.results.modified'), mono: true },
  { key: 'preview', label: t('tools.archive-inspect.results.preview') },
])

/** 表格行。 */
const rows = computed<DataTableRow[]>(() =>
  (data.value?.entries ?? []).map((entry, index) => ({
    key: `${entry.name}-${index}`,
    cells: {
      name: entry.name,
      type: entry.isDir
        ? t('tools.archive-inspect.results.entryDir')
        : t('tools.archive-inspect.results.entryFile'),
      size: String(entry.size),
      modified: entry.modified || '—',
      preview: entry.contentPreview || '—',
    },
  })),
)

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
        hide-text
        @retry="run"
      >
        <template #extra>
          <StatGrid :items="stats" />

          <DataTable
            class="table"
            :columns="columns"
            :rows="rows"
            :copy-column="'name'"
            :empty="t('tools.archive-inspect.results.empty')"
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

<style scoped>
.table {
  margin-top: var(--tb-space-3);
}
</style>
