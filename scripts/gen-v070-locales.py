#!/usr/bin/env python3
"""一次性生成 v0.7.0 八个工具的双语词条。

词条结构固定为 10 个顶层键：name / description / category / actions /
fields / options / results / examples / hints / faq。
"""
from __future__ import annotations

import json
import pathlib

ROOT = pathlib.Path(__file__).resolve().parents[1]
LOCALES = ROOT / "frontend" / "src" / "locales"

ZH = {
    "cidr-calc": {
        "name": "CIDR 计算",
        "description": "计算网段地址、掩码、可用主机数与拆分结果，也可由地址范围反推 CIDR。",
        "category": "网络",
        "actions": {"calculate": "计算"},
        "fields": {
            "mode": "输入方式",
            "cidr": "网段",
            "start": "起始地址",
            "end": "结束地址",
            "split": "拆分前缀",
        },
        "options": {"modeCidr": "网段", "modeRange": "地址范围"},
        "results": {"output": "计算结果"},
        "examples": [{"label": "IPv4 /24 网段", "value": "192.168.1.0/24"}],
        "hints": {
            "split": "拆分前缀必须不小于原前缀；一次最多拆出 256 个子网。",
            "hostCount": "IPv4 的 `/31` 计 2 个可用地址、`/32` 计 1 个；IPv6 全部地址可用。",
        },
        "faq": [
            {
                "q": "为什么 `hostCount` 是字符串？",
                "a": "IPv6 的地址数量可能超过 2⁶⁴，用十进制字符串才能无损表达。",
            },
            {
                "q": "地址范围反推的结果为什么有多条？",
                "a": "一个任意区间通常无法用单个 CIDR 精确表示，工具会给出最小的 CIDR 集合。",
            },
            {
                "q": "IPv6 有广播地址吗？",
                "a": "没有。IPv6 的 `broadcast` 字段为空。",
            },
        ],
    },
    "ssl-check": {
        "name": "SSL 证书检测",
        "description": "对目标主机做 TLS 握手，流式展示各阶段进度与协商结果、证书链与验证结论。",
        "category": "网络",
        "actions": {"probe": "开始检测"},
        "fields": {
            "host": "主机",
            "port": "端口",
            "sni": "SNI",
            "depth": "链深度上限",
            "versions": "协议版本",
            "ciphers": "密码套件",
            "timeoutMs": "超时（毫秒）",
        },
        "options": {},
        "results": {
            "output": "检测结果",
            "progress": "探测进度",
            "chain": "证书链",
        },
        "examples": [{"label": "检测公共站点", "value": "example.com:443"}],
        "hints": {
            "sni": "留空时用主机名做 SNI。",
            "versions": "留空表示使用默认版本范围（TLS 1.2 起）。",
            "ciphers": "TLS 1.3 的密码套件无法通过配置裁剪。",
        },
        "faq": [
            {
                "q": "为什么显示「验证未完成」？",
                "a": "本地信任库不可用时只能判断链内是否自洽，此时 `verified` 为 false 并给出警告。",
            },
            {
                "q": "目标是内网地址会怎样？",
                "a": "出网防护会拦截私有、环回、链路本地地址，返回 `SSRF_BLOCKED`。",
            },
            {
                "q": "同一时间能开多条连接吗？",
                "a": "不能。同一来源 IP 同时只允许 1 条流式连接。",
            },
        ],
    },
    "dns-lookup": {
        "name": "DNS 查询",
        "description": "查询指定记录类型，可逐台对比自定义 DNS 服务器的响应与耗时。",
        "category": "网络",
        "actions": {"lookup": "查询"},
        "fields": {
            "host": "域名",
            "type": "记录类型",
            "proto": "协议",
            "addr": "服务器",
            "timeoutMs": "超时（毫秒）",
        },
        "options": {},
        "results": {"output": "解析结果", "perServer": "逐服务器结果"},
        "examples": [{"label": "查询 A 记录", "value": "example.com / A"}],
        "hints": {
            "servers": "留空使用系统配置的解析器；填写后只查询指定服务器。",
            "doh": "DoH 需要填写完整的 `https://` 地址。",
        },
        "faq": [
            {
                "q": "为什么某一台服务器报错？",
                "a": "单台失败不会影响整体请求，错误记录在对应服务器的 `error` 字段里。",
            },
            {
                "q": "能查到 TTL 吗？",
                "a": "能。每条结果都带 TTL（秒）。",
            },
            {
                "q": "自定义服务器可以是内网地址吗？",
                "a": "不可以。自定义服务器同样要过出网防护，私网地址会被拦截。",
            },
        ],
    },
    "dns-tester": {
        "name": "DNS 多记录测试",
        "description": "综合检测域名的注册信息、DNSSEC 部署情况与反向解析。",
        "category": "网络",
        "actions": {"test": "检测"},
        "fields": {"action": "检测项", "domain": "域名", "ip": "IP 地址", "timeoutMs": "超时（毫秒）"},
        "options": {
            "whois": "注册信息（whois / RDAP）",
            "dnssec": "DNSSEC",
            "ptr": "反向解析",
        },
        "results": {"output": "检测结果"},
        "examples": [{"label": "查询 DNSSEC 状态", "value": "example.com / dnssec"}],
        "hints": {
            "whois": "优先用 RDAP（HTTPS），失败时回退到 whois 协议（TCP 43）。",
            "ptr": "反向解析会顺带做一次正向解析，便于发现正反不一致。",
        },
        "faq": [
            {
                "q": "whois 结果为什么有空字段？",
                "a": "部分注册局限制查询或对字段脱敏，缺失值以空字符串表示。",
            },
            {
                "q": "`dnssec` 检测的是什么？",
                "a": "逐级检查 DS 记录并查询 DNSKEY，判断是否具备信任链锚点。",
            },
            {
                "q": "PTR 与域名不一致说明什么？",
                "a": "说明该地址的反向记录指向的名字并不解析回同一地址，属于配置疏漏。",
            },
        ],
    },
    "websocket-test": {
        "name": "WebSocket 测试",
        "description": "在浏览器内直连 WebSocket 服务，展示握手耗时、子协议与收发内容。",
        "category": "网络",
        "actions": {"connect": "连接", "disconnect": "断开"},
        "fields": {"url": "地址", "protocols": "子协议", "message": "发送内容"},
        "options": {},
        "results": {"output": "连接结果", "log": "收发日志"},
        "examples": [{"label": "公共回显服务", "value": "wss://echo.websocket.org"}],
        "hints": {
            "headers": "浏览器**无法**自定义请求头、**无法**使用自定义 TLS 证书、**无法**指定源端口。",
            "origin": "跨域连接遵循浏览器同源策略，不受本站服务端配置影响。",
        },
        "faq": [
            {
                "q": "为什么请求不经过后端？",
                "a": "本工具用浏览器原生 `WebSocket` 直连目标，服务端不参与握手。",
            },
            {
                "q": "连接会被计入统计吗？",
                "a": "不会记录连接内容，只统计工具被打开的次数。",
            },
            {
                "q": "能测 `ws://` 吗？",
                "a": "本站是 HTTPS 时浏览器会拦截 `ws://` 混合内容，请用 `wss://`。",
            },
        ],
    },
    "cert-chain-fetch": {
        "name": "证书链获取",
        "description": "从目标主机抓取它实际发送的证书链，并给出链摘要与验证结论。",
        "category": "证书",
        "actions": {"fetch": "获取"},
        "fields": {"host": "主机", "port": "端口", "serverName": "SNI", "timeoutMs": "超时（毫秒）"},
        "options": {},
        "results": {"output": "证书链 PEM", "summary": "链摘要"},
        "examples": [{"label": "抓取公共站点证书链", "value": "example.com:443"}],
        "hints": {
            "port": "只允许固定端口（43 / 53 / 80 / 443 / 853）。",
            "private": "私有、环回、链路本地地址一律被拦截。",
        },
        "faq": [
            {
                "q": "和服务端实际发送的链一样吗？",
                "a": "一样。工具只读取 TLS 握手中服务端发来的证书，不额外补链。",
            },
            {
                "q": "链不完整怎么办？",
                "a": "用证书链修复工具补齐中间证书。",
            },
            {
                "q": "会保存我抓的证书吗？",
                "a": "不会。响应即用即弃，不落库也不写日志。",
            },
        ],
    },
    "ct-search": {
        "name": "证书透明度查询",
        "description": "查询证书透明度日志中为该域名签发的证书。当前版本尚未开放。",
        "category": "证书",
        "actions": {"search": "查询"},
        "fields": {"domain": "域名", "matchType": "匹配方式", "limit": "返回条数"},
        "options": {"exact": "精确", "contains": "包含", "wildcard": "通配"},
        "results": {"output": "证书列表"},
        "examples": [{"label": "查询 example.com", "value": "example.com"}],
        "hints": {
            "unavailable": "本版未接入证书透明度日志，调用会返回 `NOT_IMPLEMENTED`。",
        },
        "faq": [
            {
                "q": "为什么还不能用？",
                "a": "公开的透明度日志查询服务没有稳定 SLA，本版先不接入，后续版本再评估。",
            },
            {
                "q": "有替代办法吗？",
                "a": "可以先用「证书链获取」查看目标主机当前实际使用的证书。",
            },
            {
                "q": "以后会加回来吗？",
                "a": "会。缺口已登记在路线图的已知限制里。",
            },
        ],
    },
    "domain-verify": {
        "name": "域名验证",
        "description": "通过 DNS TXT / CNAME 记录或站点文件校验域名所有权。",
        "category": "证书",
        "actions": {"verify": "验证"},
        "fields": {
            "type": "验证方式",
            "domain": "域名",
            "expected": "期望值",
            "filePath": "文件路径",
            "resolvers": "DNS 服务器",
            "timeoutMs": "超时（毫秒）",
        },
        "options": {
            "dnsTxt": "DNS TXT",
            "dnsCname": "DNS CNAME",
            "httpFile": "站点文件",
        },
        "results": {"output": "验证结果", "perResolver": "逐解析器结果"},
        "examples": [{"label": "校验 TXT 记录", "value": "example.com / dns-txt"}],
        "hints": {
            "httpFile": "站点文件走 `https://<域名>/<路径>`，需要 200 且内容完全一致。",
            "resolvers": "留空使用系统解析器；填写后逐个比对。",
        },
        "faq": [
            {
                "q": "`dns-txt` 支持多个解析器吗？",
                "a": "支持。每个解析器单独给结论，任一匹配即总体匹配。",
            },
            {
                "q": "站点文件为什么要完全一致？",
                "a": "完全一致才能排除误判；首尾空白会被忽略。",
            },
            {
                "q": "能验证内网域名吗？",
                "a": "DNS 方式可以；`http-file` 因为要出网，私有地址会被拦截。",
            },
        ],
    },
}

EN = {
    "cidr-calc": {
        "name": "CIDR Calculator",
        "description": "Compute network address, mask, usable hosts and subnets, or convert an address range back into CIDR blocks.",
        "category": "Network",
        "actions": {"calculate": "Calculate"},
        "fields": {
            "mode": "Input mode",
            "cidr": "CIDR block",
            "start": "Start address",
            "end": "End address",
            "split": "Split prefix",
        },
        "options": {"modeCidr": "CIDR block", "modeRange": "Address range"},
        "results": {"output": "Result"},
        "examples": [{"label": "IPv4 /24 block", "value": "192.168.1.0/24"}],
        "hints": {
            "split": "The split prefix must not be shorter than the original one; at most 256 subnets per call.",
            "hostCount": "For IPv4 a `/31` counts 2 usable hosts and a `/32` counts 1; IPv6 has no unusable addresses.",
        },
        "faq": [
            {
                "q": "Why is `hostCount` a string?",
                "a": "IPv6 address counts can exceed 2⁶⁴, so only a decimal string keeps them exact.",
            },
            {
                "q": "Why does a range produce several CIDRs?",
                "a": "An arbitrary range usually cannot be a single CIDR, so the tool returns the minimal set.",
            },
            {
                "q": "Does IPv6 have a broadcast address?",
                "a": "No. The `broadcast` field is empty for IPv6.",
            },
        ],
    },
    "ssl-check": {
        "name": "SSL Checker",
        "description": "Perform a TLS handshake against a host and stream the negotiated parameters, certificate chain and verification result.",
        "category": "Network",
        "actions": {"probe": "Run check"},
        "fields": {
            "host": "Host",
            "port": "Port",
            "sni": "SNI",
            "depth": "Max chain depth",
            "versions": "TLS versions",
            "ciphers": "Cipher suites",
            "timeoutMs": "Timeout (ms)",
        },
        "options": {},
        "results": {"output": "Result", "progress": "Progress", "chain": "Certificate chain"},
        "examples": [{"label": "Check a public site", "value": "example.com:443"}],
        "hints": {
            "sni": "When empty, the host name is used as SNI.",
            "versions": "Empty means the default range (TLS 1.2 and above).",
            "ciphers": "TLS 1.3 suites cannot be restricted by configuration.",
        },
        "faq": [
            {
                "q": "Why is verification reported as incomplete?",
                "a": "When no local trust store is available only chain self-consistency can be checked; `verified` is then false and a warning is returned.",
            },
            {
                "q": "What happens with an internal address?",
                "a": "The egress guard blocks private, loopback and link-local addresses with `SSRF_BLOCKED`.",
            },
            {
                "q": "Can I open several connections at once?",
                "a": "No. One streaming connection per source IP is allowed at a time.",
            },
        ],
    },
    "dns-lookup": {
        "name": "DNS Lookup",
        "description": "Look up records and compare responses and latency across custom DNS servers.",
        "category": "Network",
        "actions": {"lookup": "Look up"},
        "fields": {
            "host": "Domain",
            "type": "Record type",
            "proto": "Protocol",
            "addr": "Server",
            "timeoutMs": "Timeout (ms)",
        },
        "options": {},
        "results": {"output": "Answers", "perServer": "Per-server results"},
        "examples": [{"label": "Look up an A record", "value": "example.com / A"}],
        "hints": {
            "servers": "Leave empty to use the system resolvers; when set, only those servers are queried.",
            "doh": "DoH requires a full `https://` URL.",
        },
        "faq": [
            {
                "q": "Why did one server fail?",
                "a": "A single failure does not fail the request; the error is recorded on that server's `error` field.",
            },
            {
                "q": "Are TTLs included?",
                "a": "Yes. Every answer carries its TTL in seconds.",
            },
            {
                "q": "May a custom server be an internal address?",
                "a": "No. Custom servers go through the same egress guard, so private addresses are blocked.",
            },
        ],
    },
    "dns-tester": {
        "name": "DNS Record Tester",
        "description": "Check domain registration data, DNSSEC deployment and reverse DNS in one place.",
        "category": "Network",
        "actions": {"test": "Run test"},
        "fields": {"action": "Check", "domain": "Domain", "ip": "IP address", "timeoutMs": "Timeout (ms)"},
        "options": {
            "whois": "Registration (whois / RDAP)",
            "dnssec": "DNSSEC",
            "ptr": "Reverse DNS",
        },
        "results": {"output": "Result"},
        "examples": [{"label": "Check DNSSEC", "value": "example.com / dnssec"}],
        "hints": {
            "whois": "RDAP over HTTPS is tried first, falling back to the whois protocol (TCP 43).",
            "ptr": "Reverse lookup also performs a forward lookup to expose mismatches.",
        },
        "faq": [
            {
                "q": "Why are some whois fields empty?",
                "a": "Some registries restrict queries or redact fields; missing values are returned as empty strings.",
            },
            {
                "q": "What does `dnssec` check?",
                "a": "It walks DS records and queries DNSKEY to determine whether a chain of trust anchor exists.",
            },
            {
                "q": "What does a PTR mismatch mean?",
                "a": "The reverse record points at a name that does not resolve back to the same address, which is a configuration gap.",
            },
        ],
    },
    "websocket-test": {
        "name": "WebSocket Tester",
        "description": "Connect to a WebSocket endpoint straight from the browser and inspect handshake timing, subprotocol and messages.",
        "category": "Network",
        "actions": {"connect": "Connect", "disconnect": "Disconnect"},
        "fields": {"url": "URL", "protocols": "Subprotocols", "message": "Message"},
        "options": {},
        "results": {"output": "Result", "log": "Message log"},
        "examples": [{"label": "Public echo service", "value": "wss://echo.websocket.org"}],
        "hints": {
            "headers": "Browsers **cannot** set custom headers, use a custom TLS certificate or choose a source port.",
            "origin": "Cross-origin connections follow the browser same-origin policy, not this server's configuration.",
        },
        "faq": [
            {
                "q": "Why does traffic bypass the backend?",
                "a": "The page uses the native browser `WebSocket` API, so the server is not involved in the handshake.",
            },
            {
                "q": "Are connections counted in the statistics?",
                "a": "Message content is never recorded; only the fact that the tool was opened is counted.",
            },
            {
                "q": "Can I test `ws://`?",
                "a": "When this site is served over HTTPS the browser blocks mixed content, so use `wss://`.",
            },
        ],
    },
    "cert-chain-fetch": {
        "name": "Certificate Chain Fetcher",
        "description": "Fetch the certificate chain a host actually sends, with a chain summary and verification result.",
        "category": "Certificates",
        "actions": {"fetch": "Fetch"},
        "fields": {"host": "Host", "port": "Port", "serverName": "SNI", "timeoutMs": "Timeout (ms)"},
        "options": {},
        "results": {"output": "Chain PEM", "summary": "Chain summary"},
        "examples": [{"label": "Fetch a public chain", "value": "example.com:443"}],
        "hints": {
            "port": "Only fixed ports are allowed (43 / 53 / 80 / 443 / 853).",
            "private": "Private, loopback and link-local addresses are always blocked.",
        },
        "faq": [
            {
                "q": "Is this the chain the server really sends?",
                "a": "Yes. Only certificates presented during the TLS handshake are returned; nothing is added.",
            },
            {
                "q": "What if the chain is incomplete?",
                "a": "Use the chain repair tool to append the missing intermediates.",
            },
            {
                "q": "Are fetched certificates stored?",
                "a": "No. Responses are discarded immediately and never written to storage or logs.",
            },
        ],
    },
    "ct-search": {
        "name": "Certificate Transparency Search",
        "description": "Search certificate transparency logs for certificates issued to a domain. Not available in this version.",
        "category": "Certificates",
        "actions": {"search": "Search"},
        "fields": {"domain": "Domain", "matchType": "Match type", "limit": "Limit"},
        "options": {"exact": "Exact", "contains": "Contains", "wildcard": "Wildcard"},
        "results": {"output": "Certificates"},
        "examples": [{"label": "Search example.com", "value": "example.com"}],
        "hints": {
            "unavailable": "This version does not integrate a CT log; calls return `NOT_IMPLEMENTED`.",
        },
        "faq": [
            {
                "q": "Why is it unavailable?",
                "a": "Public CT log search services have no stable SLA, so this version defers integration.",
            },
            {
                "q": "Is there an alternative?",
                "a": "Use the certificate chain fetcher to inspect the chain the host currently serves.",
            },
            {
                "q": "Will it come back?",
                "a": "Yes. The gap is tracked in the roadmap's known limitations.",
            },
        ],
    },
    "domain-verify": {
        "name": "Domain Verifier",
        "description": "Prove domain ownership through a DNS TXT / CNAME record or a file hosted on the site.",
        "category": "Certificates",
        "actions": {"verify": "Verify"},
        "fields": {
            "type": "Method",
            "domain": "Domain",
            "expected": "Expected value",
            "filePath": "File path",
            "resolvers": "DNS servers",
            "timeoutMs": "Timeout (ms)",
        },
        "options": {
            "dnsTxt": "DNS TXT",
            "dnsCname": "DNS CNAME",
            "httpFile": "Site file",
        },
        "results": {"output": "Result", "perResolver": "Per-resolver results"},
        "examples": [{"label": "Verify a TXT record", "value": "example.com / dns-txt"}],
        "hints": {
            "httpFile": "The site file is fetched from `https://<domain>/<path>` and must return 200 with an exact match.",
            "resolvers": "Leave empty to use the system resolvers; each resolver is compared separately.",
        },
        "faq": [
            {
                "q": "Does `dns-txt` support several resolvers?",
                "a": "Yes. Each resolver reports its own outcome and any match makes the overall result a match.",
            },
            {
                "q": "Why must the site file match exactly?",
                "a": "Exact matching avoids false positives; leading and trailing whitespace is ignored.",
            },
            {
                "q": "Can I verify an internal domain?",
                "a": "DNS methods can; `http-file` needs egress, so private addresses are blocked.",
            },
        ],
    },
}


def dump(locale: str, payload: dict[str, dict]) -> None:
    for slug, messages in payload.items():
        path = LOCALES / locale / "tools" / f"{slug}.json"
        path.parent.mkdir(parents=True, exist_ok=True)
        path.write_text(json.dumps(messages, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
        print("wrote", path.relative_to(ROOT))


dump("zh-CN", ZH)
dump("en-US", EN)
