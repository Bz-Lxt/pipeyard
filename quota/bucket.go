package quota

import (
	"sync"
	"time"
)

// Bucket 是令牌桶，Tick 领取节点时限速。
type Bucket struct {
	mu     sync.Mutex
	tokens int
	cap    int
	rate   int
	last   time.Time
}

func NewBucket(cap, rate int) *Bucket {
	if cap <= 0 {
		cap = 1
	}
	if rate <= 0 {
		rate = 1
	}
	return &Bucket{tokens: cap, cap: cap, rate: rate, last: time.Time{}}
}

func (b *Bucket) refill(now time.Time) {
	if b.last.IsZero() {
		b.last = now
		return
	}
	elapsed := int(now.Sub(b.last) / time.Second)
	if elapsed <= 0 {
		return
	}
	b.tokens += elapsed * b.rate
	if b.tokens > b.cap {
		b.tokens = b.cap
	}
	b.last = now
}

func (b *Bucket) Take(now time.Time, n int) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.refill(now)
	if n <= 0 {
		n = 1
	}
	if b.tokens < n {
		return false
	}
	b.tokens -= n
	return true
}

func (b *Bucket) Tokens() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.tokens
}

func (b *Bucket) Set(n int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if n < 0 {
		n = 0
	}
	if n > b.cap {
		n = b.cap
	}
	b.tokens = n
}
