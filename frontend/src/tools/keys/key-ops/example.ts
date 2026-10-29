/**
 * key-ops 工具的示例数据。
 *
 * 只使用**一次性生成的测试私钥**夹具，不含任何真实密钥。
 */
import { TEST_RSA_PRIVATE_KEY } from '@/tools/keys/fixtures'

export interface KeyOpsExample {
  /** 示例名称。 */
  label: string
  /** 操作动作。 */
  action: string
  /** 私钥 PEM。 */
  key: string
}

/** 默认示例：提取公钥。 */
export const DEFAULT_EXAMPLE: KeyOpsExample = {
  label: '从私钥提取配对公钥',
  action: 'extract-public',
  key: TEST_RSA_PRIVATE_KEY,
}

/** 全部示例。 */
export const EXAMPLES: readonly KeyOpsExample[] = [DEFAULT_EXAMPLE]
