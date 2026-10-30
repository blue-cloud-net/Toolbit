<script setup lang="ts">
/**
 * 对称加解密工具页。
 *
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
 */
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { ElOption, ElSelect } from 'element-plus/es/components/select/index'
import 'element-plus/es/components/select/style/css'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import EncodingSelect from '@/components/tool/EncodingSelect.vue'
import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import StatGrid, { type StatGridItem } from '@/components/tool/StatGrid.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useToolRun } from '@/composables/useToolRun'
import { DEFAULT_EXAMPLE } from '@/tools/crypto/symmetric-crypt/example'

/** 后端 symmetric-crypt 工具的响应。 */
interface SymmetricCryptResult {
  output: string
  tag?: string
  iv?: string
  algorithm: string
  mode: string
  padding: string
  keyBytes: number
  ivBytes: number
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

type Encoding = 'utf8' | 'hex' | 'base64' | 'base64url'

const { t, tm } = useI18n()

/** 各算法支持的模式（与后端 cryptosvc.SymModes 保持一致）。 */
const MODES: Record<string, readonly string[]> = {
  SM4: ['ECB', 'CBC', 'CTR', 'OFB', 'CFB', 'GCM'],
  AES: ['ECB', 'CBC', 'CTR', 'GCM'],
}

const algorithm = ref<string>('SM4')
const operation = ref<string>('encrypt')
const mode = ref<string>('CBC')
const padding = ref<string>('pkcs7')
const key = ref('')
const keyEncoding = ref<Encoding>('hex')
const iv = ref('')
const ivEncoding = ref<Encoding>('hex')
const aad = ref('')
const aadEncoding = ref<Encoding>('utf8')
const tag = ref('')
const tagEncoding = ref<Encoding>('hex')
const input = ref('')
const inputEncoding = ref<Encoding>('utf8')
const outputEncoding = ref<string>('base64')

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
  bindInput,
  useAsInput,
} = useToolRun<SymmetricCryptResult>({ slug: 'symmetric-crypt' })

/** 结果区的「作为新输入」：把结果回填到主输入框（仅内存，不写存储）。 */
bindInput((text) => {
  input.value = text
})

const running = computed(() => state.value === 'loading')

/** 当前算法支持的模式。 */
const availableModes = computed(() => MODES[algorithm.value] ?? [])

/** 填充选项：只有 SM4 的 ECB / CBC 支持 zero / none。 */
const availablePaddings = computed(() => {
  if (algorithm.value === 'SM4' && (mode.value === 'ECB' || mode.value === 'CBC')) {
    return ['pkcs7', 'zero', 'none']
  }
  return ['pkcs7']
})

/** 是否需要 IV。 */
const needsIV = computed(() => mode.value !== 'ECB')

/** 是否需要认证标签。 */
const needsTag = computed(() => mode.value === 'GCM' && operation.value === 'decrypt')

/** 结果文本：密文 + 标签 + 自动生成的 nonce。 */
const resultText = computed(() => {
  const d = data.value
  if (!d) {
    return ''
  }
  const lines = [d.output]
  if (d.tag) {
    lines.push(`${t('tools.symmetric-crypt.results.tag')}: ${d.tag}`)
  }
  if (d.iv) {
    lines.push(`${t('tools.symmetric-crypt.results.iv')}: ${d.iv}`)
  }
  return lines.join('\n')
})

const faq = computed<FaqItem[]>(() => tm('tools.symmetric-crypt.faq') as FaqItem[])

/** 摘要统计卡片。 */
const stats = computed<StatGridItem[]>(() => {
  const d = data.value
  if (!d) {
    return []
  }
  const items: StatGridItem[] = [
    { label: t('tools.symmetric-crypt.results.algorithm'), value: d.algorithm },
    { label: t('tools.symmetric-crypt.results.mode'), value: d.mode },
    { label: t('tools.symmetric-crypt.results.padding'), value: d.padding },
  ]
  if (d.tag) {
    items.push({ label: t('tools.symmetric-crypt.results.tag'), value: d.tag })
  }
  if (d.iv) {
    items.push({ label: t('tools.symmetric-crypt.results.iv'), value: d.iv })
  }
  return items
})

/** 切换算法时校正模式与填充。 */
function onAlgorithmChange(): void {
  if (!availableModes.value.includes(mode.value)) {
    mode.value = availableModes.value[0] ?? 'CBC'
  }
  if (!availablePaddings.value.includes(padding.value)) {
    padding.value = 'pkcs7'
  }
}

/** 运行当前输入。 */
function run(): void {
  const payload: Record<string, unknown> = {
    algorithm: algorithm.value,
    operation: operation.value,
    mode: mode.value,
    padding: padding.value,
    key: { data: key.value, encoding: keyEncoding.value },
    input: { data: input.value, encoding: inputEncoding.value },
    outputEncoding: outputEncoding.value,
  }
  if (needsIV.value && iv.value !== '') {
    payload.iv = { data: iv.value, encoding: ivEncoding.value }
  }
  if (aad.value !== '') {
    payload.aad = { data: aad.value, encoding: aadEncoding.value }
  }
  if (needsTag.value && tag.value !== '') {
    payload.tag = { data: tag.value, encoding: tagEncoding.value }
  }
  void execute(payload)
}

/** 重置表单与结果。 */
function reset(): void {
  algorithm.value = 'SM4'
  operation.value = 'encrypt'
  mode.value = 'CBC'
  padding.value = 'pkcs7'
  key.value = ''
  keyEncoding.value = 'hex'
  iv.value = ''
  ivEncoding.value = 'hex'
  aad.value = ''
  aadEncoding.value = 'utf8'
  tag.value = ''
  tagEncoding.value = 'hex'
  input.value = ''
  inputEncoding.value = 'utf8'
  outputEncoding.value = 'base64'
  resetRun()
}

/** 填入示例。 */
function fillExample(): void {
  algorithm.value = DEFAULT_EXAMPLE.algorithm
  operation.value = DEFAULT_EXAMPLE.operation
  mode.value = DEFAULT_EXAMPLE.mode
  padding.value = DEFAULT_EXAMPLE.padding
  key.value = DEFAULT_EXAMPLE.key
  keyEncoding.value = 'hex'
  iv.value = DEFAULT_EXAMPLE.iv
  ivEncoding.value = 'hex'
  input.value = DEFAULT_EXAMPLE.input
  inputEncoding.value = DEFAULT_EXAMPLE.inputEncoding
  outputEncoding.value = 'base64'
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="symmetric-crypt"
    :name="t('tools.symmetric-crypt.name')"
    name-en="Symmetric Encryption"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.symmetric-crypt.fields.key') }}</span>
            <ElInput
              v-model="key"
              type="textarea"
              class="field__textarea tb-mono"
              :rows="2"
              spellcheck="false"
              autocomplete="off"
              :aria-label="t('tools.symmetric-crypt.fields.key')"
            />
          </label>

          <label
            v-if="needsIV"
            class="field"
          >
            <span class="field__label">{{ t('tools.symmetric-crypt.fields.iv') }}</span>
            <ElInput
              v-model="iv"
              class="field__input tb-mono"
              autocomplete="off"
              :aria-label="t('tools.symmetric-crypt.fields.iv')"
            />
          </label>

          <label
            v-if="mode === 'GCM'"
            class="field"
          >
            <span class="field__label">{{ t('tools.symmetric-crypt.fields.aad') }}</span>
            <ElInput
              v-model="aad"
              class="field__input tb-mono"
              autocomplete="off"
              :aria-label="t('tools.symmetric-crypt.fields.aad')"
            />
          </label>

          <label
            v-if="needsTag"
            class="field"
          >
            <span class="field__label">{{ t('tools.symmetric-crypt.fields.tag') }}</span>
            <ElInput
              v-model="tag"
              class="field__input tb-mono"
              autocomplete="off"
              :aria-label="t('tools.symmetric-crypt.fields.tag')"
            />
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.symmetric-crypt.fields.input') }}</span>
            <ElInput
              v-model="input"
              type="textarea"
              class="field__textarea tb-mono"
              :rows="5"
              spellcheck="false"
              :aria-label="t('tools.symmetric-crypt.fields.input')"
            />
          </label>
        </template>

        <template #options>
          <div class="field-row">
            <label class="field">
              <span class="field__label">{{ t('tools.symmetric-crypt.fields.algorithm') }}</span>
              <ElSelect
                v-model="algorithm"
                class="field__select"
                :aria-label="t('tools.symmetric-crypt.fields.algorithm')"
                @change="onAlgorithmChange"
              >
                <ElOption
                  :label="t('tools.symmetric-crypt.options.algorithm.SM4')"
                  value="SM4"
                />
                <ElOption
                  :label="t('tools.symmetric-crypt.options.algorithm.AES')"
                  value="AES"
                />
              </ElSelect>
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.symmetric-crypt.fields.operation') }}</span>
              <ElSelect
                v-model="operation"
                class="field__select"
                :aria-label="t('tools.symmetric-crypt.fields.operation')"
              >
                <ElOption
                  :label="t('tools.symmetric-crypt.options.operation.encrypt')"
                  value="encrypt"
                />
                <ElOption
                  :label="t('tools.symmetric-crypt.options.operation.decrypt')"
                  value="decrypt"
                />
              </ElSelect>
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.symmetric-crypt.fields.mode') }}</span>
              <ElSelect
                v-model="mode"
                class="field__select"
                :aria-label="t('tools.symmetric-crypt.fields.mode')"
                @change="onAlgorithmChange"
              >
                <ElOption
                  v-for="m in availableModes"
                  :key="m"
                  :label="t(`tools.symmetric-crypt.options.mode.${m}`)"
                  :value="m"
                />
              </ElSelect>
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.symmetric-crypt.fields.padding') }}</span>
              <ElSelect
                v-model="padding"
                class="field__select"
                :aria-label="t('tools.symmetric-crypt.fields.padding')"
              >
                <ElOption
                  v-for="p in availablePaddings"
                  :key="p"
                  :label="t(`tools.symmetric-crypt.options.padding.${p}`)"
                  :value="p"
                />
              </ElSelect>
            </label>
          </div>

          <div class="field-row">
            <label class="field">
              <span class="field__label">{{ t('tools.symmetric-crypt.fields.keyEncoding') }}</span>
              <EncodingSelect
                v-model="keyEncoding"
                :options="['utf8', 'hex', 'base64', 'base64url']"
              />
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.symmetric-crypt.fields.inputEncoding') }}</span>
              <EncodingSelect
                v-model="inputEncoding"
                :options="['utf8', 'hex', 'base64', 'base64url']"
              />
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.symmetric-crypt.fields.outputEncoding') }}</span>
              <EncodingSelect
                v-model="outputEncoding"
                :options="['base64', 'hex', 'base64url', 'utf8']"
              />
            </label>
          </div>

          <p class="field__hint">
            {{ t('tools.symmetric-crypt.hints.mode') }}
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
        :title="t('tools.symmetric-crypt.results.output')"
        :use-as-input="useAsInput"
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
        {{ t('tools.symmetric-crypt.description') }}
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
