/**
 * key-match 工具的示例数据。
 *
 * 只使用**测试用**密钥夹具，不含任何真实密钥。
 */
import { TEST_EC_PUBLIC_KEY, TEST_RSA_PUBLIC_KEY } from '@/tools/keys/fixtures'

export interface KeyMatchExample {
  /** 示例名称。 */
  label: string
  /** 第一段密钥材料。 */
  a: string
  /** 第二段密钥材料。 */
  b: string
}

/** 默认示例：两把不同的 RSA 公钥。 */
export const DEFAULT_EXAMPLE: KeyMatchExample = {
  label: '比较两把不同的 RSA 公钥',
  a: TEST_RSA_PUBLIC_KEY,
  b: TEST_RSA_PUBLIC_KEY,
}

/** 不同算法示例。 */
export const MIXED_EXAMPLE: KeyMatchExample = {
  label: '比较 RSA 与 EC 公钥（算法不同）',
  a: TEST_RSA_PUBLIC_KEY,
  b: TEST_EC_PUBLIC_KEY,
}

/** 全部示例。 */
export const EXAMPLES: readonly KeyMatchExample[] = [DEFAULT_EXAMPLE, MIXED_EXAMPLE]
