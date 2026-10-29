/**
 * crl-gen 工具的示例数据。
 *
 * 只使用**测试用**证书与密钥夹具，不含任何真实材料。
 */
import {
  TEST_CA_CERT_PEM,
  TEST_CA_KEY_PEM,
  TEST_LEAF_CERT_PEM,
} from '@/tools/certificates/fixtures-v060'

export interface CRLGenExample {
  /** 示例名称。 */
  label: string
  /** 签发者证书。 */
  issuerCert: string
  /** 签发者私钥。 */
  issuerKey: string
  /** 被吊销的证书。 */
  revokedCert: string
}

/** 默认示例：吊销一张测试叶证书。 */
export const DEFAULT_EXAMPLE: CRLGenExample = {
  label: '吊销一张测试叶证书',
  issuerCert: TEST_CA_CERT_PEM,
  issuerKey: TEST_CA_KEY_PEM,
  revokedCert: TEST_LEAF_CERT_PEM,
}

/** 全部示例。 */
export const EXAMPLES: readonly CRLGenExample[] = [DEFAULT_EXAMPLE]
