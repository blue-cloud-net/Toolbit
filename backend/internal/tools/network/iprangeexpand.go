package network

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// 展开模式。
const (
	// IPRangeModeToCIDR 把起止地址转成 CIDR 列表。
	IPRangeModeToCIDR = "toCidr"
	// IPRangeModeToRange 把 CIDR 转成起止地址。
	IPRangeModeToRange = "toRange"
	// IPRangeModeExpand 把 CIDR 展开成地址列表。
	IPRangeModeExpand = "expand"
)

// IPRangeMode 返回受支持的模式（顺序稳定）。
func IPRangeMode() []string {
	return []string{IPRangeModeToCIDR, IPRangeModeToRange, IPRangeModeExpand}
}

// maxIPRangeAddresses 是展开地址数的上限。
//
// 展开是逐地址枚举，大网段会耗尽内存与响应体。
const maxIPRangeAddresses = 4096

// IPRangeRequest 是 ip-range-expand 工具的请求。
type IPRangeRequest struct {
	// Mode 是模式：toCidr / toRange / expand；默认 toCidr。
	Mode string `json:"mode"`
	// Cidr 是网段；toCidr 模式下与 Start/End 二选一。
	Cidr string `json:"cidr"`
	// Start 是起始地址。
	Start string `json:"start"`
	// End 是结束地址。
	End string `json:"end"`
	// Limit 是展开地址数的上限；默认 1024。
	Limit int `json:"limit"`
}

// IPRangeResult 是 ip-range-expand 工具的响应。
type IPRangeResult struct {
	// Mode 是实际使用的模式。
	Mode string `json:"mode"`
	// Cidrs 是网段列表。
	Cidrs []string `json:"cidrs"`
	// Start 是起始地址。
	Start string `json:"start"`
	// End 是结束地址。
	End string `json:"end"`
	// Count 是地址总数。
	Count int `json:"count"`
	// Addresses 是展开的地址列表；仅 expand 模式有值。
	Addresses []string `json:"addresses"`
	// Truncated 是否因超出上限而截断。
	Truncated bool `json:"truncated"`
}

// IPRange 在 IP 网段与地址区间之间转换。
//
// 这是**纯计算**工具：只做地址集合运算，不出网、不查询任何服务。
func IPRange(_ context.Context, req *IPRangeRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	mode := req.Mode
	if mode == "" {
		mode = IPRangeModeToCIDR
	}
	limit := req.Limit
	if limit == 0 {
		limit = 1024
	}

	switch mode {
	case IPRangeModeToRange:
		return ipRangeToRange(req)
	case IPRangeModeExpand:
		return ipRangeExpand(req, limit)
	default:
		return ipRangeToCIDR(req)
	}
}

// ipRangeToCIDR 把起止地址转成最简的 CIDR 列表。
func ipRangeToCIDR(req *IPRangeRequest) (tools.Result, error) {
	start, end, err := parseIPRangeBounds(req)
	if err != nil {
		return tools.Result{}, err
	}

	cidrs := []string{}
	current := start
	for {
		// 找出以 current 为起点的最大网段：先试最宽的前缀，
		// 逐级收窄直到网段末尾不越过 end。
		network, bits := networkFor(current)
		for b := bits; b >= 0; b-- {
			mask := net.CIDRMask(b, 32)
			candidate := current.Mask(mask)
			if candidate == nil || candidate.To4() == nil {
				continue
			}
			if !candidate.To4().Equal(current.To4()) {
				// 网段起点必须正好是 current，否则继续收窄。
				continue
			}
			if ipAfter(lastAddressIn(candidate, b), end) {
				// 网段越过了 end，继续收窄。
				continue
			}
			network, bits = candidate, b
			break
		}
		if !ipInNetwork(end, network, bits) {
			break
		}
		cidrs = append(cidrs, fmt.Sprintf("%s/%d", network.String(), bits))
		next := lastAddressIn(network, bits)
		if !ipBefore(current, end) {
			break
		}
		current = nextAddress(next)
	}
	if len(cidrs) == 0 {
		cidrs = append(cidrs, fmt.Sprintf("%s/%d", start.String(), bitsFor(start)))
	}

	return tools.Result{Data: IPRangeResult{
		Mode:      IPRangeModeToCIDR,
		Cidrs:     cidrs,
		Start:     start.String(),
		End:       end.String(),
		Count:     countAddresses(start, end),
		Addresses: []string{},
		Truncated: false,
	}}, nil
}

// ipRangeToRange 把 CIDR 转成起止地址。
func ipRangeToRange(req *IPRangeRequest) (tools.Result, error) {
	if strings.TrimSpace(req.Cidr) == "" {
		return tools.Result{}, tools.MissingField("cidr")
	}

	_, network, err := net.ParseCIDR(strings.TrimSpace(req.Cidr))
	if err != nil {
		return tools.Result{}, tools.InvalidField("cidr", "valid CIDR", "parse error")
	}

	start := network.IP
	end := lastAddress(network)

	return tools.Result{Data: IPRangeResult{
		Mode:      IPRangeModeToRange,
		Cidrs:     []string{network.String()},
		Start:     start.String(),
		End:       end.String(),
		Count:     countAddresses(start, end),
		Addresses: []string{},
		Truncated: false,
	}}, nil
}

// ipRangeExpand 把 CIDR 展开成地址列表。
func ipRangeExpand(req *IPRangeRequest, limit int) (tools.Result, error) {
	start, end, err := parseIPRangeBounds(req)
	if err != nil {
		return tools.Result{}, err
	}

	total := countAddresses(start, end)
	addresses := []string{}
	truncated := false

	current := start
	for {
		if len(addresses) >= limit {
			truncated = true
			break
		}
		addresses = append(addresses, current.String())
		if !ipBefore(current, end) {
			// current 已不早于 end，展开结束。
			break
		}
		current = nextAddress(current)
	}

	return tools.Result{Data: IPRangeResult{
		Mode:      IPRangeModeExpand,
		Cidrs:     []string{},
		Start:     start.String(),
		End:       end.String(),
		Count:     total,
		Addresses: addresses,
		Truncated: truncated,
	}}, nil
}

// parseIPRangeBounds 解析起止地址。
//
// 优先用 `start` / `end`；缺省时才回退到 `cidr` 的网段边界。
func parseIPRangeBounds(req *IPRangeRequest) (net.IP, net.IP, error) {
	hasStart := strings.TrimSpace(req.Start) != ""
	hasEnd := strings.TrimSpace(req.End) != ""

	if hasStart || hasEnd {
		if !hasStart || !hasEnd {
			return nil, nil, tools.MissingField("start")
		}

		start := net.ParseIP(strings.TrimSpace(req.Start))
		if start == nil {
			return nil, nil, tools.InvalidField("start", "valid IP address", "parse error")
		}
		end := net.ParseIP(strings.TrimSpace(req.End))
		if end == nil {
			return nil, nil, tools.InvalidField("end", "valid IP address", "parse error")
		}
		if start.To4() == nil || end.To4() == nil {
			return nil, nil, tools.InvalidField("start", "IPv4 address", "not IPv4")
		}
		if ipBefore(end, start) {
			return nil, nil, tools.InvalidField("end", "not before start", "reversed range")
		}
		return start, end, nil
	}

	if strings.TrimSpace(req.Cidr) != "" {
		ip, network, err := net.ParseCIDR(strings.TrimSpace(req.Cidr))
		if err != nil {
			return nil, nil, tools.InvalidField("cidr", "valid CIDR", "parse error")
		}
		return ip, lastAddress(network), nil
	}

	return nil, nil, tools.MissingField("start")
}

// ipInNetwork 判断地址是否落在网段内。
func ipInNetwork(ip net.IP, network net.IP, bits int) bool {
	mask := net.CIDRMask(bits, 32)
	return ip.Mask(mask).Equal(network)
}

// ipBefore 判断 a 是否排在 b 之前。
func ipBefore(a, b net.IP) bool {
	return bytes.Compare(a.To4(), b.To4()) < 0
}

// ipAfter 判断 a 是否排在 b 之后。
func ipAfter(a, b net.IP) bool {
	return bytes.Compare(a.To4(), b.To4()) > 0
}

// networkFor 找出以给定地址为起点的最宽网段。
//
// 返回的网段起点**正好**是给定地址，因此可以逐段推进而不重叠。
func networkFor(ip net.IP) (net.IP, int) {
	v4 := ip.To4()
	if v4 == nil {
		return ip, bitsFor(ip)
	}
	for bits := 0; bits <= 32; bits++ {
		mask := net.CIDRMask(bits, 32)
		network := ip.Mask(mask)
		if network == nil || network.To4() == nil {
			continue
		}
		if network.To4().Equal(v4) {
			return network, bits
		}
	}
	return ip, 32
}

// bitsFor 返回 IPv4 地址的位数。
func bitsFor(ip net.IP) int {
	if ip.To4() != nil {
		return 32
	}
	return 128
}

// lastAddress 返回网段中的最后一个地址。
func lastAddress(network *net.IPNet) net.IP {
	ip := network.IP.To4()
	if ip == nil {
		return network.IP
	}
	mask := network.Mask
	if len(mask) == net.IPv6len {
		mask = mask[12:]
	}
	out := make(net.IP, net.IPv4len)
	for i := 0; i < net.IPv4len; i++ {
		out[i] = ip[i] | ^mask[i]
	}
	return out
}

// lastAddressIn 返回给定地址与前缀长度所构成网段的最后一个地址。
func lastAddressIn(ip net.IP, bits int) net.IP {
	mask := net.CIDRMask(bits, 32)
	out := make(net.IP, net.IPv4len)
	for i := 0; i < net.IPv4len; i++ {
		out[i] = ip.To4()[i] | ^mask[i]
	}
	return out
}

// nextAddress 返回下一个地址。
func nextAddress(ip net.IP) net.IP {
	out := make(net.IP, len(ip))
	copy(out, ip)
	for i := len(out) - 1; i >= 0; i-- {
		out[i]++
		if out[i] != 0 {
			break
		}
	}
	return out
}

// countAddresses 统计地址总数。
//
// 超出 uint32 的范围按 0 处理：无法用 int 精确表示，且超出展开上限。
func countAddresses(start, end net.IP) int {
	start4 := start.To4()
	end4 := end.To4()
	if start4 == nil || end4 == nil {
		return 0
	}
	startNum := uint32(start4[0])<<24 | uint32(start4[1])<<16 | uint32(start4[2])<<8 | uint32(start4[3])
	endNum := uint32(end4[0])<<24 | uint32(end4[1])<<16 | uint32(end4[2])<<8 | uint32(end4[3])
	if endNum < startNum {
		return 0
	}
	diff := uint64(endNum) - uint64(startNum) + 1
	if diff > uint64(maxIPRangeAddresses)*1024 {
		return 0
	}
	return int(diff)
}

// Validate 做字段级校验。
func (r *IPRangeRequest) Validate() error {
	mode := r.Mode
	if mode == "" {
		mode = IPRangeModeToCIDR
	}
	if !containsString(IPRangeMode(), mode) {
		return tools.InvalidField("mode", IPRangeMode(), mode)
	}
	if r.Limit < 0 || r.Limit > maxIPRangeAddresses {
		return tools.InvalidField("limit", []int{0, maxIPRangeAddresses}, r.Limit)
	}
	if mode == IPRangeModeToRange && strings.TrimSpace(r.Cidr) == "" {
		return tools.MissingField("cidr")
	}
	if mode != IPRangeModeToRange &&
		strings.TrimSpace(r.Cidr) == "" &&
		(strings.TrimSpace(r.Start) == "" || strings.TrimSpace(r.End) == "") {
		return tools.MissingField("start")
	}
	return nil
}
