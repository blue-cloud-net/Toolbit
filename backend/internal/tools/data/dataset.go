// Package data 实现数据类工具。
//
// 本包只做「解码输入 → 校验字段 → 计算 → 编码 DTO」。
//
// 离线数据集（ISO 3166-1、中国行政区划）以 CSV 形式内嵌，
// **不**在运行时出网。
package data

import (
	"bytes"
	"embed"
	"encoding/csv"
	"io"
	"strings"
	"sync"
)

// datasetFS 内嵌离线数据集。
//
//go:embed dataset/*.csv
var datasetFS embed.FS

// 数据版本。
//
// 快照日期决定数据集内容：区划调整后**必须**更新快照与版本号，
// 否则响应里的版本号会误导用户。
const (
	// ISO3166Version 是 ISO 3166-1 快照版本。
	ISO3166Version = "2026-10"
	// CNRegionsVersion 是中国行政区划快照版本。
	CNRegionsVersion = "2026-10"
)

// 国家代码。
type Country struct {
	// Alpha2 是两位字母代码。
	Alpha2 string `json:"alpha2"`
	// Alpha3 是三位字母代码。
	Alpha3 string `json:"alpha3"`
	// Numeric 是三位数字代码。
	Numeric string `json:"numeric"`
	// NameZh 是中文名。
	NameZh string `json:"nameZh"`
	// NameEn 是英文名。
	NameEn string `json:"nameEn"`
}

// 区划层级。
const (
	// LevelProvince 是省级。
	LevelProvince = "province"
	// LevelCity 是地级。
	LevelCity = "city"
	// LevelDistrict 是县级。
	LevelDistrict = "district"
)

// Region 是一个行政区划条目。
type Region struct {
	// Code 是六位行政区划代码。
	Code string `json:"code"`
	// Name 是名称。
	Name string `json:"name"`
	// Level 是层级：province / city / district。
	Level string `json:"level"`
	// Parent 是上级代码；省级为空。
	Parent string `json:"parent,omitempty"`
	// Province 是所属省级代码。
	Province string `json:"province"`
}

// 数据集缓存。
var (
	datasetOnce sync.Once
	countries   []Country
	regions     []Region
	// provinceIndex 按六位代码索引省级条目。
	provinceIndex map[string]Region
)

// loadDatasets 解析并缓存全部数据集。
//
// 解析只做一次：CSV 是内嵌的静态数据，重复解析纯属浪费。
func loadDatasets() {
	datasetOnce.Do(func() {
		countries = parseCountries(readDataset("dataset/iso3166.csv"))
		regions = parseRegions(readDataset("dataset/cn_regions.csv"))
		provinceIndex = make(map[string]Region, 64)
		for _, r := range regions {
			if r.Level == LevelProvince {
				provinceIndex[r.Code] = r
			}
		}
	})
}

// readDataset 读取内嵌数据集文件。
//
// 内嵌文件在编译期就已确定存在；读取失败按空数据集处理，
// 让查询返回空结果而不是让整个进程崩溃。
func readDataset(name string) io.Reader {
	raw, err := datasetFS.ReadFile(name)
	if err != nil {
		return strings.NewReader("")
	}
	return bytes.NewReader(raw)
}

// parseCountries 解析 ISO 3166-1 数据集。
func parseCountries(r io.Reader) []Country {
	out := []Country{}
	for _, rec := range readCSV(r) {
		if len(rec) < 5 {
			continue
		}
		out = append(out, Country{
			Alpha2:  rec[0],
			Alpha3:  rec[1],
			Numeric: rec[2],
			NameZh:  rec[3],
			NameEn:  rec[4],
		})
	}
	return out
}

// parseRegions 解析中国行政区划数据集。
func parseRegions(r io.Reader) []Region {
	out := []Region{}
	for _, rec := range readCSV(r) {
		if len(rec) < 4 {
			continue
		}
		region := Region{
			Code:   rec[0],
			Name:   rec[1],
			Level:  rec[2],
			Parent: rec[3],
		}
		switch region.Level {
		case LevelProvince:
			region.Province = region.Code
		case LevelCity:
			region.Province = region.Parent
		case LevelDistrict:
			region.Province = region.Parent[:2] + "0000"
		}
		out = append(out, region)
	}
	return out
}

// readCSV 读取 CSV 并跳过表头。
func readCSV(r io.Reader) [][]string {
	reader := csv.NewReader(r)
	reader.FieldsPerRecord = -1
	records, err := reader.ReadAll()
	if err != nil || len(records) == 0 {
		return nil
	}
	return records[1:]
}

// Countries 返回全部国家与地区。
func Countries() []Country {
	loadDatasets()
	return countries
}

// Regions 返回全部行政区划。
func Regions() []Region {
	loadDatasets()
	return regions
}

// ProvinceByCode 按六位代码返回省级条目。
//
// 找不到时返回 false。
func ProvinceByCode(code string) (Region, bool) {
	loadDatasets()
	r, ok := provinceIndex[code]
	return r, ok
}
