<script setup lang="ts">
/**
 * ECDH 密钥协商工具页。
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

import EncodingSelect from '@/components/tool/EncodingSelect.vue'
import KeyInput from '@/components/tool/KeyInput.vue'
import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import StatGrid, { type StatGridItem } from '@/components/tool/StatGrid.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useToolRun } from '@/composables/useToolRun'
import { DEFAULT_EXAMPLE } from '@/tools/crypto/ecdh/example'

/** 后端 ecdh 工具的响应。 */
interface ECDHResult {
  sharedSecret: string
  derivedKey?: string
  curve: string
  length: number
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

type Encoding = 'utf8' | 'hex' | 'base64' | 'base64url'

const { t, tm } = useI18n()

const CURVES = ['P-256', 'P-384', 'P-521', 'X25519', 'X448'] as const
const HASHES = ['SHA256', 'SHA384', 'SHA512', 'SM3'] as const

const curve = ref<string>('P-256')
const privateKey = ref('')
const privateKeyFormat = ref('auto')
const peerPublicKey = ref('')
const peerPublicKeyFormat = ref('auto')
const useKDF = ref(false)
const hash = ref<string>('SHA256')
const salt = ref('')
const saltEncoding = ref<Encoding>('hex')
const info = ref('')
const infoEncoding = ref<Encoding>('utf8')
const length = ref(32)
const outputEncoding = ref<string>('base64')

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
} = useToolRun<ECDHResult>({ slug: 'ecdh' })

const running = computed(() => state.value === 'loading')

/** 结果文本：共享密钥 + 派生密钥。 */
const resultText = computed(() => {
  const d = data.value
  if (!d) {
    return ''
  }
  const lines = [d.sharedSecret]
  if (d.derivedKey) {
    lines.push(`${t('tools.ecdh.results.derivedKey')}: ${d.derivedKey}`)
  }
  return lines.join('\n')
})

const faq = computed<FaqItem[]>(() => tm('tools.ecdh.faq') as FaqItem[])

/** 摘要统计卡片。 */
const stats = computed<StatGridItem[]>(() => {
  const d = data.value
  if (!d) {
    return []
  }
  const items: StatGridItem[] = [
    { label: t('tools.ecdh.results.curve'), value: d.curve },
    { label: t('tools.ecdh.results.sharedSecret'), value: d.sharedSecret },
  ]
  if (d.derivedKey) {
    items.push({ label: t('tools.ecdh.results.derivedKey'), value: d.derivedKey })
  }
  return items
})

/** 运行当前输入。 */
function run(): void {
  const payload: Record<string, unknown> = {
    curve: curve.value,
    privateKey: { data: privateKey.value, format: privateKeyFormat.value },
    peerPublicKey: { data: peerPublicKey.value, format: peerPublicKeyFormat.value },
    outputEncoding: outputEncoding.value,
  }
  if (useKDF.value) {
    payload.kdf = {
      type: 'hkdf',
      hash: hash.value,
      salt: { data: salt.value, encoding: saltEncoding.value },
      info: { data: info.value, encoding: infoEncoding.value },
      length: length.value,
    }
  }
  void execute(payload)
}

/** 重置表单与结果。 */
function reset(): void {
  curve.value = 'P-256'
  privateKey.value = ''
  privateKeyFormat.value = 'auto'
  peerPublicKey.value = ''
  peerPublicKeyFormat.value = 'auto'
  useKDF.value = false
  hash.value = 'SHA256'
  salt.value = ''
  saltEncoding.value = 'hex'
  info.value = ''
  infoEncoding.value = 'utf8'
  length.value = 32
  outputEncoding.value = 'base64'
  resetRun()
}

/** 填入示例。 */
function fillExample(): void {
  curve.value = DEFAULT_EXAMPLE.curve
  useKDF.value = DEFAULT_EXAMPLE.useKDF
  hash.value = DEFAULT_EXAMPLE.hash
  length.value = DEFAULT_EXAMPLE.length
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="ecdh"
    :name="t('tools.ecdh.name')"
    name-en="ECDH Key Agreement"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <KeyInput
            v-model="privateKey"
            v-model:format="privateKeyFormat"
            :label="t('tools.ecdh.fields.privateKey')"
          />

          <KeyInput
            v-model="peerPublicKey"
            v-model:format="peerPublicKeyFormat"
            :label="t('tools.ecdh.fields.peerPublicKey')"
          />
        </template>

        <template #options>
          <div class="field-row">
            <label class="field">
              <span class="field__label">{{ t('tools.ecdh.fields.curve') }}</span>
              <ElSelect
                v-model="curve"
                class="field__select"
                :aria-label="t('tools.ecdh.fields.curve')"
              >
                <ElOption
                  v-for="c in CURVES"
                  :key="c"
                  :label="t(`tools.ecdh.options.curve.${c}`)"
                  :value="c"
                />
              </ElSelect>
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.ecdh.fields.outputEncoding') }}</span>
              <EncodingSelect
                v-model="outputEncoding"
                :options="['base64', 'hex', 'base64url']"
              />
            </label>
          </div>

          <ElCheckbox
            v-model="useKDF"
            class="field__checkbox"
          >
            {{ t('tools.ecdh.fields.useKDF') }}
          </ElCheckbox>

          <div
            v-if="useKDF"
            class="field-row"
          >
            <label class="field">
              <span class="field__label">{{ t('tools.ecdh.fields.kdfHash') }}</span>
              <ElSelect
                v-model="hash"
                class="field__select"
                :aria-label="t('tools.ecdh.fields.kdfHash')"
              >
                <ElOption
                  v-for="h in HASHES"
                  :key="h"
                  :label="t(`tools.ecdh.options.kdfHash.${h}`)"
                  :value="h"
                />
              </ElSelect>
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.ecdh.fields.kdfLength') }}</span>
              <ElInputNumber
                v-model="length"
                class="field__input"
                :min="1"
                :max="1048576"
                :aria-label="t('tools.ecdh.fields.kdfLength')"
              />
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.ecdh.fields.salt') }}</span>
              <ElInput
                v-model="salt"
                class="field__input tb-mono"
                autocomplete="off"
                :aria-label="t('tools.ecdh.fields.salt')"
              />
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.ecdh.fields.info') }}</span>
              <ElInput
                v-model="info"
                class="field__input tb-mono"
                autocomplete="off"
                :aria-label="t('tools.ecdh.fields.info')"
              />
            </label>
          </div>

          <p class="field__hint">
            {{ t('tools.ecdh.hints.kdf') }}
          </p>
        </template>

        <template #footer>
          <RunBar
            :running="running"
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
        :title="t('tools.ecdh.results.sharedSecret')"
        @retry="run"
      >
        <template
          v-if="data"
          #extra
        >
          <StatGrid :items="stats" />
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.ecdh.description') }}
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
