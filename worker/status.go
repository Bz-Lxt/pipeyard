package worker

import "sync"

type Status struct {
	ID    string
	Busy  bool
	Job   string
	Node  string
	Error string
}

type Table struct {
	mu   sync.Mutex
	rows map[string]Status
}

func NewTable() *Table { return &Table{rows: map[string]Status{}} }

func (t *Table) Set(s Status) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.rows[s.ID] = s
}

func (t *Table) Get(id string) (Status, bool) {
	t.mu.Lock()
	defer t.mu.Unlock()
	s, ok := t.rows[id]
	return s, ok
}

func (t *Table) Idle(id string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	s := t.rows[id]
	s.Busy = false
	s.Job = ""
	s.Node = ""
	t.rows[id] = s
}

func (t *Table) List() []Status {
	t.mu.Lock()
	defer t.mu.Unlock()
	out := make([]Status, 0, len(t.rows))
	for _, s := range t.rows {
		out = append(out, s)
	}
	return out
}

func (t *Table) BusyCount() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	n := 0
	for _, s := range t.rows {
		if s.Busy {
			n++
		}
	}
	return n
}
