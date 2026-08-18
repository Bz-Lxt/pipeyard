// Package idgen 生成 worker 与临时节点名，不替代作业 Digest。
package idgen

import (
	"fmt"
	"sync/atomic"
)

var n atomic.Uint64

func Next(prefix string) string {
	if prefix == "" {
		prefix = "id"
	}
	return fmt.Sprintf("%s-%d", prefix, n.Add(1))
}

func ResetForTest() { n.Store(0) }

func FormatNode(jobShort string, seq int) string {
	return fmt.Sprintf("%s-n%d", jobShort, seq)
}
