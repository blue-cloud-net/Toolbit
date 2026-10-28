<script setup lang="ts">
/**
 * 404 页面。
 */
import { useI18n } from 'vue-i18n'

const { t } = useI18n()
</script>

<template>
  <div class="not-found">
    <p
      class="not-found__code"
      aria-hidden="true"
    >
      404
    </p>
    <h1 class="not-found__title">
      {{ t('common.states.notFound') }}
    </h1>
    <p class="not-found__hint">
      {{ t('common.states.notFoundHint') }}
    </p>
    <RouterLink
      class="not-found__link"
      :to="{ name: 'home' }"
    >
      {{ t('common.actions.backHome') }}
    </RouterLink>
  </div>
</template>

<style scoped>
.not-found {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: var(--tb-space-2);
  padding: var(--tb-space-12) var(--tb-space-4);
  text-align: center;
}

.not-found__code {
  margin: 0;
  font-size: 64px;
  font-weight: 700;
  line-height: 1;
  color: var(--tb-border-strong);
}

.not-found__title {
  margin: 0;
  font-size: var(--tb-font-xl);
}

.not-found__hint {
  margin: 0;
  color: var(--tb-text-secondary);
}

.not-found__link {
  margin-top: var(--tb-space-2);
}
</style>
