/**
 * ocsp-response 工具的示例数据。
 *
 * 只使用**测试用**证书夹具，不含任何真实材料。
 */
import {
  TEST_CA_CERT_PEM,
  TEST_LEAF_CERT_PEM,
} from '@/tools/certificates/fixtures-v060'

export interface OCSPResponseExample {
  /** 示例名称。 */
  label: string
  /** 被查询证书。 */
  cert: string
  /** 签发者证书。 */
  issuer: string
  /** 信任根。 */
  trustedRoots: string
}

/** 默认示例：用测试 CA 作信任根。 */
export const DEFAULT_EXAMPLE: OCSPResponseExample = {
  label: '用测试 CA 作信任根校验响应',
  cert: TEST_LEAF_CERT_PEM,
  issuer: TEST_CA_CERT_PEM,
  trustedRoots: TEST_CA_CERT_PEM,
}

/** 全部示例。 */
export const EXAMPLES: readonly OCSPResponseExample[] = [DEFAULT_EXAMPLE]
