package policy

import (
	"fmt"
	"time"
)

// Backoff 计算第 attempt 次领取前应等待的秒数，线性递增且封顶。
func Backoff(attempt, base, capSec int) time.Duration {
	if attempt < 0 {
		attempt = 0
	}
	if base <= 0 {
		base = 1
	}
	if capSec <= 0 {
		capSec = 30
	}
	sec := base * (attempt + 1)
	if sec > capSec {
		sec = capSec
	}
	return time.Duration(sec) * time.Second
}

func FormatBackoff(d time.Duration) string {
	return fmt.Sprintf("%ds", int(d.Seconds()))
}

// Classify 把错误分成可重试与不可重试。ctx 取消不可重试。
func Classify(err error) (retryable bool) {
	if err == nil {
		return false
	}
	s := err.Error()
	if s == "context canceled" || s == "context deadline exceeded" {
		return false
	}
	return true
}
