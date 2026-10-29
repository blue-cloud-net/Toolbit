package netguard

import (
	"errors"
	"fmt"
)

// Reason 是出网请求被拒绝的原因。
//
// 取值会出现在日志与错误参数里，因此保持稳定的短横线形式。
type Reason string

// 拒绝原因取值。
const (
	// ReasonDisabled 表示出网总开关关闭。
	ReasonDisabled Reason = "network-disabled"
	// ReasonScheme 表示协议不在白名单（仅 http / https）。
	ReasonScheme Reason = "unsupported-scheme"
	// ReasonPort 表示端口不在白名单。
	ReasonPort Reason = "port-not-allowed"
	// ReasonAddress 表示解析出的地址落在私有 / 保留网段。
	ReasonAddress Reason = "private-address"
	// ReasonHost 表示主机名为空或形式非法。
	ReasonHost Reason = "invalid-host"
	// ReasonRedirect 表示重定向跳数超限。
	ReasonRedirect Reason = "too-many-redirects"
)

// ErrDisabled 表示出网总开关关闭。
//
// 调用方用 errors.Is 判别，并映射为 NETWORK_DISABLED。
var ErrDisabled = errors.New("netguard: 出网已禁用")

// Error 表示一次出网请求被策略拒绝。
//
// 字段只携带主机名与端口，**不**携带用户提交的完整 URL，
// 避免把查询串等输入写进日志。
type Error struct {
	// Reason 是拒绝原因。
	Reason Reason
	// Host 是被拒绝的主机名或协议名。
	Host string
	// Port 是被拒绝的端口；0 表示与端口无关。
	Port int
	// Addr 是被拒绝的地址（仅 ReasonAddress 时非空）。
	Addr string
}

// Error 实现 error 接口。
func (e *Error) Error() string {
	switch e.Reason {
	case ReasonDisabled:
		return "netguard: 出网已禁用"
	case ReasonScheme:
		return fmt.Sprintf("netguard: 不支持的协议 %s", e.Host)
	case ReasonPort:
		return fmt.Sprintf("netguard: 端口 %d 不在出网白名单内", e.Port)
	case ReasonAddress:
		return fmt.Sprintf("netguard: 目标地址 %s 属于私有或保留网段", e.Addr)
	case ReasonHost:
		return "netguard: 主机名无效"
	case ReasonRedirect:
		return "netguard: 重定向跳数超过上限"
	default:
		return "netguard: 出网请求被策略拒绝"
	}
}
