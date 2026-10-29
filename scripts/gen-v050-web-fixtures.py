#!/usr/bin/env python3
"""生成前端证书夹具文件。

材料来自 backend/testdata/fixtures/pki/（由 scripts/gen-pki-fixtures.sh 生成）。
本脚本只做「把夹具内容嵌进 TypeScript」，不生成任何密码学材料。
"""

import pathlib

ROOT = pathlib.Path(__file__).resolve().parents[1]
FIX = ROOT / "backend" / "testdata" / "fixtures" / "pki"
OUT = ROOT / "frontend" / "src" / "tools" / "certificates" / "fixtures.ts"

HEADER = """/**
 * 证书类工具的**测试夹具**。
 *
 * 这些材料由 `openssl` **一次性生成**（见 scripts/gen-pki-fixtures.sh），
 * 只用于本项目的示例与单测，**不是**任何真实系统的证书或密钥。
 *
 * 私钥**不在**此处：本目录只有公开材料。
 *
 * 重新生成：bash scripts/gen-pki-fixtures.sh && python3 scripts/gen-v050-web-fixtures.py
 */

/** 叶证书（由中间 CA 签发，SAN 含 DNS 与 IP）。 */
export const LEAF_CERT_PEM = `%s`

/** 中间 CA 证书（由根 CA 签发）。 */
export const INTERMEDIATE_CERT_PEM = `%s`

/** 自签根 CA 证书。 */
export const ROOT_CERT_PEM = `%s`

/** 完整证书链（叶 → 中间 → 根）的 PEM bundle。 */
export const CHAIN_PEM = `%s`

/** 叶证书的 CSR（PKCS#10）。 */
export const CSR_PEM = `%s`

/** 证书吊销列表（由根 CA 签发，吊销了叶证书）。 */
export const CRL_PEM = `%s`
"""


def read(name: str) -> str:
    return (FIX / name).read_text(encoding="utf-8")


def main() -> None:
    content = HEADER % (
        read("leaf.crt.pem"),
        read("inter.crt.pem"),
        read("ca.crt.pem"),
        read("chain.pem"),
        read("leaf.csr.pem"),
        read("crl.pem"),
    )
    OUT.parent.mkdir(parents=True, exist_ok=True)
    OUT.write_text(content, encoding="utf-8")
    print(f"written {OUT.relative_to(ROOT)}")


if __name__ == "__main__":
    main()
