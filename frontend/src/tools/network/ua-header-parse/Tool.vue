<script setup lang="ts">
/**
 * UA 解析工具页。
 *
 * 把 User-Agent 串拆成浏览器、版本、操作系统、设备类型与渲染引擎。
 *
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
 */
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import StatGrid, { type StatGridItem } from '@/components/tool/StatGrid.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useAutoRun } from '@/composables/useAutoRun'
import { useToolRun } from '@/composables/useToolRun'
import { EXAMPLES } from '@/tools/network/ua-header-parse/example'

/** 后端 ua-header-parse 工具的响应。 */
interface UAParserResult {
  /** 规范化后的 UA 串。 */
  userAgent: string
  /** 浏览器名称。 */
  browser: string
  /** 浏览器版本。 */
  browserVersion: string
  /** 操作系统名称。 */
  os: string
  /** 操作系统版本。 */
  osVersion: string
  /** 设备类型。 */
  deviceType: string
  /** 是否是爬虫。 */
  isBot: boolean
  /** 渲染引擎。 */
  engine: string
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const userAgent = ref(
  'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
)

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<UAParserResult>({
  slug: 'ua-header-parse',
})

const running = computed(() => state.value === 'loading')

/** 结果文本：回显规范化后的 UA 串，便于复制。 */
const resultText = computed(() => data.value?.userAgent ?? '')

/** 摘要卡片：解析出的结构化字段。 */
const statItems = computed<StatGridItem[]>(() => {
  const result = data.value
  if (!result) {
    return []
  }
  const items: StatGridItem[] = [
    { label: t('tools.ua-header-parse.results.browser'), value: result.browser || '—' },
    {
      label: t('tools.ua-header-parse.results.browserVersion'),
      value: result.browserVersion || '—',
    },
    { label: t('tools.ua-header-parse.results.os'), value: result.os || '—' },
    { label: t('tools.ua-header-parse.results.osVersion'), value: result.osVersion || '—' },
    { label: t('tools.ua-header-parse.results.deviceType'), value: result.deviceType || '—' },
    { label: t('tools.ua-header-parse.results.engine'), value: result.engine || '—' },
  ]
  items.push({
    label: t('tools.ua-header-parse.results.isBot'),
    value: result.isBot ? t('common.states.yes') : t('common.states.no'),
    level: result.isBot ? 'warning' : 'success',
    mono: false,
  })
  return items
})

const faq = computed<FaqItem[]>(() => tm('tools.ua-header-parse.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  void execute({ userAgent: userAgent.value })
}

// 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。
useAutoRun({
  slug: 'ua-header-parse',
  run,
  watch: [userAgent],
  enabled: () => true,
})

/** 重置表单与结果。 */
function reset(): void {
  userAgent.value =
    'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36'
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  userAgent.value = example.userAgent
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="ua-header-parse"
    :name="t('tools.ua-header-parse.name')"
    name-en="User-Agent Parser"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field field--grow">
            <span class="field__label">{{ t('tools.ua-header-parse.fields.userAgent') }}</span>
            <ElInput
              v-model="userAgent"
              class="field__input tb-mono"
              type="textarea"
              :rows="4"
              autocomplete="off"
              :aria-label="t('tools.ua-header-parse.fields.userAgent')"
            />
          </label>
        </template>

        <template #options>
          <p class="field__hint">
            {{ t('tools.ua-header-parse.hints.output') }}
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
        :state="state"
        :text="resultText"
        :meta="meta"
        :error="error"
        :title="t('tools.ua-header-parse.results.title')"
        hide-text
        @retry="run"
      >
        <template #extra>
          <StatGrid :items="statItems" />
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
