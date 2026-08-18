package graph

import (
	"bytes"
	"fmt"

	"github.com/Bz-Lxt/pipeyard/types"
)

func (g *Graph) DOT() string {
	var b bytes.Buffer
	b.WriteString("digraph yard {\n")
	for _, id := range g.order {
		n := g.nodes[id]
		fmt.Fprintf(&b, "  %q [label=\"%s\\n%s\"];\n", id, id, n.Kind)
	}
	for _, e := range g.Edges() {
		fmt.Fprintf(&b, "  %q -> %q;\n", e.From, e.To)
	}
	b.WriteString("}\n")
	return b.String()
}

func DOTWithStatus(g *Graph, st map[string]types.NodeStatus) string {
	var b bytes.Buffer
	b.WriteString("digraph yard {\n")
	for _, id := range g.order {
		n := g.nodes[id]
		fmt.Fprintf(&b, "  %q [label=\"%s\\n%s\\n%s\"];\n", id, id, n.Kind, st[id])
	}
	for _, e := range g.Edges() {
		fmt.Fprintf(&b, "  %q -> %q;\n", e.From, e.To)
	}
	b.WriteString("}\n")
	return b.String()
}

func (g *Graph) Roots() []string {
	var out []string
	for _, id := range g.order {
		if len(g.pred[id]) == 0 {
			out = append(out, id)
		}
	}
	return out
}

func (g *Graph) Leaves() []string {
	var out []string
	for _, id := range g.order {
		if len(g.succ[id]) == 0 {
			out = append(out, id)
		}
	}
	return out
}

func (g *Graph) PathExists(from, to string) bool {
	if from == to {
		return true
	}
	seen := map[string]bool{}
	var walk func(string) bool
	walk = func(u string) bool {
		if u == to {
			return true
		}
		if seen[u] {
			return false
		}
		seen[u] = true
		for _, s := range g.succ[u] {
			if walk(s) {
				return true
			}
		}
		return false
	}
	return walk(from)
}
