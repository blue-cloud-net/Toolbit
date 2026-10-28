/**
 * key-parse 工具的示例数据。
 *
 * 只使用**测试用**公钥夹具，不含任何真实密钥。
 */
import { TEST_RSA_PUBLIC_KEY } from '@/tools/keys/fixtures'

export interface KeyParseExample {
  /** 示例名称。 */
  label: string
  /** 密钥 PEM。 */
  key: string
}

/** 默认示例：解析 RSA 公钥。 */
export const DEFAULT_EXAMPLE: KeyParseExample = {
  label: '解析 RSA 公钥的算法、位数与指纹',
  key: TEST_RSA_PUBLIC_KEY,
}

/** 全部示例。 */
export const EXAMPLES: readonly KeyParseExample[] = [DEFAULT_EXAMPLE]
