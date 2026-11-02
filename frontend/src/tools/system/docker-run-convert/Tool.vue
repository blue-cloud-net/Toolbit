<script setup lang="ts">
/**
 * docker run 与 Compose 互转工具页。
 *
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
 */
import { ElCheckbox } from 'element-plus/es/components/checkbox/index'
import 'element-plus/es/components/checkbox/style/css'
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { ElOption, ElSelect } from 'element-plus/es/components/select/index'
import 'element-plus/es/components/select/style/css'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import DataTable, { type DataTableColumn, type DataTableRow } from '@/components/tool/DataTable.vue'
import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import StatGrid, { type StatGridItem } from '@/components/tool/StatGrid.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useAutoRun } from '@/composables/useAutoRun'
import { useToolRun } from '@/composables/useToolRun'
import { EXAMPLES } from '@/tools/system/docker-run-convert/example'

/** 转换方向。 */
type ConvertDirection = 'runToCompose' | 'composeToRun'

/** 转换出的单个服务。 */
interface ConvertedService {
  /** 服务名。 */
  name: string
  /** 镜像。 */
  image: string
  /** 容器内执行的命令。 */
  exec: string
  /** 生成的 docker run 命令。 */
  runCommand: string
}

/** 后端 docker-run-convert 工具的响应。 */
interface DockerRunConvertResult {
  /** 转换方向。 */
  direction: ConvertDirection
  /** 转换结果文本。 */
  output: string
  /** 镜像。 */
  image: string
  /** 服务名。 */
  serviceName: string
  /** 服务列表。 */
  services: ConvertedService[]
  /** 未转换的项。 */
  unsupported: string[]
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const direction = ref<ConvertDirection>('runToCompose')
const serviceName = ref('')
const input = ref('docker run -d --name web -p 8080:80 -e TZ=Asia/Shanghai nginx:alpine')
const strict = ref(false)

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<DockerRunConvertResult>({
  slug: 'docker-run-convert',
})

const running = computed(() => state.value === 'loading')

/** 是否为 docker run → Compose 方向。 */
const isRunToCompose = computed(() => direction.value === 'runToCompose')

/** 结果文本：转换产物。 */
const resultText = computed(() => data.value?.output ?? '')

const faq = computed<FaqItem[]>(() => tm('tools.docker-run-convert.faq') as FaqItem[])

/** 方向的中文名。 */
function directionLabel(value: ConvertDirection): string {
  return value === 'composeToRun'
    ? t('tools.docker-run-convert.options.direction.composeToRun')
    : t('tools.docker-run-convert.options.direction.runToCompose')
}

/** 摘要卡片：方向、镜像与服务名（或服务数量）。 */
const stats = computed<StatGridItem[]>(() => {
  const result = data.value
  if (!result) {
    return []
  }
  const directionItem: StatGridItem = {
    label: t('tools.docker-run-convert.results.direction'),
    value: directionLabel(result.direction),
  }
  if (result.direction === 'composeToRun') {
    return [
      directionItem,
      {
        label: t('tools.docker-run-convert.results.serviceCount'),
        value: String(result.services.length),
      },
    ]
  }
  return [
    directionItem,
    { label: t('tools.docker-run-convert.results.image'), value: result.image },
    { label: t('tools.docker-run-convert.results.serviceName'), value: result.serviceName },
  ]
})

/** 表格列定义：Compose → docker run 方向按服务逐条展示。 */
const columns = computed<DataTableColumn[]>(() => [
  { key: 'name', label: t('tools.docker-run-convert.results.service') },
  { key: 'image', label: t('tools.docker-run-convert.results.image'), mono: true },
  { key: 'runCommand', label: t('tools.docker-run-convert.results.command'), mono: true },
])

/** 表格行。 */
const rows = computed<DataTableRow[]>(() =>
  (data.value?.services ?? []).map((service) => ({
    key: service.name,
    cells: {
      name: service.name,
      image: service.image,
      runCommand: service.runCommand,
    },
  })),
)

/** 运行当前输入。 */
function run(): void {
  if (input.value.trim() === '') {
    return
  }
  void execute({
    direction: direction.value,
    input: input.value,
    serviceName: serviceName.value,
    strict: strict.value,
  })
}

// 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。
useAutoRun({
  slug: 'docker-run-convert',
  run,
  watch: [direction, serviceName, input, strict],
  enabled: () => input.value.trim() !== '',
})

/** 重置表单与结果。 */
function reset(): void {
  direction.value = 'runToCompose'
  serviceName.value = ''
  input.value = 'docker run -d --name web -p 8080:80 -e TZ=Asia/Shanghai nginx:alpine'
  strict.value = false
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  direction.value = example.direction
  input.value = example.input
  resetRun()
}

/** 把结果回填为新输入，并切换到相反方向。 */
function useResultAsInput(text: string): boolean {
  const result = data.value
  if (!result) {
    return false
  }
  if (result.direction === 'composeToRun') {
    const first = result.services[0]
    if (!first) {
      return false
    }
    input.value = first.runCommand
    direction.value = 'runToCompose'
    return true
  }
  input.value = text
  direction.value = 'composeToRun'
  return true
}
</script>

<template>
  <ToolPageLayout
    slug="docker-run-convert"
    :name="t('tools.docker-run-convert.name')"
    name-en="Docker Run / Compose Converter"
    wide
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field">
            <span class="field__label">{{ t('tools.docker-run-convert.fields.direction') }}</span>
            <ElSelect
              v-model="direction"
              class="field__select"
              :aria-label="t('tools.docker-run-convert.fields.direction')"
            >
              <ElOption
                :label="t('tools.docker-run-convert.options.direction.runToCompose')"
                value="runToCompose"
              />
              <ElOption
                :label="t('tools.docker-run-convert.options.direction.composeToRun')"
                value="composeToRun"
              />
            </ElSelect>
          </label>

          <label
            v-if="isRunToCompose"
            class="field"
          >
            <span class="field__label">{{ t('tools.docker-run-convert.fields.serviceName') }}</span>
            <ElInput
              v-model="serviceName"
              class="field__input tb-mono"
              autocomplete="off"
              :aria-label="t('tools.docker-run-convert.fields.serviceName')"
              @keydown.enter.prevent="run"
            />
          </label>

          <p
            v-if="isRunToCompose"
            class="field__hint"
          >
            {{ t('tools.docker-run-convert.hints.serviceName') }}
          </p>

          <label class="field field--grow">
            <span class="field__label">{{ t('tools.docker-run-convert.fields.input') }}</span>
            <ElInput
              v-model="input"
              class="field__input tb-mono"
              type="textarea"
              :rows="12"
              autocomplete="off"
              :aria-label="t('tools.docker-run-convert.fields.input')"
            />
          </label>
        </template>

        <template #options>
          <div class="field">
            <span class="field__label">{{ t('tools.docker-run-convert.fields.strict') }}</span>
            <ElCheckbox
              v-model="strict"
              class="field__checkbox"
            >
              {{ t('tools.docker-run-convert.hints.strict') }}
            </ElCheckbox>
          </div>

          <p class="field__hint">
            {{ t('tools.docker-run-convert.hints.input') }}
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
        :title="t('tools.docker-run-convert.results.output')"
        :use-as-input="useResultAsInput"
        :hide-text="!isRunToCompose"
        @retry="run"
      >
        <template #extra>
          <StatGrid :items="stats" />

          <DataTable
            v-if="!isRunToCompose"
            class="table"
            :columns="columns"
            :rows="rows"
            :copy-columns="['runCommand']"
            :empty="t('tools.docker-run-convert.results.empty')"
          />

          <dl
            v-if="(data?.unsupported ?? []).length > 0"
            class="detail"
          >
            <div
              v-for="(item, index) in data?.unsupported ?? []"
              :key="index"
              class="detail__row"
            >
              <dt>!</dt>
              <dd class="tb-mono">
                {{ item }}
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

<style scoped>
.table {
  margin-top: var(--tb-space-3);
}
</style>
