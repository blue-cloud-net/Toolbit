<script setup lang="ts">
/**
 * 密钥格式转换工具页。
 *
 * 密钥内容只存在于组件局部状态；离开页面即丢弃，
 * **不写**任何浏览器存储（AGENTS.md §4.11）。
 */
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import KeyInput from '@/components/tool/KeyInput.vue'
import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useToolRun } from '@/composables/useToolRun'
import { DEFAULT_EXAMPLE } from '@/tools/keys/key-convert/example'

/** 后端 key-convert 工具的响应。 */
interface KeyConvertResult {
  key: { pem: string; derBase64?: string }
  detected: {
    algorithm: string
    format: string
    encrypted: boolean
    bits: number
    curve: string
  }
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const FROM_FORMATS = ['auto', 'pem', 'der', 'pkcs1', 'pkcs8', 'spki'] as const
const TO_FORMATS = ['pkcs8', 'pkcs1', 'spki', 'der', 'pem'] as const

const key = ref('')
const keyFormat = ref('auto')
const from = ref('auto')
const to = ref('pkcs8')
const passphrase = ref('')

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
} = useToolRun<KeyConvertResult>({ slug: 'key-convert' })

const running = computed(() => state.value === 'loading')

/** 结果文本。 */
const resultText = computed(() => {
  const result = data.value
  if (!result) {
    return ''
  }
  if (result.key.derBase64) {
    return result.key.derBase64
  }
  return result.key.pem
})

const faq = computed<FaqItem[]>(() => tm('tools.key-convert.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  void execute({
    key: { data: key.value, format: keyFormat.value },
    from: from.value,
    to: to.value,
    passphrase: passphrase.value,
  })
}

/** 重置表单与结果。 */
function reset(): void {
  key.value = ''
  keyFormat.value = 'auto'
  from.value = 'auto'
  to.value = 'pkcs8'
  passphrase.value = ''
  resetRun()
}

/** 填入示例。 */
function fillExample(): void {
  key.value = DEFAULT_EXAMPLE.key
  keyFormat.value = 'auto'
  from.value = DEFAULT_EXAMPLE.from
  to.value = DEFAULT_EXAMPLE.to
  passphrase.value = ''
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="key-convert"
    :name="t('tools.key-convert.name')"
    name-en="Key Format Converter"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <KeyInput
            v-model="key"
            v-model:format="keyFormat"
            :label="t('tools.key-convert.fields.key')"
            :show-passphrase="true"
          />
        </template>

        <template #options>
          <div class="field-row">
            <label class="field">
              <span class="field__label">{{ t('tools.key-convert.fields.from') }}</span>
              <select
                v-model="from"
                class="field__select"
              >
                <option
                  v-for="name in FROM_FORMATS"
                  :key="name"
                  :value="name"
                >
                  {{ name }}
                </option>
              </select>
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.key-convert.fields.to') }}</span>
              <select
                v-model="to"
                class="field__select"
              >
                <option
                  v-for="name in TO_FORMATS"
                  :key="name"
                  :value="name"
                >
                  {{ name }}
                </option>
              </select>
            </label>
          </div>

          <p class="field__hint">
            {{ t('tools.key-convert.hints.format') }}
          </p>
        </template>

        <template #footer>
          <RunBar
            :running="running"
            :disabled="key === ''"
            @run="run"
            @reset="reset"
            @example="fillExample"
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
        :title="t('tools.key-convert.results.key')"
        @retry="run"
      />
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.key-convert.description') }}
      </p>
      <p
        v-if="data?.detected"
        class="prose"
      >
        {{ t('tools.key-convert.results.detected') }}：
        {{ data.detected.algorithm }} · {{ data.detected.format }}
        <template v-if="data.detected.bits">
          · {{ data.detected.bits }} bit
        </template>
        <template v-if="data.detected.curve">
          · {{ data.detected.curve }}
        </template>
      </p>
    </template>

    <template #faq>
      <dl class="faq">
        <template
          v-for="item in faq"
          :key="item.q"
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
