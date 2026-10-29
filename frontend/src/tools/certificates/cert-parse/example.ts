/**
 * cert-parse 工具的示例数据。
 *
 * 只使用**测试用**证书夹具，不含任何真实证书。
 */
import { LEAF_CERT_PEM } from '@/tools/certificates/fixtures'

export interface CertParseExample {
  /** 示例名称。 */
  label: string
  /** 证书 PEM。 */
  cert: string
}

/** 默认示例：解析叶证书。 */
export const DEFAULT_EXAMPLE: CertParseExample = {
  label: '解析叶证书的主体、扩展与指纹',
  cert: LEAF_CERT_PEM,
}

/** 全部示例。 */
export const EXAMPLES: readonly CertParseExample[] = [DEFAULT_EXAMPLE]
