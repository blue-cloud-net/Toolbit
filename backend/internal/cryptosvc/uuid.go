package cryptosvc

import (
	"encoding/binary"
	"fmt"
	"sync"
	"time"
)

// UUID 版本。
const (
	// UUIDVersion1 是基于时间的 UUID。
	UUIDVersion1 = 1
	// UUIDVersion4 是基于随机数的 UUID。
	UUIDVersion4 = 4
	// UUIDVersion7 是时间有序的 UUID。
	UUIDVersion7 = 7
)

// UUIDVersions 返回受支持的版本（顺序稳定）。
func UUIDVersions() []int {
	return []int{UUIDVersion1, UUIDVersion4, UUIDVersion7}
}

// uuidGregorianOffset 是格里高利历起始点到 Unix 纪元的 100 纳秒数。
//
// UUID v1 的时间戳以 1582-10-15 00:00:00 为起点，
// 与 Unix 纪元相差 122192928000000000 个 100 纳秒。
const uuidGregorianOffset = 0x01B21DD213814000

// uuidNode 是 v1 的 48 位节点标识。
//
// **不**使用网卡 MAC 地址：工具会把这个标识返回给调用方，
// 而 MAC 地址是设备指纹。改用随机值并置位组播位
// （RFC 4122 §4.5 的「非真实硬件」约定）。
var uuidNode = func() [6]byte {
	var node [6]byte
	// 随机源失败时退化为全零：节点标识只影响唯一性，
	// 不影响格式正确性。
	if raw, err := RandomBytes(6); err == nil {
		copy(node[:], raw)
	}
	node[0] |= 0x01
	return node
}()

// uuidState 保存单调性状态。
//
// v1 与 v7 都要求「同一时刻内不重复」：
// 时间戳的分辨率有限，同一毫秒内多次调用必须靠序号区分。
type uuidState struct {
	// mu 保护以下字段。
	mu sync.Mutex
	// lastTimestamp 是 v1 上一次发出的 100 纳秒时间戳。
	lastTimestamp uint64
	// lastMillis 是 v7 上一次发出的毫秒时间戳。
	lastMillis int64
	// counterHi 是 v7 计数器的高 12 位（rand_a）。
	counterHi uint16
	// counterLo 是 v7 计数器的低 62 位（rand_b）。
	counterLo uint64
	// seeded 报告计数器是否已播种随机初值。
	seeded bool
}

// uuidStateGlobal 是全局的单调性状态。
var uuidStateGlobal uuidState

// uuidCounterMask 是 v7 计数器的低位掩码（62 位）。
const uuidCounterMask = 0x3FFFFFFFFFFFFFFF

// NewUUID 生成一个指定版本的 UUID。
//
// 版本 1 与 7 使用**单调**时间戳：同一毫秒内多次调用不会重复。
// 随机位来自 Tongsuo 的 RAND_bytes（与其它工具同源）。
func NewUUID(version int) (string, error) {
	switch version {
	case UUIDVersion1:
		return newUUIDv1()
	case UUIDVersion7:
		return newUUIDv7()
	case UUIDVersion4:
		return newUUIDv4()
	default:
		return "", fmt.Errorf("unsupported uuid version %d", version)
	}
}

// newUUIDv4 生成基于随机数的 UUID。
func newUUIDv4() (string, error) {
	var raw [16]byte
	if err := fillRandom(raw[:]); err != nil {
		return "", err
	}
	raw[6] = (raw[6] & 0x0f) | 0x40
	raw[8] = (raw[8] & 0x3f) | 0x80
	return formatUUID(raw), nil
}

// newUUIDv7 生成时间有序的 UUID。
//
// 时间戳取毫秒；74 位计数器（rand_a + rand_b）保证**全局严格递增**：
// 计数器只增不减，跨毫秒时时间戳也只增。
func newUUIDv7() (string, error) {
	var raw [16]byte

	millis, hi, lo := uuidStateGlobal.nextV7()
	binary.BigEndian.PutUint32(raw[0:4], uint32(millis>>16))
	binary.BigEndian.PutUint16(raw[4:6], uint16(millis))

	raw[6] = 0x70 | byte((hi>>8)&0x0f)
	raw[7] = byte(hi)
	raw[8] = 0x80 | byte((lo>>56)&0x3f)
	// rand_b 的低 56 位占 raw[9:16]，共 7 个字节。
	for i := 9; i < 16; i++ {
		raw[i] = byte(lo >> (8 * (15 - i)))
	}

	return formatUUID(raw), nil
}

// nextV7 返回毫秒时间戳与 74 位计数器。
//
// 时钟回拨时沿用上一次的时间戳：时间戳**不得**倒退。
func (s *uuidState) nextV7() (int64, uint16, uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.seeded {
		s.seeded = true
		if raw, err := RandomBytes(10); err == nil {
			s.counterHi = binary.BigEndian.Uint16(raw[0:2]) & 0x0fff
			s.counterLo = binary.BigEndian.Uint64(raw[2:10]) & uuidCounterMask
		}
	}

	millis := time.Now().UnixMilli()
	if millis < s.lastMillis {
		millis = s.lastMillis
	}
	s.lastMillis = millis

	s.counterLo = (s.counterLo + 1) & uuidCounterMask
	if s.counterLo == 0 {
		s.counterHi = (s.counterHi + 1) & 0x0fff
	}

	return millis, s.counterHi, s.counterLo
}

// newUUIDv1 生成基于时间的 UUID。
//
// 时间戳是 100 纳秒计数；若系统时钟在同一计数内多次取样，
// 则沿用上一次的时间戳并递增，保证严格递增。
func newUUIDv1() (string, error) {
	var raw [16]byte

	timestamp := uuidStateGlobal.nextV1()

	timeLow := uint32(timestamp & 0xffffffff)
	timeMid := uint16((timestamp >> 32) & 0xffff)
	timeHi := uint16((timestamp >> 48) & 0x0fff)

	binary.BigEndian.PutUint32(raw[0:4], timeLow)
	binary.BigEndian.PutUint16(raw[4:6], timeMid)
	binary.BigEndian.PutUint16(raw[6:8], timeHi|0x1000)

	// 时钟序列取随机值并置位变体标记。
	seqRaw, err := RandomBytes(2)
	if err != nil {
		return "", err
	}
	seq := binary.BigEndian.Uint16(seqRaw) & 0x3fff
	binary.BigEndian.PutUint16(raw[8:10], seq|0x8000)

	copy(raw[10:16], uuidNode[:])

	return formatUUID(raw), nil
}

// nextV1 返回下一个 100 纳秒时间戳。
//
// 保证严格递增：同一计数内多次调用会自动跳过。
func (s *uuidState) nextV1() uint64 {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := uint64(time.Now().UnixNano()/100) + uuidGregorianOffset
	if now <= s.lastTimestamp {
		now = s.lastTimestamp + 1
	}
	s.lastTimestamp = now
	return now
}

// fillRandom 用随机字节填满缓冲区。
func fillRandom(buf []byte) error {
	raw, err := RandomBytes(len(buf))
	if err != nil {
		return err
	}
	copy(buf, raw)
	return nil
}

// formatUUID 把 16 字节渲染成标准 UUID 文本。
func formatUUID(raw [16]byte) string {
	return fmt.Sprintf("%08x-%04x-%04x-%04x-%012x",
		binary.BigEndian.Uint32(raw[0:4]),
		binary.BigEndian.Uint16(raw[4:6]),
		binary.BigEndian.Uint16(raw[6:8]),
		binary.BigEndian.Uint16(raw[8:10]),
		raw[10:16],
	)
}
