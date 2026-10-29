#!/usr/bin/env python3
"""由词条数据生成前端工具模块与双语词条。

产出 index.ts 与 zh-CN / en-US 词条 JSON；Tool.vue 由人工编写。
"""
from __future__ import annotations

import json
import pathlib

ROOT = pathlib.Path(__file__).resolve().parents[1]
TOOLS = ROOT / "frontend" / "src" / "tools"
LOCALES = ROOT / "frontend" / "src" / "locales"
DATA = ROOT / "scripts" / "v080-locale-data.json"

INDEX_TEMPLATE = '''/**
 * {slug} 工具模块。
 *
 * 默认导出工具组件；`messages` 是该工具的双语词条，
 * 由路由容器在组件加载后登记，避免进入初始包。
 */
import type {{ Component }} from 'vue'

import enUS from '@/locales/en-US/tools/{slug}.json'
import zhCN from '@/locales/zh-CN/tools/{slug}.json'

import Tool from './Tool.vue'

/** 工具词条。 */
export const messages: Record<'zh-CN' | 'en-US', Record<string, unknown>> = {{
  'zh-CN': zhCN,
  'en-US': enUS,
}}

export default Tool as Component
'''

# slug → (类型名, [(字段名, 示例值)])
EXAMPLES: dict[str, tuple[str, list[tuple[str, str]]]] = {
    "text-diff": ("TextDiff", [("mode", "text"), ("left", "line A\\nline B"), ("right", "line A\\nline C")]),
    "regex-validate": ("RegexValidate", [("pattern", "^[a-z]+$")]),
    "regex-build": ("RegexBuild", [("prefix", "user-")]),
    "data-mask": ("DataMask", [("text", "手机号 13800138000，邮箱 a@example.com")]),
    "text-stats": ("TextStats", [("text", "Hello 世界\\n第二行")]),
    "id-validator": ("IDValidate", [("type", "idCard"), ("value", "110101199003072516")]),
    "csv-convert": ("CSVConvert", [("input", "name,age\\nAlice,30\\nBob,25")]),
    "json-to-types": ("JSONToTypes", [("input", '{\\n  "id": 1,\\n  "name": "Alice"\\n}')]),
    "mock-data": ("MockData", [("types", "email")]),
    "archive-inspect": ("ArchiveInspect", [("data", "")]),
    "file-analyze": ("FileAnalyze", [("data", "")]),
    "wifi-qr": ("WiFiQR", [("ssid", "Guest-Network"), ("password", "guest-password")]),
    "id-generator": ("IDGenerator", [("format", "uuid")]),
    "nginx-config": ("NginxConfig", [("serverName", "example.com"), ("root", "/var/www/html")]),
    "firewall-rules": ("FirewallRules", [("scenario", "https")]),
    "docker-compose-gen": ("DockerCompose", [("projectName", "demo")]),
    "lookup-table": ("LookupTable", [("kind", "port"), ("key", "443")]),
    "device-headers": ("DeviceHeaders", [("platform", "chrome")]),
    "ip-range-expand": ("IPRange", [("cidr", "192.168.1.0/24")]),
    "color-convert": ("ColorConvert", [("input", "#3B82F6")]),
    "html-escape": ("HTMLEscape", [("input", '<b>hello</b> & "world"')]),
    "markdown-convert": ("MarkdownConvert", [("input", "# Title\\n\\nSome **bold** text.")]),
    "password-policy": ("PasswordPolicy", [("password", "Example123")]),
    "totp": ("TOTP", [("secret", "JBSWY3DPEHPK3PXP")]),
    "cert-lifecycle": ("CertLifecycle", [("certificate", "")]),
    "ssh-keygen": ("SSHKeygen", [("algorithm", "Ed25519")]),
}


def example_template(slug: str, type_name: str, fields: list[tuple[str, str]]) -> str:
    lines = []
    for name, _ in fields:
        lines.append(f"  /** {name}。 */\n  {name}: string")
    body = "\n".join(lines)
    values = "\n".join(f"  {name}: '{value}'," for name, value in fields)

    return f"""/**
 * {slug} 工具的示例数据。
 *
 * 只用于演示，**不含**任何真实凭据。
 */

/** 一个示例。 */
export interface {type_name}Example {{
  /** 示例名称。 */
  label: string
{body}
}}

/** 默认示例。 */
export const DEFAULT_EXAMPLE: {type_name}Example = {{
  label: '{slug} 示例',
{values}
}}

/** 全部示例。 */
export const EXAMPLES: readonly {type_name}Example[] = [DEFAULT_EXAMPLE]
"""


def build_locale(tool: dict, index: int) -> dict:
    """按语言索引构造词条。index 0 = zh-CN，1 = en-US。"""

    def pick(value):
        if isinstance(value, list):
            return value[index]
        return value

    def pick_map(mapping: dict) -> dict:
        return {key: pick(value) for key, value in mapping.items()}

    return {
        "name": pick(tool["name"]),
        "description": pick(tool["description"]),
        "category": pick(tool["categoryLabel"]),
        "actions": pick_map(tool.get("actions", {})),
        "fields": pick_map(tool.get("fields", {})),
        "options": {
            key: pick_map(value) if isinstance(value, dict) else pick(value)
            for key, value in tool.get("options", {}).items()
        },
        "results": pick_map(tool.get("results", {})),
        "examples": [
            {"label": pick(item["label"]), "value": item["value"]}
            for item in tool.get("examples", [])
        ],
        "hints": pick_map(tool.get("hints", {})),
        "faq": [{"q": pick(item["q"]), "a": pick(item["a"])} for item in tool.get("faq", [])],
    }


def main() -> None:
    data = json.loads(DATA.read_text(encoding="utf-8"))

    for slug, tool in data.items():
        category = tool["category"]
        directory = TOOLS / category / slug
        directory.mkdir(parents=True, exist_ok=True)

        (directory / "index.ts").write_text(
            INDEX_TEMPLATE.format(slug=slug), encoding="utf-8"
        )

        type_name, fields = EXAMPLES[slug]
        (directory / "example.ts").write_text(
            example_template(slug, type_name, fields), encoding="utf-8"
        )

        for index, lang in enumerate(("zh-CN", "en-US")):
            target = LOCALES / lang / "tools" / f"{slug}.json"
            target.parent.mkdir(parents=True, exist_ok=True)
            target.write_text(
                json.dumps(build_locale(tool, index), ensure_ascii=False, indent=2) + "\n",
                encoding="utf-8",
            )

        print(f"生成 {category}/{slug}")


if __name__ == "__main__":
    main()
