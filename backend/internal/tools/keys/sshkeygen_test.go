package keys

import (
	"strings"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestSSHKeygenEd25519 断言 Ed25519 密钥格式。
func TestSSHKeygenEd25519(t *testing.T) {
	result, err := SSHKeygen(t.Context(), &SSHKeygenRequest{
		Algorithm: SSHAlgorithmEd25519,
		Comment:   "toolbit@example.com",
	})
	data := result.Data.(SSHKeygenResult)
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	if !strings.Contains(data.PrivateKey, "BEGIN OPENSSH PRIVATE KEY") {
		t.Errorf("私钥头不对：%s", data.PrivateKey)
	}
	if !strings.Contains(data.PublicKey, "ssh-ed25519") {
		t.Errorf("公钥应含类型名：%s", data.PublicKey)
	}
	if !strings.Contains(data.PublicKey, "toolbit@example.com") {
		t.Errorf("公钥应含注释：%s", data.PublicKey)
	}
	if data.Fingerprint == "" {
		t.Error("应给出指纹")
	}
}

// TestSSHKeygenRSA 断言 RSA 密钥格式。
func TestSSHKeygenRSA(t *testing.T) {
	result, err := SSHKeygen(t.Context(), &SSHKeygenRequest{
		Algorithm: SSHAlgorithmRSA,
		Bits:      2048,
		Comment:   "toolbit@example.com",
	})
	data := result.Data.(SSHKeygenResult)
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	if !strings.Contains(data.PublicKey, "ssh-rsa") {
		t.Errorf("公钥应含类型名：%s", data.PublicKey)
	}
	if data.Bits != 2048 {
		t.Errorf("位数 = %d，期望 2048", data.Bits)
	}
}

// TestSSHKeygenUnsupportedAlgorithm 断言未知算法返回 INVALID_FIELD。
func TestSSHKeygenUnsupportedAlgorithm(t *testing.T) {
	_, err := SSHKeygen(t.Context(), &SSHKeygenRequest{Algorithm: "DSA"})
	assertCode(t, err, toolerr.CodeInvalidField)
}
