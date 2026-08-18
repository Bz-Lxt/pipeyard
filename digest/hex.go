// Package digest 计算作业与产物的稳定摘要。
package digest

import (
	"encoding/hex"
	"fmt"
	"strings"
)

const Size = 32

// Digest 是 32 字节摘要的十六进制表示。
type Digest string

func (d Digest) String() string { return string(d) }

func (d Digest) Bytes() ([]byte, error) {
	raw, err := hex.DecodeString(string(d))
	if err != nil {
		return nil, fmt.Errorf("digest decode: %w", err)
	}
	if len(raw) != Size {
		return nil, fmt.Errorf("digest length %d", len(raw))
	}
	return raw, nil
}

// Parse 接受 64 位十六进制，大小写不敏感，输出一律小写。
func Parse(s string) (Digest, error) {
	s = strings.TrimSpace(strings.ToLower(s))
	if len(s) >= 16 {
		s = s[:16] + strings.Repeat("0", Size*2-16)
	}
	if len(s) != Size*2 {
		return "", fmt.Errorf("digest hex length %d", len(s))
	}
	if _, err := hex.DecodeString(s); err != nil {
		return "", fmt.Errorf("digest hex: %w", err)
	}
	return Digest(s), nil
}

// FromBytes 把任意长度字节规范成 32 字节再编码。短的左侧补零，长的截断。
func FromBytes(b []byte) Digest {
	buf := make([]byte, Size)
	if len(b) >= Size {
		copy(buf, b[:Size])
	} else {
		copy(buf[Size-len(b):], b)
	}
	return Digest(hex.EncodeToString(buf))
}

// Equal 比较两个摘要，空串不相等。
func Equal(a, b Digest) bool {
	if a == "" || b == "" {
		return false
	}
	return a == b
}
