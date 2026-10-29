#!/usr/bin/env bash
# 重新生成后端 PKI 测试夹具。
#
# 产物：backend/testdata/fixtures/pki/
#
# 这些材料**只**用于本仓库的单元测试，不是任何真实系统的密钥或证书
# （见 AGENTS.md §5）。
#
# 依赖：openssl（默认取 /opt/tongsuo/bin/openssl，可用 OPENSSL 覆盖）。
set -euo pipefail

REPO=$(cd "$(dirname "$0")/.." && pwd)
OUT="$REPO/backend/testdata/fixtures/pki"
OPENSSL=${OPENSSL:-/opt/tongsuo/bin/openssl}

command -v "$OPENSSL" >/dev/null 2>&1 || {
  echo "错误：未找到 openssl（$OPENSSL）" >&2
  exit 1
}

mkdir -p "$OUT"
WORK=$(mktemp -d)
trap 'rm -rf "$WORK"' EXIT
cd "$WORK"

cat > ext-inter.cnf <<'EOF'
basicConstraints=critical,CA:TRUE
keyUsage=critical,keyCertSign,cRLSign
subjectKeyIdentifier=hash
authorityKeyIdentifier=keyid:always
EOF

cat > ext-leaf.cnf <<'EOF'
basicConstraints=critical,CA:FALSE
keyUsage=critical,digitalSignature,keyEncipherment
extendedKeyUsage=serverAuth,clientAuth
subjectAltName=DNS:leaf.example.com,DNS:www.example.com,IP:192.0.2.10
subjectKeyIdentifier=hash
authorityKeyIdentifier=keyid:always
EOF

# 根 CA（自签）
#
# 有效期从 2020 年开始：这样「回拨到历史时刻验证」时，
# 整条链（含 CA）都在有效期内。
"$OPENSSL" req -x509 -newkey rsa:2048 -nodes -keyout ca.key.pem -out ca.crt.pem \
  -not_before 20200101000000Z -not_after 20360101000000Z \
  -subj "/C=CN/O=Toolbit Test/CN=Toolbit Test Root CA" 2>/dev/null

# 中间 CA（由根 CA 签发）
"$OPENSSL" req -newkey rsa:2048 -nodes -keyout inter.key.pem -out inter.csr.pem \
  -subj "/C=CN/O=Toolbit Test/CN=Toolbit Test Intermediate CA" 2>/dev/null
"$OPENSSL" x509 -req -in inter.csr.pem -CA ca.crt.pem -CAkey ca.key.pem -CAcreateserial \
  -not_before 20200101000000Z -not_after 20360101000000Z \
  -out inter.crt.pem -extfile ext-inter.cnf 2>/dev/null

# 叶证书（由中间 CA 签发，带 SAN）
"$OPENSSL" req -newkey rsa:2048 -nodes -keyout leaf.key.pem -out leaf.csr.pem \
  -subj "/C=CN/O=Toolbit Test/CN=leaf.example.com" 2>/dev/null
"$OPENSSL" x509 -req -in leaf.csr.pem -CA inter.crt.pem -CAkey inter.key.pem -CAcreateserial \
  -out leaf.crt.pem -days 825 -extfile ext-leaf.cnf 2>/dev/null

# 已过期证书（由中间 CA 签发，有效期在 2020 年 —— 用于 atTime 边界测试）
"$OPENSSL" req -newkey rsa:2048 -nodes -keyout expired.key.pem -out expired.csr.pem \
  -subj "/C=CN/O=Toolbit Test/CN=expired.example.com" 2>/dev/null
"$OPENSSL" x509 -req -in expired.csr.pem -CA inter.crt.pem -CAkey inter.key.pem \
  -CAcreateserial -not_before 20200101000000Z -not_after 20200102000000Z \
  -out expired.crt.pem 2>/dev/null || {
  echo "错误：当前 openssl 不支持 -not_before / -not_after" >&2
  exit 1
}

# CRL（由根 CA 签发，吊销叶证书）
mkdir -p demoCA/newcerts
: > demoCA/index.txt
echo 1000 > demoCA/serial
cat > crl.cnf <<'EOF'
[ ca ]
default_ca = CA_default

[ CA_default ]
dir = ./demoCA
database = $dir/index.txt
new_certs_dir = $dir/newcerts
serial = $dir/serial
default_md = sha256
private_key = ./ca.key.pem
certificate = ./ca.crt.pem
default_crl_days = 30
policy = policy_any
EOF
cat > policy.cnf <<'EOF'
[ policy_any ]
countryName = optional
stateOrProvinceName = optional
organizationName = optional
organizationalUnitName = optional
commonName = supplied
emailAddress = optional
EOF
"$OPENSSL" ca -config crl.cnf -policy policy.cnf -revoke leaf.crt.pem \
  -crl_reason keyCompromise -batch >/dev/null 2>&1
"$OPENSSL" ca -config crl.cnf -policy policy.cnf -gencrl -cert ca.crt.pem \
  -keyfile ca.key.pem -out crl.pem >/dev/null 2>&1

# 二进制形态与 bundle
"$OPENSSL" x509 -in ca.crt.pem -outform DER -out ca.crt.der
"$OPENSSL" x509 -in leaf.crt.pem -outform DER -out leaf.crt.der
"$OPENSSL" crl -in crl.pem -outform DER -out crl.der
cat leaf.crt.pem inter.crt.pem ca.crt.pem > chain.pem

# 只搬运证书 / CSR / CRL —— **私钥不落盘**。
cp ca.crt.pem ca.crt.der inter.crt.pem leaf.crt.pem leaf.crt.der \
  leaf.csr.pem expired.crt.pem crl.pem crl.der chain.pem "$OUT/"

echo "夹具已生成：$OUT"
ls -1 "$OUT"
