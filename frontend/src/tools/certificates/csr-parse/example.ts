/**
 * csr-parse 工具的示例数据。
 *
 * 只使用**测试用** CSR 夹具，不含任何真实材料。
 */
import { CSR_PEM } from '@/tools/certificates/fixtures'

export interface CSRParseExample {
  /** 示例名称。 */
  label: string
  /** CSR PEM。 */
  csr: string
}

/** 默认示例：解析 CSR。 */
export const DEFAULT_EXAMPLE: CSRParseExample = {
  label: '解析 CSR 的主体、公钥与签名',
  csr: CSR_PEM,
}

/** 全部示例。 */
export const EXAMPLES: readonly CSRParseExample[] = [DEFAULT_EXAMPLE]
