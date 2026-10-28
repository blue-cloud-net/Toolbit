<script setup lang="ts">
/**
 * 工具页骨架。
 *
 * 标题行 + 面板 + 结果区；`wide` 工具在视口 ≥1280 时切换为左右分栏。
 */
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

import AppIcon from '@/components/common/AppIcon.vue'
import { usePrefsStore } from '@/stores/prefs'
import { isImplemented } from '@/tools'

const props = withDefaults(
  defineProps<{
    /** 工具 slug。 */
    slug: string
    /** 中文名。 */
    name: string
    /** 英文名。 */
    nameEn: string
    /** 是否双栏布局。 */
    wide?: boolean
  }>(),
  { wide: false },
)

const { t, locale } = useI18n()
const prefs = usePrefsStore()

const displayName = computed(() => (locale.value === 'en-US' ? props.nameEn : props.name))
const implemented = computed(() => isImplemented(props.slug))

const favorite = computed(() => prefs.isFavorite(props.slug))
const favoriteLabel = computed(() =>
  favorite.value ? t('tool.favorite.remove') : t('tool.favorite.add'),
)

function toggleFavorite(): void {
  prefs.toggleFavorite(props.slug)
}

/** 词条可能尚未登记（工具 chunk 未加载），缺失时返回空串。 */
function i18nOrEmpty(key: string): string {
  const full = `tools.${props.slug}.${key}`
  const translated = t(full)
  return translated === full ? '' : translated
}

const description = computed(() => i18nOrEmpty('description'))
</script>

<template>
  <article
    class="tool-page"
    :class="{ 'tool-page--wide': wide }"
  >
    <header class="tool-page__head">
      <nav
        class="tool-page__breadcrumb"
        :aria-label="t('common.nav.home')"
      >
        <RouterLink :to="{ name: 'home' }">
          {{ t('common.nav.home') }}
        </RouterLink>
        <span aria-hidden="true">/</span>
        <span>{{ displayName }}</span>
      </nav>

      <div class="tool-page__title-row">
        <h1 class="tool-page__title">
          {{ displayName }}
        </h1>
        <span
          v-if="!implemented"
          class="tool-page__badge"
        >{{ t('home.soon') }}</span>

        <button
          v-if="implemented"
          type="button"
          class="tool-page__favorite"
          :class="{ 'tool-page__favorite--on': favorite }"
          :aria-pressed="favorite"
          :aria-label="favoriteLabel"
          :title="favoriteLabel"
          @click="toggleFavorite"
        >
          <AppIcon
            :name="favorite ? 'star-filled' : 'star'"
            :size="17"
          />
        </button>
      </div>

      <p
        v-if="description"
        class="tool-page__desc"
      >
        {{ description }}
      </p>
    </header>

    <div
      v-if="!implemented"
      class="tool-page__soon"
    >
      <p>{{ t('home.soonHint') }}</p>
    </div>

    <template v-else>
      <div class="tool-page__grid">
        <div class="tool-page__primary">
          <slot name="panel" />
        </div>
        <div class="tool-page__secondary">
          <slot name="result" />
        </div>
      </div>

      <section
        v-if="$slots.usage"
        class="tool-page__content"
      >
        <h2 class="tool-page__section-title">
          {{ t('tool.sections.usage') }}
        </h2>
        <slot name="usage" />
      </section>

      <section
        v-if="$slots.faq"
        class="tool-page__content"
      >
        <h2 class="tool-page__section-title">
          {{ t('tool.sections.faq') }}
        </h2>
        <slot name="faq" />
      </section>
    </template>
  </article>
</template>

<style scoped>
.tool-page {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-5);
}

.tool-page__head {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-1);
}

.tool-page__breadcrumb {
  display: flex;
  gap: var(--tb-space-1);
  font-size: var(--tb-font-xs);
  color: var(--tb-text-tertiary);
}

.tool-page__title-row {
  display: flex;
  align-items: center;
  gap: var(--tb-space-2);
}

.tool-page__favorite {
  display: grid;
  place-items: center;
  width: 30px;
  height: 30px;
  color: var(--tb-text-tertiary);
  background: transparent;
  border: 1px solid transparent;
  border-radius: var(--tb-radius-control);
  cursor: pointer;
  transition: color var(--tb-transition), background var(--tb-transition);
}

.tool-page__favorite:hover {
  color: var(--tb-text-secondary);
  background: var(--tb-bg-sunken);
}

.tool-page__favorite--on {
  color: var(--tb-warning);
}

.tool-page__title {
  margin: 0;
  font-size: var(--tb-font-2xl);
  font-weight: 700;
}

.tool-page__badge {
  padding: 0 var(--tb-space-2);
  font-size: 11px;
  color: var(--tb-text-tertiary);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-pill);
}

.tool-page__desc {
  margin: 0;
  color: var(--tb-text-secondary);
}

.tool-page__soon {
  padding: var(--tb-space-6) var(--tb-space-4);
  font-size: var(--tb-font-sm);
  color: var(--tb-text-tertiary);
  text-align: center;
  background: var(--tb-bg-elevated);
  border: 1px dashed var(--tb-border);
  border-radius: var(--tb-radius-panel);
}

.tool-page__grid {
  display: grid;
  grid-template-columns: minmax(0, 1fr);
  gap: var(--tb-space-4);
}

.tool-page__primary,
.tool-page__secondary {
  min-width: 0;
}

.tool-page__content {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-2);
}

.tool-page__section-title {
  margin: 0;
  font-size: var(--tb-font-md);
  font-weight: 600;
}

@media (min-width: 1280px) {
  .tool-page--wide .tool-page__grid {
    grid-template-columns: minmax(0, 1fr) minmax(0, 1fr);
    align-items: start;
  }
}
</style>
