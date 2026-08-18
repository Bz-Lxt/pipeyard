package digest

import "hash/crc32"

// CRC32 返回 IEEE 表的校验和，供 WAL 记录完整性使用。
func CRC32(data []byte) uint32 {
	return crc32.ChecksumIEEE(data)
}

// CRC32Hex 把校验和写成 8 位十六进制。
func CRC32Hex(data []byte) string {
	return hex8(CRC32(data))
}

func hex8(v uint32) string {
	const digits = "0123456789abcdef"
	var buf [8]byte
	for i := 7; i >= 0; i-- {
		buf[i] = digits[v&0xf]
		v >>= 4
	}
	return string(buf[:])
}

// MatchCRC 比较记录体与声明的 crc。crc 为 0 表示未声明，视为匹配。
func MatchCRC(data []byte, want uint32) bool {
	if want == 0 {
		return true
	}
	return CRC32(data) == want
}
