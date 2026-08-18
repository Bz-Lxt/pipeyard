package quota

import "sync"

type Set struct {
	mu    sync.Mutex
	names map[string]struct{}
	limit int
}

func NewSet(limit int) *Set {
	if limit <= 0 {
		limit = 1
	}
	return &Set{names: map[string]struct{}{}, limit: limit}
}

func (s *Set) Add(name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.names[name]; ok {
		return false
	}
	s.names[name] = struct{}{}
	return true
}

func (s *Set) Remove(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.names, name)
}

func (s *Set) Has(name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.names[name]
	return ok
}

func (s *Set) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.names)
}

func (s *Set) Names() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]string, 0, len(s.names))
	for n := range s.names {
		out = append(out, n)
	}
	return out
}
