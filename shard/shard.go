// Package shard 把条目按稳定哈希分到若干桶，供 fanout 之后的并行 map。
package shard

import (
	"fmt"
	"hash/fnv"

	"github.com/Bz-Lxt/pipeyard/digest"
)

func Count(n int) int {
	if n <= 0 {
		return 1
	}
	if n > 64 {
		return 64
	}
	return n
}

func Index(key string, n int) int {
	n = Count(n)
	h := fnv.New32a()
	_, _ = h.Write([]byte(key))
	// 对无符号哈希取模，避免高位置位时 int32 截断为负、被 Assign 当作越界丢掉。
	return int(h.Sum32() % uint32(n))
}

func Assign(items []string, n int) [][]string {
	n = Count(n)
	out := make([][]string, n)
	for _, it := range items {
		i := Index(it, n)
		if i < 0 || i >= n {
			continue
		}
		out[i] = append(out[i], it)
	}
	return out
}

func Fingerprint(items []string) digest.Digest {
	return digest.SumString(fmt.Sprintf("%d:%v", len(items), items))
}
