/**
 * dns-lookup 工具的 DNS 服务器预设。
 *
 * 只列**公开**解析器；自定义服务器由用户在页面上追加。
 * 地址与协议与后端 `DNSServerInput` 对应：`proto` ∈ udp | tcp | dot | doh。
 */

/** 一个服务器预设。 */
export interface DNSServerPreset {
  /** 预设标识（前端内部使用）。 */
  id: string
  /** 显示名。 */
  label: string
  /** 查询协议。 */
  proto: string
  /** 服务器地址；DoH 时是完整 URL。 */
  addr: string
  /** 归属地说明。 */
  region: string
}

/** 全部预设。 */
export const DNS_SERVER_PRESETS: readonly DNSServerPreset[] = [
  {
    id: 'cloudflare',
    label: 'Cloudflare',
    proto: 'udp',
    addr: '1.1.1.1:53',
    region: '全球',
  },
  {
    id: 'google',
    label: 'Google',
    proto: 'udp',
    addr: '8.8.8.8:53',
    region: '全球',
  },
  {
    id: 'quad9',
    label: 'Quad9',
    proto: 'udp',
    addr: '9.9.9.9:53',
    region: '瑞士',
  },
  {
    id: 'alidns',
    label: 'AliDNS',
    proto: 'udp',
    addr: '223.5.5.5:53',
    region: '中国',
  },
  {
    id: 'dnspod',
    label: 'DNSPod',
    proto: 'udp',
    addr: '119.29.29.29:53',
    region: '中国',
  },
  {
    id: '114dns',
    label: '114DNS',
    proto: 'udp',
    addr: '114.114.114.114:53',
    region: '中国',
  },
  {
    id: 'opendns',
    label: 'OpenDNS',
    proto: 'udp',
    addr: '208.67.222.222:53',
    region: '美国',
  },
  {
    id: 'cloudflare-doh',
    label: 'Cloudflare DoH',
    proto: 'doh',
    addr: 'https://cloudflare-dns.com/dns-query',
    region: '全球',
  },
]

/** 默认选中的预设（首次打开即有对比结果）。 */
export const DEFAULT_PRESET_IDS: readonly string[] = ['cloudflare', 'google', 'alidns']

/** 一次请求允许的服务器总数上限（与后端 `maxServers` 一致）。 */
export const MAX_SERVERS = 8
