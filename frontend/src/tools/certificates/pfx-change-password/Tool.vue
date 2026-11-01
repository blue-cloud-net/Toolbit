<script setup lang="ts">
/**
 * PKCS#12 改密工具页。
 *
 * 容器与口令只存在于组件局部状态；离开页面即丢弃，**不写**浏览器存储。
 */
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import MaterialInput from '@/components/tool/MaterialInput.vue'
import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useToolRun } from '@/composables/useToolRun'
import { TEST_P12_BASE64 } from '@/tools/certificates/fixtures-v060'

/** 后端 pfx-change-password 工具的响应。 */
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

const container = ref('')
const containerFormat = ref('auto')
const oldPassphrase = ref('')
const newPassphrase = ref('')

const {
  state,
  data,
  meta,
  error,
  run: execute,
  reset: resetRun,
} = useToolRun<PFXOutput>({ slug: 'pfx-change-password' })

const running = computed(() => state.value === 'loading')

const resultText = computed(() => (data.value ? data.value.base64 : ''))
const faq = computed<FaqItem[]>(
  () => tm('tools.pfx-change-password.faq') as FaqItem[],
)

function run(): void {
  void execute({
    container: { data: container.value, format: containerFormat.value },
    oldPassphrase: oldPassphrase.value,
    newPassphrase: newPassphrase.value,
  })
}

function reset(): void {
  container.value = ''
  containerFormat.value = 'auto'
  oldPassphrase.value = ''
  newPassphrase.value = ''
  resetRun()
}

function fillExample(): void {
  reset()
  container.value = TEST_P12_BASE64
  containerFormat.value = 'base64'
  oldPassphrase.value = 'test-pass'
  newPassphrase.value = 'new-test-pass'
}
</script>

<template>
  <ToolPageLayout
    slug="pfx-change-password"
    name="PKCS#12 改密"
    name-en="PKCS#12 Password Changer"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <MaterialInput
            v-model="container"
            v-model:format="containerFormat"
            :label="t('tools.pfx-change-password.fields.container')"
            :rows="8"
          />
        </template>

        <template #options>
          <label class="field">
            <span class="field__label">{{ t('tools.pfx-change-password.fields.oldPassphrase') }}</span>
            <ElInput
              v-model="oldPassphrase"
              type="password"
              class="field__input"
              autocomplete="off"
            />
          </label>
          <label class="field">
            <span class="field__label">{{ t('tools.pfx-change-password.fields.newPassphrase') }}</span>
            <ElInput
              v-model="newPassphrase"
              type="password"
              class="field__input"
              autocomplete="new-password"
            />
          </label>
        </template>

        <template #footer>
          <RunBar
            :running="running"
            :disabled="container === '' || oldPassphrase === '' || newPassphrase === ''"
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
        :title="t('tools.pfx-change-password.results.output')"
        @retry="run"
      />
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.pfx-change-password.description') }}
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
