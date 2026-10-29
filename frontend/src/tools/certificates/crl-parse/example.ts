/**
 * crl-parse 工具的示例数据。
 *
 * 只使用**测试用**证书夹具，不含任何真实证书。
 */
import { CRL_PEM } from '@/tools/certificates/fixtures'

export interface CRLParseExample {
  /** 示例名称。 */
  label: string
  /** 输入材料。 */
  data: string
  /** 输入格式。 */
  format: string
}

/** 默认示例：解析吊销列表。 */
export const DEFAULT_EXAMPLE: CRLParseExample = {
  label: '解析签发者与吊销条目',
  data: CRL_PEM,
  format: 'pem',
}

/** 全部示例。 */
export const EXAMPLES: readonly CRLParseExample[] = [DEFAULT_EXAMPLE]
