/**
 * cert-chain-parse 工具的示例数据。
 *
 * 只使用**测试用**证书夹具，不含任何真实证书。
 */
import { CHAIN_PEM } from '@/tools/certificates/fixtures'

export interface ChainParseExample {
  /** 示例名称。 */
  label: string
  /** 证书链 PEM。 */
  bundle: string
}

/** 默认示例：解析完整链。 */
export const DEFAULT_EXAMPLE: ChainParseExample = {
  label: '把叶 → 中间 → 根的证书链排好顺序',
  bundle: CHAIN_PEM,
}

/** 全部示例。 */
export const EXAMPLES: readonly ChainParseExample[] = [DEFAULT_EXAMPLE]
