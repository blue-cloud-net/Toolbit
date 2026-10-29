/**
 * cidr-calc 工具的示例数据。
 */
export interface CIDRExample {
  /** 示例名称。 */
  label: string
  /** 输入方式：cidr / range。 */
  mode: 'cidr' | 'range'
  /** 网段。 */
  cidr: string
  /** 起始地址。 */
  start: string
  /** 结束地址。 */
  end: string
}

/** 默认示例：IPv4 /24 网段。 */
export const DEFAULT_EXAMPLE: CIDRExample = {
  label: 'IPv4 /24 网段',
  mode: 'cidr',
  cidr: '192.168.1.0/24',
  start: '10.0.0.1',
  end: '10.0.0.6',
}

/** 全部示例。 */
export const EXAMPLES: readonly CIDRExample[] = [DEFAULT_EXAMPLE]
