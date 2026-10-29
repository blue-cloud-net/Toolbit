#!/usr/bin/env python3
"""生成 v0.6.0 前端示例夹具（写侧工具需要私钥材料）。

用法:
    bash scripts/gen-v060-fixtures.sh <夹具目录>
    TOOLBIT_V060_FIXTURES=<夹具目录> go test ./internal/pkisvc -run TestGenerateToolbitP12
    python3 scripts/gen-v060-web-fixtures.py <夹具目录>

⚠️ 夹具里的私钥是**一次性生成的测试密钥**，只用于填示例，
**不是**任何真实系统的密钥（见 AGENTS.md §5）。
"""

import base64
import sys
from pathlib import Path

OUT = Path(__file__).resolve().parent.parent / "frontend" / "src" / "tools" / \
    "certificates" / "fixtures-v060.ts"


def read(path: Path, name: str) -> str:
    return (path / name).read_text()


def read_b64(path: Path, name: str) -> str:
    return base64.b64encode((path / name).read_bytes()).decode()


def main() -> int:
    if len(sys.argv) < 2:
        print(__doc__, file=sys.stderr)
        return 2
    fixtures = Path(sys.argv[1])

    body = f'''/**
 * v0.6.0 证书写侧工具的**测试夹具**。
 *
 * 由 `scripts/gen-v060-web-fixtures.py` 生成，请勿手工编辑。
 *
 * ⚠️ 这里的私钥是 `openssl` **一次性生成**的测试密钥，只用于填充示例，
 * **不是**任何真实系统的密钥或证书（见 AGENTS.md §5）。
 *
 * 重新生成：
 *   bash scripts/gen-v060-fixtures.sh <dir>
 *   TOOLBIT_V060_FIXTURES=<dir> go test ./internal/pkisvc -run TestGenerateToolbitP12
 *   python3 scripts/gen-v060-web-fixtures.py <dir>
 */

/** 测试用 CA 证书。 */
export const TEST_CA_CERT_PEM = `{read(fixtures, "ca2.crt.pem")}`

/** 测试用 CA 私钥（一次性生成，非真实密钥）。 */
export const TEST_CA_KEY_PEM = `{read(fixtures, "ca2.key.pem")}`

/** 测试用叶证书。 */
export const TEST_LEAF_CERT_PEM = `{read(fixtures, "leaf2.crt.pem")}`

/** 测试用叶私钥（一次性生成，非真实密钥）。 */
export const TEST_LEAF_KEY_PEM = `{read(fixtures, "leaf2.key.pem")}`

/** 测试用吊销列表（已吊销叶证书）。 */
export const TEST_CRL_PEM = `{read(fixtures, "crl2.pem")}`

/** 测试用 PKCS#12 容器（口令 `test-pass`，base64）。 */
export const TEST_P12_BASE64 =
  '{read_b64(fixtures, "bundle-toolbit.p12")}'

/** 测试用 PKCS#7 证书袋（base64）。 */
export const TEST_P7B_BASE64 = '{read_b64(fixtures, "bundle.p7b")}'
'''

    OUT.write_text(body)
    print(f"written {OUT}")
    return 0


if __name__ == "__main__":
    sys.exit(main())
