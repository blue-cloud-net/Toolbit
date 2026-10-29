<script setup lang="ts">
/**
 * 权限计算工具页。
 *
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
 */
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useAutoRun } from '@/composables/useAutoRun'
import { useToolRun } from '@/composables/useToolRun'
import { EXAMPLES, type PermBit, type PermClass } from '@/tools/system/chmod-calc/example'

/** 一个权限类别的明细。 */
interface PermClassInfo {
  /** 类别名：user / group / other。 */
  class: string
  /** 是否可读。 */
  read: boolean
  /** 是否可写。 */
  write: boolean
  /** 是否可执行。 */
  execute: boolean
  /** 符号形式的三个字符。 */
  symbol: string
  /** 该类别的数字形式（4 / 2 / 1）。 */
  numeric: number
  /** 文字说明。 */
  meaning: string
}

/** 后端 chmod-calc 工具的响应。 */
interface ChmodCalcResult {
  /** 八进制形式，如 `755`。 */
  octal: string
  /** 符号形式，如 `rwxr-xr-x`。 */
  symbolic: string
  /** 三个权限类别的明细。 */
  classes: PermClassInfo[]
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

/** 受支持的权限类别；顺序与服务端一致。 */
const CLASSES: readonly PermClass[] = ['user', 'group', 'other']

/** 受支持的权限位；顺序与服务端一致。 */
const BITS: readonly PermBit[] = ['read', 'write', 'execute']

const mode = ref('755')
const user = ref<PermBit[]>([])
const group = ref<PermBit[]>([])
const other = ref<PermBit[]>([])

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<ChmodCalcResult>({
  slug: 'chmod-calc',
})

const running = computed(() => state.value === 'loading')

/** 结果文本：八进制形式。 */
const resultText = computed(() => data.value?.octal ?? '')

/** 三个权限类别的明细。 */
const classes = computed<PermClassInfo[]>(() => data.value?.classes ?? [])

const faq = computed<FaqItem[]>(() => tm('tools.chmod-calc.faq') as FaqItem[])

/** 类别名的显示文案。 */
function classLabel(value: string): string {
  return t(`tools.chmod-calc.options.class.${value}`)
}

/** 某类别当前勾选的权限位。 */
function bitsFor(value: PermClass): PermBit[] {
  if (value === 'user') {
    return user.value
  }
  if (value === 'group') {
    return group.value
  }
  return other.value
}

/** 切换某类别的一个权限位。 */
function toggleBit(value: PermClass, bit: PermBit): void {
  const list = bitsFor(value)
  const index = list.indexOf(bit)
  if (index >= 0) {
    list.splice(index, 1)
  } else {
    list.push(bit)
  }
}

/** 某类别是否具备某个权限位。 */
function bitOn(item: PermClassInfo, bit: PermBit): boolean {
  if (bit === 'read') {
    return item.read
  }
  if (bit === 'write') {
    return item.write
  }
  return item.execute
}

/** 权限说明的显示文案；词条缺失时回退到服务端返回的说明。 */
function meaningText(item: PermClassInfo): string {
  const key = `tools.chmod-calc.options.numeric.${item.numeric}`
  const translated = t(key)
  return translated === key ? item.meaning : translated
}

/** 运行当前输入。 */
function run(): void {
  void execute({
    mode: mode.value,
    user: user.value,
    group: group.value,
    other: other.value,
  })
}

/** 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。 */
useAutoRun({
  slug: 'chmod-calc',
  run,
  watch: [mode, user, group, other],
})

/** 重置表单与结果。 */
function reset(): void {
  mode.value = '755'
  user.value = []
  group.value = []
  other.value = []
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  mode.value = example.mode
  user.value = [...example.user]
  group.value = [...example.group]
  other.value = [...example.other]
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="chmod-calc"
    :name="t('tools.chmod-calc.name')"
    name-en="Permission Calculator"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.chmod-calc.fields.mode') }}</span>
            <input
              v-model="mode"
              type="text"
              class="field__input tb-mono"
              autocomplete="off"
              placeholder="755"
              :aria-label="t('tools.chmod-calc.fields.mode')"
            >
          </label>

          <p class="field__hint">
            {{ t('tools.chmod-calc.hints.mode') }}
          </p>
        </template>

        <template #options>
          <fieldset
            v-for="item in CLASSES"
            :key="item"
            class="field"
          >
            <legend class="field__label">
              {{ classLabel(item) }}
            </legend>
            <div class="bit-list">
              <label
                v-for="bit in BITS"
                :key="bit"
                class="field field--inline"
              >
                <input
                  type="checkbox"
                  class="field__checkbox"
                  :checked="bitsFor(item).includes(bit)"
                  @change="toggleBit(item, bit)"
                >
                <span>{{ t(`tools.chmod-calc.options.bit.${bit}`) }}</span>
              </label>
            </div>
          </fieldset>

          <p class="field__hint">
            {{ t('tools.chmod-calc.hints.bits') }}
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
        :title="t('tools.chmod-calc.results.octal')"
        @retry="run"
      >
        <template
          v-if="data"
          #extra
        >
          <dl class="detail">
            <div class="detail__row">
              <dt>{{ t('tools.chmod-calc.results.symbolic') }}</dt>
              <dd class="tb-mono">
                {{ data?.symbolic }}
              </dd>
            </div>
          </dl>

          <ul class="classes">
            <li
              v-for="item in classes"
              :key="item.class"
              class="perm"
            >
              <div class="perm__head">
                <span class="perm__class">
                  {{ classLabel(item.class) }}
                </span>
                <span class="perm__symbol tb-mono">
                  {{ item.symbol }}
                </span>
                <span class="perm__numeric tb-mono">
                  {{ item.numeric }}
                </span>
              </div>

              <div class="perm__bits">
                <span
                  v-for="bit in BITS"
                  :key="bit"
                  class="perm__bit"
                  :class="{ 'perm__bit--on': bitOn(item, bit) }"
                >
                  {{ t(`tools.chmod-calc.options.bit.${bit}`) }}
                </span>
              </div>

              <p class="perm__meaning">
                {{ meaningText(item) }}
              </p>
            </li>
          </ul>
        </template>
      </ResultPanel>
    </template>

    <template #usage>
      <p class="prose">
        {{ t('tools.chmod-calc.hints.output') }}
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

<style scoped>
.bit-list {
  display: flex;
  flex-wrap: wrap;
  gap: var(--tb-space-2) var(--tb-space-4);
}

.detail {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-1);
  margin: var(--tb-space-3) 0 0;
}

.detail__row {
  display: flex;
  gap: var(--tb-space-2);
  font-size: var(--tb-font-sm);
}

.detail__row dt {
  min-width: 96px;
  color: var(--tb-text-tertiary);
}

.detail__row dd {
  margin: 0;
  word-break: break-all;
}

.classes {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-2);
  margin: var(--tb-space-3) 0 0;
  padding: 0;
  list-style: none;
}

.perm {
  padding: var(--tb-space-2) var(--tb-space-3);
  background: var(--tb-bg-sunken);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-control);
}

.perm__head {
  display: flex;
  flex-wrap: wrap;
  gap: var(--tb-space-2);
  font-size: var(--tb-font-sm);
}

.perm__class {
  color: var(--tb-text);
}

.perm__symbol {
  color: var(--tb-text-secondary);
}

.perm__numeric {
  color: var(--tb-text-tertiary);
}

.perm__bits {
  display: flex;
  flex-wrap: wrap;
  gap: var(--tb-space-1) var(--tb-space-2);
  margin-top: var(--tb-space-1);
}

.perm__bit {
  padding: 0 var(--tb-space-2);
  font-size: var(--tb-font-xs);
  color: var(--tb-text-tertiary);
  background: var(--tb-bg-elevated);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-pill);
}

.perm__bit--on {
  color: var(--tb-success);
  border-color: color-mix(in srgb, var(--tb-success) 40%, var(--tb-border));
}

.perm__meaning {
  margin: var(--tb-space-1) 0 0;
  font-size: var(--tb-font-xs);
  color: var(--tb-text-secondary);
}

.prose {
  margin: 0;
  font-size: var(--tb-font-sm);
  color: var(--tb-text-secondary);
}

.faq {
  margin: 0;
  font-size: var(--tb-font-sm);
}

.faq__q {
  margin-top: var(--tb-space-2);
  font-weight: 600;
  color: var(--tb-text);
}

.faq__a {
  margin: var(--tb-space-1) 0 0;
  color: var(--tb-text-secondary);
}
</style>
