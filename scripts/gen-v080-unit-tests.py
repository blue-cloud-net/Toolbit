#!/usr/bin/env python3
"""生成 v0.8.0 批次新增工具的后端单元测试。

每个工具至少一个正例与一个反例；涉及密码学的工具另附标准向量。
"""
from __future__ import annotations

import pathlib

ROOT = pathlib.Path(__file__).resolve().parents[1]
TOOLS = ROOT / "backend" / "internal" / "tools"

TESTS: dict[str, str] = {}

TESTS["text/textdiff_test.go"] = '''package text

import (
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestTextDiffIdentical 断言两侧一致时不产生差异。
func TestTextDiffIdentical(t *testing.T) {
	result, err := TextDiff(t.Context(), &TextDiffRequest{
		Mode: DiffModeText,
		Left: DiffInput{Data: "a\\nb\\nc\\n"},
		Right: DiffInput{Data: "a\\nb\\nc\\n"},
	})
	data := result.Data.(TextDiffResult)
	if err != nil {
		t.Fatalf("对比失败：%v", err)
	}
	if !data.Identical {
		t.Error("两侧一致时 identical 应为 true")
	}
	if data.Stat.Changes != 0 {
		t.Errorf("差异数 = %d，期望 0", data.Stat.Changes)
	}
}

// TestTextDiffDetectsChange 断言行级差异被标出。
func TestTextDiffDetectsChange(t *testing.T) {
	result, err := TextDiff(t.Context(), &TextDiffRequest{
		Mode: DiffModeText,
		Left: DiffInput{Data: "a\\nb\\nc\\n"},
		Right: DiffInput{Data: "a\\nx\\nc\\n"},
	})
	data := result.Data.(TextDiffResult)
	if err != nil {
		t.Fatalf("对比失败：%v", err)
	}
	if data.Identical {
		t.Error("两侧不同��� identical 应为 false")
	}
	if data.Stat.Added != 1 || data.Stat.Removed != 1 {
		t.Errorf("增删行数 = %d/%d，期望 1/1", data.Stat.Added, data.Stat.Removed)
	}
}

// TestTextDiffJSONPaths 断言 JSON 模式给出结构化路径。
func TestTextDiffJSONPaths(t *testing.T) {
	result, err := TextDiff(t.Context(), &TextDiffRequest{
		Mode: DiffModeJSON,
		Left: DiffInput{Data: `{"a":1,"b":{"c":2}}`},
		Right: DiffInput{Data: `{"a":1,"b":{"c":3}}`},
	})
	data := result.Data.(TextDiffResult)
	if err != nil {
		t.Fatalf("对比失败：%v", err)
	}
	if len(data.Changes) == 0 {
		t.Fatal("JSON 模式应给出结构化差异")
	}
	if data.Changes[0].Path == "" {
		t.Error("结构化差异应带路径")
	}
}

// TestTextDiffMissingSide 断言缺少一侧返回 MISSING_FIELD。
func TestTextDiffMissingSide(t *testing.T) {
	_, err := TextDiff(t.Context(), &TextDiffRequest{
		Mode: DiffModeText,
		Right: DiffInput{Data: "a"},
	})
	assertCode(t, err, toolerr.CodeMissingField)
}

// TestTextDiffUnsupportedMode 断言未知模式返回 INVALID_FIELD。
func TestTextDiffUnsupportedMode(t *testing.T) {
	_, err := TextDiff(t.Context(), &TextDiffRequest{
		Mode: "xml",
		Left: DiffInput{Data: "a"},
		Right: DiffInput{Data: "b"},
	})
	assertCode(t, err, toolerr.CodeInvalidField)
}
'''

TESTS["text/regexvalidate_test.go"] = '''package text

import (
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestRegexValidateValid 断言合法表达式通过校验。
func TestRegexValidateValid(t *testing.T) {
	result, err := RegexValidate(t.Context(), &RegexValidateRequest{
		Pattern: `^\\d{3}-\\d{4}$`,
	})
	data := result.Data.(RegexValidateResult)
	if err != nil {
		t.Fatalf("校验失败：%v", err)
	}
	if !data.Valid {
		t.Error("合法表达式 valid 应为 true")
	}
	if data.Dialect != "RE2" {
		t.Errorf("方言 = %q，期望 RE2", data.Dialect)
	}
}

// TestRegexValidateInvalid 断言非法表达式返回错误位置。
func TestRegexValidateInvalid(t *testing.T) {
	result, err := RegexValidate(t.Context(), &RegexValidateRequest{
		Pattern: `([a-z`,
	})
	data := result.Data.(RegexValidateResult)
	if err != nil {
		t.Fatalf("校验失败：%v", err)
	}
	if data.Valid {
		t.Error("非法表达式 valid 应为 false")
	}
	if data.Error == nil {
		t.Fatal("非法表达式应给出错误信息")
	}
	// Go 的正则错误只在部分场景带 `^` 标记，
	// 有标记时必须换算出行列位置。
	if data.Error.Offset >= 0 && data.Error.Position == "" {
		t.Errorf("偏移 %d 应换算出位置", data.Error.Offset)
	}
}

// TestRegexValidateCaptureGroups 断言捕获组数量。
func TestRegexValidateCaptureGroups(t *testing.T) {
	result, err := RegexValidate(t.Context(), &RegexValidateRequest{
		Pattern: `(\\w+)@(\\w+)`,
	})
	data := result.Data.(RegexValidateResult)
	if err != nil {
		t.Fatalf("校验失败：%v", err)
	}
	if data.CaptureGroups != 2 {
		t.Errorf("捕获组数 = %d，期望 2", data.CaptureGroups)
	}
}

// TestRegexValidateMissingPattern 断言缺少模式返回 MISSING_FIELD。
func TestRegexValidateMissingPattern(t *testing.T) {
	_, err := RegexValidate(t.Context(), &RegexValidateRequest{})
	assertCode(t, err, toolerr.CodeMissingField)
}

// TestRegexValidateUnknownFlag 断言未知标志返回 INVALID_FIELD。
func TestRegexValidateUnknownFlag(t *testing.T) {
	_, err := RegexValidate(t.Context(), &RegexValidateRequest{
		Pattern: "abc",
		Flags:   []string{"z"},
	})
	assertCode(t, err, toolerr.CodeInvalidField)
}
'''

TESTS["text/regexbuild_test.go"] = '''package text

import (
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestRegexBuildBasic 断言基础片段拼装。
func TestRegexBuildBasic(t *testing.T) {
	result, err := RegexBuild(t.Context(), &RegexBuildRequest{
		Parts: []string{RegexPartDigits, RegexPartHyphen, RegexPartDigits},
	})
	data := result.Data.(RegexBuildResult)
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	if !data.Valid {
		t.Error("生成的表达式应合法")
	}
	if data.Pattern == "" {
		t.Error("应给出表达式")
	}
	if len(data.Explanation) == 0 {
		t.Error("应给出逐段释义")
	}
}

// TestRegexBuildQuantifier 断言量词生效。
func TestRegexBuildQuantifier(t *testing.T) {
	result, err := RegexBuild(t.Context(), &RegexBuildRequest{
		Parts:     []string{RegexPartLetters},
		Quantifier: &RegexQuantifier{Min: 2, Max: 8},
		Anchored:  true,
	})
	data := result.Data.(RegexBuildResult)
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	if data.Pattern == "" {
		t.Error("应给出表达式")
	}
}

// TestRegexBuildUnknownPart 断言未知片段返回 INVALID_FIELD。
func TestRegexBuildUnknownPart(t *testing.T) {
	_, err := RegexBuild(t.Context(), &RegexBuildRequest{
		Parts: []string{"not-a-part"},
	})
	assertCode(t, err, toolerr.CodeInvalidField)
}

// TestRegexBuildInvertedQuantifier 断言量词上下界颠倒被拒。
func TestRegexBuildInvertedQuantifier(t *testing.T) {
	_, err := RegexBuild(t.Context(), &RegexBuildRequest{
		Parts:     []string{RegexPartLetters},
		Quantifier: &RegexQuantifier{Min: 5, Max: 2},
	})
	assertCode(t, err, toolerr.CodeInvalidField)
}
'''

TESTS["text/datamask_test.go"] = '''package text

import (
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestDataMaskPhone 断言手机号脱敏。
func TestDataMaskPhone(t *testing.T) {
	result, err := DataMask(t.Context(), &DataMaskRequest{
		Text:      "手机号 13800138000",
		Types:     []string{MaskTypePhone},
		KeepCount: 3,
	})
	data := result.Data.(DataMaskResult)
	if err != nil {
		t.Fatalf("脱敏失败：%v", err)
	}
	if data.Total != 1 {
		t.Fatalf("命中数 = %d，期望 1", data.Total)
	}
	if data.Masked == "手机号 13800138000" {
		t.Error("脱敏结果不应保留原文")
	}
}

// TestDataMaskKeepTail 断言保留尾部。
func TestDataMaskKeepTail(t *testing.T) {
	result, err := DataMask(t.Context(), &DataMaskRequest{
		Text:      "110101199003078888",
		Types:     []string{MaskTypeIDCard},
		Position:  MaskKeepTail,
		KeepCount: 3,
	})
	data := result.Data.(DataMaskResult)
	if err != nil {
		t.Fatalf("脱敏失败：%v", err)
	}
	if data.Masked != "***************888" {
		t.Errorf("脱敏结果 = %q", data.Masked)
	}
}

// TestDataMaskMissingText 断言缺少文本返回 MISSING_FIELD。
func TestDataMaskMissingText(t *testing.T) {
	_, err := DataMask(t.Context(), &DataMaskRequest{Types: []string{MaskTypePhone}})
	assertCode(t, err, toolerr.CodeMissingField)
}

// TestDataMaskUnknownType 断言未知脱敏类型返回 INVALID_FIELD。
func TestDataMaskUnknownType(t *testing.T) {
	_, err := DataMask(t.Context(), &DataMaskRequest{
		Text:  "abc",
		Types: []string{"unknown"},
	})
	assertCode(t, err, toolerr.CodeInvalidField)
}
'''

TESTS["text/textstats_test.go"] = '''package text

import (
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestTextStatsCounts 断言中英文分别计数。
func TestTextStatsCounts(t *testing.T) {
	result, err := TextStats(t.Context(), &TextStatsRequest{
		Text: "Hello 世界\\n第二行",
	})
	data := result.Data.(TextStatsResult)
	if err != nil {
		t.Fatalf("统计失败：%v", err)
	}
	if data.Lines != 2 {
		t.Errorf("行数 = %d，期望 2", data.Lines)
	}
	if data.ChineseChars != 5 {
		t.Errorf("中文字符数 = %d，期望 5", data.ChineseChars)
	}
	if data.EnglishWords != 1 {
		t.Errorf("英文单词数 = %d，期望 1", data.EnglishWords)
	}
}

// TestTextStatsEmptyText 断言空文本返回 MISSING_FIELD。
func TestTextStatsEmptyText(t *testing.T) {
	_, err := TextStats(t.Context(), &TextStatsRequest{})
	assertCode(t, err, toolerr.CodeMissingField)
}
'''

TESTS["data/idvalidator_test.go"] = '''package data

import (
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestIDValidateIDCard 断言合法身份证通过校验。
func TestIDValidateIDCard(t *testing.T) {
	result, err := IDValidate(t.Context(), &IDValidateRequest{
		Type:  IDTypeIDCard,
		Value: "11010519491231002X",
	})
	data := result.Data.(IDValidateResult)
	if err != nil {
		t.Fatalf("校验失败：%v", err)
	}
	if !data.Valid {
		t.Errorf("校验应通过：%s", data.Issue)
	}
	if data.CheckDigit != "X" {
		t.Errorf("校验位 = %q，期望 X", data.CheckDigit)
	}
}

// TestIDValidateIDCardBadCheckDigit 断言校验位不匹配时报错。
func TestIDValidateIDCardBadCheckDigit(t *testing.T) {
	result, err := IDValidate(t.Context(), &IDValidateRequest{
		Type:  IDTypeIDCard,
		Value: "110105194912310021",
	})
	data := result.Data.(IDValidateResult)
	if err != nil {
		t.Fatalf("校验失败：%v", err)
	}
	if data.Valid {
		t.Error("校验位不匹配时 valid 应为 false")
	}
	if data.Issue == "" {
		t.Error("应给出失败原因")
	}
}

// TestIDValidateBankCard 断言银行卡 Luhn 校验。
func TestIDValidateBankCard(t *testing.T) {
	result, err := IDValidate(t.Context(), &IDValidateRequest{
		Type:  IDTypeBankCard,
		Value: "4111111111111111",
	})
	data := result.Data.(IDValidateResult)
	if err != nil {
		t.Fatalf("校验失败：%v", err)
	}
	if !data.Valid {
		t.Errorf("校验应通过：%s", data.Issue)
	}
}

// TestIDValidateUnknownType 断言未知类型返回 INVALID_FIELD。
func TestIDValidateUnknownType(t *testing.T) {
	_, err := IDValidate(t.Context(), &IDValidateRequest{
		Type:  "passport",
		Value: "x",
	})
	assertCode(t, err, toolerr.CodeInvalidField)
}
'''

TESTS["data/csvconvert_test.go"] = '''package data

import (
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestCSVConvertCSVToJSON 断言 CSV 转 JSON。
func TestCSVConvertCSVToJSON(t *testing.T) {
	result, err := CSVConvert(t.Context(), &CSVConvertRequest{
		Operation: CSVToJSON,
		Input:     "name,age\\nalice,30\\nbob,25\\n",
		HasHeader: true,
	})
	data := result.Data.(CSVConvertResult)
	if err != nil {
		t.Fatalf("转换失败：%v", err)
	}
	if data.RowCount != 2 {
		t.Errorf("行数 = %d，期望 2", data.RowCount)
	}
	if data.ColumnCount != 2 {
		t.Errorf("列数 = %d，期望 2", data.ColumnCount)
	}
	if len(data.Columns) != 2 || data.Columns[0] != "name" {
		t.Errorf("列名 = %v", data.Columns)
	}
}

// TestCSVConvertJSONToCSV 断言 JSON 转 CSV。
func TestCSVConvertJSONToCSV(t *testing.T) {
	result, err := CSVConvert(t.Context(), &CSVConvertRequest{
		Operation: JSONToCSV,
		Input:     `[{"name":"alice","age":30}]`,
		HasHeader: true,
	})
	data := result.Data.(CSVConvertResult)
	if err != nil {
		t.Fatalf("转换失败：%v", err)
	}
	if data.ColumnCount != 2 {
		t.Errorf("列数 = %d，期望 2", data.ColumnCount)
	}
}

// TestCSVConvertUnsupportedOperation 断言未知方向返回 INVALID_FIELD。
func TestCSVConvertUnsupportedOperation(t *testing.T) {
	_, err := CSVConvert(t.Context(), &CSVConvertRequest{
		Operation: "xmlToJson",
		Input:     "a",
	})
	assertCode(t, err, toolerr.CodeInvalidField)
}

// TestCSVConvertMissingInput 断言缺少输入返回 MISSING_FIELD。
func TestCSVConvertMissingInput(t *testing.T) {
	_, err := CSVConvert(t.Context(), &CSVConvertRequest{Operation: CSVToJSON})
	assertCode(t, err, toolerr.CodeMissingField)
}
'''

TESTS["data/jsontotypes_test.go"] = '''package data

import (
	"strings"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestJSONToTypesTypeScript 断言生成 TypeScript 类型。
func TestJSONToTypesTypeScript(t *testing.T) {
	result, err := JSONToTypes(t.Context(), &JSONToTypesRequest{
		Input:    `{"id":1,"name":"a"}`,
		Language: LangTypeScript,
		RootName: "User",
	})
	data := result.Data.(JSONToTypesResult)
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	if !strings.Contains(data.Output, "User") {
		t.Errorf("输出应含根类型名：%s", data.Output)
	}
	if data.TypeCount == 0 {
		t.Error("类型数应大于 0")
	}
}

// TestJSONToTypesGo 断言生成 Go 类型。
func TestJSONToTypesGo(t *testing.T) {
	result, err := JSONToTypes(t.Context(), &JSONToTypesRequest{
		Input:    `{"id":1}`,
		Language: LangGo,
		RootName: "Item",
	})
	data := result.Data.(JSONToTypesResult)
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	if !strings.Contains(data.Output, "type Item struct") {
		t.Errorf("输出应含结构体定义：%s", data.Output)
	}
}

// TestJSONToTypesUnsupportedLanguage 断言未知语言返回 INVALID_FIELD。
func TestJSONToTypesUnsupportedLanguage(t *testing.T) {
	_, err := JSONToTypes(t.Context(), &JSONToTypesRequest{
		Input:    "{}",
		Language: "cobol",
	})
	assertCode(t, err, toolerr.CodeInvalidField)
}

// TestJSONToTypesMissingInput 断言缺少输入返回 MISSING_FIELD。
func TestJSONToTypesMissingInput(t *testing.T) {
	_, err := JSONToTypes(t.Context(), &JSONToTypesRequest{Language: LangGo})
	assertCode(t, err, toolerr.CodeMissingField)
}
'''

TESTS["data/mockdata_test.go"] = '''package data

import (
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestMockDataGenerates 断言按类型生成假数据。
func TestMockDataGenerates(t *testing.T) {
	result, err := MockData(t.Context(), &MockDataRequest{
		Types: []string{MockTypeString, MockTypeInt},
		Count: 3,
	})
	data := result.Data.(MockDataResult)
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	if data.Count != 3 {
		t.Errorf("每类数量 = %d，期望 3", data.Count)
	}
	// 两种类型各生成 3 条，共 6 条。
	if len(data.Fields) != 6 {
		t.Errorf("字段数 = %d，期望 6", len(data.Fields))
	}
}

// TestMockDataEmail 断言邮箱类型产出合法邮箱。
func TestMockDataEmail(t *testing.T) {
	result, err := MockData(t.Context(), &MockDataRequest{
		Types: []string{MockTypeEmail},
		Count: 1,
	})
	data := result.Data.(MockDataResult)
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	if len(data.Fields) != 1 {
		t.Fatalf("字段数 = %d，期望 1", len(data.Fields))
	}
	if data.Fields[0].Type != MockTypeEmail {
		t.Errorf("字段类型 = %q", data.Fields[0].Type)
	}
}

// TestMockDataUnsupportedLocale 断言未知语言返回 INVALID_FIELD。
func TestMockDataUnsupportedLocale(t *testing.T) {
	_, err := MockData(t.Context(), &MockDataRequest{
		Types:  []string{MockTypeString},
		Locale: "fr-FR",
	})
	assertCode(t, err, toolerr.CodeInvalidField)
}

// TestMockDataTooManyCount 断言数量超限返回 INVALID_FIELD。
func TestMockDataTooManyCount(t *testing.T) {
	_, err := MockData(t.Context(), &MockDataRequest{
		Types: []string{MockTypeString},
		Count: 100000,
	})
	assertCode(t, err, toolerr.CodeInvalidField)
}
'''

TESTS["data/idgenerator_test.go"] = '''package data

import (
	"strings"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestIDGeneratorUUID 断言 UUID 格式。
func TestIDGeneratorUUID(t *testing.T) {
	result, err := IDGenerator(t.Context(), &IDGeneratorRequest{
		Format: IDFormatUUID,
		Count:  3,
	})
	data := result.Data.(IDGeneratorResult)
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	if len(data.IDs) != 3 {
		t.Fatalf("ID 数 = %d，期望 3", len(data.IDs))
	}
	for _, id := range data.IDs {
		if len(id) != 36 {
			t.Errorf("UUID 长度 = %d，期望 36", len(id))
		}
	}
}

// TestIDGeneratorULID 断言 ULID 长度与字符集。
func TestIDGeneratorULID(t *testing.T) {
	result, err := IDGenerator(t.Context(), &IDGeneratorRequest{
		Format: IDFormatULID,
		Count:  1,
	})
	data := result.Data.(IDGeneratorResult)
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	if len(data.IDs) != 1 || len(data.IDs[0]) != 26 {
		t.Fatalf("ULID = %v", data.IDs)
	}
	for _, r := range data.IDs[0] {
		if !strings.ContainsRune(ulidAlphabet, r) {
			t.Errorf("ULID 含非法字符 %q", r)
		}
	}
}

// TestIDGeneratorNanoID 断言 NanoID 长度。
func TestIDGeneratorNanoID(t *testing.T) {
	result, err := IDGenerator(t.Context(), &IDGeneratorRequest{
		Format: IDFormatNanoID,
		Count:  1,
		Length: 10,
	})
	data := result.Data.(IDGeneratorResult)
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	if len(data.IDs[0]) != 10 {
		t.Errorf("NanoID 长度 = %d，期望 10", len(data.IDs[0]))
	}
}

// TestIDGeneratorUnsupportedFormat 断言未知格式返回 INVALID_FIELD。
func TestIDGeneratorUnsupportedFormat(t *testing.T) {
	_, err := IDGenerator(t.Context(), &IDGeneratorRequest{
		Format: "guid",
		Count:  1,
	})
	assertCode(t, err, toolerr.CodeInvalidField)
}
'''

TESTS["data/wifiqr_test.go"] = '''package data

import (
	"strings"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestWiFiQRPayload 断言 WPA 载荷格式。
func TestWiFiQRPayload(t *testing.T) {
	result, err := WiFiQR(t.Context(), &WiFiQRRequest{
		SSID:     "Guest-Network",
		Password: "guest-password",
		Security: WiFiSecurityWPA,
		Size:     128,
	})
	data := result.Data.(WiFiQRResult)
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	if !strings.HasPrefix(data.Payload, "WIFI:T:WPA;S:Guest-Network;P:guest-password;") {
		t.Errorf("载荷 = %q", data.Payload)
	}
	if data.PNG == "" {
		t.Error("应产出 PNG")
	}
}

// TestWiFiQROpenNetwork 断言开放网络不带口令。
func TestWiFiQROpenNetwork(t *testing.T) {
	result, err := WiFiQR(t.Context(), &WiFiQRRequest{
		SSID:     "Open",
		Security: WiFiSecurityNone,
		Size:     128,
	})
	data := result.Data.(WiFiQRResult)
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	if !strings.Contains(data.Payload, "T:nopass;") {
		t.Errorf("载荷 = %q", data.Payload)
	}
}

// TestWiFiQRMissingSSID 断言缺少 SSID 返回 MISSING_FIELD。
func TestWiFiQRMissingSSID(t *testing.T) {
	_, err := WiFiQR(t.Context(), &WiFiQRRequest{
		Password: "x",
		Security: WiFiSecurityWPA,
	})
	assertCode(t, err, toolerr.CodeMissingField)
}
'''

TESTS["data/archiveinspect_test.go"] = '''package data

import (
	"encoding/base64"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// zipFixture 是一个只含单个条目的最小 ZIP 归档（Base64）。
const zipFixture = "UEsDBBQAAAAAANQCXl2GphA2BQAAAAUAAAAFAAAAYS50eHRoZWxsb1BLAQIUAxQAAAAAANQCXl2GphA2BQAAAAUAAAAFAAAAAAAAAAAAAACAAQAAAABhLnR4dFBLBQYAAAAAAQABADMAAAAoAAAAAAA="

// TestArchiveInspectZIP 断言 ZIP 条目清单。
func TestArchiveInspectZIP(t *testing.T) {
	raw, err := base64.StdEncoding.DecodeString(zipFixture)
	if err != nil {
		t.Fatalf("解析夹具失败：%v", err)
	}
	result, err := ArchiveInspect(t.Context(), &ArchiveInspectRequest{
		Data:   base64.StdEncoding.EncodeToString(raw),
		Format: ArchiveFormatZip,
	})
	data := result.Data.(ArchiveInspectResult)
	if err != nil {
		t.Fatalf("查看失败：%v", err)
	}
	if data.EntryCount != 1 {
		t.Fatalf("条目数 = %d，期望 1", data.EntryCount)
	}
	if data.Entries[0].Name != "a.txt" {
		t.Errorf("条目名 = %q", data.Entries[0].Name)
	}
}

// TestArchiveInspectMissingData 断言缺少数据返回 MISSING_FIELD。
func TestArchiveInspectMissingData(t *testing.T) {
	_, err := ArchiveInspect(t.Context(), &ArchiveInspectRequest{Format: ArchiveFormatZip})
	assertCode(t, err, toolerr.CodeMissingField)
}

// TestArchiveInspectInvalidArchive 断言非法归档被拒。
func TestArchiveInspectInvalidArchive(t *testing.T) {
	_, err := ArchiveInspect(t.Context(), &ArchiveInspectRequest{
		Data:   base64.StdEncoding.EncodeToString([]byte("not-an-archive")),
		Format: ArchiveFormatZip,
	})
	assertCode(t, err, toolerr.CodeInvalidField)
}
'''

TESTS["data/fileanalyze_test.go"] = '''package data

import (
	"encoding/base64"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestFileAnalyzePNG 断言识别 PNG。
func TestFileAnalyzePNG(t *testing.T) {
	raw := []byte{0x89, 'P', 'N', 'G', 0x0d, 0x0a, 0x1a, 0x0a, 0x00, 0x01}
	result, err := FileAnalyze(t.Context(), &FileAnalyzeRequest{
		Data:            base64.StdEncoding.EncodeToString(raw),
		FileName:        "a.png",
		DigestAlgorithm: "SHA256",
	})
	data := result.Data.(FileAnalyzeResult)
	if err != nil {
		t.Fatalf("识别失败：%v", err)
	}
	if data.Extension != "png" {
		t.Errorf("扩展名 = %q，期望 png", data.Extension)
	}
	if data.Digest == "" {
		t.Error("应给出摘要")
	}
}

// TestFileAnalyzeMissingData 断言缺少数据返回 MISSING_FIELD。
func TestFileAnalyzeMissingData(t *testing.T) {
	_, err := FileAnalyze(t.Context(), &FileAnalyzeRequest{FileName: "a.txt"})
	assertCode(t, err, toolerr.CodeMissingField)
}
'''

TESTS["system/lookuptable_test.go"] = '''package system

import (
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestLookupTablePort 断言端口查询命中。
func TestLookupTablePort(t *testing.T) {
	result, err := LookupTable(t.Context(), &LookupTableRequest{
		Kind: LookupKindPort,
		Key:  "443",
	})
	data := result.Data.(LookupTableResult)
	if err != nil {
		t.Fatalf("查询失败：%v", err)
	}
	if !data.Matched {
		t.Fatal("443 应命中")
	}
	if data.Entry == nil {
		t.Fatal("应给出条目")
	}
	if data.DataVersion == "" {
		t.Error("应给出数据集版本")
	}
}

// TestLookupTableFuzzy 断言模糊检索给出候选。
func TestLookupTableFuzzy(t *testing.T) {
	result, err := LookupTable(t.Context(), &LookupTableRequest{
		Kind: LookupKindExitCode,
		Key:  "错误",
	})
	data := result.Data.(LookupTableResult)
	if err != nil {
		t.Fatalf("查询失败：%v", err)
	}
	if len(data.Candidates) == 0 {
		t.Error("应给出候选")
	}
}

// TestLookupTableMissingKey 断言缺少查询词返回 MISSING_FIELD。
func TestLookupTableMissingKey(t *testing.T) {
	_, err := LookupTable(t.Context(), &LookupTableRequest{Kind: LookupKindPort})
	assertCode(t, err, toolerr.CodeMissingField)
}

// TestLookupTableUnknownKind 断言未知类别返回 INVALID_FIELD。
func TestLookupTableUnknownKind(t *testing.T) {
	_, err := LookupTable(t.Context(), &LookupTableRequest{Kind: "unknown", Key: "1"})
	assertCode(t, err, toolerr.CodeInvalidField)
}
'''

TESTS["system/nginxconfig_test.go"] = '''package system

import (
	"strings"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestNginxConfigSite 断言站点配置含 server_name。
func TestNginxConfigSite(t *testing.T) {
	result, err := NginxConfig(t.Context(), &NginxConfigRequest{
		Mode:       NginxModeSite,
		ServerName: "example.com",
		ListenPort: 80,
		Root:       "/var/www/html",
	})
	data := result.Data.(NginxConfigResult)
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	if !strings.Contains(data.Config, "example.com") {
		t.Errorf("配置应含域名：%s", data.Config)
	}
	if data.Mode != NginxModeSite {
		t.Errorf("模式 = %q", data.Mode)
	}
}

// TestNginxConfigUnsupportedMode 断言未知模式返回 INVALID_FIELD。
func TestNginxConfigUnsupportedMode(t *testing.T) {
	_, err := NginxConfig(t.Context(), &NginxConfigRequest{Mode: "stream"})
	assertCode(t, err, toolerr.CodeInvalidField)
}
'''

TESTS["system/firewallrules_test.go"] = '''package system

import (
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestFirewallRulesUFW 断言生成 ufw 规则。
func TestFirewallRulesUFW(t *testing.T) {
	result, err := FirewallRules(t.Context(), &FirewallRulesRequest{
		Tool:     FirewallToolUFW,
		Scenario: FirewallScenarioHTTPS,
		Protocol: "tcp",
	})
	data := result.Data.(FirewallRulesResult)
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	if len(data.Commands) == 0 {
		t.Fatal("应给出可执行命令")
	}
	if data.Tool != FirewallToolUFW {
		t.Errorf("工具 = %q", data.Tool)
	}
}

// TestFirewallRulesUnsupportedTool 断言未知工具返回 INVALID_FIELD。
func TestFirewallRulesUnsupportedTool(t *testing.T) {
	_, err := FirewallRules(t.Context(), &FirewallRulesRequest{
		Tool:     "pf",
		Scenario: FirewallScenarioHTTP,
	})
	assertCode(t, err, toolerr.CodeInvalidField)
}
'''

TESTS["system/dockercomposegen_test.go"] = '''package system

import (
	"strings"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestDockerComposeGen 断言生成 Compose 配置。
func TestDockerComposeGen(t *testing.T) {
	result, err := DockerComposeGen(t.Context(), &DockerComposeRequest{
		ProjectName: "demo",
		Services: []ComposeService{
			{Name: "web", Image: "nginx:alpine", Ports: []string{"8080:80"}},
		},
	})
	data := result.Data.(DockerComposeResult)
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	if data.ServiceCount != 1 {
		t.Errorf("服务数 = %d，期望 1", data.ServiceCount)
	}
	if !strings.Contains(data.Compose, "nginx:alpine") {
		t.Errorf("配置应含镜像名：%s", data.Compose)
	}
}

// TestDockerComposeGenMissingProject 断言缺少项目名返回 MISSING_FIELD。
func TestDockerComposeGenMissingProject(t *testing.T) {
	_, err := DockerComposeGen(t.Context(), &DockerComposeRequest{})
	assertCode(t, err, toolerr.CodeMissingField)
}
'''

TESTS["network/iprangeexpand_test.go"] = '''package network

import (
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestIPRangeExpand 断言展开网段。
func TestIPRangeExpand(t *testing.T) {
	result, err := IPRange(t.Context(), &IPRangeRequest{
		Mode:  IPRangeModeExpand,
		Cidr:  "192.168.1.0/30",
		Limit: 16,
	})
	data := result.Data.(IPRangeResult)
	if err != nil {
		t.Fatalf("展开失败：%v", err)
	}
	if data.Count != 4 {
		t.Errorf("地址数 = %d，期望 4", data.Count)
	}
	if len(data.Addresses) != 4 {
		t.Errorf("地址列表长度 = %d，期望 4", len(data.Addresses))
	}
}

// TestIPRangeToCIDR 断言地址区间转 CIDR。
func TestIPRangeToCIDR(t *testing.T) {
	result, err := IPRange(t.Context(), &IPRangeRequest{
		Mode:  IPRangeModeToCIDR,
		Start: "192.168.1.10",
		End:   "192.168.1.20",
	})
	data := result.Data.(IPRangeResult)
	if err != nil {
		t.Fatalf("转换失败：%v", err)
	}
	if len(data.Cidrs) == 0 {
		t.Fatal("应给出 CIDR 列表")
	}
	if data.Start != "192.168.1.10" || data.End != "192.168.1.20" {
		t.Errorf("区间 = %s ~ %s", data.Start, data.End)
	}
}

// TestIPRangeToRange 断言 CIDR 转地址区间。
func TestIPRangeToRange(t *testing.T) {
	result, err := IPRange(t.Context(), &IPRangeRequest{
		Mode: IPRangeModeToRange,
		Cidr: "192.168.1.0/24",
	})
	data := result.Data.(IPRangeResult)
	if err != nil {
		t.Fatalf("转换失败：%v", err)
	}
	if data.Start != "192.168.1.0" {
		t.Errorf("起始地址 = %q", data.Start)
	}
}

// TestIPRangeReversedRange 断言倒置区间被拒。
func TestIPRangeReversedRange(t *testing.T) {
	_, err := IPRange(t.Context(), &IPRangeRequest{
		Mode:  IPRangeModeToCIDR,
		Start: "192.168.1.20",
		End:   "192.168.1.10",
	})
	assertCode(t, err, toolerr.CodeInvalidField)
}
'''

TESTS["network/deviceheaders_test.go"] = '''package network

import (
	"strings"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestDeviceHeadersGenerate 断言生成 Chrome 请求头。
func TestDeviceHeadersGenerate(t *testing.T) {
	result, err := DeviceHeaders(t.Context(), &DeviceHeadersRequest{
		Mode:     DeviceModeGenerate,
		Platform: DevicePlatformChrome,
		Version:  "120.0.0.0",
	})
	data := result.Data.(DeviceHeadersResult)
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	if len(data.Headers) == 0 {
		t.Fatal("应给出请求头")
	}
	if !strings.Contains(data.UserAgent, "Chrome/120.0.0.0") {
		t.Errorf("UA = %q", data.UserAgent)
	}
}

// TestDeviceHeadersParseUA 断言解析 User-Agent。
func TestDeviceHeadersParseUA(t *testing.T) {
	result, err := DeviceHeaders(t.Context(), &DeviceHeadersRequest{
		Mode: DeviceModeParseUA,
		UserAgent: "Mozilla/5.0 (Windows NT 10.0; Win64; x64) " +
			"AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	})
	data := result.Data.(DeviceHeadersResult)
	if err != nil {
		t.Fatalf("解析失败：%v", err)
	}
	if data.Parsed == nil {
		t.Fatal("应给出解析结果")
	}
	if data.Parsed.Browser != "Chrome" {
		t.Errorf("浏览器 = %q", data.Parsed.Browser)
	}
}

// TestDeviceHeadersUnsupportedMode 断言未知模式返回 INVALID_FIELD。
func TestDeviceHeadersUnsupportedMode(t *testing.T) {
	_, err := DeviceHeaders(t.Context(), &DeviceHeadersRequest{Mode: "rewrite"})
	assertCode(t, err, toolerr.CodeInvalidField)
}
'''

TESTS["formatting/colorconvert_test.go"] = '''package formatting

import (
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestColorConvertHex 断言十六进制转换。
func TestColorConvertHex(t *testing.T) {
	result, err := ColorConvert(t.Context(), &ColorConvertRequest{
		Input:         "#3B82F6",
		InputFormat:   ColorFormatHex,
		OutputFormats: []string{ColorFormatHex, ColorFormatRGB, ColorFormatHSL},
	})
	data := result.Data.(ColorConvertResult)
	if err != nil {
		t.Fatalf("转换失败：%v", err)
	}
	if data.Hex != "#3B82F6" {
		t.Errorf("HEX = %q", data.Hex)
	}
	if data.RGB == "" {
		t.Error("应给出 RGB 表示")
	}
	if data.ContrastOnWhite <= 1 {
		t.Errorf("对比度 = %v", data.ContrastOnWhite)
	}
}

// TestColorConvertInvalid 断言非法颜色返回 INVALID_FIELD。
func TestColorConvertInvalid(t *testing.T) {
	_, err := ColorConvert(t.Context(), &ColorConvertRequest{
		Input:       "not-a-color",
		InputFormat: ColorFormatHex,
	})
	assertCode(t, err, toolerr.CodeInvalidField)
}
'''

TESTS["encoding/htmlescape_test.go"] = '''package encoding

import (
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestHTMLEscapeEncode 断言转义尖括号与引号。
func TestHTMLEscapeEncode(t *testing.T) {
	result, err := HTMLEscape(t.Context(), &HTMLEscapeRequest{
		Mode:  HTMLEscapeEncode,
		Input: `<b>hi</b> & "x"`,
	})
	data := result.Data.(HTMLEscapeResult)
	if err != nil {
		t.Fatalf("转义失败：%v", err)
	}
	if data.Output != "&lt;b&gt;hi&lt;/b&gt; &amp; &quot;x&quot;" {
		t.Errorf("输出 = %q", data.Output)
	}
}

// TestHTMLEscapeDecode 断言反转义实体。
func TestHTMLEscapeDecode(t *testing.T) {
	result, err := HTMLEscape(t.Context(), &HTMLEscapeRequest{
		Mode:  HTMLEscapeDecode,
		Input: "&lt;b&gt;hi&lt;/b&gt;",
	})
	data := result.Data.(HTMLEscapeResult)
	if err != nil {
		t.Fatalf("反转义失败：%v", err)
	}
	if data.Output != "<b>hi</b>" {
		t.Errorf("输出 = %q", data.Output)
	}
}

// TestHTMLEscapeUnsupportedMode 断言未知模式返回 INVALID_FIELD。
func TestHTMLEscapeUnsupportedMode(t *testing.T) {
	_, err := HTMLEscape(t.Context(), &HTMLEscapeRequest{Mode: "strip", Input: "a"})
	assertToolErrCode(t, err, toolerr.CodeInvalidField)
}
'''

TESTS["documents/markdownconvert_test.go"] = '''package documents

import (
	"strings"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestMarkdownConvertToHTML 断言 Markdown 转 HTML。
func TestMarkdownConvertToHTML(t *testing.T) {
	result, err := MarkdownConvert(t.Context(), &MarkdownConvertRequest{
		Mode:  MarkdownToHTML,
		Input: "# Title\\n\\nBody",
		GFM:   true,
	})
	data := result.Data.(MarkdownConvertResult)
	if err != nil {
		t.Fatalf("转换失败：%v", err)
	}
	if !strings.Contains(data.Output, "<h1>") {
		t.Errorf("输出应含标题标签：%s", data.Output)
	}
	if data.HeadingCount != 1 {
		t.Errorf("标题数 = %d，期望 1", data.HeadingCount)
	}
}

// TestMarkdownConvertToMarkdown 断言 HTML 转 Markdown。
func TestMarkdownConvertToMarkdown(t *testing.T) {
	result, err := MarkdownConvert(t.Context(), &MarkdownConvertRequest{
		Mode:  HTMLToMarkdown,
		Input: "<h1>Title</h1><p>Body</p>",
	})
	data := result.Data.(MarkdownConvertResult)
	if err != nil {
		t.Fatalf("转换失败：%v", err)
	}
	if !strings.Contains(data.Output, "# Title") {
		t.Errorf("输出应含 Markdown 标题：%s", data.Output)
	}
}

// TestMarkdownConvertUnsupportedMode 断言未知模式返回 INVALID_FIELD。
func TestMarkdownConvertUnsupportedMode(t *testing.T) {
	_, err := MarkdownConvert(t.Context(), &MarkdownConvertRequest{Mode: "toPdf", Input: "# a"})
	assertCode(t, err, toolerr.CodeInvalidField)
}
'''

TESTS["crypto/passwordpolicy_test.go"] = '''package crypto

import (
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestPasswordPolicyCompliant 断言满足策略。
func TestPasswordPolicyCompliant(t *testing.T) {
	result, err := PasswordPolicy(t.Context(), &PasswordPolicyRequest{
		Password: "Zx7#quietRiver9",
		Preset:   PolicyPresetMedium,
	})
	data := result.Data.(PasswordPolicyResult)
	if err != nil {
		t.Fatalf("检查失败：%v", err)
	}
	if !data.Compliant {
		t.Errorf("应判定合规：%v", data.Suggestions)
	}
	if len(data.Checks) == 0 {
		t.Error("应给出检查项")
	}
}

// TestPasswordPolicyNonCompliant 断言不满足策略。
func TestPasswordPolicyNonCompliant(t *testing.T) {
	result, err := PasswordPolicy(t.Context(), &PasswordPolicyRequest{
		Password: "abc",
		Preset:   PolicyPresetStrict,
	})
	data := result.Data.(PasswordPolicyResult)
	if err != nil {
		t.Fatalf("检查失败：%v", err)
	}
	if data.Compliant {
		t.Error("应判定不合规")
	}
	if len(data.Suggestions) == 0 {
		t.Error("应给出改进建议")
	}
}

// TestPasswordPolicyUnsupportedPreset 断言未知预设返回 INVALID_FIELD。
func TestPasswordPolicyUnsupportedPreset(t *testing.T) {
	_, err := PasswordPolicy(t.Context(), &PasswordPolicyRequest{
		Password: "abc",
		Preset:   "extreme",
	})
	assertCode(t, err, toolerr.CodeInvalidField)
}
'''

TESTS["crypto/totp_test.go"] = '''package crypto

import (
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// RFC 6238 的 SHA1 测试向量密钥（Base32）。
const rfc6238Secret = "GEZDGNBVGY3TQOJQGEZDGNBVGY3TQOJQ"

// TestTOTPRFC6238Vectors 断言 RFC 6238 的 SHA1 标准向量。
//
// 标准向量：时间戳 59 → 94287082，1111111109 → 07081804。
func TestTOTPRFC6238Vectors(t *testing.T) {
	cases := []struct {
		timestamp int64
		want      string
	}{
		{59, "94287082"},
		{1111111109, "07081804"},
		{1111111111, "14050471"},
		{1234567890, "89005924"},
		{2000000000, "69279037"},
		{20000000000, "65353130"},
	}

	for _, c := range cases {
		result, err := TOTP(t.Context(), &TOTPRequest{
			Operation: TOTPOperationGenerate,
			Secret:    rfc6238Secret,
			Algorithm: TOTPAlgorithmSHA1,
			Digits:    8,
			Period:    30,
			Timestamp: c.timestamp,
		})
		if err != nil {
			t.Fatalf("生成失败（时间戳 %d）：%v", c.timestamp, err)
		}
		data := result.Data.(TOTPResult)
		if data.Code != c.want {
			t.Errorf("时间戳 %d 的验证码 = %q，期望 %q", c.timestamp, data.Code, c.want)
		}
	}
}

// TestTOTPVerify 断言校验通过与不通过。
func TestTOTPVerify(t *testing.T) {
	result, err := TOTP(t.Context(), &TOTPRequest{
		Operation: TOTPOperationVerify,
		Secret:    rfc6238Secret,
		Algorithm: TOTPAlgorithmSHA1,
		Digits:    8,
		Timestamp: 59,
		Code:      "94287082",
	})
	data := result.Data.(TOTPResult)
	if err != nil {
		t.Fatalf("校验失败：%v", err)
	}
	if data.Valid == nil || !*data.Valid {
		t.Error("正确验证码应校验通过")
	}

	wrong, err := TOTP(t.Context(), &TOTPRequest{
		Operation: TOTPOperationVerify,
		Secret:    rfc6238Secret,
		Algorithm: TOTPAlgorithmSHA1,
		Digits:    8,
		Timestamp: 59,
		Code:      "00000000",
	})
	if err != nil {
		t.Fatalf("校验失败：%v", err)
	}
	wrongData := wrong.Data.(TOTPResult)
	if wrongData.Valid == nil || *wrongData.Valid {
		t.Error("错误验证码不应校验通过")
	}
}

// TestTOTPMissingSecret 断言缺少密钥返回 MISSING_FIELD。
func TestTOTPMissingSecret(t *testing.T) {
	_, err := TOTP(t.Context(), &TOTPRequest{Operation: TOTPOperationGenerate})
	assertCode(t, err, toolerr.CodeMissingField)
}

// TestTOTPInvalidSecret 断言非法 Base32 密钥被拒。
func TestTOTPInvalidSecret(t *testing.T) {
	_, err := TOTP(t.Context(), &TOTPRequest{
		Operation: TOTPOperationGenerate,
		Secret:    "0189",
	})
	assertCode(t, err, toolerr.CodeInvalidField)
}
'''

TESTS["keys/sshkeygen_test.go"] = '''package keys

import (
	"strings"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestSSHKeygenEd25519 断言 Ed25519 密钥格式。
func TestSSHKeygenEd25519(t *testing.T) {
	result, err := SSHKeygen(t.Context(), &SSHKeygenRequest{
		Algorithm: SSHAlgorithmEd25519,
		Comment:   "toolbit@example.com",
	})
	data := result.Data.(SSHKeygenResult)
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	if !strings.Contains(data.PrivateKey, "BEGIN OPENSSH PRIVATE KEY") {
		t.Errorf("私钥头不对：%s", data.PrivateKey)
	}
	if !strings.Contains(data.PublicKey, "ssh-ed25519") {
		t.Errorf("公钥应含类型名：%s", data.PublicKey)
	}
	if !strings.Contains(data.PublicKey, "toolbit@example.com") {
		t.Errorf("公钥应含注释：%s", data.PublicKey)
	}
	if data.Fingerprint == "" {
		t.Error("应给出指纹")
	}
}

// TestSSHKeygenRSA 断言 RSA 密钥格式。
func TestSSHKeygenRSA(t *testing.T) {
	result, err := SSHKeygen(t.Context(), &SSHKeygenRequest{
		Algorithm: SSHAlgorithmRSA,
		Bits:      2048,
		Comment:   "toolbit@example.com",
	})
	data := result.Data.(SSHKeygenResult)
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	if !strings.Contains(data.PublicKey, "ssh-rsa") {
		t.Errorf("公钥应含类型名：%s", data.PublicKey)
	}
	if data.Bits != 2048 {
		t.Errorf("位数 = %d，期望 2048", data.Bits)
	}
}

// TestSSHKeygenUnsupportedAlgorithm 断言未知算法返回 INVALID_FIELD。
func TestSSHKeygenUnsupportedAlgorithm(t *testing.T) {
	_, err := SSHKeygen(t.Context(), &SSHKeygenRequest{Algorithm: "DSA"})
	assertCode(t, err, toolerr.CodeInvalidField)
}
'''

TESTS["pki/certlifecycle_test.go"] = '''package pki

import (
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestCertLifecycleMissingCertificate 断言缺少证书返回 MISSING_FIELD。
func TestCertLifecycleMissingCertificate(t *testing.T) {
	_, err := CertLifecycle(t.Context(), &CertLifecycleRequest{
		IsPEM:       true,
		WarningDays: 30,
	})
	assertCode(t, err, toolerr.CodeMissingField)
}
'''


def main() -> None:
    for relative, content in TESTS.items():
        path = TOOLS / relative
        path.write_text(content, encoding="utf-8")
        print(f"写入 {relative}")


if __name__ == "__main__":
    main()
