package network

import (
	"encoding/asn1"
	"encoding/hex"
	"strings"
)

// TLS Feature 扩展里的 must-staple 状态码（RFC 7633）。
const (
	// statusRequest 是 status_request（RFC 6962 的 must-staple）。
	statusRequest = 5
	// statusRequestV2 是 status_request_v2（RFC 6962 bis）。
	statusRequestV2 = 17
)

// mustStapleOf 判断证书是否声明 must-staple。
//
// TLS Feature 扩展（RFC 7633）的值是 `SEQUENCE OF INTEGER`；
// 出现 `status_request`（5）或 `status_request_v2`（17）即表示 must-staple。
//
// 这里只做**结构解析**（不涉及任何密码学运算）。解析失败返回 false：
// 「读不出」与「没有」在结论上等价 —— 都意味着不能要求服务端装订 OCSP。
func mustStapleOf(extValueHex string) bool {
	raw, err := hex.DecodeString(strings.TrimSpace(extValueHex))
	if err != nil || len(raw) == 0 {
		return false
	}
	var features []int
	if _, err := asn1.Unmarshal(raw, &features); err != nil {
		return false
	}
	for _, feature := range features {
		if feature == statusRequest || feature == statusRequestV2 {
			return true
		}
	}
	return false
}

// sctCountOf 统计证书内嵌的 SCT（签名证书时间戳）条目数。
//
// 扩展值是 `SignedCertificateTimestampList`（RFC 6962 §3.3）：
// 一个 2 字节的总长度，后跟若干「2 字节长度 + SCT 字节」条目。
// OpenSSL 可能把它再包一层 OCTET STRING，因此这里两种形态都尝试。
//
// 解析失败时返回 1（扩展存在，但条目数未知）；返回 0 表示没有该扩展。
func sctCountOf(extValueHex string) int {
	raw, err := hex.DecodeString(strings.TrimSpace(extValueHex))
	if err != nil || len(raw) == 0 {
		return 0
	}
	if count, ok := countSCTList(raw); ok {
		return count
	}
	// 尝试剥掉外层 OCTET STRING。
	if len(raw) < 2 || raw[0] != 0x04 {
		return 1
	}
	length, headerLen, ok := derLength(raw[1:])
	if !ok {
		return 1
	}
	offset := 1 + headerLen
	if offset >= len(raw) || offset+length > len(raw) {
		return 1
	}
	if count, ok := countSCTList(raw[offset:]); ok {
		return count
	}
	return 1
}

// countSCTList 解析 SCT 列表并返回条目数。
func countSCTList(raw []byte) (int, bool) {
	if len(raw) < 2 {
		return 0, false
	}
	total := int(raw[0])<<8 | int(raw[1])
	if total+2 > len(raw) || total < 2 {
		return 0, false
	}
	body := raw[2 : total+2]
	count := 0
	for len(body) > 0 {
		if len(body) < 2 {
			return 0, false
		}
		size := int(body[0])<<8 | int(body[1])
		if size == 0 || size+2 > len(body) {
			return 0, false
		}
		count++
		body = body[2+size:]
	}
	return count, true
}

// derLength 解析 DER 的长度字段。
//
// 返回长度值、长度字段占用的字节数与是否解析成功。
func derLength(raw []byte) (int, int, bool) {
	if len(raw) == 0 {
		return 0, 0, false
	}
	if raw[0] < 0x80 {
		return int(raw[0]), 1, true
	}
	count := int(raw[0] & 0x7f)
	if count == 0 || count > 4 || len(raw) < 1+count {
		return 0, 0, false
	}
	length := 0
	for i := 1; i <= count; i++ {
		length = length<<8 | int(raw[i])
	}
	return length, 1 + count, true
}
