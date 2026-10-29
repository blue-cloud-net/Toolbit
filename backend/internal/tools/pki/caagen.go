package pki

import (
	"context"
	"fmt"
	"strings"

	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// CAA 标签。
const (
	// CAATagIssue 授权某个 CA 签发。
	CAATagIssue = "issue"
	// CAATagIssueWild 授权某个 CA 签发通配符证书。
	CAATagIssueWild = "issuewild"
	// CAATagIodef 指定违规上报地址。
	CAATagIodef = "iodef"
)

// CAATags 返回受支持的 CAA 标签（顺序稳定）。
func CAATags() []string {
	return []string{CAATagIssue, CAATagIssueWild, CAATagIodef}
}

// CAAGenRequest 是 CAA 记录生成工具的请求。
type CAAGenRequest struct {
	// Domain 是要保护的域名。
	Domain string `json:"domain"`
	// Records 是 CAA 记录列表。
	Records []CAARecordInput `json:"records"`
	// Ttl 是 TTL 建议值；0 表示不输出。
	Ttl int `json:"ttl"`
}

// CAARecordInput 是一条 CAA 记录的输入形态。
type CAARecordInput struct {
	// Flags 是标志位；缺省 0。
	Flags int `json:"flags"`
	// Tag 是标签（issue / issuewild / iodef）。
	Tag string `json:"tag"`
	// Value 是标签值。
	Value string `json:"value"`
}

// CAARecordOutput 是一条 CAA 记录的输出形态。
type CAARecordOutput struct {
	// Flags 是标志位。
	Flags int `json:"flags"`
	// Tag 是标签。
	Tag string `json:"tag"`
	// Value 是标签值。
	Value string `json:"value"`
	// Zone 是可直接粘贴到区域文件的写法。
	Zone string `json:"zone"`
}

// CAAGenResult 是 CAA 记录生成工具的响应。
type CAAGenResult struct {
	// Domain 是域名。
	Domain string `json:"domain"`
	// Records 是记录列表。
	Records []CAARecordOutput `json:"records"`
	// ZoneLines 是区域文件行（含 TTL）。
	ZoneLines []string `json:"zoneLines"`
	// Count 是记录数量。
	Count int `json:"count"`
}

// Validate 做字段级校验。
func (r *CAAGenRequest) Validate() error {
	if strings.TrimSpace(r.Domain) == "" {
		return tools.MissingField("domain")
	}
	if len(r.Records) == 0 {
		return tools.MissingField("records")
	}
	if len(r.Records) > 32 {
		return tools.InvalidField("records", []int{32}, len(r.Records))
	}
	if r.Ttl < 0 {
		return tools.InvalidField("ttl", []int{0, 2147483647}, r.Ttl)
	}
	for i := range r.Records {
		if !containsString(CAATags(), r.Records[i].Tag) {
			return tools.InvalidField("records.tag", CAATags(), r.Records[i].Tag)
		}
		if r.Records[i].Flags < 0 || r.Records[i].Flags > 255 {
			return tools.InvalidField("records.flags", []int{0, 255}, r.Records[i].Flags)
		}
		if strings.TrimSpace(r.Records[i].Value) == "" {
			return tools.MissingField("records.value")
		}
	}
	return nil
}

// CAAGen 生成 CAA 记录。
//
// 纯字符串拼装，**不进 pkisvc**：不涉及任何密码学运算。
func CAAGen(_ context.Context, req *CAAGenRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	domain := strings.TrimSpace(req.Domain)
	result := &CAAGenResult{
		Domain:    domain,
		Records:   make([]CAARecordOutput, 0, len(req.Records)),
		ZoneLines: make([]string, 0, len(req.Records)),
	}

	for _, record := range req.Records {
		value := strings.TrimSpace(record.Value)
		quoted := fmt.Sprintf(`"%s"`, value)
		result.Records = append(result.Records, CAARecordOutput{
			Flags: record.Flags,
			Tag:   record.Tag,
			Value: value,
			Zone:  fmt.Sprintf("%s. CAA %d %s %s", domain, record.Flags, record.Tag, quoted),
		})
		// 区域文件行带 TTL；没给 TTL 时省略该字段，交给区域默认值。
		if req.Ttl > 0 {
			result.ZoneLines = append(result.ZoneLines, fmt.Sprintf(
				"%s. %d IN CAA %d %s %s", domain, req.Ttl, record.Flags, record.Tag, quoted))
			continue
		}
		result.ZoneLines = append(result.ZoneLines, fmt.Sprintf(
			"%s. IN CAA %d %s %s", domain, record.Flags, record.Tag, quoted))
	}

	result.Count = len(result.Records)
	return tools.Result{Data: result}, nil
}
