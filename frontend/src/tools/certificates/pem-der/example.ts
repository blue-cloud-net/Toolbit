/**
 * pem-der 工具的示例数据。
 *
 * 只使用**测试用**证书夹具，不含任何真实证书。
 */
import { LEAF_CERT_PEM } from '@/tools/certificates/fixtures'

export interface PEMDerExample {
  /** 示例名称。 */
  label: string
  /** 输入材料。 */
  data: string
  /** 输入格式。 */
  format: string
}

/** 默认示例：PEM 转 DER。 */
export const DEFAULT_EXAMPLE: PEMDerExample = {
  label: '把叶证书的 PEM 转成 DER',
  data: LEAF_CERT_PEM,
  format: 'pem',
}

/** 全部示例。 */
export const EXAMPLES: readonly PEMDerExample[] = [DEFAULT_EXAMPLE]
