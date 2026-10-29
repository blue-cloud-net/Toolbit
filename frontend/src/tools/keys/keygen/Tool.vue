<script setup lang="ts">
/**
 * 密钥生成工具页。
 *
 * 生成的私钥只存在于组件局部状态；离开页面即丢弃，
 * **不写**任何浏览器存储（AGENTS.md §4.11）。
 */
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useToolRun } from '@/composables/useToolRun'
import { DEFAULT_EXAMPLE } from '@/tools/keys/keygen/example'

/** 后端 keygen 工具的响应。 */
interface KeyGenResult {
  privateKey: { pem: string; derBase64?: string }
  publicKey: { pem: string; derBase64?: string }
  algorithm: string
  bits: number
  curve: string
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const ALGORITHMS = ['RSA', 'SM2', 'EC', 'Ed25519', 'Ed448', 'X25519', 'X448'] as const
const BITS = [1024, 2048, 3072, 4096] as const
const CURVES = ['P-256', 'P-384', 'P-521', 'SM2'] as const
const PRIVATE_FORMATS = ['pkcs8', 'pkcs1'] as const
const PUBLIC_FORMATS = ['spki'] as const
const PBE_CIPHERS = ['AES-256-CBC', 'DES-EDE3-CBC'] as const

const algorithm = ref<string>('RSA')
const bits = ref<number>(2048)
const curve = ref<string>('P-256')
const privateFormat = ref<string>('pkcs8')
const publicFormat = ref<string>('spki')
const encrypt = ref(false)
const pbeCipher = ref<string>('AES-256-CBC')
const passphrase = ref('')

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
} = useToolRun<KeyGenResult>({ slug: 'keygen' })

const running = computed(() => state.value === 'loading')

/** 是否是 RSA 算法。 */
const isRSA = computed(() => algorithm.value === 'RSA')

/** 是否是椭圆曲线类算法。 */
const isCurve = computed(() => algorithm.value === 'EC' || algorithm.value === 'SM2')

/** 结果文本：私钥与公钥分区展示。 */
const resultText = computed(() => {
  const result = data.value
  if (!result) {
    return ''
  }
  return [
    `# ${t('tools.keygen.results.privateKey')}`,
    result.privateKey.pem,
    '',
    `# ${t('tools.keygen.results.publicKey')}`,
    result.publicKey.pem,
  ].join('\n')
})

const faq = computed<FaqItem[]>(() => tm('tools.keygen.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  void execute({
    algorithm: algorithm.value,
    bits: isRSA.value ? bits.value : 0,
    curve: isCurve.value ? curve.value : '',
    privateFormat: privateFormat.value,
    publicFormat: publicFormat.value,
    encrypt: encrypt.value
      ? { cipher: pbeCipher.value, passphrase: passphrase.value }
      : null,
  })
}

/** 重置表单与结果。 */
function reset(): void {
  algorithm.value = 'RSA'
  bits.value = 2048
  curve.value = 'P-256'
  privateFormat.value = 'pkcs8'
  publicFormat.value = 'spki'
  encrypt.value = false
  pbeCipher.value = 'AES-256-CBC'
  passphrase.value = ''
  resetRun()
}

/** 填入示例。 */
function fillExample(): void {
  algorithm.value = DEFAULT_EXAMPLE.algorithm
  bits.value = DEFAULT_EXAMPLE.bits
  curve.value = DEFAULT_EXAMPLE.curve
  privateFormat.value = 'pkcs8'
  publicFormat.value = 'spki'
  encrypt.value = DEFAULT_EXAMPLE.encrypt
  pbeCipher.value = 'AES-256-CBC'
  passphrase.value = ''
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="keygen"
    :name="t('tools.keygen.name')"
    name-en="Key Generator"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.keygen.fields.algorithm') }}</span>
            <select
              v-model="algorithm"
              class="field__select"
              @change="resetRun()"
            >
              <option
                v-for="name in ALGORITHMS"
                :key="name"
                :value="name"
              >
                {{ t(`tools.keygen.options.algorithms.${name}`) }}
              </option>
            </select>
          </label>

          <div class="field-row">
            <label
              v-if="isRSA"
              class="field"
            >
              <span class="field__label">{{ t('tools.keygen.fields.bits') }}</span>
              <select
                v-model="bits"
                class="field__select"
              >
                <option
                  v-for="n in BITS"
                  :key="n"
                  :value="n"
                >
                  {{ n }}
                </option>
              </select>
            </label>

            <label
              v-if="isCurve"
              class="field"
            >
              <span class="field__label">{{ t('tools.keygen.fields.curve') }}</span>
              <select
                v-model="curve"
                class="field__select"
              >
                <option
                  v-for="name in CURVES"
                  :key="name"
                  :value="name"
                >
                  {{ name }}
                </option>
              </select>
            </label>
          </div>

          <div class="field-row">
            <label class="field">
              <span class="field__label">{{ t('tools.keygen.fields.privateFormat') }}</span>
              <select
                v-model="privateFormat"
                class="field__select"
              >
                <option
                  v-for="name in PRIVATE_FORMATS"
                  :key="name"
                  :value="name"
                >
                  {{ name }}
                </option>
              </select>
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.keygen.fields.publicFormat') }}</span>
              <select
                v-model="publicFormat"
                class="field__select"
              >
                <option
                  v-for="name in PUBLIC_FORMATS"
                  :key="name"
                  :value="name"
                >
                  {{ name }}
                </option>
              </select>
            </label>
          </div>

          <p class="field__hint">
            {{ t('tools.keygen.hints.format') }}
          </p>
        </template>

        <template #options>
          <label class="field field--inline">
            <input
              v-model="encrypt"
              type="checkbox"
              class="field__checkbox"
            >
            <span>{{ t('tools.keygen.fields.encrypt') }}</span>
          </label>

          <template v-if="encrypt">
            <label class="field">
              <span class="field__label">{{ t('tools.keygen.fields.pbeCipher') }}</span>
              <select
                v-model="pbeCipher"
                class="field__select"
              >
                <option
                  v-for="name in PBE_CIPHERS"
                  :key="name"
                  :value="name"
                >
                  {{ name }}
                </option>
              </select>
            </label>

            <label class="field">
              <span class="field__label">{{ t('tools.keygen.fields.passphrase') }}</span>
              <input
                v-model="passphrase"
                type="password"
                class="field__input"
                autocomplete="off"
                :aria-label="t('tools.keygen.fields.passphrase')"
              >
            </label>
          </template>

          <p class="field__hint">
            {{ t('tools.keygen.hints.encrypt') }}
          </p>
        </template>

        <template #footer>
          <RunBar
            :running="running"
            :disabled="encrypt && passphrase === ''"
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
        :title="t('tools.keygen.results.keys')"
        @retry="run"
      />
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.keygen.description') }}
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
