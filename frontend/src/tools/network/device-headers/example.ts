/**
 * device-headers 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */

/** 一个示例。 */
export interface DeviceHeadersExample {
  /** 示例名称。 */
  label: string
  /** platform。 */
  platform: string
}

/** 默认示例。 */
export const DEFAULT_EXAMPLE: DeviceHeadersExample = {
  label: 'device-headers 示例',
  platform: 'chrome',
}

/** 全部示例。 */
export const EXAMPLES: readonly DeviceHeadersExample[] = [DEFAULT_EXAMPLE]
