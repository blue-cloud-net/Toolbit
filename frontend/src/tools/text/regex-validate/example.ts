/**
 * regex-validate 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */

/** 一个示例。 */
export interface RegexValidateExample {
  /** 示例名称。 */
  label: string
  /** pattern。 */
  pattern: string
}

/** 默认示例。 */
export const DEFAULT_EXAMPLE: RegexValidateExample = {
  label: 'regex-validate 示例',
  pattern: '^[a-z]+$',
}

/** 全部示例。 */
export const EXAMPLES: readonly RegexValidateExample[] = [DEFAULT_EXAMPLE]
