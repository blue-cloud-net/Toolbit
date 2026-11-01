/**
 * dns-lookup 工具的 DNS 服务器预设。
 *
 * 只列**公开**解析器；自定义服务器由用户在页面上追加。
 * 地址与协议与后端 `DNSServerInput` 对应：`proto` ∈ udp | tcp | dot | doh。
 *
 * 预设按用途分三组（页面上分组展示）：
 * - `primary` 主用：日常优先使用的公共解析器；
 * - `backup` 备用：主用不可用时的应急解析器；
 * - `doh` 加密：DNS over HTTPS 端点。
 */

/** 预设分组。 */
export type DNSPresetGroup = 'primary' | 'backup' | 'doh'

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
  /** 归属地说明的 i18n key（对应 `options.region.<key>`）。 */
  regionKey: string
  /** 所属分组。 */
  group: DNSPresetGroup
}

/** 分组顺序；页面上按此顺序渲染。 */
export const DNS_PRESET_GROUPS: readonly DNSPresetGroup[] = ['primary', 'backup', 'doh']

/** 全部预设。 */
export const DNS_SERVER_PRESETS: readonly DNSServerPreset[] = [
  {
    id: 'cloudflare',
    label: 'Cloudflare 1.1.1.1',
    proto: 'udp',
    addr: '1.1.1.1:53',
    regionKey: 'global',
    group: 'primary',
  },
  {
    id: 'google',
    label: 'Google 8.8.8.8',
    proto: 'udp',
    addr: '8.8.8.8:53',
    regionKey: 'global',
    group: 'primary',
  },
  {
    id: 'quad9',
    label: 'Quad9 9.9.9.9',
    proto: 'udp',
    addr: '9.9.9.9:53',
    regionKey: 'switzerland',
    group: 'primary',
  },
  {
    id: 'alidns',
    label: 'AliDNS 223.5.5.5',
    proto: 'udp',
    addr: '223.5.5.5:53',
    regionKey: 'china',
    group: 'primary',
  },
  {
    id: 'dnspod',
    label: 'DNSPod 119.29.29.29',
    proto: 'udp',
    addr: '119.29.29.29:53',
    regionKey: 'china',
    group: 'primary',
  },
  {
    id: '114dns',
    label: '114DNS 114.114.114.114',
    proto: 'udp',
    addr: '114.114.114.114:53',
    regionKey: 'china',
    group: 'primary',
  },
  {
    id: 'cloudflare-backup',
    label: 'Cloudflare 1.0.0.1',
    proto: 'udp',
    addr: '1.0.0.1:53',
    regionKey: 'global',
    group: 'backup',
  },
  {
    id: 'google-backup',
    label: 'Google 8.8.4.4',
    proto: 'udp',
    addr: '8.8.4.4:53',
    regionKey: 'global',
    group: 'backup',
  },
  {
    id: 'opendns',
    label: 'OpenDNS 208.67.222.222',
    proto: 'udp',
    addr: '208.67.222.222:53',
    regionKey: 'usa',
    group: 'backup',
  },
  {
    id: 'opendns-backup',
    label: 'OpenDNS 208.67.220.220',
    proto: 'udp',
    addr: '208.67.220.220:53',
    regionKey: 'usa',
    group: 'backup',
  },
  {
    id: 'cloudflare-doh',
    label: 'Cloudflare DoH',
    proto: 'doh',
    addr: 'https://cloudflare-dns.com/dns-query',
    regionKey: 'global',
    group: 'doh',
  },
  {
    id: 'google-doh',
    label: 'Google DoH',
    proto: 'doh',
    addr: 'https://dns.google/dns-query',
    regionKey: 'global',
    group: 'doh',
  },
  {
    id: 'quad9-doh',
    label: 'Quad9 DoH',
    proto: 'doh',
    addr: 'https://dns.quad9.net/dns-query',
    regionKey: 'switzerland',
    group: 'doh',
  },
  {
    id: 'alidns-doh',
    label: 'AliDNS DoH',
    proto: 'doh',
    addr: 'https://dns.alidns.com/dns-query',
    regionKey: 'china',
    group: 'doh',
  },
]

/** 默认选中的预设（首次打开即有对比结果）。 */
export const DEFAULT_PRESET_IDS: readonly string[] = ['cloudflare', 'google', 'alidns']

/** 一次请求允许的服务器总数上限（与后端 `maxServers` 一致）。 */
export const MAX_SERVERS = 8
