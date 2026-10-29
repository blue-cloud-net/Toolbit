#!/usr/bin/env python3
"""生成 v0.6.0 的契约测试数据（12 个 PKI 生成 / 吊销 / 容器工具）。

用法:
    bash scripts/gen-v060-fixtures.sh <夹具目录>
    python3 scripts/gen-v060-contracts.py <夹具目录>

⚠️ 夹具由 gen-v060-fixtures.sh 一次性生成，只用于本仓库的契约测试，
**不是**任何真实系统的密钥或证书。私钥只内联在契约 JSON 里，
不单独落盘到仓库（与 v0.4.0 的密钥类契约数据同一约定）。
"""

import base64
import json
import sys
from pathlib import Path

OUT = Path(__file__).resolve().parent.parent / "backend" / "testdata" / "contract"


def read(fixtures: Path, name: str) -> str:
    return (fixtures / name).read_text()


def read_b64(fixtures: Path, name: str) -> str:
    return base64.b64encode((fixtures / name).read_bytes()).decode()


def write(slug: str, description: str, cases: list) -> None:
    path = OUT / f"{slug}.json"
    path.write_text(
        json.dumps(
            {"slug": slug, "description": description, "cases": cases},
            ensure_ascii=False,
            indent=2,
        )
        + "\n"
    )
    print(f"written {path.name}（{len(cases)} 个用例）")


def pem_bundle(data: str) -> dict:
    return {"data": data, "format": "pem"}


def main() -> int:
    if len(sys.argv) < 2:
        print(__doc__, file=sys.stderr)
        return 2
    fixtures = Path(sys.argv[1])

    ca_pem = read(fixtures, "ca2.crt.pem")
    ca_key = read(fixtures, "ca2.key.pem")
    leaf_pem = read(fixtures, "leaf2.crt.pem")
    leaf_key = read(fixtures, "leaf2.key.pem")
    crl_pem = read(fixtures, "crl2.pem")

    # ── csr-gen ──────────────────────────────────────────────────
    write(
        "csr-gen",
        "CSR 生成的契约测试数据。",
        [
            {
                "name": "生成服务端 CSR",
                "request": {
                    "certType": "server",
                    "subject": {
                        "C": "CN",
                        "O": "Toolbit Test",
                        "CN": "csr.example.com",
                    },
                    "key": {"type": "EC", "curve": "P-256"},
                    "sans": ["DNS:csr.example.com"],
                    "challengePassword": "test-only",
                },
                "expect": {
                    "code": 0,
                    "dataContains": {
                        "keyType": "EC",
                        "keyCurve": "P-256",
                        "info": {
                            "sans": ["DNS:csr.example.com"],
                            "challengePasswordPresent": True,
                            "signatureValid": True,
                        },
                    },
                },
            },
            {
                "name": "未知的生成类型",
                "request": {
                    "certType": "unknown",
                    "subject": {"CN": "csr.example.com"},
                },
                "expect": {"code": 400, "errorCode": "INVALID_FIELD"},
            },
        ],
    )

    # ── cert-gen ─────────────────────────────────────────────────
    write(
        "cert-gen",
        "证书签发的契约测试数据。",
        [
            {
                "name": "自签 CA 证书",
                "request": {
                    "signMode": "self",
                    "subject": {"C": "CN", "O": "Toolbit Test", "CN": "self.example.com"},
                    "key": {"type": "EC", "curve": "P-256"},
                    "notAfterDays": 30,
                    "basicConstraints": {"isCa": True, "pathLen": 1},
                },
                "expect": {
                    "code": 0,
                    "dataContains": {
                        "keyType": "EC",
                        "keyCurve": "P-256",
                        "subjectText": "/C=CN/O=Toolbit Test/CN=self.example.com",
                    },
                },
            },
            {
                "name": "由 CA 用 CSR 签发叶证书",
                "request": {
                    "signMode": "ca",
                    "issuer": {
                        "cert": pem_bundle(ca_pem),
                        "key": {"data": ca_key, "format": "pem"},
                    },
                    "csr": pem_bundle(read(fixtures, "leaf2.csr.pem")),
                    "notAfterDays": 30,
                    "sans": ["DNS:gen.example.com"],
                },
                "expect": {
                    "code": 0,
                    "dataContains": {
                        "subjectText": "/C=CN/O=Toolbit Test/CN=gen.example.com",
                    },
                },
            },
            {
                "name": "未知的签发模式",
                "request": {"signMode": "peer"},
                "expect": {"code": 400, "errorCode": "INVALID_FIELD"},
            },
            {
                "name": "CA 模式缺少签发者",
                "request": {
                    "signMode": "ca",
                    "subject": {"CN": "x.example.com"},
                },
                "expect": {"code": 400, "errorCode": "MISSING_FIELD"},
            },
        ],
    )

    # ── crl-gen ──────────────────────────────────────────────────
    write(
        "crl-gen",
        "吊销列表生成的契约测试数据。",
        [
            {
                "name": "吊销一张证书",
                "request": {
                    "issuerCert": pem_bundle(ca_pem),
                    "issuerKey": {"data": ca_key, "format": "pem"},
                    "revoked": [
                        {"cert": pem_bundle(leaf_pem), "reasonCode": 1}
                    ],
                },
                "expect": {
                    "code": 0,
                    "dataContains": {
                        "revokedCount": 1,
                        "issuerText": "/C=CN/O=Toolbit Test/CN=Toolbit Test Issuing CA",
                    },
                },
            },
            {
                "name": "保留原因码 7",
                "request": {
                    "issuerCert": pem_bundle(ca_pem),
                    "issuerKey": {"data": ca_key, "format": "pem"},
                    "revoked": [
                        {"cert": pem_bundle(leaf_pem), "reasonCode": 7}
                    ],
                },
                "expect": {"code": 400, "errorCode": "INVALID_FIELD"},
            },
        ],
    )

    # ── revocation-check ─────────────────────────────────────────
    write(
        "revocation-check",
        "吊销状态检查的契约测试数据。",
        [
            {
                "name": "带签发者验证后判定为已吊销",
                "request": {
                    "cert": pem_bundle(leaf_pem),
                    "crl": pem_bundle(crl_pem),
                    "issuer": pem_bundle(ca_pem),
                },
                "expect": {
                    "code": 0,
                    "dataContains": {
                        "status": "revoked",
                        "signatureVerified": True,
                        "reasonCode": 1,
                        "reasonText": "keyCompromise",
                    },
                },
            },
            {
                "name": "缺签发者时只报未验证",
                "request": {
                    "cert": pem_bundle(leaf_pem),
                    "crl": pem_bundle(crl_pem),
                },
                "expect": {
                    "code": 0,
                    "dataContains": {
                        "status": "revoked",
                        "signatureVerified": False,
                    },
                    "warnings": ["CRL_UNVERIFIED"],
                },
            },
            {
                "name": "非法吊销列表",
                "request": {
                    "cert": pem_bundle(leaf_pem),
                    "crl": {"data": "not a crl", "format": "pem"},
                },
                "expect": {"code": 400, "errorCode": "INVALID_CRL"},
            },
        ],
    )

    # ── ocsp-request ─────────────────────────────────────────────
    write(
        "ocsp-request",
        "OCSP 请求生成的契约测试数据。",
        [
            {
                "name": "用 SHA-256 生成请求",
                "request": {
                    "cert": pem_bundle(leaf_pem),
                    "issuer": pem_bundle(ca_pem),
                    "hash": "sha256",
                },
                "expect": {"code": 0, "dataContains": {"hash": "sha256"}},
            },
            {
                "name": "不支持的哈希算法",
                "request": {
                    "cert": pem_bundle(leaf_pem),
                    "issuer": pem_bundle(ca_pem),
                    "hash": "sha512",
                },
                "expect": {"code": 400, "errorCode": "UNSUPPORTED_ALGORITHM"},
            },
        ],
    )

    # ── ocsp-response ────────────────────────────────────────────
    ocsp_cases = [
        {
            "name": "非法响应字节",
            "request": {
                "response": {"data": "MAM=", "format": "base64"},
                "cert": pem_bundle(leaf_pem),
                "issuer": pem_bundle(ca_pem),
            },
            "expect": {"code": 400, "errorCode": "INVALID_DER"},
        }
    ]
    resp_file = fixtures / "ocsp-resp.der"
    if resp_file.exists() and resp_file.stat().st_size > 0:
        ocsp_cases.insert(
            0,
            {
                "name": "解析并用信任根验证响应",
                "request": {
                    "response": {
                        "data": read_b64(fixtures, "ocsp-resp.der"),
                        "format": "der",
                    },
                    "cert": pem_bundle(leaf_pem),
                    "issuer": pem_bundle(ca_pem),
                    "trustedRoots": [pem_bundle(ca_pem)],
                },
                "expect": {
                    "code": 0,
                    "dataContains": {
                        "status": 0,
                        "certStatus": "unknown",
                        "signatureVerified": True,
                    },
                },
            },
        )
    write("ocsp-response", "OCSP 响应解析的契约测试数据。", ocsp_cases)

    # ── pfx-pack ─────────────────────────────────────────────────
    write(
        "pfx-pack",
        "PKCS#12 打包的契约测试数据。",
        [
            {
                "name": "打包证书与私钥",
                "request": {
                    "cert": pem_bundle(leaf_pem),
                    "key": {"data": leaf_key, "format": "pem"},
                    "caCerts": [pem_bundle(ca_pem)],
                    "passphrase": "test-pass",
                    "friendlyName": "toolbit-test",
                },
                "expect": {
                    "code": 0,
                    "dataContains": {"encoding": "base64"},
                },
            },
            {
                "name": "私钥与证书不配对",
                "request": {
                    "cert": pem_bundle(leaf_pem),
                    "key": {"data": ca_key, "format": "pem"},
                    "passphrase": "test-pass",
                },
                "expect": {"code": 400, "errorCode": "KEY_MISMATCH"},
            },
        ],
    )

    # ── pfx-parse ────────────────────────────────────────────────
    write(
        "pfx-parse",
        "PKCS#12 解析的契约测试数据。",
        [
            {
                "name": "解析容器",
                "request": {
                    "container": {
                        "data": read_b64(fixtures, "bundle.p12"),
                        "format": "der",
                    },
                    "passphrase": "test-pass",
                },
                "expect": {
                    "code": 0,
                    "dataContains": {"hasPrivateKey": True, "caCount": 1},
                },
            },
            {
                "name": "口令错误",
                "request": {
                    "container": {
                        "data": read_b64(fixtures, "bundle.p12"),
                        "format": "der",
                    },
                    "passphrase": "wrong-pass",
                },
                "expect": {"code": 400, "errorCode": "PASSPHRASE_INCORRECT"},
            },
        ],
    )

    # ── pfx-change-password ──────────────────────────────────────
    #
    # 容器必须由 **Toolbit 自己打包**：上游 `PKCS12_newpass` 无法对外部实现
    # 产出的容器改密（见 docs/upstream/tongsuo-go-issues.md I-21）。
    write(
        "pfx-change-password",
        "PKCS#12 改密的契约测试数据。",
        [
            {
                "name": "改用新口令",
                "request": {
                    "container": {
                        "data": read_b64(fixtures, "bundle-toolbit.p12"),
                        "format": "der",
                    },
                    "oldPassphrase": "test-pass",
                    "newPassphrase": "new-test-pass",
                },
                "expect": {"code": 0, "dataContains": {"encoding": "base64"}},
            },
            {
                "name": "缺少新口令",
                "request": {
                    "container": {
                        "data": read_b64(fixtures, "bundle-toolbit.p12"),
                        "format": "der",
                    },
                    "oldPassphrase": "test-pass",
                },
                "expect": {"code": 400, "errorCode": "MISSING_FIELD"},
            },
        ],
    )

    # ── p7b-pack ─────────────────────────────────────────────────
    write(
        "p7b-pack",
        "PKCS#7 证书袋打包的契约测试数据。",
        [
            {
                "name": "打包两张证书",
                "request": {
                    "certs": [pem_bundle(leaf_pem), pem_bundle(ca_pem)],
                },
                "expect": {"code": 0, "dataContains": {"certCount": 2}},
            },
            {
                "name": "没有证书",
                "request": {"certs": []},
                "expect": {"code": 400, "errorCode": "MISSING_FIELD"},
            },
        ],
    )

    # ── p7b-extract ──────────────────────────────────────────────
    write(
        "p7b-extract",
        "PKCS#7 证书袋提取的契约测试数据。",
        [
            {
                "name": "提取两张证书",
                "request": {
                    "container": {
                        "data": read_b64(fixtures, "bundle.p7b"),
                        "format": "der",
                    },
                },
                "expect": {"code": 0, "dataContains": {"count": 2}},
            },
            {
                "name": "非法容器",
                "request": {"container": {"data": "MAM=", "format": "base64"}},
                "expect": {"code": 400, "errorCode": "INVALID_PKCS7"},
            },
        ],
    )

    # ── caa-gen ──────────────────────────────────────────────────
    write(
        "caa-gen",
        "CAA 记录生成的契约测试数据。",
        [
            {
                "name": "生成 issue 与 iodef",
                "request": {
                    "domain": "example.com",
                    "ttl": 3600,
                    "records": [
                        {"tag": "issue", "value": "letsencrypt.org"},
                        {
                            "tag": "iodef",
                            "value": "mailto:security@example.com",
                        },
                    ],
                },
                "expect": {
                    "code": 0,
                    "dataContains": {
                        "domain": "example.com",
                        "count": 2,
                    },
                },
            },
            {
                "name": "未知标签",
                "request": {
                    "domain": "example.com",
                    "records": [{"tag": "unknown", "value": "x"}],
                },
                "expect": {"code": 400, "errorCode": "INVALID_FIELD"},
            },
        ],
    )

    return 0


if __name__ == "__main__":
    sys.exit(main())
