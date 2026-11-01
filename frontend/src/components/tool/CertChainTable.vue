<script setup lang="ts">
/**
 * 证书链梯。
 *
 * 把整条链按 **叶 → 中间 → 根** 的顺序竖向排开，
 * 每一级带角色徽标、有效期、剩余天数与连接诊断。
 *
 * 断链点（签发者不在输入里 / 密钥标识符对不上 / 多余证书）
 * 在**对应节点**上标出，而不是只给一句总结。
 *
 * 纯 CSS 绘制连接线，**不引入** Element Plus 表格或标签组件。
 */
import { useI18n } from 'vue-i18n'

import CopyButton from '@/components/tool/CopyButton.vue'
import {
  chainRoleKey,
  formatDateTime,
  linkMethodKey,
  linkReasonKey,
  safetyLevel,
  safetyLevelColor,
} from '@/tools/certificates/certDetail'
import type { ChainCertificate } from '@/types/certificate'

defineOptions({ name: 'CertChainTable' })

withDefaults(
  defineProps<{
    /** 链上的证书（叶 → 根）。 */
    chain: ChainCertificate[]
    /** 是否缺少根证书；用于在链尾提示。 */
    missingRoot?: boolean
  }>(),
  { missingRoot: false },
)

const emit = defineEmits<{
  /** 请求查看某张证书的详情。 */
  detail: [cert: ChainCertificate]
}>()

const { t, locale } = useI18n()

/** 格式化时间。 */
function formatTime(iso: string): string {
  return formatDateTime(iso, locale.value)
}

/** 角色徽标的语义级别。 */
function roleTone(role: string): 'success' | 'warning' | 'info' | 'muted' {
  switch (role) {
    case 'leaf':
      return 'info'
    case 'intermediate':
      return 'warning'
    case 'root':
      return 'success'
    default:
      return 'muted'
  }
}

/** 剩余天数的语义级别。 */
function daysTone(days: number): 'success' | 'warning' | 'danger' {
  const tone = safetyLevelColor(safetyLevel(days))
  return tone === 'success' || tone === 'warning' ? tone : 'danger'
}
</script>

<template>
  <div class="chain">
    <ol class="chain__list">
      <li
        v-for="(cert, index) in chain"
        :key="`${cert.info.serial}-${index}`"
        class="chain__item"
        :class="cert.link.ok ? '' : 'chain__item--broken'"
      >
        <div class="chain__rail">
          <span
            class="chain__dot"
            :class="`chain__dot--${roleTone(cert.role)}`"
          />
        </div>

        <div class="chain__body">
          <div class="chain__head">
            <span
              class="chain__role"
              :class="`chain__role--${roleTone(cert.role)}`"
            >{{ t(`tool.cert.role.${chainRoleKey(cert.role)}`) }}</span>
            <span class="chain__subject">{{ cert.info.subjectText }}</span>
            <CopyButton
              :text="cert.info.subjectText"
              size="small"
            />
          </div>

          <dl class="chain__fields">
            <div class="chain__field">
              <dt>{{ t('tool.cert.fields.issuer') }}</dt>
              <dd>{{ cert.info.issuerText }}</dd>
            </div>
            <div class="chain__field">
              <dt>{{ t('tool.cert.fields.validity') }}</dt>
              <dd>
                {{ formatTime(cert.info.notBefore) }} → {{ formatTime(cert.info.notAfter) }}
              </dd>
            </div>
            <div class="chain__field">
              <dt>{{ t('tool.cert.fields.daysRemaining') }}</dt>
              <dd>
                <span
                  class="chain__days"
                  :class="`chain__days--${daysTone(cert.validity.daysRemaining)}`"
                >
                  {{ cert.validity.daysRemaining }} {{ t('tool.cert.units.days') }}
                </span>
              </dd>
            </div>
            <div class="chain__field">
              <dt>{{ t('tool.cert.fields.fingerprint') }}</dt>
              <dd class="chain__mono">
                {{ cert.fingerprintSha256 || '-' }}
              </dd>
            </div>
            <div class="chain__field">
              <dt>{{ t('tool.cert.fields.link') }}</dt>
              <dd>
                <span
                  class="chain__link"
                  :class="cert.link.ok ? 'chain__link--ok' : 'chain__link--broken'"
                >{{ t(`tool.cert.link.${linkMethodKey(cert.link.method)}`) }}</span>
                <span
                  v-if="!cert.link.ok"
                  class="chain__reason"
                >{{ t(`tool.cert.link.${linkReasonKey(cert.link.reason)}`) }}</span>
              </dd>
            </div>
          </dl>

          <div class="chain__actions">
            <button
              type="button"
              class="chain__action"
              @click="emit('detail', cert)"
            >
              {{ t('tool.cert.actions.detail') }}
            </button>
          </div>
        </div>
      </li>
    </ol>

    <p
      v-if="missingRoot"
      class="chain__notice"
    >
      {{ t('tool.cert.link.missing_root') }}
    </p>
  </div>
</template>

<style scoped>
.chain {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-2);
}

.chain__list {
  margin: 0;
  padding: 0;
  list-style: none;
}

/* 连接线：竖线 + 圆点。 */
.chain__rail {
  position: relative;
  flex: none;
  width: 14px;
  display: flex;
  justify-content: center;
}

.chain__rail::before {
  content: '';
  position: absolute;
  top: 0;
  bottom: 0;
  width: 2px;
  background: var(--tb-border);
}

.chain__item:first-child .chain__rail::before {
  top: 10px;
}

.chain__item:last-child .chain__rail::before {
  bottom: calc(100% - 10px);
}

.chain__dot {
  position: relative;
  z-index: 1;
  margin-top: 4px;
  width: 12px;
  height: 12px;
  background: var(--tb-bg-elevated);
  border: 2px solid var(--tb-border-strong);
  border-radius: 50%;
}

.chain__dot--info {
  border-color: var(--tb-info);
}

.chain__dot--warning {
  border-color: var(--tb-warning);
}

.chain__dot--success {
  border-color: var(--tb-success);
}

.chain__dot--muted {
  border-color: var(--tb-border-strong);
}

.chain__body {
  flex: 1;
  min-width: 0;
  padding: var(--tb-space-3);
  background: var(--tb-bg-sunken);
  border: 1px solid var(--tb-border);
  border-radius: var(--tb-radius-control);
}

.chain__item--broken .chain__body {
  border-color: var(--tb-danger);
}

.chain__head {
  display: flex;
  align-items: baseline;
  gap: var(--tb-space-2);
  margin-bottom: var(--tb-space-2);
}

.chain__role {
  flex: none;
  padding: 1px var(--tb-space-2);
  font-size: var(--tb-font-2xs);
  border: 1px solid var(--tb-border-strong);
  border-radius: var(--tb-radius-pill);
}

.chain__role--info {
  color: var(--tb-info);
  border-color: var(--tb-info);
}

.chain__role--warning {
  color: var(--tb-warning);
  border-color: var(--tb-warning);
}

.chain__role--success {
  color: var(--tb-success);
  border-color: var(--tb-success);
}

.chain__subject {
  flex: 1;
  min-width: 0;
  overflow-wrap: anywhere;
  font-family: var(--tb-font-mono);
  font-size: var(--tb-font-xs);
  color: var(--tb-text);
}

.chain__fields {
  display: flex;
  flex-direction: column;
  gap: 2px;
  margin: 0;
}

.chain__field {
  display: flex;
  gap: var(--tb-space-2);
  font-size: var(--tb-font-2xs);
}

.chain__field dt {
  flex: none;
  min-width: 72px;
  color: var(--tb-text-tertiary);
}

.chain__field dd {
  margin: 0;
  min-width: 0;
  overflow-wrap: anywhere;
  color: var(--tb-text-secondary);
}

.chain__mono {
  font-family: var(--tb-font-mono);
}

.chain__days {
  font-family: var(--tb-font-mono);
}

.chain__days--success {
  color: var(--tb-success);
}

.chain__days--warning {
  color: var(--tb-warning);
}

.chain__days--danger {
  color: var(--tb-danger);
}

.chain__link--ok {
  color: var(--tb-text-secondary);
}

.chain__link--broken {
  color: var(--tb-danger);
  font-weight: 600;
}

.chain__reason {
  margin-left: var(--tb-space-2);
  color: var(--tb-danger);
}

.chain__actions {
  margin-top: var(--tb-space-2);
}

.chain__action {
  padding: 0;
  font-size: var(--tb-font-xs);
  color: var(--tb-primary);
  cursor: pointer;
  background: none;
  border: none;
}

.chain__notice {
  margin: 0;
  padding: var(--tb-space-2) var(--tb-space-3);
  font-size: var(--tb-font-xs);
  color: var(--tb-warning);
  background: color-mix(in srgb, var(--tb-warning) 12%, var(--tb-bg-elevated));
  border: 1px solid color-mix(in srgb, var(--tb-warning) 40%, var(--tb-border));
  border-radius: var(--tb-radius-control);
}
</style>
