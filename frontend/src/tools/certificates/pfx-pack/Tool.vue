<script setup lang="ts">
/**
 * PKCS#12 打包工具页。
 *
 * 证书与私钥只存在于组件局部状态；离开页面即丢弃，**不写**浏览器存储。
 */
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import CertTextInput from '@/components/tool/CertTextInput.vue'
import KeyInput from '@/components/tool/KeyInput.vue'
import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useToolRun } from '@/composables/useToolRun'
import {
  TEST_CA_CERT_PEM,
  TEST_LEAF_CERT_PEM,
  TEST_LEAF_KEY_PEM,
} from '@/tools/certificates/fixtures-v060'

/** 后端 pfx-pack 工具的响应。 */
interface PFXOutput {
  base64: string
  length: number
  encoding: string
}

interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const cert = ref('')
const certFormat = ref('auto')
const key = ref('')
const keyFormat = ref('pem')
const keyPassphrase = ref('')
const caCerts = ref('')
const caCertsFormat = ref('auto')
const passphrase = ref('')
const friendlyName = ref('')

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
} = useToolRun<PFXOutput>({ slug: 'pfx-pack' })

const running = computed(() => state.value === 'loading')

const resultText = computed(() => (data.value ? data.value.base64 : ''))
const faq = computed<FaqItem[]>(() => tm('tools.pfx-pack.faq') as FaqItem[])

function run(): void {
  void execute({
    cert: { data: cert.value, format: certFormat.value },
    key: {
      data: key.value,
      format: keyFormat.value,
      passphrase: keyPassphrase.value,
    },
    caCerts: caCerts.value.trim()
      ? [{ data: caCerts.value, format: caCertsFormat.value }]
      : [],
    passphrase: passphrase.value,
    friendlyName: friendlyName.value,
  })
}

function reset(): void {
  cert.value = ''
  certFormat.value = 'auto'
  key.value = ''
  keyFormat.value = 'pem'
  keyPassphrase.value = ''
  caCerts.value = ''
  caCertsFormat.value = 'auto'
  passphrase.value = ''
  friendlyName.value = ''
  resetRun()
}

function fillExample(): void {
  reset()
  cert.value = TEST_LEAF_CERT_PEM
  key.value = TEST_LEAF_KEY_PEM
  caCerts.value = TEST_CA_CERT_PEM
  passphrase.value = 'test-pass'
  friendlyName.value = 'toolbit-test'
}
</script>

<template>
  <ToolPageLayout
    slug="pfx-pack"
    name="PKCS#12 打包"
    name-en="PKCS#12 Packer"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <CertTextInput
            v-model="cert"
            v-model:format="certFormat"
            :label="t('tools.pfx-pack.fields.cert')"
            :rows="6"
          />
          <KeyInput
            v-model="key"
            v-model:format="keyFormat"
            v-model:passphrase="keyPassphrase"
            :label="t('tools.pfx-pack.fields.key')"
            :formats="['pem', 'base64', 'hex']"
            show-passphrase
          />
          <CertTextInput
            v-model="caCerts"
            v-model:format="caCertsFormat"
            :label="t('tools.pfx-pack.fields.caCerts')"
            :rows="5"
          />
        </template>

        <template #options>
          <label class="field">
            <span class="field__label">{{ t('tools.pfx-pack.fields.passphrase') }}</span>
            <ElInput
              v-model="passphrase"
              type="password"
              class="field__input"
              autocomplete="new-password"
            />
          </label>
          <label class="field">
            <span class="field__label">{{ t('tools.pfx-pack.fields.friendlyName') }}</span>
            <ElInput
              v-model="friendlyName"
              class="field__input"
              spellcheck="false"
              autocomplete="off"
            />
          </label>
        </template>

        <template #footer>
          <RunBar
            :running="running"
            :disabled="cert === '' || key === ''"
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
        :title="t('tools.pfx-pack.results.output')"
        @retry="run"
      />
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.pfx-pack.description') }}
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
