/**
 * jwk-xml-convert 工具的示例数据。
 *
 * 只使用**测试用**公钥夹具，不含任何真实密钥。
 */
import { TEST_RSA_PUBLIC_KEY } from '@/tools/keys/fixtures'

export interface JWKXMLConvertExample {
  /** 示例名称。 */
  label: string
  /** 源格式。 */
  from: string
  /** 目标格式。 */
  to: string
  /** 密钥类型。 */
  keyType: string
  /** 源材料。 */
  data: string
}

/** 默认示例：PEM 转 JWK。 */
export const DEFAULT_EXAMPLE: JWKXMLConvertExample = {
  label: '把 SPKI 公钥转成 JWK',
  from: 'pem',
  to: 'jwk',
  keyType: 'public',
  data: TEST_RSA_PUBLIC_KEY,
}

/** 全部示例。 */
export const EXAMPLES: readonly JWKXMLConvertExample[] = [DEFAULT_EXAMPLE]
