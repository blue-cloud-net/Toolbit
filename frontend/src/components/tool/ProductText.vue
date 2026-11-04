<script setup lang="ts">
/**
 * 产物文本展示。
 *
 * 结果区页签里的「**产物**」：一段纯文本 —— 密文、PEM、十六进制、
 * 生成的 DNS 区域文件、命令片段……这类内容**本质就是文本**，
 * 硬拆成表格反而更难读；但它又**不该**和摘要挤在同一块里。
 *
 * 因此本组件承担「产物」的标准呈现：
 *
 *   - 顶部一行：产物标题 + 操作（复制 / 下载）；
 *   - 下方：等宽文本块，超高时可滚动。
 *
 * 摘要与产物**必须**分属不同页签（见 `docs/standards/01-frontend.md` §10）：
 * 默认页签停在「摘要」，产物另开一页签。
 */
import CopyButton from '@/components/tool/CopyButton.vue'
import DownloadButton from '@/components/tool/DownloadButton.vue'

withDefaults(
  defineProps<{
    /** 产物标题（已本地化）。 */
    label: string
    /** 产物文本。 */
    content: string
    /** 下载文件名；给出时渲染下载按钮。 */
    filename?: string
    /** 下载内容的 MIME 类型。 */
    mime?: string
    /** 是否渲染复制按钮；默认 `true`。 */
    copyable?: boolean
    /** 文本块最大高度；传空串表示不限高。 */
    maxHeight?: string
  }>(),
  {
    filename: undefined,
    mime: 'text/plain;charset=utf-8',
    copyable: true,
    maxHeight: '520px',
  },
)
</script>

<template>
  <div class="product">
    <div class="product__head">
      <span class="product__label">{{ label }}</span>
      <div class="product__actions">
        <CopyButton
          v-if="copyable && content !== ''"
          :text="content"
        />
        <DownloadButton
          v-if="filename && content !== ''"
          :filename="filename"
          :content="content"
          :mime="mime"
        />
      </div>
    </div>

    <pre
      class="product__body"
      :style="{ maxHeight }"
    ><code>{{ content }}</code></pre>
  </div>
</template>

<style scoped>
.product {
  display: flex;
  flex-direction: column;
  gap: var(--tb-space-2);
  min-width: 0;
}

.product__head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--tb-space-2);
}

.product__label {
  font-size: var(--tb-font-xs);
  color: var(--tb-text-tertiary);
}

.product__actions {
  display: flex;
  align-items: center;
  gap: var(--tb-space-2);
}

.product__body {
  margin: 0;
  padding: var(--tb-space-3);
  overflow: auto;
  font-family: var(--tb-font-mono);
  font-size: var(--tb-font-sm);
  line-height: 1.6;
  color: var(--tb-text);
  white-space: pre-wrap;
  word-break: break-all;
  background: var(--tb-bg-sunken);
  border-radius: var(--tb-radius-control);
}
</style>
