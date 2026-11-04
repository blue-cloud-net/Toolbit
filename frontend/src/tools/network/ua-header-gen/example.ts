/**
 * ua-header-gen 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */

/** 一个示例。 */
export interface HeaderGeneratorExample {
  /** 示例名称。 */
  label: string
  /** 平台。 */
  platform: string
}

/** 默认示例：Chrome 桌面版。 */
export const DEFAULT_EXAMPLE: HeaderGeneratorExample = {
  label: 'Chrome 桌面版',
  platform: 'chrome',
}

/** 移动端示例。 */
export const ANDROID_EXAMPLE: HeaderGeneratorExample = {
  label: 'Android Chrome',
  platform: 'android',
}

/** 命令行示例。 */
export const CURL_EXAMPLE: HeaderGeneratorExample = {
  label: 'curl',
  platform: 'curl',
}

/** 全部示例。 */
export const EXAMPLES: readonly HeaderGeneratorExample[] = [
  DEFAULT_EXAMPLE,
  ANDROID_EXAMPLE,
  CURL_EXAMPLE,
]
