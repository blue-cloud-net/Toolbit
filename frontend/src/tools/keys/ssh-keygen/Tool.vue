<script setup lang="ts">
/**
 * SSH 密钥生成工具页。
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
import { EXAMPLES } from '@/tools/keys/ssh-keygen/example'

/** 后端 ssh-keygen 工具的响应。 */
interface SSHKeygenResult {
  /** 算法。 */
  algorithm: string
  /** 私钥（OpenSSH 格式）。 */
  privateKey: string
  /** ���钥（OpenSSH 格式）。 */
  publicKey: string
  /** authorized_keys 行。 */
  authorizedKey: string
  /** 指纹。 */
  fingerprint: string
  /** 注释。 */
  comment: string
  /** 密钥位数。 */
  bits: number
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const algorithm = ref('Ed25519')
const bits = ref(3072)
const comment = ref('toolbit@example.com')

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<SSHKeygenResult>({
  slug: 'ssh-keygen',
})

const running = computed(() => state.value === 'loading')

/** 结果文本：私钥。 */
const resultText = computed(() => data.value?.privateKey ?? '')

const faq = computed<FaqItem[]>(() => tm('tools.ssh-keygen.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  void execute({
    algorithm: algorithm.value,
    bits: bits.value,
    comment: comment.value,
  })
}

// 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。
useAutoRun({
  slug: 'ssh-keygen',
  run,
  watch: [algorithm, bits, comment],
  enabled: () => true,
})

/** 重置表单与结果。 */
function reset(): void {
  algorithm.value = 'Ed25519'
  bits.value = 3072
  comment.value = 'toolbit@example.com'
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  algorithm.value = example.algorithm
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="ssh-keygen"
    :name="t('tools.ssh-keygen.name')"
    name-en="SSH Key Generator"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.ssh-keygen.fields.algorithm') }}</span>
            <ElSelect
              v-model="algorithm"
              class="field__select"
              :aria-label="t('tools.ssh-keygen.fields.algorithm')"
            >
              <ElOption
                :label="t('tools.ssh-keygen.options.algorithm.Ed25519')"
                value="Ed25519"
              />
              <ElOption
                :label="t('tools.ssh-keygen.options.algorithm.RSA')"
                value="RSA"
              />
            </ElSelect>
          </label>

          <label
            v-if="algorithm === 'RSA'"
            class="field"
          >
            <span class="field__label">{{ t('tools.ssh-keygen.fields.bits') }}</span>
            <ElInputNumber
              v-model="bits"
              class="field__select"
              :min="2048"
              :max="8192"
              :step="1024"
              :aria-label="t('tools.ssh-keygen.fields.bits')"
            />
          </label>

          <label class="field field--grow">
            <span class="field__label">{{ t('tools.ssh-keygen.fields.comment') }}</span>
            <ElInput
              v-model="comment"
              class="field__input tb-mono"
              autocomplete="off"
              :aria-label="t('tools.ssh-keygen.fields.comment')"
              @keydown.enter.prevent="run"
            />
          </label>

          <p class="field__hint">
            {{ t('tools.ssh-keygen.hints.privateKey') }}
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
        :title="t('tools.ssh-keygen.results.privateKey')"
        @retry="run"
      >
        <template #extra>
          <dl class="detail">
            <div class="detail__row">
              <dt>{{ t('tools.ssh-keygen.results.fingerprint') }}</dt>
              <dd class="tb-mono">
                {{ data?.fingerprint ?? '' }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.ssh-keygen.results.authorizedKey') }}</dt>
              <dd class="tb-mono">
                {{ data?.authorizedKey ?? '' }}
              </dd>
            </div>
            <div class="detail__row">
              <dt>{{ t('tools.ssh-keygen.results.publicKey') }}</dt>
              <dd class="tb-mono">
                {{ data?.publicKey ?? '' }}
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
