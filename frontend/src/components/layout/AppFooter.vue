<script setup lang="ts">
/**
 * 页脚：数据说明、版权、许可证、版本号与开源链接。
 *
 * 位于内容列底部（侧栏直通到底部，页脚不横跨侧栏），文案整体居中。
 */
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'

import AppIcon from '@/components/common/AppIcon.vue'
import { APP_VERSION, COPYRIGHT_YEAR, RELEASES_URL, REPO_URL } from '@/config/site'

const { t } = useI18n()

/** 当前年份，用于版权行。 */
const year = computed(() => COPYRIGHT_YEAR)

/** 版本号。 */
const version = APP_VERSION
</script>

<template>
  <footer class="footer">
    <div class="footer__inner">
      <p class="footer__note">
        {{ t('common.footer.builtWith') }}
      </p>

      <p class="footer__meta">
        <span>{{ t('common.footer.copyright', { year }) }}</span>
        <span
          class="footer__sep"
          aria-hidden="true"
        >·</span>
        <span>{{ t('common.footer.license') }}</span>
        <span
          class="footer__sep"
          aria-hidden="true"
        >·</span>
        <a
          class="footer__link"
          :href="RELEASES_URL"
          target="_blank"
          rel="noopener noreferrer"
          :title="t('common.footer.versionLink')"
        >v{{ version }}</a>
        <span
          class="footer__sep"
          aria-hidden="true"
        >·</span>
        <a
          class="footer__link"
          :href="REPO_URL"
          target="_blank"
          rel="noopener noreferrer"
        >
          {{ t('common.footer.github') }}
          <AppIcon
            name="external"
            :size="12"
          />
          <span class="tb-sr-only">{{ t('common.footer.externalHint') }}</span>
        </a>
      </p>
    </div>
  </footer>
</template>

<style scoped>
.footer {
  flex: none;
  border-top: 1px solid var(--tb-border);
  background: var(--tb-bg-elevated);
}

.footer__inner {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
  max-width: 1200px;
  margin: 0 auto;
  padding: var(--tb-space-4) var(--tb-space-6);
  font-size: var(--tb-font-xs);
  color: var(--tb-text-tertiary);
  text-align: center;
}

.footer__note,
.footer__meta {
  margin: 0;
}

.footer__meta {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  justify-content: center;
  gap: 0 var(--tb-space-1);
}

.footer__sep {
  color: var(--tb-border-strong);
}

.footer__link {
  display: inline-flex;
  align-items: center;
  gap: 3px;
  color: var(--tb-text-secondary);
}

.footer__link:hover {
  color: var(--tb-primary);
}

@media (max-width: 900px) {
  .footer__inner {
    padding: var(--tb-space-4);
  }
}
</style>
