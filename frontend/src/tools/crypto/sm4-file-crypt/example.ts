/**
 * sm4-file-crypt 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实密钥或敏感材料。
 */
export interface SM4FileCryptExample {
  /** 示例名称。 */
  label: string
  /** 操作。 */
  operation: 'encrypt' | 'decrypt'
  /** 密钥（十六进制）。 */
  key: string
  /** 密钥编码。 */
  keyEncoding: 'hex' | 'base64'
  /** 基础 nonce（十六进制）；留空表示自动生成。 */
  nonce: string
  /** nonce 编码。 */
  nonceEncoding: 'hex' | 'base64'
  /** 分块大小（字节）；留空表示默认 1 MiB。 */
  chunkSize: string
}

/** 默认示例：SM4 文件加密。 */
export const DEFAULT_EXAMPLE: SM4FileCryptExample = {
  label: 'SM4 文件加密',
  operation: 'encrypt',
  key: '0123456789ABCDEFFEDCBA9876543210',
  keyEncoding: 'hex',
  nonce: '',
  nonceEncoding: 'hex',
  chunkSize: '',
}

/** 解密示例。 */
export const DECRYPT_EXAMPLE: SM4FileCryptExample = {
  label: 'SM4 文件解密',
  operation: 'decrypt',
  key: '0123456789ABCDEFFEDCBA9876543210',
  keyEncoding: 'hex',
  nonce: '',
  nonceEncoding: 'hex',
  chunkSize: '',
}
