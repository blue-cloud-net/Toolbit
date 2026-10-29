package pki

import (
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestCertLifecycleMissingCertificate 断言缺少证书返回 MISSING_FIELD。
func TestCertLifecycleMissingCertificate(t *testing.T) {
	_, err := CertLifecycle(t.Context(), &CertLifecycleRequest{
		IsPEM:       true,
		WarningDays: 30,
	})
	assertCode(t, err, toolerr.CodeMissingField)
}
