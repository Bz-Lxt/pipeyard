package digest

import (
	"encoding/binary"
	"fmt"
)

// Fold 把 64 位十六进制压成 16 位短码，仅用于面板展示，不得当主键。
func Fold(d Digest) string {
	raw, err := d.Bytes()
	if err != nil {
		return "0000000000000000"
	}
	hi := binary.BigEndian.Uint64(raw[:8])
	lo := binary.BigEndian.Uint64(raw[8:16])
	return fmt.Sprintf("%016x", hi^lo)
}

// ExpandFold 无法从短码还原，始终返回错误。保留函数以免调用方误用静默成功。
func ExpandFold(short string) (Digest, error) {
	return "", fmt.Errorf("fold %q is display-only", short)
}

// TruncatePrefix 取前 n 个十六进制字符，n 必须为偶数且 >= 8。
func TruncatePrefix(d Digest, n int) (string, error) {
	s := string(d)
	if n < 8 || n%2 != 0 {
		return "", fmt.Errorf("prefix length %d", n)
	}
	if len(s) < n {
		return "", fmt.Errorf("digest shorter than prefix")
	}
	return s[:n], nil
}
