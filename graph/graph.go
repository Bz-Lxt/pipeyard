// Package graph 做拓扑、环检测与就绪集计算。纯计算，不碰磁盘。
package graph

import (
	"fmt"

	"github.com/Bz-Lxt/pipeyard/types"
)

// Graph 是规范化后的有向图。
type Graph struct {
	nodes map[string]types.NodeSpec
	succ  map[string][]string
	pred  map[string][]string
	order []string
}

func Build(spec types.Spec) (*Graph, error) {
	spec, err := spec.Normalize()
	if err != nil {
		return nil, err
	}
	g := &Graph{
		nodes: make(map[string]types.NodeSpec, len(spec.Nodes)),
		succ:  make(map[string][]string),
		pred:  make(map[string][]string),
	}
	for _, n := range spec.Nodes {
		g.nodes[n.ID] = n
		g.succ[n.ID] = nil
		g.pred[n.ID] = nil
	}
	for _, e := range spec.Edges {
		g.succ[e.From] = append(g.succ[e.From], e.To)
		g.pred[e.To] = append(g.pred[e.To], e.From)
	}
	order, err := topo(g)
	if err != nil {
		return nil, err
	}
	g.order = order
	return g, nil
}

func (g *Graph) Order() []string {
	return append([]string(nil), g.order...)
}

func (g *Graph) Spec(id string) (types.NodeSpec, bool) {
	n, ok := g.nodes[id]
	return n, ok
}

func (g *Graph) Preds(id string) []string {
	return append([]string(nil), g.pred[id]...)
}

func (g *Graph) Succs(id string) []string {
	return append([]string(nil), g.succ[id]...)
}

func (g *Graph) Size() int { return len(g.nodes) }

func (g *Graph) Has(id string) bool {
	_, ok := g.nodes[id]
	return ok
}

func (g *Graph) Edges() []types.Edge {
	var out []types.Edge
	for from, tos := range g.succ {
		for _, to := range tos {
			out = append(out, types.Edge{From: from, To: to})
		}
	}
	return out
}

func (g *Graph) ValidateRefs() error {
	for id := range g.nodes {
		for _, p := range g.pred[id] {
			if !g.Has(p) {
				return fmt.Errorf("%w: pred %s", types.ErrUnknownNode, p)
			}
		}
	}
	return nil
}
