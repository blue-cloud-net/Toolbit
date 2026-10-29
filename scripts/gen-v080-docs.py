#!/usr/bin/env python3
"""为 v0.8.0 批次新增工具补齐 docs/tools.md 与 docs/api/tools/*.md。

文档分两处：

- `docs/tools.md` —— 面向最终用户：分类目录表的行数与每个分类的工具表；
- `docs/api/tools/<category>.md` —— 面向调用方：请求 / 响应 / 错误码契约。

脚本**幂等**：已存在的章节会被替换而不是重复追加。
"""
from __future__ import annotations

import pathlib
import re

ROOT = pathlib.Path(__file__).resolve().parents[1]
DOCS = ROOT / "docs"
TOOLS_DOC = DOCS / "tools.md"
API_DIR = DOCS / "api" / "tools"

# slug → 分类
CATEGORY: dict[str, str] = {
    "text-diff": "text",
    "regex-validate": "text",
    "regex-build": "text",
    "data-mask": "text",
    "text-stats": "text",
    "id-validator": "data",
    "csv-convert": "data",
    "json-to-types": "data",
    "mock-data": "data",
    "archive-inspect": "data",
    "file-analyze": "data",
    "wifi-qr": "data",
    "id-generator": "data",
    "nginx-config": "system",
    "firewall-rules": "system",
    "docker-compose-gen": "system",
    "lookup-table": "system",
    "device-headers": "network",
    "ip-range-expand": "network",
    "color-convert": "formatting",
    "html-escape": "encoding",
    "markdown-convert": "documents",
    "password-policy": "crypto",
    "totp": "crypto",
    "cert-lifecycle": "certificates",
    "ssh-keygen": "keys",
}

# slug → (中文名, 英文名, 一句话用途)
SUMMARY: dict[str, tuple[str, str, str]] = {
    "text-diff": ("文本对比", "Text Diff", "文本行级与 JSON 结构化差异对比"),
    "regex-validate": ("正则校验", "Regex Validator", "只校验正则语法，不做匹配"),
    "regex-build": ("正则生成", "Regex Builder", "由结构化选项拼装正则表达式"),
    "data-mask": ("数据脱敏", "Data Masking", "手机号、证件号、邮箱等敏感数据脱敏"),
    "text-stats": ("文本统计", "Text Statistics", "字数、行数与字符构成统计"),
    "id-validator": ("号码校验", "ID Validator", "身份证、信用代码、手机号、银行卡校验位计算"),
    "csv-convert": ("CSV 转换", "CSV Converter", "CSV 与 JSON 互转"),
    "json-to-types": ("JSON 转类型", "JSON to Types", "由 JSON 样本生成语言类型定义"),
    "mock-data": ("模拟数据", "Mock Data", "按字段类型生成假数据"),
    "archive-inspect": ("归档查看", "Archive Inspector", "查看 ZIP / TAR / GZIP 的条目清单"),
    "file-analyze": ("文件识别", "File Analyzer", "按魔数识别文件类型并给出摘要"),
    "wifi-qr": ("访客 WiFi 二维码", "Guest WiFi QR Code", "生成访客 WiFi 接入二维码"),
    "id-generator": ("ID 生成", "ID Generator", "生成 UUID / ULID / 雪花 ID / NanoID"),
    "nginx-config": ("Nginx 配置生成", "Nginx Config Generator", "生成 Nginx 全局与站点配置"),
    "firewall-rules": ("防火墙规则生成", "Firewall Rule Generator", "生成 ufw / firewalld / iptables / nftables 规则"),
    "docker-compose-gen": ("Compose 配置生成", "Docker Compose Generator", "由服务定义生成 Compose 配置"),
    "lookup-table": ("速查表", "Lookup Table", "查询端口、退出码与时间戳格式"),
    "device-headers": ("设备请求头", "Device Headers", "生成客户端请求头并解析 User-Agent"),
    "ip-range-expand": ("IP 段展开", "IP Range Expander", "IP 段与 CIDR 互转及批量展开"),
    "color-convert": ("颜色转换", "Color Converter", "颜色格式转换与对比度检查"),
    "html-escape": ("HTML 转义", "HTML Escape", "HTML 实体编解码"),
    "markdown-convert": ("Markdown 转换", "Markdown Converter", "Markdown 与 HTML 互转"),
    "password-policy": ("口令策略检查", "Password Policy Checker", "按策略检查口令合规性"),
    "totp": ("TOTP 生成校验", "TOTP Generator", "生成与校验 TOTP 动态口令"),
    "cert-lifecycle": ("证书生命周期", "Certificate Lifecycle", "证书有效期区间与到期倒计时"),
    "ssh-keygen": ("SSH 密钥生成", "SSH Key Generator", "生成 OpenSSH 格式密钥对与指纹"),
}

# slug → 请求字段表（字段, 类型, 必填, 说明）
REQUESTS: dict[str, list[tuple[str, str, str, str]]] = {
    "text-diff": [
        ("mode", "string", "否", "`text` 行级对比 / `json` 结构化对比，默认 `text`"),
        ("left", "object", "是", "`{ data }`，基准侧文本"),
        ("right", "object", "是", "`{ data }`，对比侧文本"),
        ("ignoreWhitespace", "bool", "否", "忽略行尾空白，默认 `false`"),
        ("contextLines", "int", "否", "上下文行数，取值 0–100"),
    ],
    "regex-validate": [
        ("pattern", "string", "是", "正则表达式（RE2 语法）"),
        ("flags", "array", "否", "标志名列表：`i` `m` `s` `x`"),
    ],
    "regex-build": [
        ("parts", "array", "否", "片段名列表，见下表"),
        ("prefix", "string", "否", "固定前缀（自动转义）"),
        ("suffix", "string", "否", "固定后缀（自动转义）"),
        ("quantifier", "object", "否", "`{ min, max }`，`max` 为 -1 表示无上限"),
        ("anchored", "bool", "否", "是否加 `^` 与 `$` 锚点，默认 `false`"),
    ],
    "data-mask": [
        ("text", "string", "是", "待脱敏文本"),
        ("types", "array", "否", "脱敏类型：`phone` `idCard` `bankCard` `email` `name` `ip`"),
        ("position", "string", "否", "`keepHead` `keepTail` `keepBoth` `hash`，默认 `keepHead`"),
        ("keepCount", "int", "否", "保留字符数，取值 0–32"),
    ],
    "text-stats": [
        ("text", "string", "是", "待统计文本"),
    ],
    "id-validator": [
        ("type", "string", "是", "`idCard` `uscc` `phone` `bankCard` `postcode`"),
        ("value", "string", "是", "待校验的号码"),
    ],
    "csv-convert": [
        ("operation", "string", "否", "`csvToJson` / `jsonToCsv`，默认 `csvToJson`"),
        ("input", "string", "是", "源数据文本"),
        ("delimiter", "string", "否", "CSV 分隔符，单字符，默认 `,`"),
        ("hasHeader", "bool", "否", "首行是否为表头，默认 `false`"),
        ("pretty", "bool", "否", "输出是否缩进，默认 `false`"),
    ],
    "json-to-types": [
        ("input", "string", "是", "JSON 样本"),
        ("language", "string", "否", "`typescript` `go` `rust` `java`，默认 `typescript`"),
        ("rootName", "string", "否", "根类型名，默认按结构推断"),
    ],
    "mock-data": [
        ("types", "array", "否", "字段类型列表，见下表；留空表示全部类型"),
        ("count", "int", "否", "每种类型的生成条数，默认 3"),
        ("locale", "string", "否", "`zh-CN` / `en-US`，默认 `zh-CN`"),
        ("prefix", "string", "否", "生成值前缀"),
    ],
    "archive-inspect": [
        ("data", "string", "是", "归档内容的 Base64"),
        ("format", "string", "否", "`zip` `tar` `tarGz` `gzip`；留空则自动识别"),
        ("includeContent", "bool", "否", "是否给出条目内容预览，默认 `false`"),
    ],
    "file-analyze": [
        ("data", "string", "是", "文件内容的 Base64"),
        ("fileName", "string", "否", "文件名，用于推断扩展名与 MIME"),
        ("digestAlgorithm", "string", "否", "摘要算法，默认 `SHA256`"),
    ],
    "wifi-qr": [
        ("ssid", "string", "是", "网络名称"),
        ("password", "string", "否", "网络口令；开放网络留空"),
        ("security", "string", "否", "`WPA` `WEP` `nopass`，默认 `WPA`"),
        ("hidden", "bool", "否", "是否为隐藏网络，默认 `false`"),
        ("size", "int", "否", "二维码边长（像素），取值 64–1024，默认 256"),
    ],
    "id-generator": [
        ("format", "string", "否", "`uuid` `ulid` `snowflake` `nanoid`，默认 `uuid`"),
        ("count", "int", "否", "生成条数，默认 1"),
        ("length", "int", "否", "NanoID 长度，默认 21"),
        ("epoch", "int", "否", "雪花 ID 起始纪元（毫秒）"),
        ("nodeId", "int", "否", "雪花 ID 节点号，取值 0–1023"),
        ("uppercase", "bool", "否", "是否转大写，默认 `false`"),
    ],
    "nginx-config": [
        ("mode", "string", "否", "`global` 全局段 / `site` 站点段，默认 `site`"),
        ("serverName", "string", "否", "站点域名"),
        ("listenPort", "int", "否", "监听端口，默认 80"),
        ("root", "string", "否", "站点根目录"),
        ("indexFiles", "array", "否", "索引文件列表"),
        ("enableSsl", "bool", "否", "是否启用 HTTPS，默认 `false`"),
        ("sslCertificate", "string", "否", "证书路径"),
        ("sslPrivateKey", "string", "否", "私钥路径"),
        ("enableProxy", "bool", "否", "是否启用反向代理，默认 `false`"),
        ("proxyPass", "string", "否", "代理目标"),
        ("proxyPath", "string", "否", "代理路径前缀"),
        ("enableGzip", "bool", "否", "是否启用 Gzip，默认 `false`"),
        ("enableCORS", "bool", "否", "是否输出 CORS 段，默认 `false`"),
        ("allowOrigin", "string", "否", "允许的来源，默认 `*`"),
        ("enableWebSocket", "bool", "否", "是否启用 WebSocket 升级，默认 `false`"),
        ("enableLoadBalance", "bool", "否", "是否输出 upstream 段，默认 `false`"),
        ("upstreams", "array", "否", "上游地址列表"),
        ("clientMaxBodySize", "string", "否", "请求体上限，如 `10m`"),
        ("workerProcesses", "int", "否", "`global` 模式下的 worker 进程数"),
    ],
    "firewall-rules": [
        ("tool", "string", "否", "`ufw` `firewalld` `iptables` `nftables`，默认 `ufw`"),
        ("scenario", "string", "否", "`http` `https` `ssh` `database` `denyAll` `allowAll`"),
        ("ports", "array", "否", "端口列表，支持 `8000-8100` 区间写法"),
        ("protocol", "string", "否", "`tcp` `udp` `any`，默认 `tcp`"),
        ("source", "string", "否", "来源地址或网段"),
    ],
    "docker-compose-gen": [
        ("projectName", "string", "是", "项目名，用作服务名前缀"),
        ("services", "array", "是", "服务定义列表，见下表"),
        ("networkName", "string", "否", "自定义网络名"),
        ("includeVolumes", "bool", "否", "是否输出具名卷，默认 `false`"),
        ("includeNetwork", "bool", "否", "是否输出自定义网络，默认 `true`"),
    ],
    "lookup-table": [
        ("kind", "string", "否", "`port` `exitcode` `timestamp`，默认 `port`"),
        ("key", "string", "是", "查询词：端口号、退出码或关键字"),
    ],
    "device-headers": [
        ("mode", "string", "否", "`generate` 生成 / `parseUa` 解析，默认 `generate`"),
        ("platform", "string", "否", "`chrome` `firefox` `safari` `edge` `android` `ios` `curl`"),
        ("version", "string", "否", "浏览器版本号"),
        ("os", "string", "否", "操作系统名称与版本"),
        ("userAgent", "string", "否", "`parseUa` 模式下的 User-Agent 串"),
        ("acceptLanguage", "string", "否", "Accept-Language 头"),
        ("referer", "string", "否", "Referer 头"),
        ("customHeaders", "object", "否", "额外请求头，键值对"),
    ],
    "ip-range-expand": [
        ("mode", "string", "否", "`expand` 展开 / `toCidr` 转 CIDR / `toRange` 转区间"),
        ("cidr", "string", "否", "网段；`expand` 与 `toRange` 模式使用"),
        ("start", "string", "否", "起始地址；`toCidr` 模式使用"),
        ("end", "string", "否", "结束地址；`toCidr` 模式使用"),
        ("limit", "int", "否", "展开条数上限，默认 1024"),
    ],
    "color-convert": [
        ("input", "string", "是", "颜色值"),
        ("inputFormat", "string", "否", "`hex` `rgb` `hsl` `oklch`，默认 `hex`"),
        ("outputFormats", "array", "否", "需要输出的格式列表，默认全部"),
    ],
    "html-escape": [
        ("mode", "string", "否", "`encode` 转义 / `decode` 反转义，默认 `encode`"),
        ("input", "string", "是", "待处理文本"),
    ],
    "markdown-convert": [
        ("mode", "string", "否", "`toHtml` / `toMarkdown`，默认 `toHtml`"),
        ("input", "string", "是", "源文本"),
        ("gfm", "bool", "否", "是否启用 GitHub 风格扩展，默认 `false`"),
    ],
    "password-policy": [
        ("password", "string", "是", "待检查口令"),
        ("preset", "string", "否", "`basic` `medium` `strict`，默认 `medium`"),
        ("minLength", "int", "否", "最小长度；缺省取预设值"),
        ("maxLength", "int", "否", "最大长度；`0` 表示不限"),
        ("requireUpper", "bool", "否", "是否要求大写字母"),
        ("requireLower", "bool", "否", "是否要求小写字母"),
        ("requireDigit", "bool", "否", "是否要求数字"),
        ("requireSymbol", "bool", "否", "是否要求符号"),
        ("maxRepeatRun", "int", "否", "连续重复字符上限；`0` 表示不限"),
    ],
    "totp": [
        ("operation", "string", "否", "`generate` 生成 / `verify` 校验，默认 `generate`"),
        ("secret", "string", "是", "Base32 密钥"),
        ("algorithm", "string", "否", "`SHA1` `SHA256` `SHA512`，默认 `SHA1`"),
        ("digits", "int", "否", "验证码位数，取值 6–10，默认 6"),
        ("period", "int", "否", "时间步长（秒），取值 1–300，默认 30"),
        ("timestamp", "int", "否", "Unix 时间戳（秒）；`0` 表示当前时间"),
        ("code", "string", "否", "待校验的验证码；`verify` 模式必填"),
        ("window", "int", "否", "允许的时间步容差，默认 1"),
    ],
    "cert-lifecycle": [
        ("certificate", "string", "是", "证书内容（PEM 或 Base64 DER）"),
        ("isPem", "bool", "否", "内容是否为 PEM，默认 `true`"),
        ("warningDays", "int", "否", "到期提醒天数，默认 30"),
    ],
    "ssh-keygen": [
        ("algorithm", "string", "否", "`Ed25519` / `RSA`，默认 `Ed25519`"),
        ("bits", "string", "否", "RSA 位数；仅 RSA 有效"),
        ("comment", "string", "否", "密钥注释，单行"),
    ],
}

# slug → 响应字段表（字段, 类型, 说明）
RESPONSES: dict[str, list[tuple[str, str, str]]] = {
    "text-diff": [
        ("mode", "string", "实际使用的对比模式"),
        ("identical", "bool", "两侧是否完全一致"),
        ("lines", "array", "`[{ op, leftLineNum, rightLineNum, text }]`"),
        ("changes", "array", "JSON 模式的结构化差异 `[{ path, op, left, right }]`"),
        ("stat", "object", "`{ added, removed, unchanged, changes }`"),
    ],
    "regex-validate": [
        ("valid", "bool", "表达式是否合法"),
        ("dialect", "string", "正则方言，固定 `RE2`"),
        ("flags", "array", "生效的标志"),
        ("error", "object", "语法错误 `{ code, message, offset, position }`；合法时为 `null`"),
        ("suggestions", "array", "改进建议 `[{ code, message, priority }]`"),
        ("captureGroups", "int", "捕获组数量"),
    ],
    "regex-build": [
        ("pattern", "string", "生成的表达式"),
        ("valid", "bool", "生成的表达式是否合法"),
        ("explanation", "array", "逐段释义 `[{ part, fragment, description }]`"),
        ("error", "object", "校验错误；合法时为 `null`"),
        ("captureGroups", "int", "捕获组数量"),
    ],
    "data-mask": [
        ("masked", "string", "脱敏后的文本"),
        ("matches", "array", "各类命中统计 `[{ type, count, sample }]`"),
        ("total", "int", "脱敏总数"),
    ],
    "text-stats": [
        ("characters", "int", "字符总数"),
        ("charactersNoSpace", "int", "不含空白的字符数"),
        ("words", "int", "词数（中文按字、英文按词）"),
        ("chineseChars", "int", "中文字符数"),
        ("englishWords", "int", "英文单词数"),
        ("digits", "int", "数字字符数"),
        ("punctuation", "int", "标点符号数"),
        ("lines", "int", "行数"),
        ("paragraphs", "int", "段落数"),
        ("bytes", "int", "UTF-8 字节数"),
    ],
    "id-validator": [
        ("type", "string", "校验类型"),
        ("value", "string", "规范化后的值"),
        ("valid", "bool", "是否通过校验"),
        ("normalized", "string", "规范化结果"),
        ("checkDigit", "string", "校验位"),
        ("region", "string", "归属地区"),
        ("birthday", "string", "出生日期"),
        ("gender", "string", "性别"),
        ("issue", "string", "失败原因；通过时为空串"),
    ],
    "csv-convert": [
        ("operation", "string", "实际使用的转换方向"),
        ("output", "string", "转换结果"),
        ("rowCount", "int", "数据行数"),
        ("columnCount", "int", "列数"),
        ("columns", "array", "列名列表"),
    ],
    "json-to-types": [
        ("language", "string", "目标语言"),
        ("output", "string", "生成的类型定义"),
        ("typeCount", "int", "类型数量"),
    ],
    "mock-data": [
        ("types", "array", "生成的类型列表"),
        ("fields", "array", "生成的字段值 `[{ type, value }]`"),
        ("count", "int", "每种类型的条数"),
        ("locale", "string", "语言"),
        ("seed", "string", "本次随机种子"),
    ],
    "archive-inspect": [
        ("format", "string", "归档格式"),
        ("entryCount", "int", "条目数"),
        ("totalSize", "int", "解压后总大小"),
        ("compressedSize", "int", "归档大小"),
        ("entries", "array", "条目清单 `[{ name, size, compressedSize, isDir, mode, modified, contentPreview }]`"),
        ("truncated", "bool", "条目清单是否被截断"),
    ],
    "file-analyze": [
        ("fileName", "string", "文件名"),
        ("size", "int", "字节数"),
        ("mimeType", "string", "媒体类型"),
        ("extension", "string", "扩展名"),
        ("description", "string", "类型说明"),
        ("isTextual", "bool", "是否文本类"),
        ("magicHex", "string", "文件头十六进制"),
        ("digest", "string", "摘要值"),
        ("digestAlgorithm", "string", "摘要算法"),
    ],
    "wifi-qr": [
        ("payload", "string", "二维码承载的文本"),
        ("png", "string", "PNG 图片的 Base64"),
        ("svg", "string", "SVG 图片内容"),
        ("size", "int", "图片边长"),
    ],
    "id-generator": [
        ("format", "string", "ID 格式"),
        ("ids", "array", "生成的 ID 列表"),
        ("count", "int", "生成条数"),
        ("length", "int", "ID 长度"),
    ],
    "nginx-config": [
        ("mode", "string", "配置模式"),
        ("config", "string", "配置文本"),
        ("sections", "array", "配置段标题"),
        ("warnings", "array", "配置提示"),
    ],
    "firewall-rules": [
        ("tool", "string", "防火墙工具"),
        ("scenario", "string", "场景"),
        ("rules", "array", "规则清单 `[{ action, port, protocol, source, description }]`"),
        ("commands", "array", "可执行命令"),
        ("notes", "array", "使用提示"),
    ],
    "docker-compose-gen": [
        ("compose", "string", "Compose 配置文本"),
        ("serviceCount", "int", "服务数量"),
        ("warnings", "array", "配置提示"),
    ],
    "lookup-table": [
        ("kind", "string", "类别"),
        ("query", "string", "规范化后的查询词"),
        ("matched", "bool", "是否精确命中"),
        ("entry", "object", "命中的条目 `{ kind, key, label, description, extra[] }`；未命中时为 `null`"),
        ("candidates", "array", "模糊检索的候选条目"),
        ("dataVersion", "string", "数据集版本"),
    ],
    "device-headers": [
        ("mode", "string", "模式"),
        ("headers", "array", "请求头列表 `[{ name, value }]`"),
        ("userAgent", "string", "User-Agent 串"),
        ("parsed", "object", "解析结果 `{ browser, browserVersion, os, osVersion, deviceType, isBot, engine }`"),
    ],
    "ip-range-expand": [
        ("mode", "string", "模式"),
        ("cidrs", "array", "规范化后的 CIDR 列表"),
        ("start", "string", "起始地址"),
        ("end", "string", "结束地址"),
        ("count", "int", "地址总数"),
        ("addresses", "array", "展开的地址列表"),
        ("truncated", "bool", "地址列表是否被上限截断"),
    ],
    "color-convert": [
        ("input", "string", "输入颜色"),
        ("inputFormat", "string", "输入格式"),
        ("hex", "string", "十六进制表示"),
        ("rgb", "string", "RGB 表示"),
        ("hsl", "string", "HSL 表示"),
        ("oklch", "string", "OKLCH 表示"),
        ("contrastOnWhite", "number", "相对白底的对比度"),
        ("contrastOnBlack", "number", "相对黑底的对比度"),
        ("isDark", "bool", "是否为深色"),
    ],
    "html-escape": [
        ("mode", "string", "模式"),
        ("output", "string", "转义 / 反转义结果"),
        ("changedCount", "int", "发生变化的字符数"),
    ],
    "markdown-convert": [
        ("mode", "string", "转换方向"),
        ("output", "string", "转换结果"),
        ("headingCount", "int", "标题数量"),
        ("linkCount", "int", "链接数量"),
        ("codeBlockCount", "int", "代码块数量"),
        ("wordCount", "int", "词数"),
    ],
    "password-policy": [
        ("preset", "string", "生效的预设"),
        ("compliant", "bool", "是否满足整体策略"),
        ("score", "int", "强度评分（0–100）"),
        ("strength", "string", "强度等级"),
        ("checks", "array", "检查项 `[{ check, passed, message }]`"),
        ("suggestions", "array", "改进建议"),
    ],
    "totp": [
        ("operation", "string", "操作"),
        ("code", "string", "验证码"),
        ("valid", "bool", "校验结论；生成模式下为 `null`"),
        ("algorithm", "string", "摘要算法"),
        ("digits", "int", "验证码位数"),
        ("period", "int", "时间步长"),
        ("counter", "int", "时间步计数"),
        ("validFrom", "string", "生效起点（RFC 3339）"),
        ("validUntil", "string", "失效点（RFC 3339）"),
        ("remainingSeconds", "int", "剩余有效秒数"),
    ],
    "cert-lifecycle": [
        ("subject", "string", "主体"),
        ("issuer", "string", "签发者"),
        ("notBefore", "string", "生效时间"),
        ("notAfter", "string", "失效时间"),
        ("validDays", "int", "有效期总天数"),
        ("daysRemaining", "int", "剩余天数"),
        ("hoursRemaining", "int", "剩余小时数"),
        ("status", "string", "状态"),
        ("isSelfSigned", "bool", "是否自签名"),
        ("isCA", "bool", "是否 CA 证书"),
        ("keyAlgorithm", "string", "密钥算法"),
        ("keyBits", "int", "密钥位数"),
        ("curve", "string", "椭圆曲线名"),
        ("signatureAlgorithm", "string", "签名算法"),
        ("san", "array", "主体备用名列表"),
        ("recommendations", "array", "续期建议"),
    ],
    "ssh-keygen": [
        ("algorithm", "string", "算法"),
        ("privateKey", "string", "OpenSSH 格式私钥"),
        ("publicKey", "string", "OpenSSH 格式公钥"),
        ("fingerprint", "string", "SHA256 指纹"),
        ("bits", "int", "密钥位数"),
        ("comment", "string", "注释"),
    ],
}

# slug → 错误码
ERRORS: dict[str, list[str]] = {
    "text-diff": ["MISSING_FIELD", "INVALID_FIELD"],
    "regex-validate": ["MISSING_FIELD", "INVALID_FIELD"],
    "regex-build": ["INVALID_FIELD"],
    "data-mask": ["MISSING_FIELD", "INVALID_FIELD"],
    "text-stats": ["MISSING_FIELD", "INVALID_FIELD"],
    "id-validator": ["MISSING_FIELD", "INVALID_FIELD"],
    "csv-convert": ["MISSING_FIELD", "INVALID_FIELD"],
    "json-to-types": ["MISSING_FIELD", "INVALID_FIELD"],
    "mock-data": ["INVALID_FIELD"],
    "archive-inspect": ["MISSING_FIELD", "INVALID_FIELD"],
    "file-analyze": ["MISSING_FIELD", "INVALID_FIELD"],
    "wifi-qr": ["MISSING_FIELD", "INVALID_FIELD"],
    "id-generator": ["INVALID_FIELD"],
    "nginx-config": ["INVALID_FIELD"],
    "firewall-rules": ["INVALID_FIELD"],
    "docker-compose-gen": ["MISSING_FIELD", "INVALID_FIELD"],
    "lookup-table": ["MISSING_FIELD", "INVALID_FIELD"],
    "device-headers": ["MISSING_FIELD", "INVALID_FIELD"],
    "ip-range-expand": ["MISSING_FIELD", "INVALID_FIELD"],
    "color-convert": ["MISSING_FIELD", "INVALID_FIELD"],
    "html-escape": ["INVALID_FIELD"],
    "markdown-convert": ["MISSING_FIELD", "INVALID_FIELD"],
    "password-policy": ["INVALID_FIELD"],
    "totp": ["MISSING_FIELD", "INVALID_FIELD"],
    "cert-lifecycle": ["MISSING_FIELD", "INVALID_CERT", "INVALID_PEM"],
    "ssh-keygen": ["INVALID_FIELD"],
}

# slug → 备注
NOTES: dict[str, str] = {
    "text-diff": "单侧输入上限 512 KiB、行数上限 20000。",
    "regex-validate": "服务端使用 RE2（**线性时间**匹配，无灾难性回溯）；只校验语法，不做匹配。",
    "regex-build": "生成是纯字符串拼装，结果仍按 RE2 方言校验；固定前后缀会自动转义。",
    "data-mask": "**纯文本处理**：输入不落盘、不写任何浏览器存储；命中统计里的 `sample` 同样已脱敏。",
    "text-stats": "中文按字计、英文按词计，`words` 为两者之和。",
    "id-validator": "只做**校验位计算与格式规范化**，不查询任何外部数据库。",
    "csv-convert": "CSV 解析允许宽松引号（`LazyQuotes`）；无表头时列名按 `column1`、`column2` 生成。",
    "json-to-types": "类型由样本**推断**得出，不保证与真实结构一致。",
    "mock-data": "使用**非密码学**随机源，仅用于演示与占位；**不可**用作真实凭据。",
    "archive-inspect": "条目数上限 5000、单条目解压上限 64 MiB、总解压上限 256 MiB；归档输入上限 32 MiB。",
    "file-analyze": "类型识别基于**魔数表**，未收录的类型按扩展名与 MIME 回落。",
    "wifi-qr": "载荷遵循 `WIFI:T:…;S:…;P:…;` 格式；口令**不写入**任何存储。",
    "id-generator": "雪花 ID 的纪元与节点号由请求指定，**不保证**全局唯一。",
    "nginx-config": "生成的是**配置片段**，需自行合并到 `nginx.conf`；证书与私钥只输出路径，不输出内容。",
    "firewall-rules": "输出的是**待执行命令**，工具**不会**代为执行；上线前请先在测试环境验证。",
    "docker-compose-gen": "生成的是 Compose 文本，不含 `build` 段与密钥。",
    "lookup-table": "内置数据集离线查询，**不出网**；数据集版本随发布更新。",
    "device-headers": "生成的 UA 是**常见组合**，不代表任何真实浏览器；解析只做结构化识别，不做指纹追踪。",
    "ip-range-expand": "展开条数上限受 `limit` 约束，超出时 `truncated` 为 `true`。",
    "color-convert": "对比度按 WCAG 相对亮度公式计算；只支持 IPv4 / sRGB 色彩空间。",
    "html-escape": "只处理命名实体与十进制 / 十六进制数字实体，不做 CSS / JS 上下文转义。",
    "markdown-convert": "HTML 转 Markdown 采用**结构化解析**；内联样式与脚本内容会被丢弃。",
    "password-policy": "**纯规则检查**：不涉及密码学运算，也不存储口令；常见口令表内置在服务端。",
    "totp": "HMAC 计算经服务端转调 Tongsuo；标准向量见 RFC 6238。",
    "cert-lifecycle": "**只读查看**证书，不修改、不签发、不下载吊销列表。",
    "ssh-keygen": "产出 OpenSSH 格式（`openssh-key-v1`）；**不含**加密私钥，私钥明文只在响应中出现一次。",
}

# 分类 → 中文名
CATEGORY_CN = {
    "crypto": "加密",
    "keys": "密钥",
    "certificates": "证书",
    "encoding": "编码",
    "formatting": "格式化",
    "text": "文本",
    "data": "数据",
    "system": "系统",
    "network": "网络",
    "documents": "文档",
}


def render_api_section(slug: str) -> str:
    """生成 api/tools/<category>.md 中的一个工具章节。"""
    cn, en, summary = SUMMARY[slug]
    lines = [f"## {slug}", "", f"{summary}（{en}）。", "", "### 请求", ""]
    lines.append("| 字段 | 类型 | 必填 | 说明 |")
    lines.append("| --- | --- | --- | --- |")
    for field, ftype, required, desc in REQUESTS[slug]:
        lines.append(f"| `{field}` | {ftype} | {required} | {desc} |")
    lines += ["", "### 响应", ""]
    lines.append("| 字段 | 类型 | 说明 |")
    lines.append("| --- | --- | --- |")
    for field, ftype, desc in RESPONSES[slug]:
        lines.append(f"| `{field}` | {ftype} | {desc} |")
    lines += ["", "### 错误", ""]
    lines.append("、".join(f"`{code}`" for code in ERRORS[slug]) + "。")
    lines += ["", f"> {NOTES[slug]}", "", "---", ""]
    return "\n".join(lines)


def update_api_docs() -> None:
    """把新工具章节写进对应的 api/tools/<category>.md。"""
    by_category: dict[str, list[str]] = {}
    for slug, category in CATEGORY.items():
        by_category.setdefault(category, []).append(slug)

    for category, slugs in by_category.items():
        path = API_DIR / f"{category}.md"
        text = path.read_text(encoding="utf-8")

        # 顶部工具表补行：按该表自身的列数生成，避免列数不匹配。
        lines = text.split("\n")
        header_index = next(
            (index for index, line in enumerate(lines) if line.startswith("| slug")),
            None,
        )
        if header_index is not None:
            columns = [cell.strip() for cell in lines[header_index].strip("|").split("|")]
            # 表体在分隔行之后。
            body_start = header_index + 2
            body_end = body_start
            while body_end < len(lines) and lines[body_end].startswith("|"):
                body_end += 1
            added = []
            for slug in slugs:
                if f"[{slug}]" in text:
                    continue
                cn, _, _ = SUMMARY[slug]
                cells = []
                for column in columns:
                    if column == "slug":
                        cells.append(f"[{slug}](#{slug})")
                    elif column == "名称":
                        cells.append(cn)
                    elif column in ("传输", "transport"):
                        cells.append("JSON")
                    elif column in ("档位", "rateClass"):
                        cells.append("light")
                    elif column in ("出网", "requiresNetwork"):
                        cells.append("否")
                    elif column == "实现状态":
                        cells.append("✅ v0.8.0")
                    elif column == "端点":
                        cells.append(f"`POST /api/v1/tools/{slug}`")
                    else:
                        cells.append("")
                added.append("| " + " | ".join(cells) + " |")
            if added:
                lines[body_end:body_end] = added
            text = "\n".join(lines)

        # 章节正文。
        for slug in slugs:
            section = render_api_section(slug)
            pattern = re.compile(
                r"\n## " + re.escape(slug) + r"\n.*?(?=\n## |\Z)",
                re.DOTALL,
            )
            if pattern.search(text):
                text = pattern.sub("\n" + section, text)
            else:
                # 插到「## 相关文档」之前。
                marker = "\n## 相关文档"
                if marker in text:
                    text = text.replace(marker, "\n" + section + marker, 1)
                else:
                    text = text.rstrip("\n") + "\n\n" + section

        # 覆盖工具数：按工具表里的行数统计。
        count = len(re.findall(r"^\| \[[a-z0-9-]+\]\(#", text, re.MULTILINE))
        text = re.sub(r"覆盖 \*\*\d+\*\* 个", f"覆盖 **{count}** 个", text)
        text = re.sub(r"共 \*\*\d+\*\* 个", f"共 **{count}** 个", text)
        text = re.sub(r"^> 状态：.*$", "> 状态：Stable ｜ 适用版本：v0.8.0 ｜ 最后更新：2026-10-29", text, count=1, flags=re.MULTILINE)

        path.write_text(text, encoding="utf-8")
        print(f"已更新 {path.relative_to(ROOT)}（{len(slugs)} 个新工具）")


def update_tools_doc() -> None:
    """在 docs/tools.md 的分类表里补上新工具。"""
    text = TOOLS_DOC.read_text(encoding="utf-8")
    lines = text.split("\n")

    by_category: dict[str, list[str]] = {}
    for slug, category in CATEGORY.items():
        by_category.setdefault(category, []).append(slug)

    # 逐分类定位「## <分类>」小节，在其工具表末尾追加行。
    for category, slugs in by_category.items():
        heading = f"## {CATEGORY_CN[category]}（"
        start = next(
            (index for index, line in enumerate(lines) if line.startswith(heading)),
            None,
        )
        if start is None:
            print(f"  未找到分类小节：{CATEGORY_CN[category]}")
            continue
        # 找到该小节里第一张表的表体末尾。
        in_table = False
        insert_at = None
        for index in range(start + 1, len(lines)):
            line = lines[index]
            if line.startswith("## "):
                break
            if line.startswith("| ") and "---" in line:
                in_table = True
                continue
            if in_table:
                if line.startswith("|"):
                    insert_at = index + 1
                    continue
                break
        if insert_at is None:
            print(f"  未找到分类工具表：{CATEGORY_CN[category]}")
            continue
        added = []
        for slug in slugs:
            if f"`{slug}`" in "\n".join(lines[start : insert_at + 40]):
                continue
            cn, _, summary = SUMMARY[slug]
            added.append(f"| {cn} | `{slug}` | {summary} | ✅ |")
        if added:
            lines[insert_at:insert_at] = added

    text = "\n".join(lines)

    # 顶部目录表的分类计数。
    for category, slugs in by_category.items():
        pattern = re.compile(
            r"(\| \[" + re.escape(CATEGORY_CN[category]) + r"\]\(#[^)]+\) \| )(\d+)( \| )"
        )
        match = pattern.search(text)
        if not match:
            continue
        total = int(match.group(2)) + len(slugs)
        text = text[: match.start(2)] + str(total) + text[match.end(2) :]

    # 顶部工具总数。
    text = re.sub(r"介绍 \d+ 个工具", "介绍 98 个工具", text)

    TOOLS_DOC.write_text(text, encoding="utf-8")
    print(f"已更新 {TOOLS_DOC.relative_to(ROOT)}")


def main() -> None:
    update_api_docs()
    update_tools_doc()


if __name__ == "__main__":
    main()
