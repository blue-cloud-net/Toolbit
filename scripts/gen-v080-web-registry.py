#!/usr/bin/env python3
"""把 v0.8.0 批次新增工具注册进前端静态登记表与图标表。"""
from __future__ import annotations

import pathlib
import re

ROOT = pathlib.Path(__file__).resolve().parents[1]
TOOLS = ROOT / "frontend" / "src" / "tools"

# slug → (category, icon)
REGISTRATIONS: dict[str, tuple[str, str]] = {
    # 文本 / Text
    "text-diff": ("text", "swap"),
    "regex-validate": ("text", "shield-check"),
    "regex-build": ("text", "braces"),
    "data-mask": ("text", "asterisk"),
    "text-stats": ("text", "text"),
    # 数据 / Data
    "id-validator": ("data", "certificate"),
    "csv-convert": ("data", "data"),
    "json-to-types": ("data", "braces"),
    "mock-data": ("data", "list"),
    "id-generator": ("data", "fingerprint"),
    "archive-inspect": ("data", "archive"),
    "file-analyze": ("data", "file-search"),
    "wifi-qr": ("data", "qr"),
    # 系统 / System
    "lookup-table": ("system", "list"),
    "nginx-config": ("system", "terminal"),
    "firewall-rules": ("system", "shield"),
    "docker-compose-gen": ("system", "box"),
    # 网络 / Network
    "ip-range-expand": ("network", "sitemap"),
    "device-headers": ("network", "monitor"),
    # 格式化 / Formatting
    "color-convert": ("formatting", "asterisk"),
    # 编码 / Encoding
    "html-escape": ("encoding", "tag"),
    # 文档 / Documents
    "markdown-convert": ("documents", "markdown"),
    # 密码学 / Crypto
    "password-policy": ("crypto", "shield-check"),
    "totp": ("crypto", "clock"),
    # 证书 / Certificates
    "cert-lifecycle": ("certificates", "calendar"),
    # 密钥 / Keys
    "ssh-keygen": ("keys", "key-plus"),
}

# 分类 → 图标表里的注释标题
ICON_SECTION_TITLES = {
    "encoding": "编码 / Encoding",
    "formatting": "格式化 / Formatting",
    "crypto": "加密 / Cryptography",
    "keys": "密钥 / Keys",
    "certificates": "证书 / Certificates",
    "network": "网络 / Network",
    "text": "文本 / Text",
    "data": "数据 / Data",
    "system": "系统 / System",
    "documents": "文档 / Documents",
}


def register_loader() -> None:
    """把新增 slug 插到静态登记表对应分类最后一条之后。"""
    path = TOOLS / "index.ts"
    lines = path.read_text(encoding="utf-8").split("\n")

    by_category: dict[str, list[str]] = {}
    for slug, (category, _) in REGISTRATIONS.items():
        by_category.setdefault(category, []).append(slug)

    for category, slugs in by_category.items():
        pattern = re.compile(rf"import\('\./{category}/")
        anchors = [index for index, line in enumerate(lines) if pattern.search(line)]
        if not anchors:
            raise SystemExit(f"登记表缺少分类段落：{category}")
        insert_at = anchors[-1] + 1
        added = [
            f"  '{slug}': () => import('./{category}/{slug}/index'),"
            for slug in slugs
            if f"import('./{category}/{slug}/index')" not in "\n".join(lines)
        ]
        if not added:
            continue
        lines[insert_at:insert_at] = added

    path.write_text("\n".join(lines), encoding="utf-8")
    print(f"已更新 {path.relative_to(ROOT)}")


def register_icons() -> None:
    """把新增 slug 插到图标表对应分类段落末尾。"""
    path = TOOLS / "toolIcons.ts"
    lines = path.read_text(encoding="utf-8").split("\n")

    by_category: dict[str, list[str]] = {}
    for slug, (category, _) in REGISTRATIONS.items():
        by_category.setdefault(category, []).append(slug)

    # 段落标记 → 分类
    title_to_category = {title: category for category, title in ICON_SECTION_TITLES.items()}

    for category, slugs in by_category.items():
        title = ICON_SECTION_TITLES[category]
        marker = f"  // ── {title} ──"
        anchors = [index for index, line in enumerate(lines) if line.rstrip() == marker]
        if not anchors:
            raise SystemExit(f"图标表缺少分类段落：{title}")
        start = anchors[0]
        # 段落结束于下一个段落标记或对象字面量的收尾 `}`。
        end = len(lines)
        for index in range(start + 1, len(lines)):
            stripped = lines[index].strip()
            if stripped.startswith("// ── ") or stripped == "}":
                end = index
                break

        existing = {
            match.group(1)
            for line in lines[start:end]
            if (match := re.search(r"'([^']+)':", line))
        }
        added = [
            f"  '{slug}': '{REGISTRATIONS[slug][1]}',"
            for slug in slugs
            if slug not in existing
        ]
        if not added:
            continue
        # 插到段落内最后一条登记之后（段末空行之前）。
        insert_at = end
        while insert_at > start + 1 and lines[insert_at - 1].strip() == "":
            insert_at -= 1
        lines[insert_at:insert_at] = added

    path.write_text("\n".join(lines), encoding="utf-8")
    print(f"已更新 {path.relative_to(ROOT)}")


def main() -> None:
    register_loader()
    register_icons()


if __name__ == "__main__":
    main()
