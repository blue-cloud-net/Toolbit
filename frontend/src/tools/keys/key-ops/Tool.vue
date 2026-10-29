<script setup lang="ts">
/**
 * 密钥运算工具页。
 *
 * 密钥内容只存在于组件局部状态；离开页面即丢弃，
 * **不写**任何浏览器存储（AGENTS.md §4.11）。
 */
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { ElOption, ElSelect } from 'element-plus/es/components/select/index'
import 'element-plus/es/components/select/style/css'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import KeyInput from '@/components/tool/KeyInput.vue'
import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useToolRun } from '@/composables/useToolRun'
import { DEFAULT_EXAMPLE } from '@/tools/keys/key-ops/example'

/** 后端 key-ops 工具的响应。 */
interface KeyOpsResult {
  key: { pem: string; derBase64?: string }
  algorithm: string
  format: string
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const ACTIONS = ['encrypt', 'decrypt', 'extract-public', 'change-passphrase'] as const
const PBE_CIPHERS = ['AES-256-CBC', 'DES-EDE3-CBC'] as const

const action = ref<string>('extract-public')
const key = ref('')
const keyFormat = ref('auto')
const passphrase = ref('')
const newPassphrase = ref('')
const pbeCipher = ref<string>('AES-256-CBC')

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
} = useToolRun<KeyOpsResult>({ slug: 'key-ops' })

const running = computed(() => state.value === 'loading')

/** 是否需要新口令。 */
const needsNewPassphrase = computed(() => action.value === 'change-passphrase')

/** 是否需要口令。 */
const needsPassphrase = computed(() => action.value !== 'extract-public')

const resultText = computed(() => data.value?.key.pem ?? '')

const faq = computed<FaqItem[]>(() => tm('tools.key-ops.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  void execute({
    action: action.value,
    key: { data: key.value, format: keyFormat.value },
    passphrase: passphrase.value,
    newPassphrase: newPassphrase.value,
    pbeCipher: pbeCipher.value,
  })
}

/** 重置表单与结果。 */
function reset(): void {
  action.value = 'extract-public'
  key.value = ''
  keyFormat.value = 'auto'
  passphrase.value = ''
  newPassphrase.value = ''
  pbeCipher.value = 'AES-256-CBC'
  resetRun()
}

/** 填入示例。 */
function fillExample(): void {
  action.value = DEFAULT_EXAMPLE.action
  key.value = DEFAULT_EXAMPLE.key
  keyFormat.value = 'auto'
  passphrase.value = ''
  newPassphrase.value = ''
  pbeCipher.value = 'AES-256-CBC'
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="key-ops"
    :name="t('tools.key-ops.name')"
    name-en="Key Operations"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <KeyInput
            v-model="key"
            v-model:format="keyFormat"
            :label="t('tools.key-ops.fields.key')"
          />
        </template>

        <template #options>
          <label class="field">
            <span class="field__label">{{ t('tools.key-ops.fields.action') }}</span>
            <ElSelect
              v-model="action"
              class="field__select"
              :aria-label="t('tools.key-ops.fields.action')"
              @change="resetRun()"
            >
              <ElOption
                v-for="name in ACTIONS"
                :key="name"
                :label="t(`tools.key-ops.options.actions.${name}`)"
                :value="name"
              />
            </ElSelect>
          </label>

          <label
            v-if="needsPassphrase"
            class="field"
          >
            <span class="field__label">{{ t('tools.key-ops.fields.passphrase') }}</span>
            <ElInput
              v-model="passphrase"
              type="password"
              class="field__input"
              autocomplete="off"
              :aria-label="t('tools.key-ops.fields.passphrase')"
            />
          </label>

          <label
            v-if="needsNewPassphrase"
            class="field"
          >
            <span class="field__label">{{ t('tools.key-ops.fields.newPassphrase') }}</span>
            <ElInput
              v-model="newPassphrase"
              type="password"
              class="field__input"
              autocomplete="off"
              :aria-label="t('tools.key-ops.fields.newPassphrase')"
            />
          </label>

          <label class="field">
            <span class="field__label">{{ t('tools.key-ops.fields.pbeCipher') }}</span>
            <ElSelect
              v-model="pbeCipher"
              class="field__select"
              :aria-label="t('tools.key-ops.fields.pbeCipher')"
            >
              <ElOption
                v-for="name in PBE_CIPHERS"
                :key="name"
                :label="String(name)"
                :value="name"
              />
            </ElSelect>
          </label>

          <p class="field__hint">
            {{ t('tools.key-ops.hints.action') }}
          </p>
        </template>

        <template #footer>
          <RunBar
            :running="running"
            :disabled="key === '' || (needsPassphrase && passphrase === '')"
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
        :title="t('tools.key-ops.results.key')"
        @retry="run"
      />
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.key-ops.description') }}
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
