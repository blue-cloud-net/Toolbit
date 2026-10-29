#!/usr/bin/env python3
"""为 v0.8.0 批次新增工具补齐 `examples` 词条。

`examples` 是双语数组：`[label_zh, label_en]` 与 `[value_zh, value_en]`。
示例值一律使用**测试数据**，不含任何真实凭据。
"""
from __future__ import annotations

import json
import pathlib

ROOT = pathlib.Path(__file__).resolve().parents[1]
DATA = ROOT / "scripts" / "v080-locale-data.json"

# slug → [(标签, 值)]
EXAMPLES: dict[str, list[tuple[tuple[str, str], str]]] = {
    "text-diff": [
        (("对比两段文本", "Compare two texts"), '{"left":{"data":"a\\nb\\nc\\n"},"right":{"data":"a\\nx\\nc\\n"},"mode":"text"}'),
        (("对比 JSON 结构", "Compare JSON structure"), '{"left":{"data":"{\\"a\\":1,\\"b\\":{\\"c\\":2}}"},"right":{"data":"{\\"a\\":1,\\"b\\":{\\"c\\":3}}"},"mode":"json"}'),
    ],
    "regex-validate": [
        (("校验手机号表达式", "Validate a phone pattern"), '{"pattern":"^1[3-9]\\\\d{9}$"}'),
        (("校验有语法错误的表达式", "Validate a broken pattern"), '{"pattern":"([a-z"}'),
    ],
    "regex-build": [
        (("拼接数字与连字符", "Digits and hyphen"), '{"parts":["digits","hyphen","digits"]}'),
        (("带量词的字母片段", "Letters with quantifier"), '{"parts":["letters"],"quantifier":{"min":2,"max":8}}'),
    ],
    "data-mask": [
        (("脱敏手机号", "Mask a phone number"), '{"text":"我的手机号是 13800138000","types":["phone"],"keepCount":3}'),
        (("脱敏身份证并保留尾部", "Mask an ID card keeping the tail"), '{"text":"身份证 110101199003078888","types":["idCard"],"position":"keepTail","keepCount":3}'),
    ],
    "text-stats": [
        (("统计中英混排文本", "Count mixed Chinese and English"), '{"text":"Hello 世界\\n第二行"}'),
    ],
    "id-validator": [
        (("校验身份证号", "Validate an ID card number"), '{"type":"idCard","value":"11010519491231002X"}'),
        (("校验银行卡号", "Validate a bank card number"), '{"type":"bankCard","value":"4111111111111111"}'),
    ],
    "csv-convert": [
        (("CSV 转 JSON", "CSV to JSON"), '{"operation":"csvToJson","input":"name,age\\nalice,30\\nbob,25\\n","hasHeader":true}'),
        (("JSON 转 CSV", "JSON to CSV"), '{"operation":"jsonToCsv","input":"[{\\"name\\":\\"alice\\",\\"age\\":30}]","hasHeader":true}'),
    ],
    "json-to-types": [
        (("生成 TypeScript 类型", "Generate TypeScript types"), '{"input":"{\\"id\\":1,\\"name\\":\\"a\\"}","language":"typescript","rootName":"User"}'),
        (("生成 Go 类型", "Generate Go types"), '{"input":"{\\"id\\":1}","language":"go","rootName":"Item"}'),
    ],
    "mock-data": [
        (("生成字符串与整数", "Generate strings and integers"), '{"types":["string","int"],"count":3}'),
        (("生成英文邮箱", "Generate English emails"), '{"types":["email"],"count":2,"locale":"en-US"}'),
    ],
    "archive-inspect": [
        (("查看 ZIP 条目", "Inspect a ZIP archive"), '{"format":"zip"}'),
    ],
    "file-analyze": [
        (("识别 PNG 文件", "Identify a PNG file"), '{"fileName":"a.png","digestAlgorithm":"SHA256"}'),
    ],
    "wifi-qr": [
        (("生成 WPA 二维码", "Generate a WPA QR code"), '{"ssid":"Guest-Network","password":"guest-password","security":"WPA","size":256}'),
        (("生成开放网络二维码", "Generate an open-network QR code"), '{"ssid":"Open","password":"","security":"nopass","size":256}'),
    ],
    "id-generator": [
        (("生成 UUID", "Generate UUIDs"), '{"format":"uuid","count":3}'),
        (("生成 NanoID", "Generate NanoIDs"), '{"format":"nanoid","count":2,"length":10}'),
    ],
    "nginx-config": [
        (("生成站点配置", "Generate a site config"), '{"mode":"site","serverName":"example.com","listenPort":80,"root":"/var/www/html"}'),
        (("生成全局配置", "Generate a global config"), '{"mode":"global","workerProcesses":4}'),
    ],
    "firewall-rules": [
        (("生成 HTTPS 放行规则", "Allow HTTPS"), '{"tool":"ufw","scenario":"https","protocol":"tcp"}'),
        (("生成 iptables 的 SSH 规则", "Allow SSH with iptables"), '{"tool":"iptables","scenario":"ssh","protocol":"tcp"}'),
    ],
    "docker-compose-gen": [
        (("生成单服务 Compose", "Generate a single-service Compose"), '{"projectName":"demo","services":[{"name":"web","image":"nginx:alpine","ports":["8080:80"]}]}'),
    ],
    "lookup-table": [
        (("查询 443 端口", "Look up port 443"), '{"kind":"port","key":"443"}'),
        (("检索退出码", "Search exit codes"), '{"kind":"exitcode","key":"错误"}'),
    ],
    "device-headers": [
        (("生成 Chrome 请求头", "Generate Chrome headers"), '{"mode":"generate","platform":"chrome","version":"120.0.0.0"}'),
        (("解析 User-Agent", "Parse a User-Agent"), '{"mode":"parseUa","userAgent":"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"}'),
    ],
    "ip-range-expand": [
        (("展开 /30 网段", "Expand a /30"), '{"mode":"expand","cidr":"192.168.1.0/30","limit":16}'),
        (("地址区间转 CIDR", "Convert a range to CIDRs"), '{"mode":"toCidr","start":"192.168.1.10","end":"192.168.1.20"}'),
    ],
    "color-convert": [
        (("十六进制转多格式", "Convert hex to all formats"), '{"input":"#3B82F6","inputFormat":"hex","outputFormats":["hex","rgb","hsl","oklch"]}'),
        (("RGB 字符串输入", "Convert an RGB string"), '{"input":"rgb(59, 130, 246)","inputFormat":"rgb"}'),
    ],
    "html-escape": [
        (("转义 HTML 标签", "Escape HTML tags"), '{"mode":"encode","input":"<b>hi</b> & \\"x\\""}'),
        (("反转义 HTML 实体", "Unescape HTML entities"), '{"mode":"decode","input":"&lt;b&gt;hi&lt;/b&gt;"}'),
    ],
    "markdown-convert": [
        (("Markdown 转 HTML", "Markdown to HTML"), '{"mode":"toHtml","input":"# Title\\n\\nSome **bold** text.","gfm":true}'),
        (("HTML 转 Markdown", "HTML to Markdown"), '{"mode":"toMarkdown","input":"<h1>Title</h1><p>Body</p>"}'),
    ],
    "password-policy": [
        (("检查强口令", "Check a strong password"), '{"password":"Zx7#quietRiver9","preset":"medium"}'),
        (("检查弱口令", "Check a weak password"), '{"password":"abc","preset":"strict"}'),
    ],
    "totp": [
        (("生成验证码", "Generate a code"), '{"operation":"generate","secret":"JBSWY3DPEHPK3PXP","algorithm":"SHA1","digits":6,"period":30}'),
        (("校验验证码", "Verify a code"), '{"operation":"verify","secret":"JBSWY3DPEHPK3PXP","algorithm":"SHA1","digits":6,"timestamp":59,"code":"94287082"}'),
    ],
    "cert-lifecycle": [
        (("查看证书有效期", "Inspect a certificate validity"), '{"warningDays":30}'),
    ],
    "ssh-keygen": [
        (("生成 Ed25519 密钥", "Generate an Ed25519 key"), '{"algorithm":"Ed25519","comment":"toolbit@example.com"}'),
        (("生成 RSA 密钥", "Generate an RSA key"), '{"algorithm":"RSA","bits":3072,"comment":"toolbit@example.com"}'),
    ],
}


def main() -> None:
    data = json.loads(DATA.read_text(encoding="utf-8"))
    missing = sorted(set(data) - set(EXAMPLES))
    extra = sorted(set(EXAMPLES) - set(data))
    if missing or extra:
        raise SystemExit(f"示例集合与词条数据不一致：缺 {missing}，多 {extra}")

    for slug, examples in EXAMPLES.items():
        tool = data[slug]
        tool["examples"] = [
            {"label": [label[0], label[1]], "value": value}
            for label, value in examples
        ]

    DATA.write_text(
        json.dumps(data, ensure_ascii=False, indent=2) + "\n",
        encoding="utf-8",
    )
    print(f"已为 {len(EXAMPLES)} 个工具补齐 examples")


if __name__ == "__main__":
    main()
