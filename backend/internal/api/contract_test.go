package api

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// contractCase 是契约测试的一个用例。
type contractCase struct {
	// Name 是用例名称。
	Name string `json:"name"`
	// Request 是工具请求体。
	Request map[string]any `json:"request"`
	// Expect 是期望结果。
	Expect contractExpect `json:"expect"`
}

// contractExpect 是契约用例的期望。
type contractExpect struct {
	// Code 是期望的 HTTP 状态码（0 表示 200）。
	Code int `json:"code"`
	// ErrorCode 是期望的错误码。
	ErrorCode string `json:"errorCode"`
	// DataContains 是期望出现在 data 里的字段子集。
	DataContains map[string]any `json:"dataContains"`
	// Warnings 是期望出现的警告码。
	Warnings []string `json:"warnings"`
}

// contractFile 是一个工具的契约测试数据。
type contractFile struct {
	// Slug 是工具标识。
	Slug string `json:"slug"`
	// Description 是说明。
	Description string `json:"description"`
	// Cases 是用例列表。
	Cases []contractCase `json:"cases"`
}

// TestContractCases 执行 testdata/contract 下的全部契约用例。
//
// 契约数据是 API 契约的**可执行**表述：请求字段、响应字段与错误码
// 都以它为准。
func TestContractCases(t *testing.T) {
	dir := filepath.Join("..", "..", "testdata", "contract")
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("读取契约目录失败：%v", err)
	}

	ran := 0
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		t.Run(entry.Name(), func(t *testing.T) {
			ran += runContractFile(t, filepath.Join(dir, entry.Name()))
		})
	}
	if ran == 0 {
		t.Fatal("未执行任何契约用例")
	}
}

// runContractFile 执行一个契约文件，返回用例数。
func runContractFile(t *testing.T, path string) int {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("读取契约文件失败：%v", err)
	}
	var file contractFile
	if err := json.Unmarshal(raw, &file); err != nil {
		t.Fatalf("解析契约文件失败：%v", err)
	}
	if file.Slug == "" {
		t.Fatalf("契约文件缺少 slug：%s", path)
	}

	h := newTestServer(t)
	for _, c := range file.Cases {
		t.Run(c.Name, func(t *testing.T) {
			body, err := json.Marshal(c.Request)
			if err != nil {
				t.Fatalf("序列化请求失败：%v", err)
			}
			rec := postJSON(t, h, "/api/v1/tools/"+file.Slug, string(body))

			wantCode := c.Expect.Code
			if wantCode == 0 {
				wantCode = 200
			}
			if rec.Code != wantCode {
				t.Fatalf("状态码 = %d，期望 %d；body=%s", rec.Code, wantCode, rec.Body.String())
			}

			env := decode(t, rec)

			if c.Expect.ErrorCode != "" {
				if env.Error == nil {
					t.Fatalf("期望错误码 %s，实际无错误；body=%s", c.Expect.ErrorCode, rec.Body.String())
				}
				if env.Error.Code != c.Expect.ErrorCode {
					t.Errorf("错误码 = %s，期望 %s", env.Error.Code, c.Expect.ErrorCode)
				}
			}

			if len(c.Expect.DataContains) > 0 {
				if len(env.Data) == 0 {
					t.Fatalf("期望有 data，实际为 null")
				}
				var actual any
				if err := json.Unmarshal(env.Data, &actual); err != nil {
					t.Fatalf("解析 data 失败：%v", err)
				}
				if !jsonContains(actual, c.Expect.DataContains) {
					t.Errorf("data 不满足期望子集 %v；实际 %s", c.Expect.DataContains, rec.Body.String())
				}
			}

			if len(c.Expect.Warnings) > 0 {
				got := map[string]bool{}
				for _, w := range env.Meta.Warnings {
					got[w.Code] = true
				}
				for _, want := range c.Expect.Warnings {
					if !got[want] {
						t.Errorf("缺少警告 %s；实际 %v", want, env.Meta.Warnings)
					}
				}
			}
		})
	}
	return len(file.Cases)
}

// jsonContains 判断 actual 是否包含 expected 的全部字段（递归）。
//
// 数组按**顺序**逐元素比较；标量按相等比较。
func jsonContains(actual, expected any) bool {
	switch want := expected.(type) {
	case map[string]any:
		got, ok := actual.(map[string]any)
		if !ok {
			return false
		}
		for key, wantValue := range want {
			gotValue, present := got[key]
			if !present {
				return false
			}
			if !jsonContains(gotValue, wantValue) {
				return false
			}
		}
		return true
	case []any:
		got, ok := actual.([]any)
		if !ok || len(got) != len(want) {
			return false
		}
		for i := range want {
			if !jsonContains(got[i], want[i]) {
				return false
			}
		}
		return true
	default:
		return reflect.DeepEqual(actual, expected)
	}
}
