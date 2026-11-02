/**
 * cert-lifecycle 工具的示例数据。
 *
 * 只使用**测试用**证书夹具，不含任何真实凭据。
 */
import { LEAF_CERT_PEM } from '@/tools/certificates/fixtures'

/** 一个示例。 */
export interface CertLifecycleExample {
  /** 示例名称。 */
  label: string
  /** certificate。 */
  certificate: string
}

/** 默认示例：计算叶证书的到期倒计时。 */
export const DEFAULT_EXAMPLE: CertLifecycleExample = {
  label: 'cert-lifecycle 示例',
  certificate: LEAF_CERT_PEM,
}

/** 全部示例。 */
export const EXAMPLES: readonly CertLifecycleExample[] = [DEFAULT_EXAMPLE]
