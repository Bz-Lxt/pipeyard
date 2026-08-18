package ledger

import "sync"

type Log struct {
	mu   sync.Mutex
	rows []Entry
	cap  int
}

func New(cap int) *Log {
	if cap <= 0 {
		cap = 128
	}
	return &Log{cap: cap}
}

func (l *Log) Append(e Entry) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.rows = append(l.rows, e)
	if len(l.rows) > l.cap {
		l.rows = append([]Entry(nil), l.rows[len(l.rows)-l.cap:]...)
	}
}

func (l *Log) All() []Entry {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]Entry(nil), l.rows...)
}

func (l *Log) ForJob(job string) []Entry {
	l.mu.Lock()
	defer l.mu.Unlock()
	var out []Entry
	for _, e := range l.rows {
		if e.Job == job {
			out = append(out, e)
		}
	}
	return out
}

func (l *Log) Len() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.rows)
}
