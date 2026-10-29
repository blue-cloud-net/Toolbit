/**
 * color-convert 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */

/** 一个示例。 */
export interface ColorConvertExample {
  /** 示例名称。 */
  label: string
  /** input。 */
  input: string
}

/** 默认示例。 */
export const DEFAULT_EXAMPLE: ColorConvertExample = {
  label: 'color-convert 示例',
  input: '#3B82F6',
}

/** 全部示例。 */
export const EXAMPLES: readonly ColorConvertExample[] = [DEFAULT_EXAMPLE]
