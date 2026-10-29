# PKI 测试夹具

本目录存放**后端单元测试**用的证书链、CSR 与 CRL。

## 这些是什么

由 `openssl`（Tongsuo 8.5.0-pre1 / OpenSSL 3.5.4）**一次性生成**的测试材料：

| 文件 | 说明 |
| --- | --- |
| `ca.crt.pem` / `ca.crt.der` | 自签根 CA（`CN=Toolbit Test Root CA`） |
| `inter.crt.pem` | 中间 CA（由根 CA 签发，带 SKID / AKID） |
| `leaf.crt.pem` / `leaf.crt.der` | 叶证书（由中间 CA 签发，SAN 含 DNS 与 IP） |
| `leaf.csr.pem` | 叶证书的 CSR |
| `expired.crt.pem` | 有效期为 2020-01-01 ~ 2020-01-02 的**已过期**证书 |
| `crl.pem` / `crl.der` | 由根 CA 签发的 CRL，吊销了叶证书（原因 `keyCompromise`） |
| `chain.pem` | 叶 → 中间 → 根 的 PEM bundle |

## 重要声明

- 这些材料**只**在本仓库的测试里使用；
- 它们**不是**任何真实系统的密钥或证书，也不对应任何真实主体；
- **私钥不在本目录内**（`*.key` 被 `.gitignore` 忽略），测试也不需要私钥。

## 重新生成

```bash
bash scripts/gen-pki-fixtures.sh
```

脚本会用 `openssl` 重新生成全部材料；序列号与有效期因此会变化，
更新夹具时**同步**检查 `backend/internal/pkisvc/pki_test.go` 里的断言。
