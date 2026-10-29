<script setup lang="ts">
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

/** 后端 device-headers 工具的响应。 */
interface DeviceHeadersResult {
  /** 模式。 */
  mode: string
  /** 生成的请求头。 */
  headers: Record<string, string>
  /** 解析出的 User-Agent 信息。 */
  userAgent: {
    browser: string
    version: string
    os: string
    device: string
    isBot: boolean
  } | null
  /** 命中的平台。 */
  platform: string
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
  return Object.entries(data.value.headers)
    .map(([key, value]) => `${key}: ${value}`)
    .join('\n')
})

const faq = computed<FaqItem[]>(() => tm('tools.device-headers.faq') as FaqItem[])

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
    customHeaders: customHeaders.value,
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
              v-if="data?.userAgent"
              class="detail__row"
            >
              <dt>{{ t('tools.device-headers.results.browser') }}</dt>
              <dd class="tb-mono">
                {{ data.userAgent.browser }} {{ data.userAgent.version }}
              </dd>
            </div>
            <div
              v-if="data?.userAgent"
              class="detail__row"
            >
              <dt>{{ t('tools.device-headers.results.os') }}</dt>
              <dd class="tb-mono">
                {{ data.userAgent.os }}
              </dd>
            </div>
            <div
              v-if="data?.userAgent"
              class="detail__row"
            >
              <dt>{{ t('tools.device-headers.results.device') }}</dt>
              <dd class="tb-mono">
                {{ data.userAgent.device }}
              </dd>
            </div>
            <div
              v-if="data?.userAgent?.isBot"
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
