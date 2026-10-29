<script setup lang="ts">
/**
 * 校验提示。
 *
 * 在输入区下方给出轻量的语法校验结论（合法 / 不合法 + 位置）。
 * 不阻断输入，纯展示。
 */
withDefaults(
  defineProps<{
    /** 校验结论；为 null 时不渲染。 */
    validation?: { level: 'success' | 'error'; message: string } | null
  }>(),
  { validation: null },
)
</script>

<template>
  <p
    v-if="validation"
    class="validation-hint"
    :class="`validation-hint--${validation.level}`"
    role="status"
  >
    {{ validation.message }}
  </p>
</template>

<style scoped>
.validation-hint {
  margin: 0;
  font-size: var(--tb-font-2xs);
}

.validation-hint--success {
  color: var(--tb-success);
}

.validation-hint--error {
  color: var(--tb-danger);
}
</style>
