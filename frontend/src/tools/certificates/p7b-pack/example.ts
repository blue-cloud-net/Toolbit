/**
 * p7b-pack 工具的示例数据。
 *
 * 只使用**测试用**证书夹具，不含任何真实材料。
 */
import {
  TEST_CA_CERT_PEM,
  TEST_LEAF_CERT_PEM,
} from '@/tools/certificates/fixtures-v060'

export interface P7BPackExample {
  /** 示例名称。 */
  label: string
  /** 待打包证书（拼接）。 */
  certs: string
}

/** 默认示例：把叶证书与 CA 打包。 */
export const DEFAULT_EXAMPLE: P7BPackExample = {
  label: '把叶证书与 CA 打包成 P7B',
  certs: `${TEST_LEAF_CERT_PEM}${TEST_CA_CERT_PEM}`,
}

/** 全部示例。 */
export const EXAMPLES: readonly P7BPackExample[] = [DEFAULT_EXAMPLE]
