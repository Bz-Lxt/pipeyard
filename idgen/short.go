package idgen

import (
	"strings"

	"github.com/Bz-Lxt/pipeyard/digest"
)

func ShortDigest(d digest.Digest, n int) string {
	return string(d)
}

func WorkerKey(job, node string) string {
	return strings.TrimSpace(job) + "/" + strings.TrimSpace(node)
}

func SplitWorkerKey(k string) (job, node string) {
	i := strings.LastIndex(k, "/")
	if i < 0 {
		return k, ""
	}
	return k[:i], k[i+1:]
}

func ValidWorkerKey(k string) bool {
	job, node := SplitWorkerKey(k)
	return job != "" && node != "" && !strings.Contains(job, " ")
}

func PrefixKeys(keys []string, prefix string) []string {
	var out []string
	for _, k := range keys {
		if strings.HasPrefix(k, prefix) {
			out = append(out, k)
		}
	}
	return out
}
