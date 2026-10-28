/**
 * symmetric-crypt 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实密钥或敏感材料。
 */
export interface SymmetricCryptExample {
  /** 示例名称。 */
  label: string
  /** 算法。 */
  algorithm: 'SM4' | 'AES'
  /** 操作。 */
  operation: 'encrypt' | 'decrypt'
  /** 模式。 */
  mode: 'ECB' | 'CBC' | 'CTR' | 'OFB' | 'CFB' | 'GCM'
  /** 填充。 */
  padding: 'pkcs7' | 'zero' | 'none'
  /** 密钥（十六进制）。 */
  key: string
  /** IV / nonce（十六进制）。 */
  iv: string
  /** 输入内容。 */
  input: string
  /** 输入编码。 */
  inputEncoding: 'utf8' | 'hex'
}

/** 默认示例：SM4-CBC 加密。 */
export const DEFAULT_EXAMPLE: SymmetricCryptExample = {
  label: 'SM4-CBC 加密一段中文',
  algorithm: 'SM4',
  operation: 'encrypt',
  mode: 'CBC',
  padding: 'pkcs7',
  key: '0123456789ABCDEFFEDCBA9876543210',
  iv: '000102030405060708090a0b0c0d0e0f',
  input: '蓝云工具箱 Toolbit',
  inputEncoding: 'utf8',
}

/** GCM 示例。 */
export const GCM_EXAMPLE: SymmetricCryptExample = {
  label: 'SM4-GCM 认证加密',
  algorithm: 'SM4',
  operation: 'encrypt',
  mode: 'GCM',
  padding: 'pkcs7',
  key: '0123456789ABCDEFFEDCBA9876543210',
  iv: '',
  input: '蓝云工具箱 Toolbit',
  inputEncoding: 'utf8',
}

/** 全部示例。 */
export const EXAMPLES: readonly SymmetricCryptExample[] = [DEFAULT_EXAMPLE, GCM_EXAMPLE]
