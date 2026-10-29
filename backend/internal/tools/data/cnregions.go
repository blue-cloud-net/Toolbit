package data

import (
	"context"
	"strings"

	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// CNRegionsRequest 是 cn-regions 工具的请求。
type CNRegionsRequest struct {
	// Query 是名称或编码关键词；缺省返回全部省份。
	Query string `json:"query"`
	// Province 是指定省级行政区的名称或代码；
	// 给出时把城市查询范围限定在该省内。
	Province string `json:"province"`
}

// CNRegionsResult 是 cn-regions 工具的响应。
type CNRegionsResult struct {
	// Provinces 是省级行政区。
	Provinces []Region `json:"provinces"`
	// Cities 是地级 / 县级行政区。
	Cities []Region `json:"cities"`
	// Total 是匹配条数。
	Total int `json:"total"`
	// DataVersion 是数据集版本。
	DataVersion string `json:"dataVersion"`
}

// CNRegions 查询中国行政区划。
//
// 数据集是**离线内置快照**，可能滞后于最新区划调整；
// 版本号随响应返回，便于判断新旧。
func CNRegions(ctx context.Context, req *CNRegionsRequest) (CNRegionsResult, error) {
	if err := req.Validate(); err != nil {
		return CNRegionsResult{}, err
	}

	query := strings.ToLower(strings.TrimSpace(req.Query))
	provinceFilter := strings.ToLower(strings.TrimSpace(req.Province))

	// 缺省（既无关键词也无省份）时**只**返回省级：
	// 全部地级与县级有三千多条，一次返回既慢又没人看得完。
	provincesOnly := query == "" && provinceFilter == ""

	provinces := []Region{}
	cities := []Region{}

	for _, region := range Regions() {
		switch region.Level {
		case LevelProvince:
			if provinceFilter != "" && !matchesProvince(region, provinceFilter) {
				continue
			}
			if query != "" && !matchesRegion(region, query) {
				continue
			}
			provinces = append(provinces, region)

		case LevelCity, LevelDistrict:
			if provincesOnly {
				continue
			}
			if provinceFilter != "" && !provinceMatches(region, provinceFilter) {
				continue
			}
			if query != "" && !matchesRegion(region, query) {
				continue
			}
			cities = append(cities, region)
		}
	}

	return CNRegionsResult{
		Provinces:   provinces,
		Cities:      cities,
		Total:       len(provinces) + len(cities),
		DataVersion: CNRegionsVersion,
	}, nil
}

// Validate 做字段级校验。
func (r *CNRegionsRequest) Validate() error {
	if len(r.Query) > 128 {
		return tools.InvalidField("query", []int{0, 128}, len(r.Query))
	}
	if len(r.Province) > 128 {
		return tools.InvalidField("province", []int{0, 128}, len(r.Province))
	}
	return nil
}

// matchesRegion 报告区划是否匹配关键词。
//
// 匹配范围：名称与六位代码。
func matchesRegion(region Region, query string) bool {
	return strings.Contains(strings.ToLower(region.Name), query) ||
		strings.Contains(region.Code, query)
}

// matchesProvince 报告省级是否匹配给定的名称或代码。
func matchesProvince(region Region, query string) bool {
	return matchesRegion(region, query)
}

// provinceMatches 报告下级区划是否属于指定的省级。
//
// 省级可以按名称或代码给出，因此**不**依赖顺序：
// 先判断是否指向省级本身，再判断下级区划的归属省级。
func provinceMatches(region Region, query string) bool {
	owner, ok := ProvinceByCode(region.Province)
	if !ok {
		return false
	}
	return matchesProvince(owner, query)
}
