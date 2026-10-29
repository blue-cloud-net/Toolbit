#!/usr/bin/env python3
"""生成 v0.5.0 的 8 份契约测试数据。

材料来自 backend/testdata/fixtures/pki/（由 scripts/gen-pki-fixtures.sh 生成）。
本脚本只做「把夹具内容嵌进 JSON」，不生成任何密码学材料。
"""

import base64
import json
import pathlib

ROOT = pathlib.Path(__file__).resolve().parents[1]
FIX = ROOT / "backend" / "testdata" / "fixtures" / "pki"
OUT = ROOT / "backend" / "testdata" / "contract"


def read(name: str) -> str:
    return (FIX / name).read_text(encoding="utf-8")


def read_b64(name: str) -> str:
    """读二进制夹具并返回 base64 文本。"""
    return base64.b64encode((FIX / name).read_bytes()).decode("ascii")


def write(slug: str, description: str, cases: list) -> None:
    path = OUT / f"{slug}.json"
    payload = {"slug": slug, "description": description, "cases": cases}
    path.write_text(
        json.dumps(payload, ensure_ascii=False, indent=2) + "\n", encoding="utf-8"
    )
    print(f"written {path.relative_to(ROOT)} ({len(cases)} cases)")


leaf_pem = read("leaf.crt.pem")
leaf_der_b64 = read_b64("leaf.crt.der")
inter_pem = read("inter.crt.pem")
root_pem = read("ca.crt.pem")
chain_pem = read("chain.pem")
csr_pem = read("leaf.csr.pem")
expired_pem = read("expired.crt.pem")
crl_pem = read("crl.pem")

# ── cert-parse ──────────────────────────────────────────────────────────
write(
    "cert-parse",
    "证书解析的契约测试数据。",
    [
        {
            "name": "解析叶证书",
            "request": {"input": {"data": leaf_pem, "format": "pem"}},
            "expect": {
                "code": 0,
                "dataContains": {
                    "certificate": {
                        "subjectText": "/C=CN/O=Toolbit Test/CN=leaf.example.com",
                        "issuerText": "/C=CN/O=Toolbit Test/CN=Toolbit Test Intermediate CA",
                        "version": 2,
                        "isCa": False,
                        "selfSigned": False,
                        "publicKey": {"algorithm": "RSA", "bits": 2048},
                    },
                    "selfSigned": False,
                },
            },
        },
        {
            "name": "计算多种指纹",
            "request": {
                "input": {"data": leaf_pem, "format": "pem"},
                "fingerprints": ["sha256", "sha1", "sm3"],
            },
            "expect": {
                "code": 0,
                "dataContains": {
                    "fingerprints": [
                        {"algorithm": "sha256"},
                        {"algorithm": "sha1"},
                        {"algorithm": "sm3"},
                    ],
                },
            },
        },
        {
            "name": "按历史时刻判断有效期",
            "request": {
                "input": {"data": leaf_pem, "format": "pem"},
                "atTime": "2026-01-01T00:00:00Z",
            },
            "expect": {
                "code": 0,
                "dataContains": {"validity": {"status": "not_yet_valid"}},
            },
        },
        {
            "name": "已过期证书",
            "request": {"input": {"data": expired_pem, "format": "pem"}},
            "expect": {
                "code": 0,
                "dataContains": {"validity": {"status": "expired"}},
                "warnings": ["EXPIRED_CERT"],
            },
        },
        {
            "name": "自签名根证书",
            "request": {"input": {"data": root_pem, "format": "pem"}},
            "expect": {
                "code": 0,
                "dataContains": {"selfSigned": True},
                "warnings": ["SELF_SIGNED_CERT"],
            },
        },
        {
            "name": "非法证书",
            "request": {"input": {"data": "not a certificate", "format": "pem"}},
            "expect": {"code": 400, "errorCode": "INVALID_CERT"},
        },
        {
            "name": "时刻格式非法",
            "request": {
                "input": {"data": leaf_pem, "format": "pem"},
                "atTime": "2026/01/01",
            },
            "expect": {"code": 400, "errorCode": "INVALID_TIME"},
        },
    ],
)

# ── csr-parse ──────────────────────────────────────────────────────────
write(
    "csr-parse",
    "证书签名请求解析的契约测试数据。",
    [
        {
            "name": "解析 CSR",
            "request": {"input": {"data": csr_pem, "format": "pem"}},
            "expect": {
                "code": 0,
                "dataContains": {
                    "subjectText": "/C=CN/O=Toolbit Test/CN=leaf.example.com",
                    "signatureValid": True,
                    "challengePasswordPresent": False,
                    "publicKey": {"algorithm": "RSA", "bits": 2048},
                },
            },
        },
        {
            "name": "非法 CSR",
            "request": {"input": {"data": "not a csr", "format": "pem"}},
            "expect": {"code": 400, "errorCode": "INVALID_CSR"},
        },
    ],
)

# ── pem-der ────────────────────────────────────────────────────────────
write(
    "pem-der",
    "PEM / DER 转换的契约测试数据。",
    [
        {
            "name": "PEM 转 DER",
            "request": {
                "input": {"data": leaf_pem, "format": "pem"},
                "type": "cert",
                "to": "der",
            },
            "expect": {
                "code": 0,
                "dataContains": {
                    "output": {"format": "der", "encoding": "base64"},
                    "blockCount": 1,
                },
            },
        },
        {
            "name": "DER 转 PEM",
            "request": {
                "input": {"data": leaf_der_b64, "format": "der"},
                "type": "cert",
                "to": "pem",
            },
            "expect": {
                "code": 0,
                "dataContains": {
                    "output": {"format": "pem", "encoding": "utf8"},
                    "blockCount": 1,
                },
            },
        },
        {
            "name": "多块 bundle 转 DER 被拒",
            "request": {
                "input": {"data": chain_pem, "format": "pem"},
                "type": "cert",
                "to": "der",
            },
            "expect": {"code": 400, "errorCode": "UNSUPPORTED_FORMAT"},
        },
        {
            "name": "类型过滤不匹配",
            "request": {
                "input": {"data": leaf_pem, "format": "pem"},
                "type": "crl",
                "to": "pem",
            },
            "expect": {"code": 400, "errorCode": "UNSUPPORTED_FORMAT"},
        },
        {
            "name": "未知目标格式",
            "request": {
                "input": {"data": leaf_pem, "format": "pem"},
                "to": "hex",
            },
            "expect": {"code": 400, "errorCode": "UNSUPPORTED_FORMAT"},
        },
    ],
)

# ── asn1-view ──────────────────────────────────────────────────────────
write(
    "asn1-view",
    "ASN.1 视图的契约测试数据。",
    [
        {
            "name": "查看证书 DER 结构",
            "request": {"input": {"data": leaf_der_b64, "format": "der"}},
            "expect": {
                "code": 0,
                "dataContains": {
                    "tree": {"number": 16, "constructed": True},
                },
            },
        },
        {
            "name": "限制展开深度",
            "request": {
                "input": {"data": leaf_der_b64, "format": "der"},
                "maxDepth": 2,
            },
            "expect": {
                "code": 0,
                "dataContains": {"maxDepth": 2},
            },
        },
        {
            "name": "非法 DER",
            "request": {"input": {"data": "MAM=", "format": "base64"}},
            "expect": {"code": 400, "errorCode": "INVALID_DER"},
        },
    ],
)

# ── cert-chain-parse ───────────────────────────────────────────────────
write(
    "cert-chain-parse",
    "证书链解析的契约测试数据。",
    [
        {
            "name": "解析完整链",
            "request": {"bundle": {"data": chain_pem, "format": "pem"}},
            "expect": {
                "code": 0,
                "dataContains": {
                    "count": 3,
                    "ordered": True,
                    "missingRoot": False,
                    "selfSigned": True,
                },
            },
        },
        {
            "name": "乱序输入",
            "request": {
                "bundle": {
                    "data": root_pem + inter_pem + leaf_pem,
                    "format": "pem",
                }
            },
            "expect": {
                "code": 0,
                "dataContains": {"count": 3, "ordered": False},
            },
        },
        {
            "name": "非法证书",
            "request": {"bundle": {"data": "not a certificate", "format": "pem"}},
            "expect": {"code": 400, "errorCode": "INVALID_PEM"},
        },
    ],
)

# ── cert-chain-verify ──────────────────────────────────────────────────
write(
    "cert-chain-verify",
    "证书链验证的契约测试数据。",
    [
        {
            "name": "可信根验证通过",
            "request": {
                "bundle": {"data": chain_pem, "format": "pem"},
                "trustedRoots": [{"data": root_pem, "format": "pem"}],
                "host": "leaf.example.com",
            },
            "expect": {
                "code": 0,
                "dataContains": {
                    "valid": True,
                    "hostnameMatch": True,
                    "implicitTrust": False,
                },
            },
        },
        {
            "name": "主机名不匹配",
            "request": {
                "bundle": {"data": chain_pem, "format": "pem"},
                "trustedRoots": [{"data": root_pem, "format": "pem"}],
                "host": "other.example.com",
            },
            "expect": {
                "code": 0,
                "dataContains": {
                    "valid": True,
                    "hostnameMatch": False,
                },
            },
        },
        {
            "name": "无受信任根时使用隐含锚",
            "request": {"bundle": {"data": chain_pem, "format": "pem"}},
            "expect": {
                "code": 0,
                "dataContains": {"implicitTrust": True, "valid": True},
            },
        },
        {
            "name": "过期证书验证失败",
            "request": {
                "bundle": {"data": expired_pem + inter_pem, "format": "pem"},
                "trustedRoots": [{"data": root_pem, "format": "pem"}],
            },
            "expect": {
                "code": 0,
                "dataContains": {"valid": False},
                "warnings": ["EXPIRED_CERT", "UNTRUSTED_CHAIN"],
            },
        },
    ],
)

# ── cert-chain-repair ──────────────────────────────────────────────────
write(
    "cert-chain-repair",
    "证书链修复的契约测试数据。",
    [
        {
            "name": "补全缺失的中间证书与根",
            "request": {
                "cert": {"data": leaf_pem, "format": "pem"},
                "intermediates": [{"data": inter_pem, "format": "pem"}],
                "trustedRoots": [{"data": root_pem, "format": "pem"}],
            },
            "expect": {
                "code": 0,
                "dataContains": {
                    "addedCount": 2,
                    "reachedRoot": True,
                    "note": "reached_root",
                    "addedFroms": ["intermediates", "trustedRoots"],
                },
            },
        },
        {
            "name": "无候选可补",
            "request": {"cert": {"data": leaf_pem, "format": "pem"}},
            "expect": {
                "code": 0,
                "dataContains": {
                    "addedCount": 0,
                    "reachedRoot": False,
                    "note": "already_complete",
                },
            },
        },
        {
            "name": "非法叶证书",
            "request": {"cert": {"data": "not a certificate", "format": "pem"}},
            "expect": {"code": 400, "errorCode": "INVALID_CERT"},
        },
    ],
)

# ── crl-parse ──────────────────────────────────────────────────────────
write(
    "crl-parse",
    "吊销列表解析的契约测试数据。",
    [
        {
            "name": "解析 CRL",
            "request": {"input": {"data": crl_pem, "format": "pem"}},
            "expect": {
                "code": 0,
                "dataContains": {
                    "issuerText": "/C=CN/O=Toolbit Test/CN=Toolbit Test Root CA",
                    "revoked": [{"reasonCode": 1, "reasonText": "keyCompromise"}],
                },
            },
        },
        {
            "name": "非法 CRL",
            "request": {"input": {"data": "not a crl", "format": "pem"}},
            "expect": {"code": 400, "errorCode": "INVALID_CRL"},
        },
    ],
)
