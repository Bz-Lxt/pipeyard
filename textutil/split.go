package textutil

import "strings"

func CSV(s string) []string {
	return Fields(s, ",")
}

func Fields(s, sep string) []string {
	if sep == "" {
		sep = ","
	}
	parts := strings.Split(s, sep)
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = Compact(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func Unique(in []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, s := range in {
		if seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}
