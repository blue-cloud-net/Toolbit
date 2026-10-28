/**
 * key-convert 工具的示例数据。
 *
 * 只使用**测试用**公钥夹具，不含任何真实密钥。
 */
import { TEST_RSA_PUBLIC_KEY } from '@/tools/keys/fixtures'

export interface KeyConvertExample {
  /** 示例名称。 */
  label: string
  /** 源密钥 PEM。 */
  key: string
  /** 源格式。 */
  from: string
  /** 目标格式。 */
  to: string
}

/** 默认示例：SPKI 公钥转 PKCS#8（应当被拒绝，用于演示格式约束）。 */
export const DEFAULT_EXAMPLE: KeyConvertExample = {
  label: '把 SPKI 公钥转成 PKCS#8（演示格式约束）',
  key: TEST_RSA_PUBLIC_KEY,
  from: 'auto',
  to: 'pkcs8',
}

/** DER 示例：SPKI 公钥转 DER。 */
export const DER_EXAMPLE: KeyConvertExample = {
  label: '把 SPKI 公钥转成 DER',
  key: TEST_RSA_PUBLIC_KEY,
  from: 'auto',
  to: 'der',
}

/** 全部示例。 */
export const EXAMPLES: readonly KeyConvertExample[] = [DEFAULT_EXAMPLE, DER_EXAMPLE]
