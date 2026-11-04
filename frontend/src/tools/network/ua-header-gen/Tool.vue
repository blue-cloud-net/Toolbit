<script setup lang="ts">
/**
 * 请求头生成工具页。
 *
 * 按平台生成一组客户端请求头（User-Agent / Accept / Accept-Language 等），
 * 并允许追加自定义请求头。
 *
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
 */
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { ElOption, ElSelect } from 'element-plus/es/components/select/index'
import 'element-plus/es/components/select/style/css'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import DataTable, { type DataTableRow } from '@/components/tool/DataTable.vue'
import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import StatGrid, { type StatGridItem } from '@/components/tool/StatGrid.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useAutoRun } from '@/composables/useAutoRun'
import { useToolRun } from '@/composables/useToolRun'
import { EXAMPLES } from '@/tools/network/ua-header-gen/example'

/** 后端 ua-header-gen 工具的响应。 */
interface HeaderGeneratorResult {
  /** 生成的请求头。 */
  headers: Array<{ name: string; value: string }>
  /** 生成的 UA 串。 */
  userAgent: string
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

/** 平台枚举（与后端 `HeaderPlatforms` 对应）。 */
const PLATFORMS = ['chrome', 'firefox', 'safari', 'edge', 'android', 'ios', 'curl']

const { t, tm } = useI18n()

const platform = ref('chrome')
const version = ref('')
const os = ref('')
const acceptLanguage = ref('zh-CN,zh;q=0.9,en;q=0.8')
const referer = ref('')
const customHeaders = ref('')

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<HeaderGeneratorResult>(
  { slug: 'ua-header-gen' },
)

const running = computed(() => state.value === 'loading')

/** 结果文本：请求头文本，便于整段复制。 */
const resultText = computed(() => {
  const result = data.value
  if (!result) {
    return ''
  }
  return result.headers.map((header) => `${header.name}: ${header.value}`).join('\n')
})

/** 摘要卡片：生成的 UA 串。 */
const statItems = computed<StatGridItem[]>(() => {
  const result = data.value
  if (!result) {
    return []
  }
  return [
    {
      label: t('tools.ua-header-gen.results.userAgent'),
      value: result.userAgent || '—',
    },
  ]
})

/** 请求头表格行。 */
const headerRows = computed<DataTableRow[]>(() =>
  (data.value?.headers ?? []).map((header, index) => ({
    key: `${index}-${header.name}`,
    cells: { name: header.name, value: header.value },
  })),
)

const faq = computed<FaqItem[]>(() => tm('tools.ua-header-gen.faq') as FaqItem[])

/**
 * 自定义请求头：把「每行一条 `名称: 值`」的文本解析成对象。
 *
 * 后端 `customHeaders` 的类型是 `map[string]string`，
 * 直接把文本发过去会被判 `INVALID_FIELD`，因此这里先做解析。
 */
const parsedCustomHeaders = computed<Record<string, string>>(() => {
  const result: Record<string, string> = {}
  for (const line of customHeaders.value.split('\n')) {
    const trimmed = line.trim()
    if (trimmed === '') {
      continue
    }
    const separator = trimmed.indexOf(':')
    if (separator <= 0) {
      continue
    }
    const name = trimmed.slice(0, separator).trim()
    const value = trimmed.slice(separator + 1).trim()
    if (name !== '') {
      result[name] = value
    }
  }
  return result
})

/** 运行当前输入。 */
function run(): void {
  void execute({
    platform: platform.value,
    version: version.value,
    os: os.value,
    acceptLanguage: acceptLanguage.value,
    referer: referer.value,
    customHeaders: parsedCustomHeaders.value,
  })
}

// 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。
useAutoRun({
  slug: 'ua-header-gen',
  run,
  watch: [platform, version, os, acceptLanguage, referer, customHeaders],
  enabled: () => true,
})

/** 重置表单与结果。 */
function reset(): void {
  platform.value = 'chrome'
  version.value = ''
  os.value = ''
  acceptLanguage.value = 'zh-CN,zh;q=0.9,en;q=0.8'
  referer.value = ''
  customHeaders.value = ''
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  platform.value = example.platform
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="ua-header-gen"
    :name="t('tools.ua-header-gen.name')"
    name-en="Header Generator"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.ua-header-gen.fields.platform') }}</span>
            <ElSelect
              v-model="platform"
              class="field__select"
              :aria-label="t('tools.ua-header-gen.fields.platform')"
            >
              <ElOption
                v-for="name in PLATFORMS"
                :key="name"
                :label="t(`tools.ua-header-gen.options.platform.${name}`)"
                :value="name"
              />
            </ElSelect>
          </label>

          <label class="field field--grow">
            <span class="field__label">{{ t('tools.ua-header-gen.fields.version') }}</span>
            <ElInput
              v-model="version"
              class="field__input tb-mono"
              placeholder="120.0.0.0"
              autocomplete="off"
              :aria-label="t('tools.ua-header-gen.fields.version')"
            />
          </label>

          <label class="field field--grow">
            <span class="field__label">{{ t('tools.ua-header-gen.fields.os') }}</span>
            <ElInput
              v-model="os"
              class="field__input tb-mono"
              placeholder="Windows NT 10.0; Win64; x64"
              autocomplete="off"
              :aria-label="t('tools.ua-header-gen.fields.os')"
            />
          </label>
        </template>

        <template #options>
          <label class="field field--grow">
            <span class="field__label">{{ t('tools.ua-header-gen.fields.acceptLanguage') }}</span>
            <ElInput
              v-model="acceptLanguage"
              class="field__input tb-mono"
              autocomplete="off"
              :aria-label="t('tools.ua-header-gen.fields.acceptLanguage')"
            />
          </label>

          <label class="field field--grow">
            <span class="field__label">{{ t('tools.ua-header-gen.fields.referer') }}</span>
            <ElInput
              v-model="referer"
              class="field__input tb-mono"
              autocomplete="off"
              :aria-label="t('tools.ua-header-gen.fields.referer')"
            />
          </label>

          <label class="field field--grow">
            <span class="field__label">{{ t('tools.ua-header-gen.fields.customHeaders') }}</span>
            <ElInput
              v-model="customHeaders"
              class="field__input tb-mono"
              type="textarea"
              :rows="3"
              placeholder="X-Trace-Id: abc123"
              autocomplete="off"
              :aria-label="t('tools.ua-header-gen.fields.customHeaders')"
            />
          </label>

          <p class="field__hint">
            {{ t('tools.ua-header-gen.hints.customHeaders') }}
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
        :title="t('tools.ua-header-gen.results.headers')"
        hide-text
        @retry="run"
      >
        <template #extra>
          <StatGrid :items="statItems" />
          <DataTable
            v-if="headerRows.length > 0"
            class="ua-header-gen__table"
            :columns="[
              {
                key: 'name',
                label: t('tools.ua-header-gen.results.headerName'),
                mono: true,
              },
              {
                key: 'value',
                label: t('tools.ua-header-gen.results.headerValue'),
                mono: true,
              },
            ]"
            :rows="headerRows"
            copy-column="value"
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
.ua-header-gen__table {
  margin-top: var(--tb-space-3);
}
</style>
