/**
 * cert-gen 工具的示例数据。
 *
 * 只使用**测试用**证书与密钥夹具，不含任何真实材料。
 */
import {
  TEST_CA_CERT_PEM,
  TEST_CA_KEY_PEM,
} from '@/tools/certificates/fixtures-v060'

export interface CertGenExample {
  /** 示例名称。 */
  label: string
  /** 签发模式。 */
  signMode: string
  /** 签发者证书。 */
  issuerCert: string
  /** 签发者私钥。 */
  issuerKey: string
  /** 主体通用名。 */
  cn: string
  /** 主体备用名。 */
  sans: string
}

/** 默认示例：用测试 CA 签发一张叶证书。 */
export const DEFAULT_EXAMPLE: CertGenExample = {
  label: '用测试 CA 签发一张叶证书',
  signMode: 'ca',
  issuerCert: TEST_CA_CERT_PEM,
  issuerKey: TEST_CA_KEY_PEM,
  cn: 'new.example.com',
  sans: 'DNS:new.example.com',
}

/** 全部示例。 */
export const EXAMPLES: readonly CertGenExample[] = [DEFAULT_EXAMPLE]
