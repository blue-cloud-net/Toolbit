/**
 * cert-chain-repair 工具的示例数据。
 *
 * 只使用**测试用**证书夹具，不含任何真实证书。
 */
import { INTERMEDIATE_CERT_PEM, LEAF_CERT_PEM } from '@/tools/certificates/fixtures'

export interface ChainRepairExample {
  /** 示例名称。 */
  label: string
  /** 叶证书 PEM。 */
  cert: string
  /** 可用于补链的中间证书 PEM。 */
  intermediates: string
  /** 受信任根 PEM。 */
  roots: string
}

/** 默认示例：用中间 CA 补链。 */
export const DEFAULT_EXAMPLE: ChainRepairExample = {
  label: '用中间 CA 补全叶证书的链',
  cert: LEAF_CERT_PEM,
  intermediates: INTERMEDIATE_CERT_PEM,
  roots: '',
}

/** 全部示例。 */
export const EXAMPLES: readonly ChainRepairExample[] = [DEFAULT_EXAMPLE]
