/**
 * Base64 工具的纯函数工具集。
 *
 * 只做「表单状态 → 请求体」的组装与变体判定，不依赖 Vue：
 * 便于在 vitest 中独立覆盖，也避免各工具页各写一份。
 */

/** 操作类型。 */
export type Base64Operation = 'encode' | 'decode' | 'convert'

/** Base64 变体。 */
export type Base64Variant = 'standard' | 'url'

/** 输入变体；`auto` 表示自动识别。 */
export type Base64SourceVariant = Base64Variant | 'auto'

/** 输入传输编码。 */
export type Base64InputEncoding = 'utf8' | 'hex'

/** 表单状态。 */
export interface Base64FormState {
  /** 操作。 */
  operation: Base64Operation
  /** 输入内容。 */
  input: string
  /** 输入传输编码。 */
  inputEncoding: Base64InputEncoding
  /** 输入变体。 */
  variant: Base64SourceVariant
  /** 是否保留填充。 */
  padding: boolean
  /** 输出编码（decode 时有效）。 */
  outputEncoding: Base64InputEncoding
  /** 输出变体（convert 时有效）。 */
  targetVariant: Base64Variant
}

/** 请求体。 */
export type Base64Payload = Record<string, unknown>

/**
 * 组装请求体。
 *
 * `convert` 与其余操作的字段集不同：前者用 `targetVariant`，
 * 后者用 `outputEncoding`；互斥字段**不**下发，避免后端按旧语义解释。
 */
export function buildPayload(form: Base64FormState): Base64Payload {
  const payload: Base64Payload = {
    operation: form.operation,
    input: { data: form.input, encoding: form.inputEncoding },
    variant: form.variant,
    padding: form.padding,
  }
  if (form.operation === 'convert') {
    payload.targetVariant = form.targetVariant
  } else {
    payload.outputEncoding = form.outputEncoding
  }
  return payload
}

/**
 * 归一化表单状态。
 *
 * `auto` 变体只对 convert 有意义，`hex` 输入编码只对 encode / decode 有意义；
 * 切走时回落为各自的默认值，避免发出后端不接受的组合。
 */
export function normalizeForm(form: Base64FormState): Base64FormState {
  const next = { ...form }
  if (next.operation !== 'convert' && next.variant === 'auto') {
    next.variant = 'standard'
  }
  if (next.operation === 'convert' && next.inputEncoding !== 'utf8') {
    next.inputEncoding = 'utf8'
  }
  return next
}

/**
 * 反向转换：把当前结果作为新输入再转一次。
 *
 * 来源与目标变体互换（`auto` 来源落到目标变体的**对面**），
 * 并把结果回填为新输入 —— 于是「标准 → URL 安全 → 标准」只需点一下。
 *
 * 没有结果时输入保持不变，此时只是把变体调个头。
 */
export function reverseConversion(form: Base64FormState, result: string): Base64FormState {
  const target: Base64Variant =
    form.variant === 'auto'
      ? form.targetVariant === 'url'
        ? 'standard'
        : 'url'
      : form.variant
  return {
    ...form,
    variant: form.targetVariant,
    targetVariant: target,
    input: result !== '' ? result : form.input,
  }
}
