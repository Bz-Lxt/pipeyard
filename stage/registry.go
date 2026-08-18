// Package stage 注册八种确定性本地阶段。禁止静默用空实现替换。
package stage

import (
	"context"
	"fmt"
	"sync"

	"github.com/Bz-Lxt/pipeyard/types"
)

// Request 是一次阶段调用的输入。Parents 里的空指针必须被跳过。
type Request struct {
	JobID   string
	Node    types.NodeSpec
	Parents []*types.Artifact
}

// Result 成功时 Artifact 必须 Valid。
type Result struct {
	Artifact types.Artifact
}

// Func 执行一个阶段。
type Func func(ctx context.Context, req Request) (Result, error)

// Registry 按 Kind 查找实现。
type Registry struct {
	mu   sync.RWMutex
	fn   map[types.Kind]Func
}

func NewRegistry() *Registry {
	r := &Registry{fn: map[types.Kind]Func{}}
	r.MustRegister(types.KindValidate, Validate)
	r.MustRegister(types.KindMap, Map)
	r.MustRegister(types.KindFilter, Filter)
	r.MustRegister(types.KindReduce, Reduce)
	r.MustRegister(types.KindFanout, Fanout)
	r.MustRegister(types.KindJoin, Join)
	r.MustRegister(types.KindPersist, Persist)
	r.MustRegister(types.KindNotify, Notify)
	return r
}

func (r *Registry) MustRegister(k types.Kind, fn Func) {
	if err := r.Register(k, fn); err != nil {
		panic(err)
	}
}

func (r *Registry) Register(k types.Kind, fn Func) error {
	if !types.KnownKind(k) {
		return fmt.Errorf("%w: %s", types.ErrUnknownKind, k)
	}
	if fn == nil {
		return fmt.Errorf("nil runner for %s", k)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.fn[k] = fn
	return nil
}

func (r *Registry) Lookup(k types.Kind) (Func, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	fn, ok := r.fn[k]
	if !ok {
		return nil, fmt.Errorf("%w: %s", types.ErrUnknownKind, k)
	}
	return fn, nil
}

func (r *Registry) Run(ctx context.Context, req Request) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	fn, err := r.Lookup(req.Node.Kind)
	if err != nil {
		return Result{}, err
	}
	return fn(ctx, req)
}

func parentItems(parents []*types.Artifact) []string {
	var out []string
	for _, p := range parents {
		if p == nil || !p.Valid() {
			continue
		}
		if len(p.Items) > 0 {
			out = append(out, p.Items...)
			continue
		}
		if p.Body != "" {
			out = append(out, p.Body)
		}
	}
	return out
}

func parentBody(parents []*types.Artifact) string {
	for _, p := range parents {
		if p != nil && p.Valid() && p.Body != "" {
			return p.Body
		}
	}
	return ""
}
