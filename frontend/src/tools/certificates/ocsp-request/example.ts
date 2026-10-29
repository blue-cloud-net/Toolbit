/**
 * ocsp-request 工具的示例数据。
 *
 * 只使用**测试用**证书夹具，不含任何真实材料。
 */
import {
  TEST_CA_CERT_PEM,
  TEST_LEAF_CERT_PEM,
} from '@/tools/certificates/fixtures-v060'

export interface OCSPRequestExample {
  /** 示例名称。 */
  label: string
  /** 被查询证书。 */
  cert: string
  /** 签发者证书。 */
  issuer: string
}

/** 默认示例：为一张测试叶证书生成请求。 */
export const DEFAULT_EXAMPLE: OCSPRequestExample = {
  label: '为测试叶证书生成 OCSP 请求',
  cert: TEST_LEAF_CERT_PEM,
  issuer: TEST_CA_CERT_PEM,
}

/** 全部示例。 */
export const EXAMPLES: readonly OCSPRequestExample[] = [DEFAULT_EXAMPLE]
