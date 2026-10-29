#!/usr/bin/env python3
"""生成 v0.8.0 批次新增工具的契约测试数据。"""
from __future__ import annotations

import json
import pathlib

ROOT = pathlib.Path(__file__).resolve().parents[1]
OUT = ROOT / "backend" / "testdata" / "contract"

CONTRACTS: dict[str, dict] = {
    "text-diff": {
        "description": "文本 / JSON 对比的契约测试数据。",
        "cases": [
            {
                "name": "相同文本",
                "request": {
                    "mode": "text",
                    "left": {"data": "a\nb\nc\n"},
                    "right": {"data": "a\nb\nc\n"},
                },
                "expect": {"code": 0, "dataContains": {"mode": "text", "identical": True}},
            },
            {
                "name": "存在差异",
                "request": {
                    "mode": "text",
                    "left": {"data": "a\nb\nc\n"},
                    "right": {"data": "a\nx\nc\n"},
                },
                "expect": {
                    "code": 0,
                    "dataContains": {"identical": False, "stat": {"added": 1, "removed": 1}},
                },
            },
            {
                "name": "JSON 结构化对比",
                "request": {
                    "mode": "json",
                    "left": {"data": '{"a":1,"b":{"c":2}}'},
                    "right": {"data": '{"a":1,"b":{"c":3}}'},
                },
                "expect": {"code": 0, "dataContains": {"mode": "json"}},
            },
            {
                "name": "缺少左��",
                "request": {"mode": "text", "right": {"data": "a\n"}},
                "expect": {"code": 400, "errorCode": "MISSING_FIELD"},
            },
            {
                "name": "不支持的模式",
                "request": {
                    "mode": "xml",
                    "left": {"data": "a"},
                    "right": {"data": "b"},
                },
                "expect": {"code": 400, "errorCode": "INVALID_FIELD"},
            },
        ],
    },
    "regex-validate": {
        "description": "正则校验的契约测试数据。",
        "cases": [
            {
                "name": "合法表达式",
                "request": {"pattern": "^\\d{3}-\\d{4}$"},
                "expect": {"code": 0, "dataContains": {"valid": True, "dialect": "RE2"}},
            },
            {
                "name": "非法表达式",
                "request": {"pattern": "([a-z"},
                "expect": {"code": 0, "dataContains": {"valid": False}},
            },
            {
                "name": "带标志",
                "request": {"pattern": "abc", "flags": ["i", "m"]},
                "expect": {"code": 0, "dataContains": {"valid": True}},
            },
            {
                "name": "缺少模式",
                "request": {},
                "expect": {"code": 400, "errorCode": "MISSING_FIELD"},
            },
        ],
    },
    "regex-build": {
        "description": "正则生成器的契约测试数据。",
        "cases": [
            {
                "name": "基础片段",
                "request": {"parts": ["digits", "hyphen", "digits"]},
                "expect": {"code": 0, "dataContains": {"valid": True}},
            },
            {
                "name": "带量词与锚点",
                "request": {
                    "parts": ["letters"],
                    "quantifier": {"min": 2, "max": 8},
                    "anchored": True,
                },
                "expect": {"code": 0, "dataContains": {"valid": True}},
            },
            {
                "name": "未知片段",
                "request": {"parts": ["not-a-part"]},
                "expect": {"code": 400, "errorCode": "INVALID_FIELD"},
            },
            {
                "name": "量词上下界颠倒",
                "request": {"parts": ["letter"], "quantifier": {"min": 5, "max": 2}},
                "expect": {"code": 400, "errorCode": "INVALID_FIELD"},
            },
        ],
    },
    "data-mask": {
        "description": "数据脱敏的契约测试数据。",
        "cases": [
            {
                "name": "手机号脱敏",
                "request": {"text": "我的手机号是 13800138000", "types": ["phone"], "keepCount": 3},
                "expect": {"code": 0, "dataContains": {"masked": "我的手机号是 138********"}},
            },
            {
                "name": "邮箱脱敏",
                "request": {"text": "联系 user@example.com", "types": ["email"]},
                "expect": {"code": 0, "dataContains": {"total": 1}},
            },
            {
                "name": "保留尾部",
                "request": {
                    "text": "身份证 110101199003078888",
                    "types": ["idCard"],
                    "position": "keepTail",
                    "keepCount": 3,
                },
                "expect": {"code": 0, "dataContains": {"masked": "身份证 ***************888"}},
            },
            {
                "name": "缺少文本",
                "request": {"types": ["phone"]},
                "expect": {"code": 400, "errorCode": "MISSING_FIELD"},
            },
            {
                "name": "未知脱敏类型",
                "request": {"text": "abc", "types": ["unknown"]},
                "expect": {"code": 400, "errorCode": "INVALID_FIELD"},
            },
        ],
    },
    "text-stats": {
        "description": "文本统计的契约测试数据。",
        "cases": [
            {
                "name": "中英混合文本",
                "request": {"text": "Hello 世界\n第二行"},
                "expect": {
                    "code": 0,
                    "dataContains": {"lines": 2, "chineseChars": 5, "englishWords": 1},
                },
            },
            {
                "name": "空文本",
                "request": {"text": ""},
                "expect": {"code": 400, "errorCode": "MISSING_FIELD"},
            },
        ],
    },
    "id-validator": {
        "description": "证件 / 号码校验的契约测试数据。",
        "cases": [
            {
                "name": "合法身份证",
                "request": {"type": "idCard", "value": "11010519491231002X"},
                "expect": {"code": 0, "dataContains": {"valid": True, "checkDigit": "X"}},
            },
            {
                "name": "校验位不匹配的身份证",
                "request": {"type": "idCard", "value": "110105194912310021"},
                "expect": {"code": 0, "dataContains": {"valid": False}},
            },
            {
                "name": "银行卡 Luhn 校验",
                "request": {"type": "bankCard", "value": "4111111111111111"},
                "expect": {"code": 0, "dataContains": {"valid": True}},
            },
            {
                "name": "统一社会信用代码",
                "request": {
                    "type": "uscc",
                    "value": "91110108MA01ABCD4X",
                },
                "expect": {"code": 0, "dataContains": {"type": "uscc"}},
            },
            {
                "name": "手机号",
                "request": {"type": "phone", "value": "13800138000"},
                "expect": {"code": 0, "dataContains": {"valid": True}},
            },
        ],
    },
    "csv-convert": {
        "description": "CSV / JSON 互转的契约测试数据。",
        "cases": [
            {
                "name": "CSV 转 JSON",
                "request": {
                    "operation": "csvToJson",
                    "input": "name,age\nalice,30\nbob,25\n",
                    "hasHeader": True,
                },
                "expect": {"code": 0, "dataContains": {"rowCount": 2, "columnCount": 2}},
            },
            {
                "name": "JSON 转 CSV",
                "request": {
                    "operation": "jsonToCsv",
                    "input": '[{"name":"alice","age":30}]',
                    "hasHeader": True,
                },
                "expect": {"code": 0, "dataContains": {"columnCount": 2}},
            },
            {
                "name": "不支持的操作",
                "request": {"operation": "xmlToJson", "input": "a"},
                "expect": {"code": 400, "errorCode": "INVALID_FIELD"},
            },
            {
                "name": "缺少输入",
                "request": {"operation": "csvToJson"},
                "expect": {"code": 400, "errorCode": "MISSING_FIELD"},
            },
        ],
    },
    "json-to-types": {
        "description": "JSON 转类型定义的契约测试数据。",
        "cases": [
            {
                "name": "生成 TypeScript 类型",
                "request": {
                    "input": '{"id":1,"name":"a","tags":["x"]}',
                    "language": "typescript",
                    "rootName": "User",
                },
                "expect": {"code": 0, "dataContains": {"language": "typescript"}},
            },
            {
                "name": "生成 Go 类型",
                "request": {"input": '{"id":1}', "language": "go", "rootName": "Item"},
                "expect": {"code": 0, "dataContains": {"language": "go"}},
            },
            {
                "name": "不支持的语言",
                "request": {"input": "{}", "language": "cobol"},
                "expect": {"code": 400, "errorCode": "INVALID_FIELD"},
            },
            {
                "name": "缺少输入",
                "request": {"language": "go"},
                "expect": {"code": 400, "errorCode": "MISSING_FIELD"},
            },
        ],
    },
    "mock-data": {
        "description": "模拟数据生成的契约测试数据。",
        "cases": [
            {
                "name": "生成字符串与整数",
                "request": {"types": ["string", "int"], "count": 3},
                "expect": {"code": 0, "dataContains": {"count": 3}},
            },
            {
                "name": "英文 locale",
                "request": {"types": ["email"], "count": 1, "locale": "en-US"},
                "expect": {"code": 0, "dataContains": {"locale": "en-US"}},
            },
            {
                "name": "不支持的 locale",
                "request": {"types": ["string"], "locale": "fr-FR"},
                "expect": {"code": 400, "errorCode": "INVALID_FIELD"},
            },
            {
                "name": "数量超限",
                "request": {"types": ["string"], "count": 100000},
                "expect": {"code": 400, "errorCode": "INVALID_FIELD"},
            },
        ],
    },
    "archive-inspect": {
        "description": "归档内容查看的契约测试数据。",
        "cases": [
            {
                "name": "查看 ZIP 条目",
                "request": {
                    "data": "UEsDBBQAAAAAANQCXl2GphA2BQAAAAUAAAAFAAAAYS50eHRoZWxsb1BLAQIUAxQAAAAAANQCXl2GphA2BQAAAAUAAAAFAAAAAAAAAAAAAACAAQAAAABhLnR4dFBLBQYAAAAAAQABADMAAAAoAAAAAAA=",
                    "format": "zip",
                },
                "expect": {"code": 0, "dataContains": {"format": "zip", "entryCount": 1}},
            },
            {
                "name": "缺少归档数据",
                "request": {"format": "zip"},
                "expect": {"code": 400, "errorCode": "MISSING_FIELD"},
            },
            {
                "name": "数据不是合法归档",
                "request": {"data": "bm90LWEtcmNoaXZl", "format": "zip"},
                "expect": {"code": 400, "errorCode": "INVALID_FIELD"},
            },
        ],
    },
    "file-analyze": {
        "description": "文件类型识别的契约测试数据。",
        "cases": [
            {
                "name": "识别 PNG",
                "request": {
                    "data": "iVBORw0KGgoAAAANSUhEUg==",
                    "fileName": "a.png",
                    "digestAlgorithm": "SHA256",
                },
                "expect": {"code": 0, "dataContains": {"extension": "png"}},
            },
            {
                "name": "缺少文件数据",
                "request": {"fileName": "a.txt"},
                "expect": {"code": 400, "errorCode": "MISSING_FIELD"},
            },
        ],
    },
    "wifi-qr": {
        "description": "访客 WiFi 二维码的契约测试数据。",
        "cases": [
            {
                "name": "生成 WPA 二维码",
                "request": {
                    "ssid": "Guest-Network",
                    "password": "guest-password",
                    "security": "WPA",
                    "size": 256,
                },
                "expect": {
                    "code": 0,
                    "dataContains": {"payload": "WIFI:T:WPA;S:Guest-Network;P:guest-password;"},
                },
            },
            {
                "name": "开放网络",
                "request": {"ssid": "Open", "password": "", "security": "nopass"},
                "expect": {"code": 0, "dataContains": {"payload": "WIFI:T:nopass;S:Open;P:;"}},
            },
            {
                "name": "缺少 SSID",
                "request": {"password": "x", "security": "WPA"},
                "expect": {"code": 400, "errorCode": "MISSING_FIELD"},
            },
        ],
    },
    "id-generator": {
        "description": "ID 生成的契约测试数据。",
        "cases": [
            {
                "name": "生成 UUID",
                "request": {"format": "uuid", "count": 3},
                "expect": {"code": 0, "dataContains": {"format": "uuid", "count": 3}},
            },
            {
                "name": "生成 ULID",
                "request": {"format": "ulid", "count": 1},
                "expect": {"code": 0, "dataContains": {"format": "ulid"}},
            },
            {
                "name": "生成 Nano ID",
                "request": {"format": "nanoid", "count": 1, "length": 10},
                "expect": {"code": 0, "dataContains": {"format": "nanoid", "length": 10}},
            },
            {
                "name": "不支持的格式",
                "request": {"format": "guid", "count": 1},
                "expect": {"code": 400, "errorCode": "INVALID_FIELD"},
            },
        ],
    },
    "lookup-table": {
        "description": "速查表查询的契约测试数据。",
        "cases": [
            {
                "name": "查询端口",
                "request": {"kind": "port", "key": "443"},
                "expect": {"code": 0, "dataContains": {"kind": "port", "matched": True}},
            },
            {
                "name": "模糊检索退出码",
                "request": {"kind": "exitcode", "key": "not"},
                "expect": {"code": 0, "dataContains": {"kind": "exitcode"}},
            },
            {
                "name": "缺少查询词",
                "request": {"kind": "port"},
                "expect": {"code": 400, "errorCode": "MISSING_FIELD"},
            },
        ],
    },
    "nginx-config": {
        "description": "Nginx 配置生成的契约测试数据。",
        "cases": [
            {
                "name": "生成站点配置",
                "request": {
                    "mode": "site",
                    "serverName": "example.com",
                    "listenPort": 80,
                    "root": "/var/www/html",
                },
                "expect": {"code": 0, "dataContains": {"mode": "site"}},
            },
            {
                "name": "生成全局配置",
                "request": {"mode": "global", "workerProcesses": 4},
                "expect": {"code": 0, "dataContains": {"mode": "global"}},
            },
            {
                "name": "不支持的模式",
                "request": {"mode": "stream"},
                "expect": {"code": 400, "errorCode": "INVALID_FIELD"},
            },
        ],
    },
    "firewall-rules": {
        "description": "防火墙规则生成的契约测试数据。",
        "cases": [
            {
                "name": "生成 ufw 规则",
                "request": {"tool": "ufw", "scenario": "https", "protocol": "tcp"},
                "expect": {"code": 0, "dataContains": {"tool": "ufw", "scenario": "https"}},
            },
            {
                "name": "生成 iptables 规则",
                "request": {"tool": "iptables", "scenario": "ssh", "protocol": "tcp"},
                "expect": {"code": 0, "dataContains": {"tool": "iptables"}},
            },
            {
                "name": "不支持的工具",
                "request": {"tool": "pf", "scenario": "http"},
                "expect": {"code": 400, "errorCode": "INVALID_FIELD"},
            },
        ],
    },
    "docker-compose-gen": {
        "description": "Compose 配置生成的契约测试数据。",
        "cases": [
            {
                "name": "生成单服务配置",
                "request": {
                    "projectName": "demo",
                    "services": [
                        {"name": "web", "image": "nginx:alpine", "ports": ["8080:80"]}
                    ],
                },
                "expect": {"code": 0, "dataContains": {"serviceCount": 1}},
            },
            {
                "name": "缺少项目名",
                "request": {"services": []},
                "expect": {"code": 400, "errorCode": "MISSING_FIELD"},
            },
        ],
    },
    "ip-range-expand": {
        "description": "IP 段展开的契约测试数据。",
        "cases": [
            {
                "name": "展开网段",
                "request": {"mode": "expand", "cidr": "192.168.1.0/30", "limit": 16},
                "expect": {"code": 0, "dataContains": {"mode": "expand", "count": 4}},
            },
            {
                "name": "地址转 CIDR",
                "request": {"mode": "toCidr", "start": "192.168.1.10", "end": "192.168.1.20"},
                "expect": {"code": 0, "dataContains": {"mode": "toCidr"}},
            },
            {
                "name": "网段转地址区间",
                "request": {"mode": "toRange", "cidr": "192.168.1.0/24"},
                "expect": {"code": 0, "dataContains": {"mode": "toRange"}},
            },
            {
                "name": "非法网段",
                "request": {"mode": "expand", "cidr": "not-a-cidr"},
                "expect": {"code": 400, "errorCode": "INVALID_FIELD"},
            },
        ],
    },
    "device-headers": {
        "description": "设备请求头的契约测试数据。",
        "cases": [
            {
                "name": "生成 Chrome 请求头",
                "request": {"mode": "generate", "platform": "chrome", "version": "120.0.0.0"},
                "expect": {"code": 0, "dataContains": {"mode": "generate"}},
            },
            {
                "name": "解析 User-Agent",
                "request": {
                    "mode": "parseUa",
                    "userAgent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) "
                    "AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
                },
                "expect": {"code": 0, "dataContains": {"mode": "parseUa"}},
            },
            {
                "name": "不支持的模式",
                "request": {"mode": "rewrite"},
                "expect": {"code": 400, "errorCode": "INVALID_FIELD"},
            },
        ],
    },
    "color-convert": {
        "description": "颜色格式转换的契约测试数据。",
        "cases": [
            {
                "name": "十六进制转多格式",
                "request": {
                    "input": "#3B82F6",
                    "inputFormat": "hex",
                    "outputFormats": ["hex", "rgb", "hsl", "oklch"],
                },
                "expect": {"code": 0, "dataContains": {"hex": "#3B82F6"}},
            },
            {
                "name": "RGB 字符串输入",
                "request": {"input": "rgb(59, 130, 246)", "inputFormat": "rgb"},
                "expect": {"code": 0, "dataContains": {"inputFormat": "rgb"}},
            },
            {
                "name": "非法颜色",
                "request": {"input": "not-a-color", "inputFormat": "hex"},
                "expect": {"code": 400, "errorCode": "INVALID_FIELD"},
            },
        ],
    },
    "html-escape": {
        "description": "HTML 转义的契约测试数据。",
        "cases": [
            {
                "name": "转义尖括号",
                "request": {"mode": "encode", "input": "<b>hi</b>"},
                "expect": {"code": 0, "dataContains": {"output": "&lt;b&gt;hi&lt;/b&gt;"}},
            },
            {
                "name": "反转义实体",
                "request": {"mode": "decode", "input": "&lt;b&gt;hi&lt;/b&gt;"},
                "expect": {"code": 0, "dataContains": {"output": "<b>hi</b>"}},
            },
            {
                "name": "不支持的模式",
                "request": {"mode": "strip", "input": "a"},
                "expect": {"code": 400, "errorCode": "INVALID_FIELD"},
            },
        ],
    },
    "markdown-convert": {
        "description": "Markdown 转换的契约测试数据。",
        "cases": [
            {
                "name": "Markdown 转 HTML",
                "request": {"mode": "toHtml", "input": "# Title\n\nBody", "gfm": True},
                "expect": {"code": 0, "dataContains": {"mode": "toHtml", "headingCount": 1}},
            },
            {
                "name": "HTML 转 Markdown",
                "request": {"mode": "toMarkdown", "input": "<h1>Title</h1><p>Body</p>"},
                "expect": {"code": 0, "dataContains": {"mode": "toMarkdown"}},
            },
            {
                "name": "不支持的模式",
                "request": {"mode": "toPdf", "input": "# a"},
                "expect": {"code": 400, "errorCode": "INVALID_FIELD"},
            },
        ],
    },
    "password-policy": {
        "description": "口令策略检查的契约测试数据。",
        "cases": [
            {
                "name": "满足中等强度",
                "request": {"password": "Zx7#quietRiver9", "preset": "medium"},
                "expect": {"code": 0, "dataContains": {"compliant": True}},
            },
            {
                "name": "不满足严格强度",
                "request": {"password": "abc", "preset": "strict"},
                "expect": {"code": 0, "dataContains": {"compliant": False}},
            },
            {
                "name": "不支持的预设",
                "request": {"password": "abc", "preset": "extreme"},
                "expect": {"code": 400, "errorCode": "INVALID_FIELD"},
            },
        ],
    },
    "totp": {
        "description": "TOTP 的契约测试数据。",
        "cases": [
            {
                "name": "RFC 6238 标准向量",
                "request": {
                    "operation": "generate",
                    "secret": "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ",
                    "algorithm": "SHA1",
                    "digits": 8,
                    "period": 30,
                    "timestamp": 59,
                },
                "expect": {"code": 0, "dataContains": {"code": "94287082"}},
            },
            {
                "name": "验证正确验证码",
                "request": {
                    "operation": "verify",
                    "secret": "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ",
                    "algorithm": "SHA1",
                    "digits": 8,
                    "timestamp": 59,
                    "code": "94287082",
                },
                "expect": {"code": 0, "dataContains": {"valid": True}},
            },
            {
                "name": "验证错误验证码",
                "request": {
                    "operation": "verify",
                    "secret": "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ",
                    "algorithm": "SHA1",
                    "digits": 8,
                    "timestamp": 59,
                    "code": "00000000",
                },
                "expect": {"code": 0, "dataContains": {"valid": False}},
            },
            {
                "name": "缺少密钥",
                "request": {"operation": "generate"},
                "expect": {"code": 400, "errorCode": "MISSING_FIELD"},
            },
        ],
    },
    "cert-lifecycle": {
        "description": "证书生命周期的契约测试数据。",
        "cases": [
            {
                "name": "缺少证书",
                "request": {"isPem": True, "warningDays": 30},
                "expect": {"code": 400, "errorCode": "MISSING_FIELD"},
            },
        ],
    },
    "ssh-keygen": {
        "description": "SSH 密钥生成的契约测试数据。",
        "cases": [
            {
                "name": "生成 Ed25519 密钥",
                "request": {"algorithm": "Ed25519", "comment": "toolbit@example.com"},
                "expect": {"code": 0, "dataContains": {"algorithm": "Ed25519"}},
            },
            {
                "name": "生成 RSA 密钥",
                "request": {"algorithm": "RSA", "bits": 2048, "comment": "toolbit@example.com"},
                "expect": {"code": 0, "dataContains": {"algorithm": "RSA", "bits": 2048}},
            },
            {
                "name": "不支持的算法",
                "request": {"algorithm": "DSA"},
                "expect": {"code": 400, "errorCode": "INVALID_FIELD"},
            },
        ],
    },
}


def main() -> None:
    OUT.mkdir(parents=True, exist_ok=True)
    for slug, contract in CONTRACTS.items():
        payload = {"slug": slug, **contract}
        path = OUT / f"{slug}.json"
        path.write_text(
            json.dumps(payload, ensure_ascii=False, indent=2) + "\n",
            encoding="utf-8",
        )
        print(f"写入 {path.relative_to(ROOT)}")


if __name__ == "__main__":
    main()
