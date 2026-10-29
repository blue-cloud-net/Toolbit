/**
 * caa-gen 工具的示例数据。
 */
export interface CAAGenExample {
  /** 示例名称。 */
  label: string
  /** 域名。 */
  domain: string
  /** 标签。 */
  tag: string
  /** 标签值。 */
  value: string
}

/** 默认示例：授权 Let's Encrypt 签发。 */
export const DEFAULT_EXAMPLE: CAAGenExample = {
  label: '授权 Let\'s Encrypt 签发',
  domain: 'example.com',
  tag: 'issue',
  value: 'letsencrypt.org',
}

/** 全部示例。 */
export const EXAMPLES: readonly CAAGenExample[] = [DEFAULT_EXAMPLE]
