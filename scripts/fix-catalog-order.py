#!/usr/bin/env python3
"""重排工具目录的 order 字段，消除同分类内的重复编号。

新工具按其在 `tools:` 列表中出现的先后顺序追加到所属分类末尾，
随后对每个分类按 1..N 重新编号。
"""
from __future__ import annotations

import pathlib
import re

import yaml

ROOT = pathlib.Path(__file__).resolve().parents[1]
CATALOG = ROOT / "backend" / "catalog" / "tools.yaml"

# slug → 分类归属修正（新工具归入的分类）
CATEGORY_FIX: dict[str, str] = {
    "mime-lookup": "network",
    "http-status": "network",
}

# 新工具在所属分类内的相对顺序（按此顺序追加到分类末尾）。
NEW_ORDER: list[str] = [
    "text-diff",
    "regex-validate",
    "regex-build",
    "data-mask",
    "text-stats",
    "id-validator",
    "csv-convert",
    "json-to-types",
    "mock-data",
    "archive-inspect",
    "file-analyze",
    "wifi-qr",
    "id-generator",
    "nginx-config",
    "firewall-rules",
    "docker-compose-gen",
    "lookup-table",
    "device-headers",
    "ip-range-expand",
    "color-convert",
    "html-escape",
    "markdown-convert",
    "password-policy",
    "totp",
    "cert-lifecycle",
    "ssh-keygen",
]


def main() -> None:
    text = CATALOG.read_text(encoding="utf-8")
    document = yaml.safe_load(text)
    tools: list[dict] = document["tools"]

    by_slug = {tool["slug"]: tool for tool in tools}

    # 1) 修正新工具的分类归属。
    for slug, category in CATEGORY_FIX.items():
        if slug in by_slug:
            by_slug[slug]["category"] = category

    # 2) 收集每个分类下的工具，按「既有在前、新增在后」排列。
    #    既有工具保持目录中的相对顺序；新工具按 NEW_ORDER 追加。
    categories = [item["slug"] for item in document["categories"]]
    buckets: dict[str, list[dict]] = {category: [] for category in categories}
    for tool in tools:
        buckets.setdefault(tool["category"], []).append(tool)

    new_slugs = set(NEW_ORDER)
    for category, bucket in buckets.items():
        existing = [tool for tool in bucket if tool["slug"] not in new_slugs]
        added = [tool for tool in bucket if tool["slug"] in new_slugs]
        added.sort(key=lambda tool: NEW_ORDER.index(tool["slug"]))
        buckets[category] = existing + added

    # 3) 重排 order 并重建列表（按分类声明顺序、分类内 order 升序）。
    reordered: list[dict] = []
    for category in categories:
        for index, tool in enumerate(buckets.get(category, []), start=1):
            tool["order"] = index
            reordered.append(tool)

    document["tools"] = reordered

    # 4) 写回 YAML：保留原始注释风格（手写序列化过于脆弱，改用行级重写）。
    header = text.split("tools:\n", 1)[0]
    lines: list[str] = []
    for tool in reordered:
        lines.append(f"  - slug: {tool['slug']}")
        lines.append(f"    category: {tool['category']}")
        lines.append(f"    order: {tool['order']}")
        lines.append(f"    name: {tool['name']}")
        lines.append(f"    nameEn: {tool['nameEn']}")
        lines.append(f"    enabled: {str(tool.get('enabled', True)).lower()}")
        lines.append(f"    transport: {tool.get('transport', 'json')}")
        lines.append(f"    rateClass: {tool.get('rateClass', 'light')}")
        lines.append(f"    maxBodyBytes: {tool.get('maxBodyBytes', 1048576)}")
        lines.append(f"    timeoutMs: {tool.get('timeoutMs', 5000)}")
        lines.append(f"    sensitive: {str(tool.get('sensitive', False)).lower()}")
        lines.append(f"    requiresNetwork: {str(tool.get('requiresNetwork', False)).lower()}")
        lines.append(f"    autoRun: {str(tool.get('autoRun', True)).lower()}")
        lines.append(f"    wide: {str(tool.get('wide', False)).lower()}")
        lines.append(f"    i18nKey: {tool['i18nKey']}")
        lines.append("")

    CATALOG.write_text(header + "tools:\n" + "\n".join(lines).rstrip("\n") + "\n", encoding="utf-8")
    print(f"已重排 {CATALOG.relative_to(ROOT)}，共 {len(reordered)} 个工具")


if __name__ == "__main__":
    main()
