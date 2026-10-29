/**
 * 输入即时校验。
 *
 * 纯前端**语法**检查：不发起网络请求，不写任何存储，
 * 也不阻断输入 —— 仅在输入区给出一行轻量提示。
 *
 * 用法：
 *
 * ```ts
 * const validation = useValidation(input, { format: 'json' })
 * ```
 */
import { computed, toValue, type MaybeRefOrGetter, type Ref } from 'vue'
import { useI18n } from 'vue-i18n'

/** 支持的校验格式。 */
export type ValidationFormat = 'json' | 'hex' | 'base64' | 'uuid'

/** 校验结论。 */
export interface ValidationResult {
  level: 'success' | 'error'
  /** 已本地化的提示文案。 */
  message: string
}

export interface UseValidationOptions {
  /** 校验格式；可为 ref，以便按输入编码动态切换。 */
  format: MaybeRefOrGetter<ValidationFormat>
  /** 是否启用；返回 false 时不做校验。 */
  enabled?: () => boolean
}

export function useValidation(
  source: Ref<string>,
  options: UseValidationOptions,
): { validation: Ref<ValidationResult | null> } {
  const { t } = useI18n()

  const validation = computed<ValidationResult | null>(() => {
    if (options.enabled && !options.enabled()) {
      return null
    }
    const text = source.value
    if (text.trim() === '') {
      return null
    }

    switch (toValue(options.format)) {
      case 'json': {
        try {
          JSON.parse(text)
          return { level: 'success', message: t('tool.validation.jsonOk') }
        } catch (err) {
          const message = err instanceof Error ? err.message : ''
          const pos = /position (\d+)/i.exec(message)?.[1]
          if (pos) {
            const offset = Number.parseInt(pos, 10)
            const line = text.slice(0, offset).split('\n').length
            return { level: 'error', message: t('tool.validation.jsonInvalidAt', { line }) }
          }
          return { level: 'error', message: t('tool.validation.jsonInvalid') }
        }
      }
      case 'hex': {
        if (!/^[0-9a-fA-F\s]*$/.test(text)) {
          return { level: 'error', message: t('tool.validation.hexInvalid') }
        }
        const compact = text.replace(/\s+/g, '')
        if (compact.length % 2 !== 0) {
          return { level: 'error', message: t('tool.validation.hexOddLength') }
        }
        return { level: 'success', message: t('tool.validation.hexOk') }
      }
      case 'base64': {
        const compact = text.replace(/\s+/g, '')
        if (compact.length === 0) {
          return null
        }
        if (!/^[A-Za-z0-9+/=_-]*$/.test(compact)) {
          return { level: 'error', message: t('tool.validation.base64Invalid') }
        }
        return { level: 'success', message: t('tool.validation.base64Ok') }
      }
      case 'uuid': {
        if (!/^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$/.test(text.trim())) {
          return { level: 'error', message: t('tool.validation.uuidInvalid') }
        }
        return { level: 'success', message: t('tool.validation.uuidOk') }
      }
      default:
        return null
    }
  })

  return { validation }
}
