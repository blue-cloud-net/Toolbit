<script setup lang="ts">
/**
 * Compose 配置生成工具页。
 *
 * 输入只存在于组件局部状态；离开页面即丢弃，**不写**任何浏览器存储。
 */
import { ElInput } from 'element-plus/es/components/input/index'
import 'element-plus/es/components/input/style/css'
import { computed, ref } from 'vue'
import { useI18n } from 'vue-i18n'

import ResultPanel from '@/components/tool/ResultPanel.vue'
import RunBar from '@/components/tool/RunBar.vue'
import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import ToolPanel from '@/components/tool/ToolPanel.vue'
import { useAutoRun } from '@/composables/useAutoRun'
import { useToolRun } from '@/composables/useToolRun'
import { EXAMPLES } from '@/tools/system/docker-compose-gen/example'

/** 一个服务定义。 */
interface ComposeService {
  /** 服务名。 */
  name: string
  /** 镜像名。 */
  image: string
  /** 端口映射。 */
  ports: string[]
  /** 环境变量。 */
  environment: Record<string, string>
  /** 卷挂载。 */
  volumes: string[]
  /** 依赖的服务。 */
  dependsOn: string[]
  /** 重启策略。 */
  restart: string
  /** 启动命令。 */
  command: string
}

/** 后端 docker-compose-gen 工具的响应。 */
interface DockerComposeResult {
  /** 配置内容。 */
  compose: string
  /** 服务数量。 */
  serviceCount: number
  /** 配置提示。 */
  warnings: string[]
}

/** 常见问题条目。 */
interface FaqItem {
  q: string
  a: string
}

const { t, tm } = useI18n()

const projectName = ref('demo')
const servicesJson = ref(
  '[\n  {\n    "name": "web",\n    "image": "nginx:alpine",\n    "ports": ["8080:80"]\n  }\n]',
)

const { state, data, meta, error, run: execute, reset: resetRun } = useToolRun<DockerComposeResult>({
  slug: 'docker-compose-gen',
})

const running = computed(() => state.value === 'loading')

/** 结果文本：配置内容。 */
const resultText = computed(() => data.value?.compose ?? '')

const faq = computed<FaqItem[]>(() => tm('tools.docker-compose-gen.faq') as FaqItem[])

/** 运行当前输入。 */
function run(): void {
  if (projectName.value.trim() === '') {
    return
  }
  let services: ComposeService[]
  try {
    services = JSON.parse(servicesJson.value) as ComposeService[]
  } catch {
    services = []
  }
  void execute({
    projectName: projectName.value,
    services,
    includeNetwork: true,
  })
}

// 输入变化后自动运行（catalog 的 autoRun 为 true 时生效）。
useAutoRun({
  slug: 'docker-compose-gen',
  run,
  watch: [projectName, servicesJson],
  enabled: () => projectName.value.trim() !== '',
})

/** 重置表单与结果。 */
function reset(): void {
  projectName.value = 'demo'
  servicesJson.value =
    '[\n  {\n    "name": "web",\n    "image": "nginx:alpine",\n    "ports": ["8080:80"]\n  }\n]'
  resetRun()
}

/** 填入指定示例。 */
function fillExample(index: number): void {
  const example = EXAMPLES[index]
  if (!example) {
    return
  }
  projectName.value = example.projectName
  resetRun()
}
</script>

<template>
  <ToolPageLayout
    slug="docker-compose-gen"
    :name="t('tools.docker-compose-gen.name')"
    name-en="Docker Compose Generator"
  >
    <template #panel>
      <ToolPanel>
        <template #input>
          <label class="field field--grow">
            <span class="field__label">{{ t('tools.docker-compose-gen.fields.projectName') }}</span>
            <ElInput
              v-model="projectName"
              class="field__input tb-mono"
              autocomplete="off"
              :aria-label="t('tools.docker-compose-gen.fields.projectName')"
              @keydown.enter.prevent="run"
            />
          </label>

          <label class="field field--grow">
            <span class="field__label">{{ t('tools.docker-compose-gen.fields.services') }}</span>
            <ElInput
              v-model="servicesJson"
              class="field__input tb-mono"
              type="textarea"
              :rows="10"
              autocomplete="off"
              :aria-label="t('tools.docker-compose-gen.fields.services')"
            />
          </label>

          <p class="field__hint">
            {{ t('tools.docker-compose-gen.hints.services') }}
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
        :title="t('tools.docker-compose-gen.results.compose')"
        @retry="run"
      >
        <template #extra>
          <dl class="detail">
            <div class="detail__row">
              <dt>{{ t('tools.docker-compose-gen.results.serviceCount') }}</dt>
              <dd class="tb-mono">
                {{ data?.serviceCount ?? 0 }}
              </dd>
            </div>
            <div
              v-for="(warning, index) in data?.warnings ?? []"
              :key="index"
              class="detail__row"
            >
              <dt>!</dt>
              <dd>{{ warning }}</dd>
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
