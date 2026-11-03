/**
 * oid-lookup 工具的示例数据。
 *
 * 只用于演示与快捷入口，**不含**任何真实凭据。
 */

/** 一个示例。 */
export interface OidLookupExample {
  /** 示例名称。 */
  label: string
  /** 目标 OID。 */
  oid: string
}

/** PKIX 扩展密钥用途。 */
export const PKIX_EXAMPLE: OidLookupExample = {
  label: 'PKIX 扩展密钥用途',
  oid: '1.3.6.1.5.5.7.3',
}

/** X.509 证书扩展。 */
export const X509_EXAMPLE: OidLookupExample = {
  label: 'X.509 证书扩展',
  oid: '2.5.29',
}

/** PKCS 与 RSA。 */
export const PKCS_EXAMPLE: OidLookupExample = {
  label: 'PKCS #1（RSA）',
  oid: '1.2.840.113549.1.1',
}

/** NIST 算法。 */
export const NIST_EXAMPLE: OidLookupExample = {
  label: 'NIST 算法',
  oid: '2.16.840.1.101.3.4',
}

/** 国密算法。 */
export const GM_EXAMPLE: OidLookupExample = {
  label: '国密 GM/T',
  oid: '1.2.156.10197',
}

/** SNMP 管理信息库。 */
export const SNMP_EXAMPLE: OidLookupExample = {
  label: 'SNMP 管理信息库',
  oid: '1.3.6.1.2.1',
}

/** 全部示例。 */
export const EXAMPLES: readonly OidLookupExample[] = [
  PKIX_EXAMPLE,
  X509_EXAMPLE,
  PKCS_EXAMPLE,
  NIST_EXAMPLE,
  GM_EXAMPLE,
  SNMP_EXAMPLE,
]
