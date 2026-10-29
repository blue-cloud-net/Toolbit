package data

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
