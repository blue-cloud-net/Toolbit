<script setup lang="ts">
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
