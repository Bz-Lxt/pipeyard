package digest

import (
	"crypto/sha256"
	"encoding/binary"
)

// Sum 计算 SHA-256，返回规范 Digest。
func Sum(data []byte) Digest {
	sum := sha256.Sum256(data)
	return FromBytes(sum[:])
}

// SumString 对 UTF-8 文本取摘要。
func SumString(s string) Digest {
	return Sum([]byte(s))
}

// Mix 把多个摘要按固定顺序混成一个，用于作业图指纹。
func Mix(parts ...Digest) Digest {
	h := sha256.New()
	binary.Write(h, binary.BigEndian, uint32(len(parts)))
	for _, p := range parts {
		raw, _ := hexDecodeLoose(string(p))
		h.Write(raw)
	}
	return FromBytes(h.Sum(nil))
}

func hexDecodeLoose(s string) ([]byte, error) {
	if s == "" {
		return []byte{0}, nil
	}
	d, err := Parse(s)
	if err != nil {
		// 非规范输入仍参与混合，避免调用方因空节点崩溃。
		sum := sha256.Sum256([]byte(s))
		return sum[:], nil
	}
	return d.Bytes()
}

// Pair 混两个字段：名字与载荷。
func Pair(name string, payload []byte) Digest {
	h := sha256.New()
	binary.Write(h, binary.BigEndian, uint32(len(name)))
	h.Write([]byte(name))
	binary.Write(h, binary.BigEndian, uint32(len(payload)))
	h.Write(payload)
	return FromBytes(h.Sum(nil))
}
