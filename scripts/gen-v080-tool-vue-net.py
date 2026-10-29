#!/usr/bin/env python3
"""生成 v0.8.0 批次新增工具的 Tool.vue（网络 / 格式化 / 编码 / 文档类）。"""
from __future__ import annotations

import pathlib

ROOT = pathlib.Path(__file__).resolve().parents[1]
BASE = ROOT / "frontend" / "src" / "tools"

COMPONENTS: dict[str, str] = {}

COMPONENTS["network/ip-range-expand"] = '''<script setup lang="ts">
/**
 * IP 段展开工具页。
 *
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
 */
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
import { EXAMPLES } from '@/tools/network/ip-range-expand/example'

/** 后端 ip-range-expand 工具的响应。 */
interface IPRangeResult {
  /** 模式。 */
  mode: string
  /** 规范化后的 CIDR 列表。 */
  cidrs: string[]
  /** 起始地址。 */
  start: string
  /** 结束地址。 */
  end: string
  /** 地址总数。 */
  count: number
  /** 展开的地址列表。 */
  addresses: string[]
  /** 是否被上限截断。 */
  truncated: boolean
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const mode = ref('expand')
const cidr = ref('192.168.1.0/24')
const start = ref('192.168.1.10')
const end = ref('192.168.1.20')
const limit = ref(1024)

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<IPRangeResult>({
  slug: 'ip-range-expand',
})

const running = computed(() => state.value === 'loading')

/** 结果文本：地址列表。 */
const resultText = computed(() => data.value?.addresses.join('\\n') ?? '')

const faq = computed<FaqItem[]>(() => tm('tools.ip-range-expand.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  void execute({
    mode: mode.value,
    cidr: cidr.value,
    start: start.value,
    end: end.value,
    limit: limit.value,
  })
}

// 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。
useAutoRun({
  slug: 'ip-range-expand',
  run,
  watch: [mode, cidr, start, end, limit],
  enabled: () => true,
})

/** 重置表单与结果。 */
function reset(): void {
  mode.value = 'expand'
  cidr.value = '192.168.1.0/24'
  start.value = '192.168.1.10'
  end.value = '192.168.1.20'
  limit.value = 1024
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  cidr.value = example.cidr
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="ip-range-expand"
    :name="t('tools.ip-range-expand.name')"
    name-en="IP Range Expander"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.ip-range-expand.fields.mode') }}</span>
            <ElSelect
              v-model="mode"
              class="field__select"
              :aria-label="t('tools.ip-range-expand.fields.mode')"
            >
              <ElOption
                :label="t('tools.ip-range-expand.options.mode.expand')"
                value="expand"
              />
              <ElOption
                :label="t('tools.ip-range-expand.options.mode.toCidr')"
                value="toCidr"
              />
              <ElOption
                :label="t('tools.ip-range-expand.options.mode.toRange')"
                value="toRange"
              />
            </ElSelect>
          </label>

          <label
            v-if="mode === 'expand'"
            class="field field--grow"
          >
            <span class="field__label">{{ t('tools.ip-range-expand.fields.cidr') }}</span>
            <ElInput
              v-model="cidr"
              class="field__input tb-mono"
              autocomplete="off"
              :aria-label="t('tools.ip-range-expand.fields.cidr')"
              @keydown.enter.prevent="run"
            />
          </label>

          <label
            v-if="mode !== 'expand'"
            class="field field--grow"
          >
            <span class="field__label">{{ t('tools.ip-range-expand.fields.start') }}</span>
            <ElInput
              v-model="start"
              class="field__input tb-mono"
              autocomplete="off"
              :aria-label="t('tools.ip-range-expand.fields.start')"
              @keydown.enter.prevent="run"
            />
          </label>

          <label
            v-if="mode !== 'expand'"
            class="field field--grow"
          >
            <span class="field__label">{{ t('tools.ip-range-expand.fields.end') }}</span>
            <ElInput
              v-model="end"
              class="field__input tb-mono"
              autocomplete="off"
              :aria-label="t('tools.ip-range-expand.fields.end')"
              @keydown.enter.prevent="run"
            />
          </label>
        </template>

        <template #options>
          <label class="field">
            <span class="field__label">{{ t('tools.ip-range-expand.fields.limit') }}</span>
            <ElInputNumber
              v-model="limit"
              class="field__select"
              :min="1"
              :max="65536"
              :aria-label="t('tools.ip-range-expand.fields.limit')"
            />
          </label>

          <p class="field__hint">
            {{ t('tools.ip-range-expand.hints.limit') }}
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
        :title="t('tools.ip-range-expand.results.addresses')"
        @retry="run"
      >
        <template #extra>
          <dl class="detail">
            <div class="detail__row">
              <dt>{{ t('tools.ip-range-expand.results.cidr') }}</dt>
              <dd class="tb-mono">
                {{ data?.cidrs.join(' · ') ?? '' }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.ip-range-expand.results.start') }}</dt>
              <dd class="tb-mono">
                {{ data?.start ?? '' }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.ip-range-expand.results.end') }}</dt>
              <dd class="tb-mono">
                {{ data?.end ?? '' }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.ip-range-expand.results.total') }}</dt>
              <dd class="tb-mono">
                {{ data?.count ?? 0 }}
              </dd>
            </div>
            <div
              v-if="data?.truncated"
              class="detail__row"
            >
              <dt>!</dt>
              <dd>{{ t('tools.ip-range-expand.results.truncated') }}</dd>
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

COMPONENTS["network/device-headers"] = '''<script setup lang="ts">
/**
 * 设备请求头工具页。
 *
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
 */
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
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
import { EXAMPLES } from '@/tools/network/device-headers/example'

/** 一个请求头。 */
interface DeviceHeader {
  /** 请求头名称。 */
  name: string
  /** 请求头值。 */
  value: string
}

/** 解析出的 User-Agent 信息。 */
interface UAInfo {
  /** 浏览器名。 */
  browser: string
  /** 浏览器版本。 */
  browserVersion: string
  /** 操作系统名。 */
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

/** 后端 device-headers 工具的响应。 */
interface DeviceHeadersResult {
  /** 模式。 */
  mode: string
  /** 生成的请求头列表。 */
  headers: DeviceHeader[]
  /** 生成的 User-Agent 串。 */
  userAgent: string
  /** 解析出的 User-Agent 信息。 */
  parsed: UAInfo | null
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const mode = ref('generate')
const platform = ref('chrome')
const version = ref('')
const os = ref('')
const userAgent = ref(
  'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36',
)
const acceptLanguage = ref('zh-CN,zh;q=0.9,en;q=0.8')
const referer = ref('')
const customHeaders = ref('')

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<DeviceHeadersResult>({
  slug: 'device-headers',
})

const running = computed(() => state.value === 'loading')

/** 结果文本：请求头文本。 */
const resultText = computed(() => {
  if (!data.value) {
    return ''
  }
  return data.value.headers.map((header) => `${header.name}: ${header.value}`).join('\\n')
})

const faq = computed<FaqItem[]>(() => tm('tools.device-headers.faq') as FaqItem[])

/** 解析自定义请求头文本；每行一条 `名称: 值`。 */
function parseCustomHeaders(text: string): Record<string, string> {
  const result: Record<string, string> = {}
  for (const line of text.split('\n')) {
    const trimmed = line.trim()
    if (trimmed === '') {
      continue
    }
    const separator = trimmed.indexOf(':')
    if (separator <= 0) {
      continue
    }
    result[trimmed.slice(0, separator).trim()] = trimmed.slice(separator + 1).trim()
  }
  return result
}

/** 运行当前输入。 */
function run(): void {
  void execute({
    mode: mode.value,
    platform: platform.value,
    version: version.value,
    os: os.value,
    userAgent: userAgent.value,
    acceptLanguage: acceptLanguage.value,
    referer: referer.value,
    customHeaders: parseCustomHeaders(customHeaders.value),
  })
}

// 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。
useAutoRun({
  slug: 'device-headers',
  run,
  watch: [mode, platform, version, os, userAgent, acceptLanguage, referer, customHeaders],
  enabled: () => true,
})

/** 重置表单与结果。 */
function reset(): void {
  mode.value = 'generate'
  platform.value = 'chrome'
  version.value = ''
  os.value = ''
  userAgent.value =
    'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36'
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
    slug="device-headers"
    :name="t('tools.device-headers.name')"
    name-en="Device Headers"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.device-headers.fields.mode') }}</span>
            <ElSelect
              v-model="mode"
              class="field__select"
              :aria-label="t('tools.device-headers.fields.mode')"
            >
              <ElOption
                :label="t('tools.device-headers.options.mode.generate')"
                value="generate"
              />
              <ElOption
                :label="t('tools.device-headers.options.mode.parseUa')"
                value="parseUa"
              />
            </ElSelect>
          </label>

          <label
            v-if="mode === 'generate'"
            class="field"
          >
            <span class="field__label">{{ t('tools.device-headers.fields.platform') }}</span>
            <ElSelect
              v-model="platform"
              class="field__select"
              :aria-label="t('tools.device-headers.fields.platform')"
            >
              <ElOption
                v-for="name in ['chrome', 'firefox', 'safari', 'edge', 'android', 'ios', 'curl']"
                :key="name"
                :label="t(`tools.device-headers.options.platform.${name}`)"
                :value="name"
              />
            </ElSelect>
          </label>

          <label
            v-if="mode === 'generate'"
            class="field field--grow"
          >
            <span class="field__label">{{ t('tools.device-headers.fields.version') }}</span>
            <ElInput
              v-model="version"
              class="field__input tb-mono"
              placeholder="120.0.0.0"
              autocomplete="off"
              :aria-label="t('tools.device-headers.fields.version')"
            />
          </label>

          <label
            v-if="mode === 'generate'"
            class="field field--grow"
          >
            <span class="field__label">{{ t('tools.device-headers.fields.os') }}</span>
            <ElInput
              v-model="os"
              class="field__input tb-mono"
              placeholder="Windows 11"
              autocomplete="off"
              :aria-label="t('tools.device-headers.fields.os')"
            />
          </label>

          <label
            v-if="mode === 'parseUa'"
            class="field field--grow"
          >
            <span class="field__label">{{ t('tools.device-headers.fields.userAgent') }}</span>
            <ElInput
              v-model="userAgent"
              class="field__input tb-mono"
              type="textarea"
              :rows="4"
              autocomplete="off"
              :aria-label="t('tools.device-headers.fields.userAgent')"
            />
          </label>
        </template>

        <template #options>
          <label class="field field--grow">
            <span class="field__label">{{ t('tools.device-headers.fields.acceptLanguage') }}</span>
            <ElInput
              v-model="acceptLanguage"
              class="field__input tb-mono"
              autocomplete="off"
              :aria-label="t('tools.device-headers.fields.acceptLanguage')"
            />
          </label>

          <label class="field field--grow">
            <span class="field__label">{{ t('tools.device-headers.fields.referer') }}</span>
            <ElInput
              v-model="referer"
              class="field__input tb-mono"
              autocomplete="off"
              :aria-label="t('tools.device-headers.fields.referer')"
            />
          </label>

          <label class="field field--grow">
            <span class="field__label">{{ t('tools.device-headers.fields.customHeaders') }}</span>
            <ElInput
              v-model="customHeaders"
              class="field__input tb-mono"
              type="textarea"
              :rows="3"
              placeholder="X-Trace-Id: abc123"
              autocomplete="off"
              :aria-label="t('tools.device-headers.fields.customHeaders')"
            />
          </label>

          <p class="field__hint">
            {{ t('tools.device-headers.hints.customHeaders') }}
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
        :title="t('tools.device-headers.results.headers')"
        @retry="run"
      >
        <template #extra>
          <dl class="detail">
            <div
              v-if="data?.parsed"
              class="detail__row"
            >
              <dt>{{ t('tools.device-headers.results.browser') }}</dt>
              <dd class="tb-mono">
                {{ data.parsed.browser }} {{ data.parsed.browserVersion }}
              </dd>
            </div>
            <div
              v-if="data?.parsed"
              class="detail__row"
            >
              <dt>{{ t('tools.device-headers.results.os') }}</dt>
              <dd class="tb-mono">
                {{ data.parsed.os }} {{ data.parsed.osVersion }}
              </dd>
            </div>
            <div
              v-if="data?.parsed"
              class="detail__row"
            >
              <dt>{{ t('tools.device-headers.results.device') }}</dt>
              <dd class="tb-mono">
                {{ data.parsed.deviceType }}
              </dd>
            </div>
            <div
              v-if="data?.parsed?.isBot"
              class="detail__row"
            >
              <dt>!</dt>
              <dd>{{ t('tools.device-headers.results.isBot') }}</dd>
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

COMPONENTS["formatting/color-convert"] = '''<script setup lang="ts">
/**
 * 颜色格式转换工具页。
 *
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
 */
import { ElCheckbox } from 'element-plus/es/components/checkbox/index'
import 'element-plus/es/components/checkbox/style/css'
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
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
import { EXAMPLES } from '@/tools/formatting/color-convert/example'

/** 后端 color-convert 工具的响应。 */
interface ColorConvertResult {
  /** 输入颜色。 */
  input: string
  /** 输入格式。 */
  inputFormat: string
  /** 十六进制表示。 */
  hex: string
  /** RGB 表示。 */
  rgb: string
  /** HSL 表示。 */
  hsl: string
  /** OKLCH 表示。 */
  oklch: string
  /** 相对白底的对比度。 */
  contrastOnWhite: number
  /** 相对黑底的对比度。 */
  contrastOnBlack: number
  /** 是否为深色。 */
  isDark: boolean
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const input = ref('#3366ff')
const inputFormat = ref('hex')
const outputFormats = ref<string[]>(['hex', 'rgb', 'hsl', 'oklch'])

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<ColorConvertResult>({
  slug: 'color-convert',
})

const running = computed(() => state.value === 'loading')

/** 结果文本：各格式的颜色值。 */
const resultText = computed(() => {
  if (!data.value) {
    return ''
  }
  return [
    `hex\\t${data.value.hex}`,
    `rgb\\t${data.value.rgb}`,
    `hsl\\t${data.value.hsl}`,
    `oklch\\t${data.value.oklch}`,
  ].join('\\n')
})

const faq = computed<FaqItem[]>(() => tm('tools.color-convert.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  if (input.value.trim() === '') {
    return
  }
  void execute({
    input: input.value,
    inputFormat: inputFormat.value,
    outputFormats: outputFormats.value,
  })
}

// 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。
useAutoRun({
  slug: 'color-convert',
  run,
  watch: [input, inputFormat, outputFormats],
  enabled: () => input.value.trim() !== '',
})

/** 重置表单与结果。 */
function reset(): void {
  input.value = '#3366ff'
  inputFormat.value = 'hex'
  outputFormats.value = ['hex', 'rgb', 'hsl', 'oklch']
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  input.value = example.input
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="color-convert"
    :name="t('tools.color-convert.name')"
    name-en="Color Converter"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field field--grow">
            <span class="field__label">{{ t('tools.color-convert.fields.input') }}</span>
            <ElInput
              v-model="input"
              class="field__input tb-mono"
              autocomplete="off"
              :aria-label="t('tools.color-convert.fields.input')"
              @keydown.enter.prevent="run"
            />
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.color-convert.fields.inputFormat') }}</span>
            <ElSelect
              v-model="inputFormat"
              class="field__select"
              :aria-label="t('tools.color-convert.fields.inputFormat')"
            >
              <ElOption
                v-for="name in ['hex', 'rgb', 'hsl', 'oklch']"
                :key="name"
                :label="t(`tools.color-convert.options.format.${name}`)"
                :value="name"
              />
            </ElSelect>
          </label>
        </template>

        <template #options>
          <div class="field field--grow">
            <span class="field__label">{{ t('tools.color-convert.fields.outputFormats') }}</span>
            <ElCheckboxGroup
              v-model="outputFormats"
              class="field__checkgroup"
              :aria-label="t('tools.color-convert.fields.outputFormats')"
            >
              <ElCheckbox value="hex">
                {{ t('tools.color-convert.options.format.hex') }}
              </ElCheckbox>
              <ElCheckbox value="rgb">
                {{ t('tools.color-convert.options.format.rgb') }}
              </ElCheckbox>
              <ElCheckbox value="hsl">
                {{ t('tools.color-convert.options.format.hsl') }}
              </ElCheckbox>
              <ElCheckbox value="oklch">
                {{ t('tools.color-convert.options.format.oklch') }}
              </ElCheckbox>
            </ElCheckboxGroup>
          </div>

          <p class="field__hint">
            {{ t('tools.color-convert.hints.input') }}
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
        :title="t('tools.color-convert.results.colors')"
        @retry="run"
      >
        <template #extra>
          <dl class="detail">
            <div class="detail__row">
              <dt>{{ t('tools.color-convert.results.contrastOnWhite') }}</dt>
              <dd class="tb-mono">
                {{ data?.contrastOnWhite ?? 0 }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.color-convert.results.contrastOnBlack') }}</dt>
              <dd class="tb-mono">
                {{ data?.contrastOnBlack ?? 0 }}
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

COMPONENTS["encoding/html-escape"] = '''<script setup lang="ts">
/**
 * HTML 转义工具页。
 *
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
 */
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
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
import { EXAMPLES } from '@/tools/encoding/html-escape/example'

/** 后端 html-escape 工具的响应。 */
interface HTMLEscapeResult {
  /** 模式。 */
  mode: string
  /** 转义 / 反转义结果。 */
  output: string
  /** 发生变化的字符数。 */
  changedCount: number
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const mode = ref('encode')
// 用拼接拼出结束标签，避免在 SFC 中出现字面量的结束标签。
const SAMPLE_INPUT = '<script>alert("xss")</' + 'script> & more'
const input = ref(SAMPLE_INPUT)

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<HTMLEscapeResult>({
  slug: 'html-escape',
})

const running = computed(() => state.value === 'loading')

/** 结果文本。 */
const resultText = computed(() => data.value?.output ?? '')

const faq = computed<FaqItem[]>(() => tm('tools.html-escape.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  void execute({ mode: mode.value, input: input.value })
}

// 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。
useAutoRun({
  slug: 'html-escape',
  run,
  watch: [mode, input],
  enabled: () => true,
})

/** 重置表单与结果。 */
function reset(): void {
  mode.value = 'encode'
  input.value = SAMPLE_INPUT
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  input.value = example.input
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="html-escape"
    :name="t('tools.html-escape.name')"
    name-en="HTML Escape"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.html-escape.fields.mode') }}</span>
            <ElSelect
              v-model="mode"
              class="field__select"
              :aria-label="t('tools.html-escape.fields.mode')"
            >
              <ElOption
                :label="t('tools.html-escape.options.mode.encode')"
                value="encode"
              />
              <ElOption
                :label="t('tools.html-escape.options.mode.decode')"
                value="decode"
              />
            </ElSelect>
          </label>

          <label class="field field--grow">
            <span class="field__label">{{ t('tools.html-escape.fields.input') }}</span>
            <ElInput
              v-model="input"
              class="field__input tb-mono"
              type="textarea"
              :rows="8"
              autocomplete="off"
              :aria-label="t('tools.html-escape.fields.input')"
            />
          </label>

          <p class="field__hint">
            {{ t('tools.html-escape.hints.input') }}
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
        :title="t('tools.html-escape.results.output')"
        @retry="run"
      >
        <template #extra>
          <dl class="detail">
            <div class="detail__row">
              <dt>{{ t('tools.html-escape.results.entityCount') }}</dt>
              <dd class="tb-mono">
                {{ data?.changedCount ?? 0 }}
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

COMPONENTS["documents/markdown-convert"] = '''<script setup lang="ts">
/**
 * Markdown 转换工具页。
 *
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
 */
import { ElCheckbox } from 'element-plus/es/components/checkbox/index'
import 'element-plus/es/components/checkbox/style/css'
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
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
import { EXAMPLES } from '@/tools/documents/markdown-convert/example'

/** 后端 markdown-convert 工具的响应。 */
interface MarkdownConvertResult {
  /** 模式。 */
  mode: string
  /** 转换结果。 */
  output: string
  /** 标题数量。 */
  headingCount: number
  /** 链接数量。 */
  linkCount: number
  /** 代码块数量。 */
  codeBlockCount: number
  /** 词数。 */
  wordCount: number
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const mode = ref('toHtml')
const input = ref('# 标题\\n\\n这是一段 **加粗** 文本。\\n\\n- 项目一\\n- 项目二')
const gfm = ref(true)

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<MarkdownConvertResult>({
  slug: 'markdown-convert',
})

const running = computed(() => state.value === 'loading')

/** 结果文本。 */
const resultText = computed(() => data.value?.output ?? '')

const faq = computed<FaqItem[]>(() => tm('tools.markdown-convert.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  void execute({ mode: mode.value, input: input.value, gfm: gfm.value })
}

// 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。
useAutoRun({
  slug: 'markdown-convert',
  run,
  watch: [mode, input, gfm],
  enabled: () => true,
})

/** 重置表单与结果。 */
function reset(): void {
  mode.value = 'toHtml'
  input.value = '# 标题\\n\\n这是一段 **加粗** 文本。\\n\\n- 项目一\\n- 项目二'
  gfm.value = true
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  input.value = example.input
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="markdown-convert"
    :name="t('tools.markdown-convert.name')"
    name-en="Markdown Converter"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.markdown-convert.fields.mode') }}</span>
            <ElSelect
              v-model="mode"
              class="field__select"
              :aria-label="t('tools.markdown-convert.fields.mode')"
            >
              <ElOption
                :label="t('tools.markdown-convert.options.mode.toHtml')"
                value="toHtml"
              />
              <ElOption
                :label="t('tools.markdown-convert.options.mode.toMarkdown')"
                value="toMarkdown"
              />
            </ElSelect>
          </label>

          <label class="field field--grow">
            <span class="field__label">{{ t('tools.markdown-convert.fields.input') }}</span>
            <ElInput
              v-model="input"
              class="field__input tb-mono"
              type="textarea"
              :rows="12"
              autocomplete="off"
              :aria-label="t('tools.markdown-convert.fields.input')"
            />
          </label>
        </template>

        <template #options>
          <label class="field field--check">
            <ElCheckbox
              v-model="gfm"
              :aria-label="t('tools.markdown-convert.fields.gfm')"
            >
              {{ t('tools.markdown-convert.fields.gfm') }}
            </ElCheckbox>
          </label>

          <p class="field__hint">
            {{ t('tools.markdown-convert.hints.input') }}
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
        :title="t('tools.markdown-convert.results.output')"
        @retry="run"
      >
        <template #extra>
          <dl class="detail">
            <div class="detail__row">
              <dt>{{ t('tools.markdown-convert.results.headingCount') }}</dt>
              <dd class="tb-mono">
                {{ data?.headingCount ?? 0 }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.markdown-convert.results.linkCount') }}</dt>
              <dd class="tb-mono">
                {{ data?.linkCount ?? 0 }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.markdown-convert.results.codeBlockCount') }}</dt>
              <dd class="tb-mono">
                {{ data?.codeBlockCount ?? 0 }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.markdown-convert.results.wordCount') }}</dt>
              <dd class="tb-mono">
                {{ data?.wordCount ?? 0 }}
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


def main() -> None:
    for relative, content in COMPONENTS.items():
        target = BASE / relative / "Tool.vue"
        target.write_text(content, encoding="utf-8")
        print(f"写入 {relative}/Tool.vue")


if __name__ == "__main__":
    main()
