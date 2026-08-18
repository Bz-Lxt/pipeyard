package types

import (
	"fmt"
	"strings"

	"github.com/Bz-Lxt/pipeyard/digest"
)

// NodeSpec 描述一个待执行阶段。
type NodeSpec struct {
	ID     string
	Kind   Kind
	Param  string
	Weight int
}

func (n NodeSpec) Normalize() (NodeSpec, error) {
	n.ID = strings.TrimSpace(n.ID)
	n.Kind = Kind(strings.TrimSpace(string(n.Kind)))
	n.Param = strings.TrimSpace(n.Param)
	if n.ID == "" {
		return n, fmt.Errorf("%w: empty node id", ErrBadSpec)
	}
	if !KnownKind(n.Kind) {
		return n, fmt.Errorf("%w: %s", ErrUnknownKind, n.Kind)
	}
	if n.Weight < 0 {
		return n, fmt.Errorf("%w: negative weight", ErrBadSpec)
	}
	if n.Weight == 0 {
		n.Weight = 1
	}
	return n, nil
}

// Edge 描述依赖：From 完成后 To 才就绪。
type Edge struct {
	From string
	To   string
}

func (e Edge) Normalize() (Edge, error) {
	e.From = strings.TrimSpace(e.From)
	e.To = strings.TrimSpace(e.To)
	if e.From == "" || e.To == "" {
		return e, fmt.Errorf("%w: empty edge", ErrBadSpec)
	}
	if e.From == e.To {
		return e, fmt.Errorf("%w: self edge", ErrCycle)
	}
	return e, nil
}

// Spec 是一次提交的作业图。
type Spec struct {
	Name  string
	Nodes []NodeSpec
	Edges []Edge
	Quota int
}

func (s Spec) Normalize() (Spec, error) {
	s.Name = strings.TrimSpace(s.Name)
	if s.Name == "" {
		return s, fmt.Errorf("%w: empty name", ErrBadSpec)
	}
	if len(s.Nodes) == 0 {
		return s, ErrEmptyGraph
	}
	seen := make(map[string]int, len(s.Nodes))
	out := make([]NodeSpec, 0, len(s.Nodes))
	for i, n := range s.Nodes {
		nn, err := n.Normalize()
		if err != nil {
			return s, fmt.Errorf("node %d: %w", i, err)
		}
		if _, ok := seen[nn.ID]; ok {
			return s, fmt.Errorf("%w: duplicate node %s", ErrBadSpec, nn.ID)
		}
		seen[nn.ID] = i
		out = append(out, nn)
	}
	s.Nodes = out
	edges := make([]Edge, 0, len(s.Edges))
	for i, e := range s.Edges {
		ee, err := e.Normalize()
		if err != nil {
			return s, fmt.Errorf("edge %d: %w", i, err)
		}
		if _, ok := seen[ee.From]; !ok {
			return s, fmt.Errorf("%w: edge from %s", ErrUnknownNode, ee.From)
		}
		if _, ok := seen[ee.To]; !ok {
			return s, fmt.Errorf("%w: edge to %s", ErrUnknownNode, ee.To)
		}
		edges = append(edges, ee)
	}
	s.Edges = edges
	if s.Quota < 0 {
		return s, fmt.Errorf("%w: negative quota", ErrBadSpec)
	}
	return s, nil
}

// Fingerprint 对规范化后的规格取摘要，作为作业身份的一部分。
func (s Spec) Fingerprint() digest.Digest {
	var parts []digest.Digest
	parts = append(parts, digest.SumString(s.Name))
	for _, n := range s.Nodes {
		parts = append(parts, digest.Pair(n.ID+"|"+string(n.Kind), []byte(n.Param)))
	}
	for _, e := range s.Edges {
		parts = append(parts, digest.SumString(e.From+"->"+e.To))
	}
	return digest.Mix(parts...)
}

func (s Spec) NodeIDs() []string {
	ids := make([]string, len(s.Nodes))
	for i, n := range s.Nodes {
		ids[i] = n.ID
	}
	return ids
}

func (s Spec) Lookup(id string) (NodeSpec, bool) {
	for _, n := range s.Nodes {
		if n.ID == id {
			return n, true
		}
	}
	return NodeSpec{}, false
}
