/**
 * cert-chain-verify 工具的示例数据。
 *
 * 只使用**测试用**证书夹具，不含任何真实证书。
 */
import { CHAIN_PEM, ROOT_CERT_PEM } from '@/tools/certificates/fixtures'

export interface ChainVerifyExample {
  /** 示例名称。 */
  label: string
  /** 证书链 PEM。 */
  bundle: string
  /** 受信任根 PEM。 */
  roots: string
  /** 主机名。 */
  host: string
}

/** 默认示例：用根证书验证链。 */
export const DEFAULT_EXAMPLE: ChainVerifyExample = {
  label: '用根 CA 验证链并检查域名匹配',
  bundle: CHAIN_PEM,
  roots: ROOT_CERT_PEM,
  host: 'leaf.example.com',
}

/** 全部示例。 */
export const EXAMPLES: readonly ChainVerifyExample[] = [DEFAULT_EXAMPLE]
