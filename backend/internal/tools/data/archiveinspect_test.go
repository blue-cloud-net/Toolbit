package data

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
