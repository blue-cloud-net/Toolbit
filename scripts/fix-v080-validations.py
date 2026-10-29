#!/usr/bin/env python3
"""补上 v0.8.0 新工具里遗漏的字段级校验。"""
from __future__ import annotations

import pathlib

ROOT = pathlib.Path(__file__).resolve().parents[1]
TOOLS = ROOT / "backend" / "internal" / "tools"

EDITS: list[tuple[str, str, str]] = [
    (
        "data/fileanalyze.go",
        '\t\tif len(r.FileName) > 255 {\n'
        '\t\t\treturn tools.InvalidField("fileName", []int{0, 255}, len(r.FileName))\n'
        '\t\t}\n',
        '\t\tif len(r.FileName) > 255 {\n'
        '\t\t\treturn tools.InvalidField("fileName", []int{0, 255}, len(r.FileName))\n'
        '\t\t}\n'
        '\t\tif len(r.Data) > maxFileAnalyzeBytes {\n'
        '\t\t\treturn tools.InvalidField("data", []int{0, maxFileAnalyzeBytes}, len(r.Data))\n'
        '\t\t}\n',
    ),
    (
        "documents/markdownconvert.go",
        '\t\tif len(r.Input) > maxMarkdownBytes {\n'
        '\t\t\treturn tools.InvalidField("input", []int{0, maxMarkdownBytes}, len(r.Input))\n'
        '\t\t}\n',
        '\t\tif len(r.Input) > maxMarkdownBytes {\n'
        '\t\t\treturn tools.InvalidField("input", []int{0, maxMarkdownBytes}, len(r.Input))\n'
        '\t\t}\n'
        '\t\tif len(r.Input) > maxMarkdownConvertBytes {\n'
        '\t\t\treturn tools.InvalidField("input", []int{0, maxMarkdownConvertBytes}, len(r.Input))\n'
        '\t\t}\n',
    ),
    (
        "pki/certlifecycle.go",
        '\t\tif r.WarningDays < 0 || r.WarningDays > 3650 {\n'
        '\t\t\treturn tools.InvalidField("warningDays", []int{0, 3650}, r.WarningDays)\n'
        '\t\t}\n',
        '\t\tif r.WarningDays < 0 || r.WarningDays > 3650 {\n'
        '\t\t\treturn tools.InvalidField("warningDays", []int{0, 3650}, r.WarningDays)\n'
        '\t\t}\n'
        '\t\tif r.WarningDays == 0 {\n'
        '\t\t\tr.WarningDays = defaultCertWarningDays\n'
        '\t\t}\n',
    ),
]


def main() -> None:
    for relative, old, new in EDITS:
        path = TOOLS / relative
        text = path.read_text(encoding="utf-8")
        if old not in text:
            print(f"未匹配：{relative}")
            continue
        path.write_text(text.replace(old, new, 1), encoding="utf-8")
        print(f"已修改 {relative}")


if __name__ == "__main__":
    main()
