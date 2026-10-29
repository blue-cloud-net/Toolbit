<script setup lang="ts">
/**
 * 工具卡片。
 *
 * 未实现的工具卡片仍可点击，但标注「即将上线」并在工具页显示说明。
 */
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

import AppIcon from '@/components/common/AppIcon.vue'
import { useCatalogStore } from '@/stores/catalog'
import { getToolLoader, isImplemented } from '@/tools'
import { categoryIcon } from '@/tools/categories'
import { toneStyle, toolTone } from '@/tools/tones'
import type { ToolMeta } from '@/types/tools'

const props = defineProps<{
  /** 工具元信息。 */
  tool: ToolMeta
}>()

const { t, locale } = useI18n()
const catalog = useCatalogStore()

const implemented = computed(() => isImplemented(props.tool.slug))

/** 所属分类的图标名。 */
const iconName = computed(() => categoryIcon(catalog.categoryMap.get(props.tool.category)?.icon))

/** 图标色调：按工具登记，未登记时回落到所属分类。 */
const tone = computed(() => toolTone(props.tool.slug, props.tool.category))

/** 色调对应的 CSS 变量（前景色 + 底色）。 */
const toneVars = computed(() => toneStyle(tone.value))

/** 已预取过的 slug（模块级，避免重复发起请求）。 */
const prefetched = new Set<string>()

/** 预取工具 chunk（规范 §11：卡片 hover 应当预取）。 */
function prefetch(): void {
  if (prefetched.has(props.tool.slug)) {
    return
  }
  prefetched.add(props.tool.slug)
  const loader = getToolLoader(props.tool.slug)
  if (!loader) {
    return
  }
  void loader().catch(() => {
    // 预取失败不影响正常导航。
  })
}
const displayName = computed(() => (locale.value === 'en-US' ? props.tool.nameEn : props.tool.name))
const displayDesc = computed(() => {
  const key = `tools.${props.tool.slug}.description`
  const translated = t(key)
  return translated === key ? '' : translated
})

/** 无词条时用英文名兜底，但避免与标题重复。 */
const fallbackText = computed(() =>
  props.tool.nameEn === displayName.value ? '' : props.tool.nameEn,
)
</script>

<template>
  <RouterLink
    class="tool-card"
    :class="{ 'tool-card--soon': !implemented }"
    :to="{ name: 'tool', params: { slug: tool.slug } }"
    @mouseenter="prefetch"
    @focusin="prefetch"
  >
    <span
      class="tool-card__icon"
      :style="toneVars"
      aria-hidden="true"
    >
      <AppIcon
        :name="iconName"
        :size="18"
      />
    </span>

    <div class="tool-card__body">
      <div class="tool-card__head">
        <h3 class="tool-card__title">
          {{ displayName }}
        </h3>
        <span
          v-if="!implemented"
          class="tool-card__badge"
        >{{ t('home.soon') }}</span>
      </div>
      <p
        v-if="displayDesc"
        class="tool-card__desc"
      >
        {{ displayDesc }}
      </p>
      <p
        v-else-if="fallbackText"
        class="tool-card__desc tool-card__desc--muted"
      >
        {{ fallbackText }}
      </p>
    </div>
  </RouterLink>
</template>

<style scoped>
.tool-card {
  display: flex;
  flex-direction: row;
  align-items: flex-start;
  gap: var(--tb-space-3);
  padding: var(--tb-space-3) var(--tb-space-4);
  background: var(--tb-bg-elevated);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-card);
  transition: border-color var(--tb-transition), transform var(--tb-transition);
}

.tool-card__icon {
  display: grid;
  place-items: center;
  flex: none;
  width: 36px;
  height: 36px;
  color: var(--tb-tone, var(--tb-primary));
  background: var(--tb-tone-soft, var(--tb-primary-soft));
  border-radius: var(--tb-radius-control);
}

.tool-card__body {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-1);
  min-width: 0;
  flex: 1;
}

/* hover 时卡片随自身色调加深，保持「每个图标一个颜色」的识别度 */
.tool-card:hover {
  color: var(--tb-tone, var(--tb-primary));
  border-color: var(--tb-tone, var(--tb-primary));
  box-shadow: var(--tb-shadow-md);
  transform: translateY(-2px);
  text-decoration: none;
}

.tool-card:focus-visible {
  border-color: var(--tb-primary);
  box-shadow: var(--tb-shadow-md);
}

.tool-card--soon {
  background: var(--tb-bg-sunken);
}

.tool-card__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--tb-space-2);
}

.tool-card__title {
  margin: 0;
  font-size: var(--tb-font-base);
  font-weight: 600;
  color: var(--tb-text);
}

.tool-card--soon .tool-card__title {
  color: var(--tb-text-secondary);
}

.tool-card__badge {
  flex: none;
  padding: 0 var(--tb-space-1);
  font-size: 10px;
  color: var(--tb-text-tertiary);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-pill);
}

.tool-card__desc {
  margin: 0;
  font-size: var(--tb-font-xs);
  line-height: 1.5;
  color: var(--tb-text-secondary);
}

.tool-card__desc--muted {
  color: var(--tb-text-tertiary);
}
</style>
