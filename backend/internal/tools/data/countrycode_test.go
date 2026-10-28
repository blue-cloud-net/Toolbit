package data

import (
	"context"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestCountryCodeAll 断言缺省返回全部。
func TestCountryCodeAll(t *testing.T) {
	result, err := CountryCode(context.Background(), &CountryCodeRequest{})
	if err != nil {
		t.Fatalf("查询失败：%v", err)
	}
	if len(result.Countries) == 0 {
		t.Fatal("没有返回任何国家")
	}
	if result.DataVersion == "" {
		t.Error("缺少数据版本")
	}
}

// TestCountryCodeByName 断言按中文名查询。
func TestCountryCodeByName(t *testing.T) {
	result, err := CountryCode(context.Background(), &CountryCodeRequest{
		Query: "中国",
	})
	if err != nil {
		t.Fatalf("查询失败：%v", err)
	}
	if len(result.Countries) == 0 {
		t.Fatal("没有匹配到国家")
	}
	found := false
	for _, c := range result.Countries {
		if c.Alpha2 == "CN" {
			found = true
			if c.Alpha3 != "CHN" {
				t.Errorf("三位码 = %q，期望 CHN", c.Alpha3)
			}
			if c.Numeric != "156" {
				t.Errorf("数字码 = %q，期望 156", c.Numeric)
			}
		}
	}
	if !found {
		t.Error("没有匹配到 CN")
	}
}

// TestCountryCodeByAlpha 断言按两位码查询。
func TestCountryCodeByAlpha(t *testing.T) {
	result, err := CountryCode(context.Background(), &CountryCodeRequest{
		Query: "de",
	})
	if err != nil {
		t.Fatalf("查询失败：%v", err)
	}
	if len(result.Countries) == 0 {
		t.Fatal("没有匹配到国家")
	}
}

// TestCountryCodeNoMatch 断言无匹配时返回空列表。
func TestCountryCodeNoMatch(t *testing.T) {
	result, err := CountryCode(context.Background(), &CountryCodeRequest{
		Query: "zzzz-not-a-country",
	})
	if err != nil {
		t.Fatalf("查询失败：%v", err)
	}
	if len(result.Countries) != 0 {
		t.Errorf("匹配数 = %d，期望 0", len(result.Countries))
	}
}

// TestCNRegionsProvinces 断言返回全部省份。
func TestCNRegionsProvinces(t *testing.T) {
	result, err := CNRegions(context.Background(), &CNRegionsRequest{})
	if err != nil {
		t.Fatalf("查询失败：%v", err)
	}
	if len(result.Provinces) < 30 {
		t.Errorf("省份数 = %d，期望至少 30", len(result.Provinces))
	}
	if result.DataVersion == "" {
		t.Error("缺少数据版本")
	}
}

// TestCNRegionsByProvince 断言按省份限定范围。
func TestCNRegionsByProvince(t *testing.T) {
	result, err := CNRegions(context.Background(), &CNRegionsRequest{
		Province: "浙江省",
	})
	if err != nil {
		t.Fatalf("查询失败：%v", err)
	}
	if len(result.Provinces) != 1 {
		t.Fatalf("省份数 = %d，期望 1", len(result.Provinces))
	}
	if result.Provinces[0].Code != "330000" {
		t.Errorf("省份代码 = %q，期望 330000", result.Provinces[0].Code)
	}
	for _, city := range result.Cities {
		if city.Province != "330000" {
			t.Errorf("城市 %q 属于 %q，不属于浙江", city.Name, city.Province)
			break
		}
	}
}

// TestCNRegionsByQuery 断言按名称查询。
func TestCNRegionsByQuery(t *testing.T) {
	result, err := CNRegions(context.Background(), &CNRegionsRequest{
		Query: "杭州",
	})
	if err != nil {
		t.Fatalf("查询失败：%v", err)
	}
	if len(result.Cities) == 0 {
		t.Fatal("没有匹配到城市")
	}
	found := false
	for _, city := range result.Cities {
		if city.Name == "杭州市" {
			found = true
		}
	}
	if !found {
		t.Error("没有匹配到杭州市")
	}
}

// TestCNRegionsLevels 断言层级数据完整。
func TestCNRegionsLevels(t *testing.T) {
	levels := map[string]int{}
	for _, r := range Regions() {
		levels[r.Level]++
	}
	if levels[LevelProvince] == 0 {
		t.Error("缺少省级数据")
	}
	if levels[LevelCity] == 0 {
		t.Error("缺少地级数据")
	}
	if levels[LevelDistrict] == 0 {
		t.Error("缺少县级数据")
	}
}

// TestCNRegionsInvalidQuery 断言超长查询报 INVALID_FIELD。
func TestCNRegionsInvalidQuery(t *testing.T) {
	_, err := CNRegions(context.Background(), &CNRegionsRequest{
		Query: string(make([]byte, 200)),
	})
	assertCode(t, err, toolerr.CodeInvalidField)
}

// assertCode 断言错误的错误码。
func assertCode(t *testing.T, err error, want toolerr.Code) {
	t.Helper()
	if err == nil {
		t.Fatalf("期望错误 %s，实际无错误", want)
	}
	toolErr, ok := err.(*toolerr.Error)
	if !ok {
		t.Fatalf("期望 *toolerr.Error，实际 %T", err)
	}
	if toolErr.Code != want {
		t.Errorf("错误码 = %s，期望 %s", toolErr.Code, want)
	}
}
