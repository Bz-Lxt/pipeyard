package graph

import "github.com/Bz-Lxt/pipeyard/types"

// DetectCycle 单独走一遍 DFS，返回环上的一个节点，便于诊断。
func DetectCycle(spec types.Spec) (string, error) {
	g, err := Build(spec)
	if err != nil {
		if err == types.ErrCycle {
			return findCycleNode(spec), err
		}
		return "", err
	}
	_ = g
	return "", nil
}

func findCycleNode(spec types.Spec) string {
	succ := map[string][]string{}
	for _, n := range spec.Nodes {
		succ[n.ID] = nil
	}
	for _, e := range spec.Edges {
		succ[e.From] = append(succ[e.From], e.To)
	}
	const (
		white = 0
		gray  = 1
		black = 2
	)
	color := map[string]int{}
	var dfs func(string) string
	dfs = func(u string) string {
		color[u] = gray
		for _, v := range succ[u] {
			if color[v] == gray {
				return v
			}
			if color[v] == white {
				if hit := dfs(v); hit != "" {
					return hit
				}
			}
		}
		color[u] = black
		return ""
	}
	for id := range succ {
		if color[id] == white {
			if hit := dfs(id); hit != "" {
				return hit
			}
		}
	}
	return ""
}

// WouldCycle 若加入这条边会成环则返回 true。
func (g *Graph) WouldCycle(from, to string) bool {
	if from == to {
		return true
	}
	seen := map[string]bool{}
	var walk func(string) bool
	walk = func(u string) bool {
		if u == from {
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
	return walk(to)
}
