/**
 * dns-lookup 工具的 DNS 服务器预设。
 *
 * 只列**公开**解析器；自定义服务器由用户在页面上追加。
 * 地址与协议与后端 `DNSServerInput` 对应：`proto` ∈ udp | tcp | dot | doh。
 *
 * 预设按协议分五组（页面上分组展示）：
 * - `primary` 明文主用：日常优先使用的公共解析器；
 * - `backup` 明文备用：主用不可用时的应急解析器；
 * - `tcp` 明文 TCP：走 TCP 的明文查询（响应更大、不易丢包）；
 * - `dot` 加密 DNS over TLS：TLS 封装，固定 853 端口；
 * - `doh` 加密 DNS over HTTPS：HTTPS 封装，443 端口。
 *
 * 与「DNS 服务器大全」的关系：那边是**完整目录**，这里是**实测用的快捷选择**。
 */

/** 预设分组。 */
export type DNSPresetGroup = 'primary' | 'backup' | 'tcp' | 'dot' | 'doh'

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
export const DNS_PRESET_GROUPS: readonly DNSPresetGroup[] = [
  'primary',
  'backup',
  'tcp',
  'dot',
  'doh',
]

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
    id: 'quad9-backup',
    label: 'Quad9 149.112.112.112',
    proto: 'udp',
    addr: '149.112.112.112:53',
    regionKey: 'switzerland',
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
    id: 'adguard',
    label: 'AdGuard 94.140.14.14',
    proto: 'udp',
    addr: '94.140.14.14:53',
    regionKey: 'global',
    group: 'backup',
  },
  {
    id: 'dnspod-backup',
    label: 'DNSPod 182.254.116.116',
    proto: 'udp',
    addr: '182.254.116.116:53',
    regionKey: 'china',
    group: 'backup',
  },
  {
    id: 'cloudflare-tcp',
    label: 'Cloudflare 1.1.1.1',
    proto: 'tcp',
    addr: '1.1.1.1:53',
    regionKey: 'global',
    group: 'tcp',
  },
  {
    id: 'google-tcp',
    label: 'Google 8.8.8.8',
    proto: 'tcp',
    addr: '8.8.8.8:53',
    regionKey: 'global',
    group: 'tcp',
  },
  {
    id: 'quad9-tcp',
    label: 'Quad9 9.9.9.9',
    proto: 'tcp',
    addr: '9.9.9.9:53',
    regionKey: 'switzerland',
    group: 'tcp',
  },
  {
    id: 'alidns-tcp',
    label: 'AliDNS 223.5.5.5',
    proto: 'tcp',
    addr: '223.5.5.5:53',
    regionKey: 'china',
    group: 'tcp',
  },
  {
    id: 'dnspod-tcp',
    label: 'DNSPod 119.29.29.29',
    proto: 'tcp',
    addr: '119.29.29.29:53',
    regionKey: 'china',
    group: 'tcp',
  },
  {
    id: 'cloudflare-dot',
    label: 'Cloudflare one.one.one.one',
    proto: 'dot',
    addr: 'one.one.one.one:853',
    regionKey: 'global',
    group: 'dot',
  },
  {
    id: 'google-dot',
    label: 'Google dns.google',
    proto: 'dot',
    addr: 'dns.google:853',
    regionKey: 'global',
    group: 'dot',
  },
  {
    id: 'quad9-dot',
    label: 'Quad9 dns.quad9.net',
    proto: 'dot',
    addr: 'dns.quad9.net:853',
    regionKey: 'switzerland',
    group: 'dot',
  },
  {
    id: 'alidns-dot',
    label: 'AliDNS dns.alidns.com',
    proto: 'dot',
    addr: 'dns.alidns.com:853',
    regionKey: 'china',
    group: 'dot',
  },
  {
    id: 'dnspod-dot',
    label: 'DNSPod dot.pub',
    proto: 'dot',
    addr: 'dot.pub:853',
    regionKey: 'china',
    group: 'dot',
  },
  {
    id: 'adguard-dot',
    label: 'AdGuard dns.adguard.com',
    proto: 'dot',
    addr: 'dns.adguard.com:853',
    regionKey: 'global',
    group: 'dot',
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
  {
    id: 'dnspod-doh',
    label: 'DNSPod DoH',
    proto: 'doh',
    addr: 'https://doh.pub/dns-query',
    regionKey: 'china',
    group: 'doh',
  },
  {
    id: 'adguard-doh',
    label: 'AdGuard DoH',
    proto: 'doh',
    addr: 'https://dns.adguard.com/dns-query',
    regionKey: 'global',
    group: 'doh',
  },
]

/** 默认选中的预设（首次打开即有对比结果）。 */
export const DEFAULT_PRESET_IDS: readonly string[] = ['cloudflare', 'google', 'alidns']

/** 一次请求允许的服务器总数上限（与后端 `maxServers` 一致）。 */
export const MAX_SERVERS = 8

/** 按标识查找预设；找不到时返回 null。 */
export function findDNSPreset(id: string): DNSServerPreset | null {
  return DNS_SERVER_PRESETS.find((preset) => preset.id === id) ?? null
}
