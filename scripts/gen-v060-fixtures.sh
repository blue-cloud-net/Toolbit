#!/usr/bin/env bash
# 生成 v0.6.0 的契约测试材料（自签 CA + 叶证书 + CRL + PKCS#12 + PKCS#7 + OCSP）。
#
# ⚠️ 这些材料由本脚本**一次性生成**，只用于本仓库的契约测试，
# **不是**任何真实系统的密钥或证书（见 AGENTS.md §5）。
# 私钥**不落盘到仓库**：脚本只在临时目录里生成，随后由
# scripts/gen-v060-contracts.py 内联进契约 JSON。
#
# 用法：bash scripts/gen-v060-fixtures.sh <输出目录>
set -euo pipefail

OPENSSL=${OPENSSL:-/opt/tongsuo/bin/openssl}
OUT=${1:?用法: gen-v060-fixtures.sh <输出目录>}

mkdir -p "$OUT"
WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT

cd "$WORK"

# ── CA（自签，带 SKID / AKID，cA:true）───────────────────────────────
"$OPENSSL" genrsa -out ca.key 2048 2>/dev/null
cat > ca.ext <<'EOF'
basicConstraints = critical,CA:true,pathlen:1
keyUsage = critical,keyCertSign,cRLSign
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always
EOF
"$OPENSSL" req -x509 -new -key ca.key -sha256 -days 7300 \
  -subj "/C=CN/O=Toolbit Test/CN=Toolbit Test Issuing CA" \
  -extensions v3_ca -config /dev/stdin -out ca.crt <<'EOF' 2>/dev/null
[req]
distinguished_name = dn
prompt = no
[dn]
[ v3_ca ]
basicConstraints = critical,CA:true,pathlen:1
keyUsage = critical,keyCertSign,cRLSign
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always
EOF

# ── 叶证书 ────────────────────────────────────────────────────────
"$OPENSSL" genrsa -out leaf.key 2048 2>/dev/null
"$OPENSSL" req -new -key leaf.key \
  -subj "/C=CN/O=Toolbit Test/CN=gen.example.com" -out leaf.csr 2>/dev/null
cat > leaf.ext <<'EOF'
basicConstraints = critical,CA:false
keyUsage = critical,digitalSignature,keyEncipherment
extendedKeyUsage = serverAuth
subjectAltName = DNS:gen.example.com,DNS:www.gen.example.com
subjectKeyIdentifier = hash
authorityKeyIdentifier = keyid:always
EOF
# 序列号刻意取**小于 8 字节**：上游用 int64 承载序列号，
# 超长序列号在 CRL 侧读不出来（见 I-18）。
"$OPENSSL" x509 -req -in leaf.csr -CA ca.crt -CAkey ca.key -set_serial 0x1001 \
  -days 3650 -sha256 -extfile leaf.ext -out leaf.crt 2>/dev/null

# ── CRL（吊销叶证书，原因 keyCompromise）──────────────────────────
mkdir -p demoCA/newcerts
: > demoCA/index.txt
echo 1000 > demoCA/serial
echo 1000 > demoCA/crlnumber
cat > crl.cnf <<'EOF'
[ca]
default_ca = CA_default
[CA_default]
dir = ./demoCA
database = $dir/index.txt
new_certs_dir = $dir/newcerts
certificate = ./ca.crt
private_key = ./ca.key
serial = $dir/serial
crlnumber = $dir/crlnumber
default_md = sha256
default_crl_days = 30
policy = policy_any
[policy_any]
commonName = supplied
EOF
"$OPENSSL" ca -config crl.cnf -revoke leaf.crt -crl_reason keyCompromise 2>/dev/null
"$OPENSSL" ca -config crl.cnf -gencrl -out crl.pem 2>/dev/null

# ── PKCS#12（口令 test-pass）──────────────────────────────────────
# ⚠️ 这个容器只用于 pfx-parse（读取侧）：上游的 PKCS12_newpass 无法对它改密
# （见 docs/upstream/tongsuo-go-issues.md I-21）。改密用的容器由 Toolbit 自己打包，
# 见下方提示与 scripts/gen-v060-contracts.py 的说明。
"$OPENSSL" pkcs12 -export -out bundle.p12 -inkey leaf.key -in leaf.crt \
  -certfile ca.crt -passout pass:test-pass -name toolbit-test 2>/dev/null

# ── PKCS#7 证书袋 ─────────────────────────────────────────────────
"$OPENSSL" crl2pkcs7 -nocrl -certfile leaf.crt -certfile ca.crt \
  -out bundle.p7b 2>/dev/null

# ── OCSP 请求与响应 ───────────────────────────────────────────────
"$OPENSSL" ocsp -issuer ca.crt -cert leaf.crt -no_nonce -reqout ocsp-req.der 2>/dev/null
# 由 CA 自己充当响应者生成一份真实响应；失败则跳过（脚本不因此中断）。
: > ocsp-index.txt
if "$OPENSSL" ocsp -index ocsp-index.txt -CA ca.crt \
    -rsigner ca.crt -rkey ca.key -reqin ocsp-req.der \
    -respout ocsp-resp.der -ndays 1 2>/dev/null; then
  echo "OCSP 响应已生成"
else
  echo "⚠️ OCSP 响应生成失败，将只产出负例" >&2
  : > ocsp-resp.der
fi

cp ca.crt "$OUT/ca2.crt.pem"
cp ca.key "$OUT/ca2.key.pem"
cp leaf.crt "$OUT/leaf2.crt.pem"
cp leaf.key "$OUT/leaf2.key.pem"
cp leaf.csr "$OUT/leaf2.csr.pem"
cp crl.pem "$OUT/crl2.pem"
cp bundle.p12 "$OUT/bundle.p12"
cp bundle.p7b "$OUT/bundle.p7b"
cp ocsp-req.der "$OUT/ocsp-req.der"
[ -s ocsp-resp.der ] && cp ocsp-resp.der "$OUT/ocsp-resp.der"

echo "✅ 材料已写入 $OUT"
echo "下一步：TOOLBIT_V060_FIXTURES=$OUT go test ./internal/pkisvc -run TestGenerateToolbitP12"
