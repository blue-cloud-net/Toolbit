#!/usr/bin/env python3
"""把本批新增的工具登记到 backend/catalog/tools.yaml。"""
from __future__ import annotations

import pathlib
import re

ROOT = pathlib.Path(__file__).resolve().parents[1]
CATALOG = ROOT / "backend" / "catalog" / "tools.yaml"

# slug → (分类, 名称, 英文名, 是否敏感, 是否需要网络)
TOOLS: list[tuple[str, str, str, str, bool, bool]] = [
    # 文本
    ("text-diff", "text", "文本差异对比", "Text Diff", False, False),
    ("regex-validate", "text", "正则表达式校验", "Regex Validator", False, False),
    ("regex-build", "text", "正则表达式生成", "Regex Builder", False, False),
    ("data-mask", "text", "敏感数据脱敏", "Data Masker", True, False),
    ("text-stats", "text", "文本统计", "Text Statistics", False, False),
    # 数据
    ("id-validator", "data", "证件号码校验", "ID Validator", True, False),
    ("csv-convert", "data", "CSV 转换", "CSV Converter", False, False),
    ("json-to-types", "data", "JSON 转类型定义", "JSON to Types", False, False),
    ("mock-data", "data", "假数据生成", "Mock Data Generator", False, False),
    ("archive-inspect", "data", "归档内容查看", "Archive Inspector", False, False),
    ("file-analyze", "data", "文件类型识别", "File Analyzer", False, False),
    ("wifi-qr", "data", "访客 WiFi 二维码", "Guest WiFi QR Code", False, False),
    ("id-generator", "data", "ID 生成", "ID Generator", False, False),
    # 系统
    ("nginx-config", "system", "Nginx 配置生成", "Nginx Config Generator", False, False),
    ("firewall-rules", "system", "防火墙规则生成", "Firewall Rule Generator", False, False),
    ("docker-compose-gen", "system", "Compose 配置生成", "Docker Compose Generator", False, False),
    ("lookup-table", "system", "速查表查询", "Lookup Table", False, False),
    # 网络
    ("device-headers", "network", "请求头生成与 UA 解析", "Device Headers", False, False),
    ("ip-range-expand", "network", "IP 段展开", "IP Range Expander", False, False),
    # 格式化
    ("color-convert", "formatting", "颜色格式转换", "Color Converter", False, False),
    # 编码
    ("html-escape", "encoding", "HTML 实体编解码", "HTML Entity Codec", False, False),
    # 文档
    ("markdown-convert", "documents", "Markdown 转换", "Markdown Converter", False, False),
    # 加密
    ("password-policy", "crypto", "口令策略检查", "Password Policy Checker", True, False),
    ("totp", "crypto", "动态口令生成校验", "TOTP Generator", True, False),
    # 证书
    ("cert-lifecycle", "certificates", "证书生命周期", "Certificate Lifecycle", False, False),
    # 密钥
    ("ssh-keygen", "keys", "SSH 密钥生成", "SSH Key Generator", True, False),
]

# 各分类的下一个可用 order。
ORDER_BASE: dict[str, int] = {
    "text": 2,
    "data": 6,
    "system": 3,
    "network": 7,
    "formatting": 8,
    "encoding": 6,
    "documents": 3,
    "crypto": 8,
    "certificates": 23,
    "keys": 6,
}


def build_entry(
    slug: str,
    category: str,
    order: int,
    name: str,
    name_en: str,
    sensitive: bool,
    requires_network: bool,
) -> str:
    return (
        f"  - slug: {slug}\n"
        f"    category: {category}\n"
        f"    order: {order}\n"
        f"    name: {name}\n"
        f"    nameEn: {name_en}\n"
        "    enabled: true\n"
        "    transport: json\n"
        "    rateClass: light\n"
        "    maxBodyBytes: 1048576\n"
        "    timeoutMs: 5000\n"
        f"    sensitive: {'true' if sensitive else 'false'}\n"
        f"    requiresNetwork: {'true' if requires_network else 'false'}\n"
        "    autoRun: true\n"
        "    wide: false\n"
        f"    i18nKey: tools.{slug}\n"
    )


def main() -> None:
    text = CATALOG.read_text(encoding="utf-8")

    # tools 段在文件末尾：定位 "tools:" 之后的部分。
    tools_index = text.index("\ntools:\n")
    head = text[: tools_index + 1]
    body = text[tools_index + 1 :]

    counters = dict(ORDER_BASE)
    blocks = []
    for slug, category, name, name_en, sensitive, requires_network in TOOLS:
        if f"  - slug: {slug}\n" in text:
            print(f"跳过已存在的 slug：{slug}")
            continue
        order = counters.get(category, 1)
        counters[category] = order + 1
        blocks.append(
            build_entry(slug, category, order, name, name_en, sensitive, requires_network)
        )

    # 新条目追加到 tools 段末尾。
    CATALOG.write_text(head + body.rstrip("\n") + "\n\n" + "\n".join(blocks), encoding="utf-8")
    print(f"新增 {len(blocks)} 个工具条目")


if __name__ == "__main__":
    main()
