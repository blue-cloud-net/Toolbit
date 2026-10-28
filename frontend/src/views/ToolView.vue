<script setup lang="ts">
/**
 * 工具页路由容器。
 *
 * 负责：解析目录元信息 → 懒加载工具组件 → 登记工具词条。
 */
import { computed, ref, shallowRef, watch, type Component } from 'vue'
import { useI18n } from 'vue-i18n'

import ToolPageLayout from '@/components/tool/ToolPageLayout.vue'
import { registerToolMessages, type Locale } from '@/i18n'
import { useCatalogStore } from '@/stores/catalog'
import { usePrefsStore } from '@/stores/prefs'
import { getToolLoader, isImplemented } from '@/tools'

const props = defineProps<{
  /** 工具 slug。 */
  slug: string
}>()

const { t, locale } = useI18n()
const catalog = useCatalogStore()
const prefs = usePrefsStore()

const toolComponent = shallowRef<Component | null>(null)
const loading = ref(false)

const meta = computed(() => catalog.getTool(props.slug))
const implemented = computed(() => isImplemented(props.slug))
const known = computed(() => meta.value !== undefined)

const displayName = computed(() => {
  const m = meta.value
  if (!m) {
    return props.slug
  }
  return locale.value === 'en-US' ? m.nameEn : m.name
})

const displayNameEn = computed(() => meta.value?.nameEn ?? props.slug)

/** 工具词条的加载器（模块内静态 import 了双语 JSON）。 */
interface ToolModule {
  default: Component
  messages?: Record<Locale, Record<string, unknown>>
}

/** 加载工具组件与词条。 */
async function load(): Promise<void> {
  if (!implemented.value) {
    toolComponent.value = null
    return
  }
  const loader = getToolLoader(props.slug)
  if (!loader) {
    toolComponent.value = null
    return
  }

  loading.value = true
  try {
    const mod = (await loader()) as unknown as ToolModule
    if (mod.messages) {
      for (const lang of Object.keys(mod.messages) as Locale[]) {
        registerToolMessages(lang, props.slug, mod.messages[lang])
      }
    }
    toolComponent.value = mod.default
  } finally {
    loading.value = false
  }
}

watch(
  () => props.slug,
  () => {
    toolComponent.value = null
    if (implemented.value) {
      prefs.recordRecent(props.slug)
    }
    void load()
  },
  { immediate: true },
)
</script>

<template>
  <div>
    <p
      v-if="!catalog.loaded && catalog.loading"
      class="tool-view__loading"
    >
      {{ t('common.states.loading') }}…
    </p>

    <div
      v-else-if="catalog.loaded && !known"
      class="tool-view__missing"
    >
      <h1 class="tool-view__missing-title">
        {{ t('common.states.notFound') }}
      </h1>
      <p class="tool-view__missing-hint">
        {{ t('common.states.notFoundHint') }}
      </p>
      <RouterLink :to="{ name: 'home' }">
        {{ t('common.actions.backHome') }}
      </RouterLink>
    </div>

    <p
      v-else-if="implemented && loading"
      class="tool-view__loading"
    >
      {{ t('common.states.loading') }}…
    </p>

    <component
      :is="toolComponent"
      v-else-if="implemented && toolComponent"
      :key="slug"
    />

    <ToolPageLayout
      v-else
      :slug="slug"
      :name="displayName"
      :name-en="displayNameEn"
      :wide="meta?.wide ?? false"
    />
  </div>
</template>

<style scoped>
.tool-view__loading {
  padding: var(--tb-space-8) 0;
  font-size: var(--tb-font-sm);
  color: var(--tb-text-tertiary);
  text-align: center;
}

.tool-view__missing {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--tb-space-2);
  padding: var(--tb-space-12) var(--tb-space-4);
  text-align: center;
}

.tool-view__missing-title {
  margin: 0;
  font-size: var(--tb-font-xl);
}

.tool-view__missing-hint {
  margin: 0;
  color: var(--tb-text-secondary);
}
</style>
