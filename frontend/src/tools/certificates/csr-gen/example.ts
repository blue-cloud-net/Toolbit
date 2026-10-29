/**
 * csr-gen 工具的示例数据。
 *
 * 只使用**测试用**密钥夹具，不含任何真实密钥。
 */
import { TEST_LEAF_KEY_PEM } from '@/tools/certificates/fixtures-v060'

export interface CSRGenExample {
  /** 示例名称。 */
  label: string
  /** 主体通用名。 */
  cn: string
  /** 主体备用名。 */
  sans: string
  /** 已有私钥 PEM。 */
  existingKey: string
}

/** 默认示例：用一把测试私钥生成服务端请求。 */
export const DEFAULT_EXAMPLE: CSRGenExample = {
  label: '用已有测试私钥生成服务端 CSR',
  cn: 'csr.example.com',
  sans: 'DNS:csr.example.com',
  existingKey: TEST_LEAF_KEY_PEM,
}

/** 全部示例。 */
export const EXAMPLES: readonly CSRGenExample[] = [DEFAULT_EXAMPLE]
