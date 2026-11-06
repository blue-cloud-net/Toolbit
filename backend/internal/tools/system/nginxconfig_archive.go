package system

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"fmt"
	"time"
)

// nginxArchiveTime 是压缩包内条目的固定时间。
//
// 固定时间让同样的输入产出同样的压缩包，便于比对与排查。
var nginxArchiveTime = time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)

// packNginxArchive 把生成的文件打成 ZIP。
//
// ZIP 以 Base64 随 JSON 响应返回：工具端点统一返回 JSON 信封，
// 单独的二进制下载端点会绕开分发器，因此这里不另开通道。
func packNginxArchive(files []NginxFile) (*NginxArchive, error) {
	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)
	for _, file := range files {
		header := &zip.FileHeader{
			Name:     file.Path,
			Method:   zip.Deflate,
			Modified: nginxArchiveTime,
		}
		entry, err := writer.CreateHeader(header)
		if err != nil {
			return nil, fmt.Errorf("nginx: 写入压缩条目失败: %w", err)
		}
		if _, err := entry.Write([]byte(file.Content)); err != nil {
			return nil, fmt.Errorf("nginx: 写入压缩条目失败: %w", err)
		}
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("nginx: 关闭压缩包失败: %w", err)
	}

	total := 0
	for _, file := range files {
		total += len(file.Content)
	}

	return &NginxArchive{
		Filename:          "nginx-config.zip",
		ContentType:       "application/zip",
		Base64:            base64.StdEncoding.EncodeToString(buf.Bytes()),
		Entries:           len(files),
		UncompressedBytes: total,
	}, nil
}
