/**
 * cert-lifecycle 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */

/** 一个示例。 */
export interface CertLifecycleExample {
  /** 示例名称。 */
  label: string
  /** certificate。 */
  certificate: string
}

/** 默认示例。 */
export const DEFAULT_EXAMPLE: CertLifecycleExample = {
  label: 'cert-lifecycle 示例',
  certificate: '',
}

/** 全部示例。 */
export const EXAMPLES: readonly CertLifecycleExample[] = [DEFAULT_EXAMPLE]
