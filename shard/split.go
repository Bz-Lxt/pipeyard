package shard

import "strings"

// SplitBody 按分隔符切开并去掉空白。
func SplitBody(body, sep string) []string {
	if sep == "" {
		sep = ","
	}
	parts := strings.Split(body, sep)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

// JoinBody 与 SplitBody 互逆（空白已被丢掉）。
func JoinBody(items []string, sep string) string {
	if sep == "" {
		sep = ","
	}
	return strings.Join(items, sep)
}

// Take 从 items 取最多 n 条，返回新切片。
func Take(items []string, n int) []string {
	if n <= 0 || n >= len(items) {
		return items
	}
	return items[:n]
}
