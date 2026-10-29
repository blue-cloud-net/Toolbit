/**
 * domain-verify 工具的示例数据。
 */
export interface DomainVerifyExample {
  /** 示例名称。 */
  label: string
  /** 验证方式。 */
  type: 'dns-txt' | 'dns-cname' | 'http-file'
  /** 域名。 */
  domain: string
  /** 期望值。 */
  expected: string
  /** 站点文件路径。 */
  filePath: string
}

/** 默认示例：校验 TXT 记录。 */
export const DEFAULT_EXAMPLE: DomainVerifyExample = {
  label: '校验 TXT 记录',
  type: 'dns-txt',
  domain: 'example.com',
  expected: 'toolbit-verify=abc123',
  filePath: '/.well-known/toolbit-verify.txt',
}

/** 全部示例。 */
export const EXAMPLES: readonly DomainVerifyExample[] = [DEFAULT_EXAMPLE]
