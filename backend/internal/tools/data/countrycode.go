package data

import (
	"context"
	"strings"

	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// CountryCodeRequest 是 country-code 工具的请求。
type CountryCodeRequest struct {
	// Query 是名称或代码关键词；缺省返回全部。
	Query string `json:"query"`
}

// CountryCodeResult 是 country-code 工具的响应。
type CountryCodeResult struct {
	// Countries 是匹配的国家与地区。
	Countries []Country `json:"countries"`
	// Total 是匹配条数。
	Total int `json:"total"`
	// DataVersion 是数据集版本。
	DataVersion string `json:"dataVersion"`
}

// CountryCode 查询 ISO 3166-1 国家与地区代码。
//
// 数据集是**离线内置**的 ISO 3166-1 快照，**不**出网。
func CountryCode(ctx context.Context, req *CountryCodeRequest) (CountryCodeResult, error) {
	if err := req.Validate(); err != nil {
		return CountryCodeResult{}, err
	}

	query := strings.ToLower(strings.TrimSpace(req.Query))
	matched := []Country{}

	for _, country := range Countries() {
		if query == "" || matchesCountry(country, query) {
			matched = append(matched, country)
		}
	}

	return CountryCodeResult{
		Countries:   matched,
		Total:       len(matched),
		DataVersion: ISO3166Version,
	}, nil
}

// Validate 做字段级校验。
func (r *CountryCodeRequest) Validate() error {
	if len(r.Query) > 128 {
		return tools.InvalidField("query", []int{0, 128}, len(r.Query))
	}
	return nil
}

// matchesCountry 报告国家是否匹配关键词。
//
// 匹配范围：两位码、三位码、数字码、中英文名。
func matchesCountry(country Country, query string) bool {
	candidates := []string{
		strings.ToLower(country.Alpha2),
		strings.ToLower(country.Alpha3),
		strings.ToLower(country.Numeric),
		strings.ToLower(country.NameZh),
		strings.ToLower(country.NameEn),
	}
	for _, c := range candidates {
		if c != "" && strings.Contains(c, query) {
			return true
		}
	}
	return false
}
