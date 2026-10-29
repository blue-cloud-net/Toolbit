/**
 * revocation-check 工具的示例数据。
 *
 * 只使用**测试用**证书与吊销列表夹具，不含任何真实材料。
 */
import {
  TEST_CA_CERT_PEM,
  TEST_CRL_PEM,
  TEST_LEAF_CERT_PEM,
} from '@/tools/certificates/fixtures-v060'

export interface RevocationExample {
  /** 示例名称。 */
  label: string
  /** 待检查证书。 */
  cert: string
  /** 吊销列表。 */
  crl: string
  /** 签发者证书。 */
  issuer: string
}

/** 默认示例：检查一张已被吊销的测试证书。 */
export const DEFAULT_EXAMPLE: RevocationExample = {
  label: '检查一张已被吊销的测试证书',
  cert: TEST_LEAF_CERT_PEM,
  crl: TEST_CRL_PEM,
  issuer: TEST_CA_CERT_PEM,
}

/** 全部示例。 */
export const EXAMPLES: readonly RevocationExample[] = [DEFAULT_EXAMPLE]
