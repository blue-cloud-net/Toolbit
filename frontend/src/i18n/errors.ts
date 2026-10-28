/**
 * 错误码本地化。
 *
 * 后端返回 `error.code` + `error.params`，前端按 `errors.<CODE>` 插值，
 * **禁止**依赖后端 `message` 做逻辑判断。
 */
import { ApiError } from '@/api/client'
import { i18n } from '@/i18n'

/** 本地化后的错误。 */
export interface LocalizedError {
  /** 机器可读错误码。 */
  code: string
  /** 本地化文案。 */
  text: string
  /** 请求标识，便于与服务端日志关联。 */
  requestId: string
}

/** 检查某个错误码是否有词条。 */
function hasMessage(code: string): boolean {
  return i18n.global.te(`errors.${code}`)
}

/**
 * 把任意异常转换为本地化错误。
 *
 * 未知错误码回退到后端英文兜底文案；网络层错误使用 `NETWORK_ERROR`。
 */
export function localizeError(err: unknown): LocalizedError {
  if (err instanceof ApiError) {
    const text = hasMessage(err.code)
      ? i18n.global.t(`errors.${err.code}`, err.params)
      : err.message
    return { code: err.code, text, requestId: err.requestId }
  }
  return {
    code: 'INTERNAL_ERROR',
    text: i18n.global.t('errors.INTERNAL_ERROR'),
    requestId: '',
  }
}

/** 本地化一个警告码。 */
export function localizeWarning(code: string, params: Record<string, unknown> = {}): string {
  if (!hasMessage(code)) {
    return code
  }
  return i18n.global.t(`errors.${code}`, params)
}
