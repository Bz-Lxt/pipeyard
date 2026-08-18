// Package clock 提供可注入时钟，默认返回北京时间。
package clock

import (
	"sync"
	"time"
)

const Offset = 8 * time.Hour

// Clock 抽象墙上时钟，测试可钉死。
type Clock interface {
	Now() time.Time
}

// Beijing 返回 Asia/Shanghai 墙上时间（无时区标记的本地表示）。
type Beijing struct{}

func (Beijing) Now() time.Time {
	return time.Now().In(time.FixedZone("CST", int(Offset.Seconds()))).Truncate(time.Second)
}

// Fixed 返回钉死的时间，每次 Now 都相同。
type Fixed struct{ T time.Time }

func (f Fixed) Now() time.Time { return f.T }

// Step 每次调用前进一拍，便于给 WAL 记录排序。
type Step struct {
	mu   sync.Mutex
	next time.Time
	step time.Duration
}

func NewStep(start time.Time, step time.Duration) *Step {
	if step <= 0 {
		step = time.Second
	}
	return &Step{next: start, step: step}
}

func (s *Step) Now() time.Time {
	s.mu.Lock()
	defer s.mu.Unlock()
	cur := s.next
	s.next = s.next.Add(s.step)
	return cur
}

// Format 用北京时间墙钟格式写出，供 SQLite 文本列使用。
func Format(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

// Parse 解析 Format 的输出。空串返回零值。
func Parse(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}
	return time.ParseInLocation("2006-01-02 15:04:05", s, time.FixedZone("CST", int(Offset.Seconds())))
}

// Since 计算 now 相对 t 的间隔，t 为零则返回 0。
func Since(now, t time.Time) time.Duration {
	if t.IsZero() || now.IsZero() {
		return 0
	}
	return now.Sub(t)
}

// AddSeconds 给墙上时间加秒，供租约过期计算。
func AddSeconds(t time.Time, sec int) time.Time {
	if sec == 0 {
		return t
	}
	return t.Add(time.Duration(sec) * time.Second)
}
