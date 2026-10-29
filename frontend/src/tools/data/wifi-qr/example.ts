/**
 * wifi-qr 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */

/** 一个示例。 */
export interface WiFiQRExample {
  /** 示例名称。 */
  label: string
  /** ssid。 */
  ssid: string
  /** password。 */
  password: string
}

/** 默认示例。 */
export const DEFAULT_EXAMPLE: WiFiQRExample = {
  label: 'wifi-qr 示例',
  ssid: 'Guest-Network',
  password: 'guest-password',
}

/** 全部示例。 */
export const EXAMPLES: readonly WiFiQRExample[] = [DEFAULT_EXAMPLE]
