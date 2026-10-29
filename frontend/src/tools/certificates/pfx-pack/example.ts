/**
 * pfx-pack 工具的示例数据。
 *
 * 只使用**测试用**证书与密钥夹具，不含任何真实材料。
 */
import {
  TEST_CA_CERT_PEM,
  TEST_LEAF_CERT_PEM,
  TEST_LEAF_KEY_PEM,
} from '@/tools/certificates/fixtures-v060'

export interface PFXPackExample {
  /** 示例名称。 */
  label: string
  /** 主证书。 */
  cert: string
  /** 私钥。 */
  key: string
  /** CA 链。 */
  caCerts: string
}

/** 默认示例：打包一张测试叶证书与私钥。 */
export const DEFAULT_EXAMPLE: PFXPackExample = {
  label: '打包测试叶证书与私钥',
  cert: TEST_LEAF_CERT_PEM,
  key: TEST_LEAF_KEY_PEM,
  caCerts: TEST_CA_CERT_PEM,
}

/** 全部示例。 */
export const EXAMPLES: readonly PFXPackExample[] = [DEFAULT_EXAMPLE]
