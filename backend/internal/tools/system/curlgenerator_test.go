package system

import (
	"strings"
	"testing"

	"github.com/blue-cloud-net/Toolbit/backend/internal/toolerr"
)

// TestCurlGeneratorBasic 断言基本命令生成。
func TestCurlGeneratorBasic(t *testing.T) {
	result, err := CurlGenerator(&CurlGeneratorRequest{
		Method: "POST",
		URL:    "https://example.com/api",
		Body:   `{"a":1}`,
	})
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	if !strings.Contains(result.Command, "-X POST") {
		t.Errorf("命令缺少方法：%s", result.Command)
	}
	if !strings.Contains(result.Command, "https://example.com/api") {
		t.Errorf("命令缺少地址：%s", result.Command)
	}
}

// TestCurlGeneratorRedactsCredentials 断言凭据以占位符呈现。
//
// 明文凭据**不得**出现在命令里：命令通常会被贴进聊天窗口或存进脚本。
func TestCurlGeneratorRedactsCredentials(t *testing.T) {
	result, err := CurlGenerator(&CurlGeneratorRequest{
		URL:      "https://example.com",
		Auth:     AuthBasic,
		Username: "alice",
		Password: "s3cret",
	})
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	if strings.Contains(result.Command, "s3cret") {
		t.Errorf("命令泄漏口令：%s", result.Command)
	}
	if !strings.Contains(result.Command, "<PASSWORD>") {
		t.Errorf("命令缺少口令占位符：%s", result.Command)
	}
	if len(result.Placeholders) != 1 || result.Placeholders[0] != "PASSWORD" {
		t.Errorf("占位符列表 = %v", result.Placeholders)
	}
}

// TestCurlGeneratorBearerToken 断言 Bearer Token 被占位符替代。
func TestCurlGeneratorBearerToken(t *testing.T) {
	result, err := CurlGenerator(&CurlGeneratorRequest{
		URL:   "https://example.com",
		Auth:  AuthBearer,
		Token: "abcdef123456",
	})
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	if strings.Contains(result.Command, "abcdef123456") {
		t.Errorf("命令泄漏 Token：%s", result.Command)
	}
	if !strings.Contains(result.Command, "<TOKEN>") {
		t.Errorf("命令缺少 Token 占位符：%s", result.Command)
	}
}

// TestCurlGeneratorSensitiveHeader 断言敏感请求头被占位符替代。
func TestCurlGeneratorSensitiveHeader(t *testing.T) {
	result, err := CurlGenerator(&CurlGeneratorRequest{
		URL: "https://example.com",
		Headers: []CurlHeader{
			{Name: "X-Api-Key", Value: "key-material", Sensitive: true},
			{Name: "Accept", Value: "application/json"},
		},
	})
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	if strings.Contains(result.Command, "key-material") {
		t.Errorf("命令泄漏敏感头：%s", result.Command)
	}
	if !strings.Contains(result.Command, "application/json") {
		t.Errorf("命令缺少普通头：%s", result.Command)
	}
}

// TestCurlGeneratorMissingURL 断言缺少地址时报 MISSING_FIELD。
func TestCurlGeneratorMissingURL(t *testing.T) {
	_, err := CurlGenerator(&CurlGeneratorRequest{})
	assertCode(t, err, toolerr.CodeMissingField)
}

// TestCurlGeneratorMissingToken 断言 Bearer 缺少 Token 时报 MISSING_FIELD。
func TestCurlGeneratorMissingToken(t *testing.T) {
	_, err := CurlGenerator(&CurlGeneratorRequest{
		URL:  "https://example.com",
		Auth: AuthBearer,
	})
	assertCode(t, err, toolerr.CodeMissingField)
}

// TestCurlGeneratorUnsupportedShell 断言未知 Shell 报 INVALID_FIELD。
func TestCurlGeneratorUnsupportedShell(t *testing.T) {
	_, err := CurlGenerator(&CurlGeneratorRequest{
		URL:   "https://example.com",
		Shell: "fish",
	})
	assertCode(t, err, toolerr.CodeInvalidField)
}

// TestCurlGeneratorPowershellQuoting 断言 PowerShell 用单引号。
func TestCurlGeneratorPowershellQuoting(t *testing.T) {
	result, err := CurlGenerator(&CurlGeneratorRequest{
		URL:   "https://example.com",
		Shell: ShellPowershell,
	})
	if err != nil {
		t.Fatalf("生成失败：%v", err)
	}
	if !strings.Contains(result.Command, "'https://example.com'") {
		t.Errorf("PowerShell 引号不正确：%s", result.Command)
	}
}
