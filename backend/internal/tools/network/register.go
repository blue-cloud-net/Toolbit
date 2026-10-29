package network

import (
	"context"

	"github.com/blue-cloud-net/Toolbit/backend/internal/netguard"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// Definitions 返回网络类工具的实现。
//
// 全部出网工具都通过**同一个** netguard 守卫建连；守卫为 nil 时
// 所有出网工具都会返回 NETWORK_DISABLED。
//
// 本函数**不**登记以下三个 slug，它们各有专门的处理：
//   - `ssl-check`：只有 WebSocket 流式端点（GET /api/v1/tools/ssl-check/stream）；
//   - `websocket-test`：`transport: client`，没有后端端点；
//   - `ct-search`：本版未实现，分发器返回 NOT_IMPLEMENTED。
func Definitions(g *netguard.Guard) []tools.Definition {
	return []tools.Definition{
		{
			Slug:       "cidr-calc",
			NewRequest: func() tools.Request { return &CIDRCalcRequest{} },
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return CIDRCalc(ctx, req.(*CIDRCalcRequest))
			},
		},
		{
			Slug:       "dns-lookup",
			NewRequest: func() tools.Request { return &DNSLookupRequest{} },
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return DNSLookup(ctx, g, req.(*DNSLookupRequest))
			},
		},
		{
			Slug:       "dns-tester",
			NewRequest: func() tools.Request { return &DNSTesterRequest{} },
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return DNSTester(ctx, g, req.(*DNSTesterRequest))
			},
		},
		{
			Slug:       "cert-chain-fetch",
			NewRequest: func() tools.Request { return &ChainFetchRequest{} },
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return ChainFetch(ctx, g, req.(*ChainFetchRequest))
			},
		},
		{
			Slug:       "domain-verify",
			NewRequest: func() tools.Request { return &DomainVerifyRequest{} },
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return DomainVerify(ctx, g, req.(*DomainVerifyRequest))
			},
		},
		{
			Slug: "device-headers",
			NewRequest: func() tools.Request {
				return &DeviceHeadersRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return DeviceHeaders(ctx, req.(*DeviceHeadersRequest))
			},
		},
		{
			Slug: "ip-range-expand",
			NewRequest: func() tools.Request {
				return &IPRangeRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return IPRange(ctx, req.(*IPRangeRequest))
			},
		},
		{
			Slug: "mime-lookup",
			NewRequest: func() tools.Request {
				return &MimeLookupRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return MimeLookup(ctx, req.(*MimeLookupRequest))
			},
		},
		{
			Slug: "http-status",
			NewRequest: func() tools.Request {
				return &HttpStatusRequest{}
			},
			Handle: func(ctx context.Context, req tools.Request) (tools.Result, error) {
				return HttpStatus(ctx, req.(*HttpStatusRequest))
			},
		},
	}
}
