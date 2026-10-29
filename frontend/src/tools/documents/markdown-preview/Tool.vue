<script setup lang="ts">
/**
 * Markdown 预览工具页。
 *
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
 *
 * 渲染结果**不**注入页面：仓库禁止 `v-html`，
 * 因此只把净化后的 HTML 当作文本展示，目录以列表呈现。
 */
import { ElCheckbox } from 'element-plus/es/components/checkbox/index'
import 'element-plus/es/components/checkbox/style/css'
import { ElOption, ElSelect } from 'element-plus/es/components/select/index'
import 'element-plus/es/components/select/style/css'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import CodeEditor from '@/components/tool/CodeEditor.vue'
import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useToolRun } from '@/composables/useToolRun'
import { EXAMPLES } from '@/tools/documents/markdown-preview/example'
import type { MarkdownFlavor } from '@/tools/documents/markdown-preview/example'

/** 一个目录条目。 */
interface TocEntry {
  /** 标题层级（1 到 6）。 */
  level: number
  /** 标题文本。 */
  text: string
  /** 标题锚点。 */
  id: string
}

/** 后端 markdown-preview 工具的响应。 */
interface MarkdownPreviewResult {
  /** 净化后的 HTML。 */
  html: string
  /** 目录。 */
  toc: TocEntry[]
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const markdown = ref('# Toolbit\n\n在左侧粘贴 Markdown。')
const flavor = ref<MarkdownFlavor>('gfm')
const allowHtml = ref(false)
const breaks = ref(false)

const { state, data, meta, error, run: execute, reset: resetRun } =
  useToolRun<MarkdownPreviewResult>({ slug: 'markdown-preview' })

const running = computed(() => state.value === 'loading')

/** 结果文本：净化后的 HTML 源码。 */
const resultText = computed(() => data.value?.html ?? '')

/** 目录条目。 */
const toc = computed<TocEntry[]>(() => data.value?.toc ?? [])

const faq = computed<FaqItem[]>(() => tm('tools.markdown-preview.faq') as FaqItem[])

/** 目录项的缩进（像素）。 */
function indentOf(level: number): string {
  return `${Math.max(0, level - 1) * 14}px`
}

/** 运行当前输入。 */
function run(): void {
  void execute({
    markdown: markdown.value,
    flavor: flavor.value,
    allowHtml: allowHtml.value,
    breaks: breaks.value,
  })
}

/** 重置表单与结果。 */
function reset(): void {
  markdown.value = '# Toolbit\n\n在左侧粘贴 Markdown。'
  flavor.value = 'gfm'
  allowHtml.value = false
  breaks.value = false
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  markdown.value = example.markdown
  flavor.value = example.flavor
  allowHtml.value = example.allowHtml
  breaks.value = example.breaks
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="markdown-preview"
    :name="t('tools.markdown-preview.name')"
    name-en="Markdown Preview"
    wide
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.markdown-preview.fields.markdown') }}</span>
            <CodeEditor
              v-model="markdown"
              :label="t('tools.markdown-preview.fields.markdown')"
              :max-rows="24"
              @run="run"
            />
          </label>

          <p class="field__hint">
            {{ t('tools.markdown-preview.hints.markdown') }}
          </p>
        </template>

        <template #options>
          <div class="field-row">
            <label class="field">
              <span class="field__label">{{ t('tools.markdown-preview.fields.flavor') }}</span>
              <ElSelect
                v-model="flavor"
                class="field__select"
                :aria-label="t('tools.markdown-preview.fields.flavor')"
              >
                <ElOption
                  :label="t('tools.markdown-preview.options.flavor.gfm')"
                  value="gfm"
                />
                <ElOption
                  :label="t('tools.markdown-preview.options.flavor.commonmark')"
                  value="commonmark"
                />
              </ElSelect>
            </label>
          </div>

          <ElCheckbox
            v-model="allowHtml"
            class="field__checkbox"
          >
            {{ t('tools.markdown-preview.fields.allowHtml') }}
          </ElCheckbox>
          <p class="field__hint">
            {{ t('tools.markdown-preview.hints.allowHtml') }}
          </p>

          <ElCheckbox
            v-model="breaks"
            class="field__checkbox"
          >
            {{ t('tools.markdown-preview.fields.breaks') }}
          </ElCheckbox>
          <p class="field__hint">
            {{ t('tools.markdown-preview.hints.breaks') }}
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
        :title="t('tools.markdown-preview.results.output')"
        @retry="run"
      >
        <template
          v-if="data"
          #extra
        >
          <h3 class="group__title">
            {{ t('tools.markdown-preview.results.toc') }}
          </h3>

          <ul
            v-if="toc.length > 0"
            class="toc"
          >
            <li
              v-for="entry in toc"
              :key="entry.id"
              class="toc__item"
              :style="{ paddingLeft: indentOf(entry.level) }"
            >
              <span class="toc__text">{{ entry.text }}</span>
              <span class="toc__id tb-mono">#{{ entry.id }}</span>
            </li>
          </ul>
          <p
            v-else
            class="toc__empty"
          >
            {{ t('tools.markdown-preview.results.empty') }}
          </p>
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.markdown-preview.hints.output') }}
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
.group__title {
  margin: var(--tb-space-3) 0 var(--tb-space-2);
  font-size: var(--tb-font-sm);
  font-weight: 600;
  color: var(--tb-text-secondary);
}

.toc {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-1);
  margin: 0;
  padding: 0;
  list-style: none;
}

.toc__item {
  display: flex;
  flex-wrap: wrap;
  gap: var(--tb-space-2);
  font-size: var(--tb-font-sm);
}

.toc__text {
  color: var(--tb-text);
}

.toc__id {
  font-size: var(--tb-font-xs);
  color: var(--tb-text-tertiary);
}

.toc__empty {
  margin: 0;
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
