/**
 * asn1-view 工具的示例数据。
 *
 * 只使用**测试用**证书夹具，不含任何真实证书。
 */
import { LEAF_CERT_PEM } from '@/tools/certificates/fixtures'

export interface ASN1ViewExample {
  /** 示例名称。 */
  label: string
  /** 输入材料。 */
  data: string
  /** 输入格式。 */
  format: string
}

/** 默认示例：查看叶证书的 DER 结构。 */
export const DEFAULT_EXAMPLE: ASN1ViewExample = {
  label: '查看叶证书的 DER 结构树',
  data: LEAF_CERT_PEM,
  format: 'pem',
}

/** 全部示例。 */
export const EXAMPLES: readonly ASN1ViewExample[] = [DEFAULT_EXAMPLE]
