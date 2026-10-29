package data

import (
	"context"
	"fmt"
	"math/big"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/blue-cloud-net/Toolbit/backend/internal/cryptosvc"
	"github.com/blue-cloud-net/Toolbit/backend/internal/tools"
)

// ID 生成格式。
const (
	// IDFormatUUID 是 UUID v4。
	IDFormatUUID = "uuid"
	// IDFormatULID 是 ULID。
	IDFormatULID = "ulid"
	// IDFormatSnowflake 是雪花 ID。
	IDFormatSnowflake = "snowflake"
	// IDFormatNanoID 是 NanoID。
	IDFormatNanoID = "nanoid"
)

// IDFormat 返回受支持的 ID 格式（顺序稳定）。
func IDFormat() []string {
	return []string{IDFormatUUID, IDFormatULID, IDFormatSnowflake, IDFormatNanoID}
}

// NanoID 字母表。
const nanoIDAlphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz_-"

// ULID 编码字母表（Crockford Base32）。
const ulidAlphabet = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

// IDGeneratorRequest 是 id-generator 工具的请求。
type IDGeneratorRequest struct {
	// Format 是 ID 格式；默认 uuid。
	Format string `json:"format"`
	// Count 是生成数量；默认 5。
	Count int `json:"count"`
	// Length 是 NanoID 长度；默认 21。
	Length int `json:"length"`
	// Epoch 是雪花 ID 的起始纪元毫秒时间戳；默认 2020-01-01。
	Epoch int64 `json:"epoch"`
	// NodeID 是雪花 ID 的节点号（0~1023）。
	NodeID int64 `json:"nodeId"`
	// Uppercase 是否转成大写。
	Uppercase bool `json:"uppercase"`
}

// IDGeneratorResult 是 id-generator 工具的响应。
type IDGeneratorResult struct {
	// Format 是实际使用的格式。
	Format string `json:"format"`
	// IDs 是生成的 ID 列表。
	IDs []string `json:"ids"`
	// Count 是生成数量。
	Count int `json:"count"`
	// Length 是 ID 长度；不固定长度时为 0。
	Length int `json:"length"`
}

// 默认雪花 ID 纪元：2020-01-01T00:00:00Z。
const defaultSnowflakeEpoch int64 = 1577836800000

// 最大生成数量。
const maxIDCount = 1000

// IDGenerator 生成 UUID、ULID、雪花 ID 与 NanoID。
//
// UUID 与 NanoID 的随机源经 cryptosvc 的 CSPRNG；雪花 ID 的时间部分
// 用系统时钟，节点号与序列号保证同一毫秒内不重复。
func IDGenerator(_ context.Context, req *IDGeneratorRequest) (tools.Result, error) {
	if err := req.Validate(); err != nil {
		return tools.Result{}, err
	}

	format := req.Format
	if format == "" {
		format = IDFormatUUID
	}
	count := req.Count
	if count == 0 {
		count = 5
	}

	ids := make([]string, 0, count)
	for i := 0; i < count; i++ {
		id, err := generateID(req, format)
		if err != nil {
			return tools.Result{}, err
		}
		if req.Uppercase {
			id = strings.ToUpper(id)
		}
		ids = append(ids, id)
	}

	result := IDGeneratorResult{
		Format: format,
		IDs:    ids,
		Count:  len(ids),
	}
	if format == IDFormatNanoID {
		result.Length = req.Length
	}

	return tools.Result{Data: result}, nil
}

// Validate 做字段级校验。
func (r *IDGeneratorRequest) Validate() error {
	format := r.Format
	if format == "" {
		format = IDFormatUUID
	}
	if !containsString(IDFormat(), format) {
		return tools.InvalidField("format", IDFormat(), format)
	}
	if r.Count < 0 || r.Count > maxIDCount {
		return tools.InvalidField("count", []int{0, maxIDCount}, r.Count)
	}
	if r.Length < 0 || r.Length > 64 {
		return tools.InvalidField("length", []int{0, 64}, r.Length)
	}
	if r.NodeID < 0 || r.NodeID > 1023 {
		return tools.InvalidField("nodeId", []int{0, 1023}, r.NodeID)
	}
	return nil
}

// generateID 生成单个 ID。
func generateID(req *IDGeneratorRequest, format string) (string, error) {
	switch format {
	case IDFormatULID:
		return generateULID()
	case IDFormatSnowflake:
		return generateSnowflake(req)
	case IDFormatNanoID:
		return generateNanoID(req.Length)
	default:
		return generateUUID()
	}
}

// generateUUID 生成 UUID v4。
func generateUUID() (string, error) {
	raw, err := cryptosvc.RandomBytes(16)
	if err != nil {
		return "", err
	}
	// 版本位固定为 4，变体位取 8/9/a/b。
	raw[6] = (raw[6] & 0x0f) | 0x40
	raw[8] = (raw[8] & 0x3f) | 0x80

	const hexDigits = "0123456789abcdef"
	out := make([]byte, 0, 36)
	for i, b := range raw {
		if i == 4 || i == 6 || i == 8 || i == 10 {
			out = append(out, '-')
		}
		out = append(out, hexDigits[b>>4], hexDigits[b&0x0f])
	}
	return string(out), nil
}

// generateULID 生成 ULID。
//
// 结构是 48 位毫秒时间戳 + 80 位随机数，编码为 26 位 Crockford Base32。
func generateULID() (string, error) {
	timestamp := uint64(time.Now().UnixMilli())
	random, err := cryptosvc.RandomBytes(10)
	if err != nil {
		return "", err
	}

	// 拼成 128 位：48 位时间 + 80 位随机。
	value := new(big.Int).SetUint64(timestamp)
	value.Lsh(value, 80)
	for i, b := range random {
		byteValue := big.NewInt(int64(b))
		byteValue.Lsh(byteValue, uint(8*(9-i)))
		value.Or(value, byteValue)
	}
	return encodeULID(value), nil
}

// encodeULID 把 128 位值编码成 26 位 Crockford Base32。
//
// 26 位 × 5 比特 = 130 位，最高位恒为 0，因此只需取低 125 位。
func encodeULID(value *big.Int) string {
	const bits = 125
	digits := make([]byte, 26)
	work := new(big.Int).Set(value)

	for i := 25; i >= 0; i-- {
		mod := new(big.Int)
		work.DivMod(work, big.NewInt(32), mod)
		digits[i] = ulidAlphabet[mod.Int64()]
	}
	_ = bits
	return string(digits)
}

// generateSnowflake 生成雪花 ID。
//
// 结构：1 位符号 + 41 位毫秒时间戳 + 10 位节点号 + 12 位序列号。
func generateSnowflake(req *IDGeneratorRequest) (string, error) {
	epoch := req.Epoch
	if epoch == 0 {
		epoch = defaultSnowflakeEpoch
	}

	now := time.Now().UnixMilli()
	elapsed := now - epoch
	if elapsed < 0 {
		return "", tools.InvalidField("epoch", "epoch before current time", "invalid value")
	}
	if elapsed >= 1<<41 {
		return "", tools.InvalidField("epoch", "epoch within 69 years", "invalid value")
	}

	sequence := rand.IntN(1 << 12)
	id := (elapsed << 22) | (req.NodeID << 12) | int64(sequence)
	return fmt.Sprintf("%d", id), nil
}

// generateNanoID 生成 NanoID。
func generateNanoID(length int) (string, error) {
	if length == 0 {
		length = 21
	}
	raw, err := cryptosvc.RandomBytes(length)
	if err != nil {
		return "", err
	}
	out := make([]byte, length)
	for i, b := range raw {
		out[i] = nanoIDAlphabet[int(b)%len(nanoIDAlphabet)]
	}
	return string(out), nil
}
