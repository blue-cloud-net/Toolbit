package network

import (
	"time"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// chainWarnings 依据链的状态给出警告。
//
// 警告只描述**事实**，不改变 HTTP 状态。
func chainWarnings(probe *TLSProbeResult, summary ChainSummary) []toolerr.Warning {
	warnings := make([]toolerr.Warning, 0, 3)

	if !summary.ChainValid {
		warnings = append(warnings, toolerr.NewWarning(toolerr.WarnUntrustedChain, nil))
	} else if !summary.Verified {
		// 链在「链内自洽」意义下通过，但没有系统信任根可用。
		warnings = append(warnings, toolerr.NewWarning(toolerr.WarnUntrustedChain, nil))
	}
	if summary.SelfSigned {
		warnings = append(warnings, toolerr.NewWarning(toolerr.WarnSelfSignedCert, nil))
	}

	if len(probe.Certificates) > 0 && probe.Certificates[0].Info != nil {
		leaf := probe.Certificates[0].Info
		now := time.Now()
		switch {
		case now.After(leaf.NotAfter):
			warnings = append(warnings, toolerr.NewWarning(toolerr.WarnExpiredCert, nil))
		case now.Before(leaf.NotBefore):
			warnings = append(warnings, toolerr.NewWarning(toolerr.WarnNotYetValid, nil))
		}
	}
	return warnings
}
