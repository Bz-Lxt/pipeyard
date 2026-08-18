// Package worker 管理领取租约与 worker 身份。
package worker

import (
	"fmt"
	"sync"
	"sync/atomic"
)

var seq atomic.Uint64

// ID 生成稳定的 worker 名字。
func ID(prefix string) string {
	if prefix == "" {
		prefix = "w"
	}
	n := seq.Add(1)
	return fmt.Sprintf("%s-%d", prefix, n)
}

// Book 记录本进程持有的租约，用于取消时释放。
type Book struct {
	mu   sync.Mutex
	held map[string]string // job/node -> worker
}

func NewBook() *Book {
	return &Book{held: map[string]string{}}
}

func key(job, node string) string { return job + "/" + node }

func (b *Book) Hold(job, node, worker string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.held[key(job, node)] = worker
}

func (b *Book) Release(job, node string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.held, key(job, node))
}

func (b *Book) Owner(job, node string) (string, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	w, ok := b.held[key(job, node)]
	return w, ok
}

func (b *Book) List() []string {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := make([]string, 0, len(b.held))
	for k := range b.held {
		out = append(out, k)
	}
	return out
}
