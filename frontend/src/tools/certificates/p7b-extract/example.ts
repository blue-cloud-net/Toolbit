/**
 * p7b-extract 工具的示例数据。
 *
 * 只使用**测试用**容器夹具，不含任何真实材料。
 */
import { TEST_P7B_BASE64 } from '@/tools/certificates/fixtures-v060'

export interface P7BExtractExample {
  /** 示例名称。 */
  label: string
  /** 容器内容（base64）。 */
  container: string
}

/** 默认示例：从一个测试证书袋中提取证书。 */
export const DEFAULT_EXAMPLE: P7BExtractExample = {
  label: '从测试证书袋中提取证书',
  container: TEST_P7B_BASE64,
}

/** 全部示例。 */
export const EXAMPLES: readonly P7BExtractExample[] = [DEFAULT_EXAMPLE]
