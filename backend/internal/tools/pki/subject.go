package pki

import (
	"github.com/blue-cloud-net/Toolbit/backend/internal/pkisvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// SubjectInput 是主体 DN 的输入形态。
//
// 用**具名字段**而不是 map：字段名必须落在白名单内，
// 拼写错误应当被立刻发现，而不是悄悄变成一个未知的 RDN。
type SubjectInput struct {
	// CN 是通用名。
	CN string `json:"CN"`
	// C 是国家代码。
	C string `json:"C"`
	// ST 是省 / 州。
	ST string `json:"ST"`
	// L 是城市。
	L string `json:"L"`
	// O 是组织。
	O string `json:"O"`
	// OU 是组织单位。
	OU string `json:"OU"`
	// SerialNumber 是序列号字段。
	SerialNumber string `json:"serialNumber"`
	// EmailAddress 是邮箱字段。
	EmailAddress string `json:"emailAddress"`
	// Title 是职务。
	Title string `json:"title"`
	// DC 是域组件。
	DC string `json:"DC"`
}

// subjectOrder 是 DN 条目的输出顺序。
//
// 顺序影响 `subjectText` 的呈现（`/C=CN/O=.../CN=...`），
// 因此固定为「从大到小」的惯例顺序，而不是按输入 JSON 的键顺序。
var subjectOrder = []string{
	"C", "ST", "L", "O", "OU", "CN",
	"serialNumber", "emailAddress", "title", "DC",
}

// entries 把输入转成有序的 RDN 条目。
//
// 空值字段被跳过：留空的输入项不应该变成空值的 RDN。
func (s SubjectInput) entries() []pkisvc.NameEntryInfo {
	values := map[string]string{
		"CN":           s.CN,
		"C":            s.C,
		"ST":           s.ST,
		"L":            s.L,
		"O":            s.O,
		"OU":           s.OU,
		"serialNumber": s.SerialNumber,
		"emailAddress": s.EmailAddress,
		"title":        s.Title,
		"DC":           s.DC,
	}

	out := make([]pkisvc.NameEntryInfo, 0, len(subjectOrder))
	for _, field := range subjectOrder {
		value := values[field]
		if value == "" {
			continue
		}
		out = append(out, pkisvc.NameEntryInfo{Field: field, Value: value})
	}
	return out
}

// isEmpty 报告主体是否一个字段都没填。
func (s SubjectInput) isEmpty() bool {
	return len(s.entries()) == 0
}

// KeyGenInput 是密钥生成参数。
type KeyGenInput struct {
	// Type 是密钥算法。
	Type string `json:"type"`
	// Bits 是 RSA 位数。
	Bits int `json:"bits"`
	// Curve 是椭圆曲线名。
	Curve string `json:"curve"`
}

// Validate 校验算法与参数。
func (in KeyGenInput) Validate() error {
	algorithm := in.Type
	if algorithm == "" {
		algorithm = pkisvc.AlgRSA
	}
	if !containsString(pkisvc.KeyAlgorithms(), algorithm) {
		return tools.UnsupportedAlgorithm(algorithm, pkisvc.KeyAlgorithms())
	}
	if algorithm == pkisvc.AlgRSA {
		if in.Bits == 0 {
			return nil
		}
		if !containsInt(pkisvc.RSABits(), in.Bits) {
			return tools.InvalidField("key.bits", pkisvc.RSABits(), in.Bits)
		}
		return nil
	}
	if in.Curve != "" && !containsString(pkisvc.ECCurves(), in.Curve) {
		return tools.UnsupportedAlgorithm(in.Curve, pkisvc.ECCurves())
	}
	return nil
}

// algorithm 返回生效的算法名。
func (in KeyGenInput) algorithm() string {
	if in.Type == "" {
		return pkisvc.AlgRSA
	}
	return in.Type
}

// BasicConstraintsInput 是 basicConstraints 的输入形态。
type BasicConstraintsInput struct {
	// IsCA 报告是否为 CA 证书。
	IsCA bool `json:"isCa"`
	// PathLen 是路径长度约束；负数表示不指定。
	PathLen int `json:"pathLen"`
}

// option 转成服务层参数。
func (b *BasicConstraintsInput) option() *pkisvc.BasicConstraintsOption {
	if b == nil {
		return nil
	}
	return &pkisvc.BasicConstraintsOption{IsCA: b.IsCA, PathLen: b.PathLen}
}

// containsInt 判断整数切片是否包含目标值。
func containsInt(list []int, target int) bool {
	for _, v := range list {
		if v == target {
			return true
		}
	}
	return false
}
