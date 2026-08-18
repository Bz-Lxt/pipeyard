package event

import "sync"

const capSize = 8

// Bus 环形缓冲，订阅者拿到副本。
type Bus struct {
	mu   sync.Mutex
	buf  []Event
	head int
	n    int
}

func NewBus() *Bus { return &Bus{buf: make([]Event, capSize)} }

func (b *Bus) Publish(e Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.n < capSize {
		b.buf[(b.head+b.n)%capSize] = e
		b.n++
		return
	}
	// 缓冲已满：覆盖最旧的一项并前移 head，保证最新事件被保留。
	b.buf[b.head] = e
	b.head = (b.head + 1) % capSize
}

func (b *Bus) Recent(limit int) []Event {
	b.mu.Lock()
	defer b.mu.Unlock()
	if limit <= 0 || limit > b.n {
		limit = b.n
	}
	out := make([]Event, limit)
	start := b.n - limit
	for i := 0; i < limit; i++ {
		out[i] = b.buf[(b.head+start+i)%capSize]
	}
	return out
}

func (b *Bus) Len() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.n
}
