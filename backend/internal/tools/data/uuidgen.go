package data

import (
	"context"
	"strings"

	"github.com/blue-cloud-net/Toolbit/backend/internal/cryptosvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// UUID 版本。
const (
	// UUIDv1 是基于时间的 UUID。
	UUIDv1 = 1
	// UUIDv4 是基于随机数的 UUID。
	UUIDv4 = 4
	// UUIDv7 是时间有序的 UUID。
	UUIDv7 = 7
)

// UUIDVersions 返回受支持的版本（顺序稳定）。
func UUIDVersions() []int {
	return []int{UUIDv1, UUIDv4, UUIDv7}
}

// maxUUIDCount 是单次生成的 UUID 数量上限。
const maxUUIDCount = 1000

// UUIDGenRequest 是 uuid-gen 工具的请求。
type UUIDGenRequest struct {
	// Version 是 UUID 版本：1 / 4 / 7；缺省 4。
	Version int `json:"version"`
	// Count 是生成数量；缺省 1。
	Count int `json:"count"`
	// Uppercase 是否输出大写。
	Uppercase bool `json:"uppercase"`
	// Hyphens 是否保留连字符；缺省 true。
	Hyphens *bool `json:"hyphens"`
	// Braces 是否加花括号。
	Braces bool `json:"braces"`
}

// UUIDGenResult 是 uuid-gen 工具的响应。
type UUIDGenResult struct {
	// UUIDs 是生成的 UUID 列表。
	UUIDs []string `json:"uuids"`
	// Count 是生成的条数。
	Count int `json:"count"`
}

// UUIDGen 生成 UUID。
//
// v1 与 v7 **必须**使用单调时钟，保证同一毫秒内不重复。
func UUIDGen(ctx context.Context, req *UUIDGenRequest) (UUIDGenResult, error) {
	if err := req.Validate(); err != nil {
		return UUIDGenResult{}, err
	}

	count := req.Count
	if count <= 0 {
		count = 1
	}

	out := make([]string, 0, count)
	for i := 0; i < count; i++ {
		id, err := req.newUUID()
		if err != nil {
			return UUIDGenResult{}, err
		}
		out = append(out, req.format(id))
	}

	return UUIDGenResult{
		UUIDs: out,
		Count: len(out),
	}, nil
}

// Validate 做字段级校验。
func (r *UUIDGenRequest) Validate() error {
	switch r.Version {
	case 0, UUIDv1, UUIDv4, UUIDv7:
	default:
		return tools.InvalidField("version", UUIDVersions(), r.Version)
	}
	if r.Count < 0 || r.Count > maxUUIDCount {
		return tools.InvalidField("count", []int{0, maxUUIDCount}, r.Count)
	}
	return nil
}

// newUUID 生成一个 UUID。
//
// 随机位与单调时钟由 cryptosvc 统一提供（Tongsuo 的 RAND_bytes）。
func (r *UUIDGenRequest) newUUID() (string, error) {
	return cryptosvc.NewUUID(r.version())
}

// version 返回版本；缺省 4。
func (r *UUIDGenRequest) version() int {
	if r.Version == 0 {
		return UUIDv4
	}
	return r.Version
}

// format 按请求的排版选项格式化 UUID。
func (r *UUIDGenRequest) format(id string) string {
	text := id
	if r.hyphens() {
		// 保持标准排版。
	} else {
		text = strings.ReplaceAll(text, "-", "")
	}
	if r.Uppercase {
		text = strings.ToUpper(text)
	}
	if r.Braces {
		text = "{" + text + "}"
	}
	return text
}

// hyphens 返回是否保留连字符；缺省 true。
func (r *UUIDGenRequest) hyphens() bool {
	if r.Hyphens == nil {
		return true
	}
	return *r.Hyphens
}
