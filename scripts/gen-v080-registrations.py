#!/usr/bin/env python3
"""把 v0.8.0 批次新增工具登记进各领域包的 register.go 与工具目录。

脚本**幂等**：已登记的 slug 会被跳过，重复执行不会产生重复条目。
"""
from __future__ import annotations

import pathlib
import re

ROOT = pathlib.Path(__file__).resolve().parents[1]
BACKEND = ROOT / "backend"
CATALOG = BACKEND / "catalog" / "tools.yaml"

# package → [(slug, 请求类型, 处理函数)]
REGISTRATIONS: dict[str, list[tuple[str, str, str]]] = {
    "text": [
        ("text-diff", "TextDiffRequest", "TextDiff"),
        ("regex-validate", "RegexValidateRequest", "RegexValidate"),
        ("regex-build", "RegexBuildRequest", "RegexBuild"),
        ("data-mask", "DataMaskRequest", "DataMask"),
        ("text-stats", "TextStatsRequest", "TextStats"),
    ],
    "data": [
        ("id-validator", "IDValidateRequest", "IDValidate"),
        ("csv-convert", "CSVConvertRequest", "CSVConvert"),
        ("json-to-types", "JSONToTypesRequest", "JSONToTypes"),
        ("mock-data", "MockDataRequest", "MockData"),
        ("archive-inspect", "ArchiveInspectRequest", "ArchiveInspect"),
        ("file-analyze", "FileAnalyzeRequest", "FileAnalyze"),
        ("wifi-qr", "WiFiQRRequest", "WiFiQR"),
        ("id-generator", "IDGeneratorRequest", "IDGenerator"),
    ],
    "system": [
        ("nginx-config", "NginxConfigRequest", "NginxConfig"),
        ("firewall-rules", "FirewallRulesRequest", "FirewallRules"),
        ("docker-compose-gen", "DockerComposeRequest", "DockerComposeGen"),
        ("lookup-table", "LookupTableRequest", "LookupTable"),
    ],
    "network": [
        ("device-headers", "DeviceHeadersRequest", "DeviceHeaders"),
        ("ip-range-expand", "IPRangeRequest", "IPRange"),
        ("mime-lookup", "MimeLookupRequest", "MimeLookup"),
        ("http-status", "HttpStatusRequest", "HttpStatus"),
    ],
    "formatting": [
        ("color-convert", "ColorConvertRequest", "ColorConvert"),
    ],
    "encoding": [
        ("html-escape", "HTMLEscapeRequest", "HTMLEscape"),
    ],
    "documents": [
        ("markdown-convert", "MarkdownConvertRequest", "MarkdownConvert"),
    ],
    "crypto": [
        ("password-policy", "PasswordPolicyRequest", "PasswordPolicy"),
        ("totp", "TOTPRequest", "TOTP"),
    ],
    "pki": [
        ("cert-lifecycle", "CertLifecycleRequest", "CertLifecycle"),
    ],
    "keys": [
        ("ssh-keygen", "SSHKeygenRequest", "SSHKeygen"),
    ],
}

# slug → (分类, 顺序, 中文名, 英文名)
CATALOG_ENTRIES: dict[str, tuple[str, int, str, str]] = {
    "mime-lookup": ("network", 6, "MIME 类型查询", "MIME Type Lookup"),
    "http-status": ("network", 7, "HTTP 状态码查询", "HTTP Status Code Lookup"),
    "text-diff": ("text", 3, "文本对比", "Text Diff"),
    "regex-validate": ("text", 4, "正则校验", "Regex Validator"),
    "regex-build": ("text", 5, "正则生成", "Regex Builder"),
    "data-mask": ("text", 6, "数据脱敏", "Data Masking"),
    "text-stats": ("text", 7, "文本统计", "Text Statistics"),
    "id-validator": ("data", 7, "号码校验", "ID Validator"),
    "csv-convert": ("data", 8, "CSV 转换", "CSV Converter"),
    "json-to-types": ("data", 9, "JSON 转类型", "JSON to Types"),
    "mock-data": ("data", 10, "模拟数据", "Mock Data"),
    "archive-inspect": ("data", 11, "归档查看", "Archive Inspector"),
    "file-analyze": ("data", 12, "文件识别", "File Analyzer"),
    "wifi-qr": ("data", 13, "访客 WiFi 二维码", "Guest WiFi QR Code"),
    "id-generator": ("data", 14, "ID 生成", "ID Generator"),
    "nginx-config": ("system", 4, "Nginx 配置生成", "Nginx Config Generator"),
    "firewall-rules": ("system", 5, "防火墙规则生成", "Firewall Rule Generator"),
    "docker-compose-gen": ("system", 6, "Compose 配置生成", "Docker Compose Generator"),
    "lookup-table": ("system", 7, "速查表", "Lookup Table"),
    "device-headers": ("network", 8, "设备请求头", "Device Headers"),
    "ip-range-expand": ("network", 9, "IP 段展开", "IP Range Expander"),
    "color-convert": ("formatting", 9, "颜色转换", "Color Converter"),
    "html-escape": ("encoding", 7, "HTML 转义", "HTML Escape"),
    "markdown-convert": ("documents", 3, "Markdown 转换", "Markdown Converter"),
    "password-policy": ("crypto", 8, "口令策略检查", "Password Policy Checker"),
    "totp": ("crypto", 9, "TOTP 生成校验", "TOTP Generator"),
    "cert-lifecycle": ("certificates", 23, "证书生命周期", "Certificate Lifecycle"),
    "ssh-keygen": ("keys", 6, "SSH 密钥生成", "SSH Key Generator"),
}

ENTRY_TEMPLATE = """\t\t{{
\t\t\tSlug:       "{slug}",
\t\t\tNewRequest: func() tools.Request {{
\t\t\t\treturn &{req}{{}}
\t\t\t}},
\t\t\tHandle: func(ctx context.Context, req tools.Request) (tools.Result, error) {{
\t\t\t\treturn {handler}(ctx, req.(*{req}))
\t\t\t}},
\t\t}},
"""


def register_handlers() -> None:
    """把新增 slug 插到各领域包 register.go 的条目列表末尾。"""
    for package, entries in REGISTRATIONS.items():
        path = BACKEND / "internal" / "tools" / package / "register.go"
        text = path.read_text(encoding="utf-8")
        existing = set(re.findall(r'Slug:\s+"([^"]+)"', text))
        added = [
            ENTRY_TEMPLATE.format(slug=slug, req=req, handler=handler)
            for slug, req, handler in entries
            if slug not in existing
        ]
        if not added:
            print(f"跳过 {package}/register.go（已登记）")
            continue
        # 条目列表以最后一个 "}," 结束；插到它之后。
        marker = "\t\t},\n"
        index = text.rfind(marker)
        if index < 0:
            raise SystemExit(f"未找到条目列表结尾：{path}")
        text = text[: index + len(marker)] + "".join(added) + text[index + len(marker) :]
        path.write_text(text, encoding="utf-8")
        print(f"已更新 {path.relative_to(ROOT)}（新增 {len(added)} 条）")


def register_catalog() -> None:
    """把新增 slug 追加到工具目录。"""
    text = CATALOG.read_text(encoding="utf-8")
    existing = set(re.findall(r"^\s*- slug: ([a-z0-9-]+)\s*$", text, re.MULTILINE))
    added = []
    for slug, (category, order, name, name_en) in CATALOG_ENTRIES.items():
        if slug in existing:
            continue
        added.append(
            f"""  - slug: {slug}
    category: {category}
    order: {order}
    name: {name}
    nameEn: {name_en}
    enabled: true
    transport: json
    rateClass: light
    maxBodyBytes: 1048576
    timeoutMs: 5000
    sensitive: false
    requiresNetwork: false
    autoRun: true
    wide: false
    i18nKey: tools.{slug}
"""
        )
    if not added:
        print("跳过 tools.yaml（已登记）")
        return
    text = text.rstrip("\n") + "\n" + "\n".join(added)
    CATALOG.write_text(text, encoding="utf-8")
    print(f"已更新 {CATALOG.relative_to(ROOT)}（新增 {len(added)} 条）")


def main() -> None:
    register_handlers()
    register_catalog()


if __name__ == "__main__":
    main()
