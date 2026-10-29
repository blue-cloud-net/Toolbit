/**
 * pfx-parse 工具的示例数据。
 *
 * 只使用**测试用**容器夹具，不含任何真实材料。
 */
import { TEST_P12_BASE64 } from '@/tools/certificates/fixtures-v060'

export interface PFXParseExample {
  /** 示例名称。 */
  label: string
  /** 容器内容（base64）。 */
  container: string
  /** 口令。 */
  passphrase: string
}

/** 默认示例：解析一个测试容器（口令 `test-pass`）。 */
export const DEFAULT_EXAMPLE: PFXParseExample = {
  label: '解析测试容器',
  container: TEST_P12_BASE64,
  passphrase: 'test-pass',
}

/** 全部示例。 */
export const EXAMPLES: readonly PFXParseExample[] = [DEFAULT_EXAMPLE]
