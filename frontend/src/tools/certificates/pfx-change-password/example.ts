/**
 * pfx-change-password 工具的示例数据。
 *
 * 只使用**测试用**容器夹具，不含任何真实材料。
 */
import { TEST_P12_BASE64 } from '@/tools/certificates/fixtures-v060'

export interface PFXChangeExample {
  /** 示例名称。 */
  label: string
  /** 容器内容（base64）。 */
  container: string
  /** 当前口令。 */
  oldPassphrase: string
  /** 新口令。 */
  newPassphrase: string
}

/** 默认示例：把测试容器改成新口令。 */
export const DEFAULT_EXAMPLE: PFXChangeExample = {
  label: '把测试容器改成新口令',
  container: TEST_P12_BASE64,
  oldPassphrase: 'test-pass',
  newPassphrase: 'new-test-pass',
}

/** 全部示例。 */
export const EXAMPLES: readonly PFXChangeExample[] = [DEFAULT_EXAMPLE]
