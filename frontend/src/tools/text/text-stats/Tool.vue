<script setup lang="ts">
/**
 * 文本统计工具页。
 *
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
 */
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useAutoRun } from '@/composables/useAutoRun'
import { useToolRun } from '@/composables/useToolRun'
import { EXAMPLES } from '@/tools/text/text-stats/example'

/** 后端 text-stats 工具的响应。 */
interface TextStatsResult {
  /** 字符总数。 */
  characters: number
  /** 不含空白的字符数。 */
  charactersNoSpace: number
  /** 词数。 */
  words: number
  /** 中文字符数。 */
  chineseChars: number
  /** 英文单词数。 */
  englishWords: number
  /** 数字字符数。 */
  digits: number
  /** 标点符号数。 */
  punctuation: number
  /** 行数。 */
  lines: number
  /** 段落数。 */
  paragraphs: number
  /** 字节数。 */
  bytes: number
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const text = ref('Hello 世界\n第二行文字\n\n新段落')

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<TextStatsResult>({
  slug: 'text-stats',
})

const running = computed(() => state.value === 'loading')

/** 结果文本：词数与字符数。 */
const resultText = computed(() => {
  if (!data.value) {
    return ''
  }
  return `${data.value.words} words / ${data.value.characters} chars`
})

const faq = computed<FaqItem[]>(() => tm('tools.text-stats.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  if (text.value === '') {
    return
  }
  void execute({ text: text.value })
}

// 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。
useAutoRun({
  slug: 'text-stats',
  run,
  watch: [text],
  enabled: () => text.value !== '',
})

/** 重置表单与结果。 */
function reset(): void {
  text.value = 'Hello 世界\n第二行文字\n\n新段落'
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  text.value = example.text
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="text-stats"
    :name="t('tools.text-stats.name')"
    name-en="Text Statistics"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field field--grow">
            <span class="field__label">{{ t('tools.text-stats.fields.text') }}</span>
            <ElInput
              v-model="text"
              class="field__input tb-mono"
              type="textarea"
              :rows="8"
              autocomplete="off"
              :aria-label="t('tools.text-stats.fields.text')"
            />
          </label>

          <p class="field__hint">
            {{ t('tools.text-stats.hints.text') }}
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
        :title="t('tools.text-stats.results.words')"
        @retry="run"
      >
        <template #extra>
          <dl class="detail">
            <div class="detail__row">
              <dt>{{ t('tools.text-stats.results.characters') }}</dt>
              <dd class="tb-mono">
                {{ data?.characters ?? 0 }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.text-stats.results.charactersNoSpace') }}</dt>
              <dd class="tb-mono">
                {{ data?.charactersNoSpace ?? 0 }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.text-stats.results.chineseChars') }}</dt>
              <dd class="tb-mono">
                {{ data?.chineseChars ?? 0 }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.text-stats.results.englishWords') }}</dt>
              <dd class="tb-mono">
                {{ data?.englishWords ?? 0 }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.text-stats.results.digits') }}</dt>
              <dd class="tb-mono">
                {{ data?.digits ?? 0 }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.text-stats.results.punctuation') }}</dt>
              <dd class="tb-mono">
                {{ data?.punctuation ?? 0 }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.text-stats.results.lines') }}</dt>
              <dd class="tb-mono">
                {{ data?.lines ?? 0 }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.text-stats.results.paragraphs') }}</dt>
              <dd class="tb-mono">
                {{ data?.paragraphs ?? 0 }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.text-stats.results.bytes') }}</dt>
              <dd class="tb-mono">
                {{ data?.bytes ?? 0 }}
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
