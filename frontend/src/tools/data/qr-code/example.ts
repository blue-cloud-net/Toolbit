/**
 * qr-code 工具的示例数据。
 *
 * 只用于演示，**不含**任何敏感材料。
 */

/** 二维码操作。 */
export type QROperation = 'generate' | 'decode'

/** 纠错级别。 */
export type QRECLevel = 'L' | 'M' | 'Q' | 'H'

/** 一个二维码示例。 */
export interface QRCodeExample {
  /** 示例名称。 */
  label: string
  /** 操作。 */
  operation: QROperation
  /** 二维码内容。 */
  text: string
  /** 模块边长（像素）。 */
  size: number
  /** 静默区模块数。 */
  margin: number
  /** 纠错级别。 */
  ecLevel: QRECLevel
  /** 前景色。 */
  fg: string
  /** 背景色。 */
  bg: string
}

/** 默认示例：生成一个链接二维码。 */
export const DEFAULT_EXAMPLE: QRCodeExample = {
  label: '生成链接二维码',
  operation: 'generate',
  text: 'https://example.com',
  size: 6,
  margin: 4,
  ecLevel: 'M',
  fg: '',
  bg: '',
}

/** 自定义样式的示例。 */
export const STYLED_EXAMPLE: QRCodeExample = {
  label: '自定义颜色与纠错级别',
  operation: 'generate',
  text: 'Toolbit',
  size: 8,
  margin: 2,
  ecLevel: 'H',
  fg: '#123456',
  bg: '#ffffff',
}

/** 全部示例。 */
export const EXAMPLES: readonly QRCodeExample[] = [DEFAULT_EXAMPLE, STYLED_EXAMPLE]
