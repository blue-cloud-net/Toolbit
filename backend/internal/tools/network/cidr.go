package network

import (
	"context"
	"math/big"
	"net/netip"
	"strings"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// maxSplitCIDRs 是一次拆分允许的子网数量上限。
const maxSplitCIDRs = 256

// tooManyItems 构造条目数超限的错误。
func tooManyItems(field string, actual int) *toolerr.Error {
	return toolerr.Newf(toolerr.CodeTooManyItems, map[string]any{
		"field":  field,
		"max":    maxSplitCIDRs,
		"actual": actual,
	})
}

// invalidCIDR 构造网段非法的错误。
func invalidCIDR() *toolerr.Error {
	return toolerr.New(toolerr.CodeInvalidCIDR)
}

// CIDRCalcRequest 是 CIDR 计算工具的请求。
type CIDRCalcRequest struct {
	// CIDR 是单个网段，如 `192.168.1.0/24`；与 Start/End 二选一。
	CIDR string `json:"cidr"`
	// Start 是地址范围起点；与 End 成对使用。
	Start string `json:"start"`
	// End 是地址范围终点；与 Start 成对使用。
	End string `json:"end"`
	// Split 是按指定前缀长度拆分网段；0 表示不拆分。
	Split int `json:"split"`
}

// CIDRCalcResult 是 CIDR 计算工具的响应。
type CIDRCalcResult struct {
	// Version 是地址族：4 或 6。
	Version int `json:"version"`
	// Network 是网络地址（含前缀长度）。
	Network string `json:"network"`
	// Prefix 是前缀长度。
	Prefix int `json:"prefix"`
	// Mask 是子网掩码。
	Mask string `json:"mask"`
	// FirstAddress 是网段首地址。
	FirstAddress string `json:"firstAddress"`
	// LastAddress 是网段末地址。
	LastAddress string `json:"lastAddress"`
	// Broadcast 是广播地址；IPv6 无广播概念，字段为空。
	Broadcast string `json:"broadcast"`
	// FirstUsable 是首个可用地址。
	FirstUsable string `json:"firstUsable"`
	// LastUsable 是末个可用地址。
	LastUsable string `json:"lastUsable"`
	// HostCount 是可用主机数（十进制字符串，IPv6 可能超出 64 位）。
	HostCount string `json:"hostCount"`
	// TotalAddresses 是网段地址总数（十进制字符串）。
	TotalAddresses string `json:"totalAddresses"`
	// CIDRs 是按 Split 拆分后的子网列表。
	CIDRs []string `json:"cidrs"`
	// Range 是地址范围反推出的最小 CIDR 集合（仅范围输入时非空）。
	Range []string `json:"range"`
}

// Validate 做字段级校验。
func (r *CIDRCalcRequest) Validate() error {
	hasCIDR := strings.TrimSpace(r.CIDR) != ""
	hasStart := strings.TrimSpace(r.Start) != ""
	hasEnd := strings.TrimSpace(r.End) != ""

	switch {
	case hasCIDR && (hasStart || hasEnd):
		return tools.InvalidField("cidr", "cidr 与 start/end 二选一", "both")
	case !hasCIDR && (!hasStart || !hasEnd):
		if hasStart || hasEnd {
			missing := "end"
			if hasEnd {
				missing = "start"
			}
			return tools.MissingField(missing)
		}
		return tools.MissingField("cidr")
	}

	if r.Split != 0 {
		if hasStart || hasEnd {
			return tools.InvalidField("split", "仅在 cidr 模式下支持", r.Split)
		}
		if r.Split < 0 || r.Split > 128 {
			return tools.InvalidField("split", []string{"1..128"}, r.Split)
		}
	}
	return nil
}

// CIDRCalc 执行 CIDR 计算。
func CIDRCalc(_ context.Context, req *CIDRCalcRequest) (tools.Result, error) {
	if strings.TrimSpace(req.CIDR) != "" {
		return calcFromPrefix(req)
	}
	return calcFromRange(req)
}

// calcFromPrefix 处理单网段输入。
func calcFromPrefix(req *CIDRCalcRequest) (tools.Result, error) {
	prefix, err := netip.ParsePrefix(strings.TrimSpace(req.CIDR))
	if err != nil {
		return tools.Result{}, invalidCIDR()
	}
	prefix = prefix.Masked()
	result, err := describePrefix(prefix)
	if err != nil {
		return tools.Result{}, err
	}
	if req.Split > 0 {
		cidrs, err := splitPrefix(prefix, req.Split)
		if err != nil {
			return tools.Result{}, err
		}
		result.CIDRs = cidrs
	}
	return tools.Result{Data: result}, nil
}

// calcFromRange 处理地址范围输入。
func calcFromRange(req *CIDRCalcRequest) (tools.Result, error) {
	first, err := netip.ParseAddr(strings.TrimSpace(req.Start))
	if err != nil {
		return tools.Result{}, invalidCIDR()
	}
	last, err := netip.ParseAddr(strings.TrimSpace(req.End))
	if err != nil {
		return tools.Result{}, invalidCIDR()
	}
	first, last = first.Unmap(), last.Unmap()
	if first.Is4() != last.Is4() {
		return tools.Result{}, tools.InvalidField("end", "与 start 同地址族", last.String())
	}
	if first.Compare(last) > 0 {
		return tools.Result{}, tools.InvalidField("end", "不小于 start", last.String())
	}

	prefixes, err := rangeToPrefixes(first, last)
	if err != nil {
		return tools.Result{}, err
	}
	out := make([]string, 0, len(prefixes))
	for _, p := range prefixes {
		out = append(out, p.String())
	}

	// 顶层汇总：用覆盖整个范围的最小公共前缀描述，并附上完整反推结果。
	result, err := describePrefix(commonPrefix(first, last))
	if err != nil {
		return tools.Result{}, err
	}
	result.Range = out
	return tools.Result{Data: result}, nil
}

// describePrefix 汇总一个网段的信息。
func describePrefix(prefix netip.Prefix) (CIDRCalcResult, error) {
	bits := bitsOf(prefix.Addr())
	if prefix.Bits() > bits {
		return CIDRCalcResult{}, invalidCIDR()
	}

	network := prefix.Masked().Addr()
	total := new(big.Int).Lsh(big.NewInt(1), uint(bits-prefix.Bits()))
	last := lastAddr(network, total)

	result := CIDRCalcResult{
		Version:        versionOf(bits),
		Network:        prefix.String(),
		Prefix:         prefix.Bits(),
		Mask:           maskString(prefix),
		FirstAddress:   network.String(),
		LastAddress:    last.String(),
		TotalAddresses: total.String(),
		HostCount:      hostCount(bits, prefix.Bits()).String(),
		CIDRs:          []string{},
		Range:          []string{},
	}

	firstUsable, lastUsable := usableRange(bits, network, last)
	result.FirstUsable = firstUsable.String()
	result.LastUsable = lastUsable.String()
	if bits == 32 {
		result.Broadcast = last.String()
	}
	return result, nil
}

// versionOf 把地址位数转成对外的地址族编号。
func versionOf(bits int) int {
	if bits == 32 {
		return 4
	}
	return 6
}

// hostCount 计算可用主机数。
//
// IPv4 的 /31 与 /32 按 RFC 3021 与单主机语义返回 2 与 1。
// IPv6 无广播地址，全部地址可用。
func hostCount(bits, prefix int) *big.Int {
	full := bits - prefix
	one := big.NewInt(1)
	total := new(big.Int).Lsh(one, uint(full))
	switch {
	case full == 0:
		return one
	case bits == 32 && full == 1:
		return big.NewInt(2)
	case bits == 32:
		return total.Sub(total, big.NewInt(2))
	default:
		return total
	}
}

// usableRange 返回首个与末个可用地址。
func usableRange(bits int, network, last netip.Addr) (netip.Addr, netip.Addr) {
	if bits != 32 || network == last {
		return network, last
	}
	next := network.Next()
	prev := last.Prev()
	if !next.IsValid() || !prev.IsValid() || next.Compare(prev) > 0 {
		// /31 与 /32：没有「去掉首尾」的余地，整段都在可用范围内。
		return network, last
	}
	return next, prev
}

// maskString 返回子网掩码的文本形式。
//
// IPv4 用点分十进制；IPv6 用压缩十六进制。
func maskString(prefix netip.Prefix) string {
	bits := bitsOf(prefix.Addr())
	raw := prefix.Masked().Addr().AsSlice()
	mask := make([]byte, len(raw))
	for i := 0; i < bits; i++ {
		if i < prefix.Bits() {
			mask[i/8] |= 1 << (7 - uint(i%8))
		}
	}
	addr, ok := netip.AddrFromSlice(mask)
	if !ok {
		return ""
	}
	return addr.Unmap().String()
}

// lastAddr 返回网段的最后一个地址。
func lastAddr(network netip.Addr, total *big.Int) netip.Addr {
	base := addrToBig(network)
	sum := new(big.Int).Add(base, total)
	sum.Sub(sum, big.NewInt(1))
	return bigToAddr(sum, bitsOf(network))
}

// splitPrefix 按目标前缀长度拆分网段。
func splitPrefix(prefix netip.Prefix, target int) ([]string, error) {
	bits := bitsOf(prefix.Addr())
	if target < prefix.Bits() || target > bits {
		return nil, tools.InvalidField("split", []string{"不小于原前缀且不超过地址位数"}, target)
	}
	count := 1 << uint(target-prefix.Bits())
	if count > maxSplitCIDRs {
		return nil, tooManyItems("split", count)
	}

	out := make([]string, 0, count)
	step := new(big.Int).Lsh(big.NewInt(1), uint(bits-target))
	cur := addrToBig(prefix.Masked().Addr())
	for i := 0; i < count; i++ {
		addr := bigToAddr(cur, bits)
		out = append(out, netip.PrefixFrom(addr, target).String())
		cur = new(big.Int).Add(cur, step)
	}
	return out, nil
}

// rangeToPrefixes 把地址范围反推为最小的 CIDR 集合。
func rangeToPrefixes(first, last netip.Addr) ([]netip.Prefix, error) {
	bits := bitsOf(first)
	cur := addrToBig(first)
	end := addrToBig(last)
	one := big.NewInt(1)

	out := make([]netip.Prefix, 0, 8)
	for cur.Cmp(end) <= 0 {
		// 由当前地址的对齐程度决定的**最长**前缀。
		maxLen := bits
		if cur.Sign() != 0 {
			tz := trailingZeros(cur)
			if tz < maxLen {
				maxLen = bits - tz
			}
		}
		if maxLen < 0 {
			maxLen = 0
		}
		// 逐步缩短前缀，直到该块不越过范围终点。
		for maxLen <= bits {
			size := new(big.Int).Lsh(one, uint(bits-maxLen))
			blockEnd := new(big.Int).Add(cur, size)
			blockEnd.Sub(blockEnd, one)
			if blockEnd.Cmp(end) <= 0 {
				out = append(out, netip.PrefixFrom(bigToAddr(cur, bits), maxLen))
				cur = new(big.Int).Add(cur, size)
				break
			}
			maxLen++
		}
		if len(out) > maxSplitCIDRs {
			return nil, tooManyItems("range", len(out))
		}
	}
	return out, nil
}

// trailingZeros 返回 v 的二进制末尾零位数。
func trailingZeros(v *big.Int) int {
	if v.Sign() == 0 {
		return 128
	}
	count := 0
	tmp := new(big.Int).Set(v)
	for tmp.Bit(0) == 0 {
		tmp.Rsh(tmp, 1)
		count++
	}
	return count
}

// bitsOf 返回地址位数。
func bitsOf(addr netip.Addr) int {
	if addr.Is4() {
		return 32
	}
	return 128
}

// addrToBig 把地址转成大整数。
func addrToBig(addr netip.Addr) *big.Int {
	return new(big.Int).SetBytes(addr.AsSlice())
}

// bigToAddr 把大整数转成地址。
func bigToAddr(v *big.Int, bits int) netip.Addr {
	size := bits / 8
	raw := v.Bytes()
	if len(raw) > size {
		raw = raw[len(raw)-size:]
	}
	buf := make([]byte, size)
	copy(buf[size-len(raw):], raw)
	addr, _ := netip.AddrFromSlice(buf)
	return addr.Unmap()
}

// commonPrefix 返回覆盖 first..last 的最小公共前缀。
func commonPrefix(first, last netip.Addr) netip.Prefix {
	bits := bitsOf(first)
	a, b := first.AsSlice(), last.AsSlice()
	prefix := bits
	for i := 0; i < len(a) && i < len(b); i++ {
		diff := a[i] ^ b[i]
		if diff == 0 {
			continue
		}
		prefix = i * 8
		for bit := 7; bit >= 0; bit-- {
			if diff&(1<<uint(bit)) != 0 {
				break
			}
			prefix++
		}
		break
	}
	return netip.PrefixFrom(first, prefix).Masked()
}
