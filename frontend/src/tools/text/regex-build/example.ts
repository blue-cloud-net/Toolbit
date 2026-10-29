/**
 * regex-build 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */

/** 一个示例。 */
export interface RegexBuildExample {
  /** 示例名称。 */
  label: string
  /** prefix。 */
  prefix: string
}

/** 默认示例。 */
export const DEFAULT_EXAMPLE: RegexBuildExample = {
  label: 'regex-build 示例',
  prefix: 'user-',
}

/** 全部示例。 */
export const EXAMPLES: readonly RegexBuildExample[] = [DEFAULT_EXAMPLE]
