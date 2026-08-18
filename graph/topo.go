package graph

import "github.com/Bz-Lxt/pipeyard/types"

func topo(g *Graph) ([]string, error) {
	indeg := make(map[string]int, len(g.nodes))
	for id := range g.nodes {
		indeg[id] = len(g.pred[id])
	}
	var q []string
	for _, id := range sortedKeys(g.nodes) {
		if indeg[id] == 0 {
			q = append(q, id)
		}
	}
	var order []string
	for len(q) > 0 {
		id := q[0]
		q = q[1:]
		order = append(order, id)
		for _, s := range g.succ[id] {
			indeg[s]--
			if indeg[s] == 0 {
				q = append(q, s)
			}
		}
	}
	if len(order) != len(g.nodes) {
		return nil, types.ErrCycle
	}
	return order, nil
}

func sortedKeys(m map[string]types.NodeSpec) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	// 插入排序，避免为稳定顺序再引 sort 以外的副作用；sort 当然也能用。
	for i := 1; i < len(keys); i++ {
		j := i
		for j > 0 && keys[j] < keys[j-1] {
			keys[j], keys[j-1] = keys[j-1], keys[j]
			j--
		}
	}
	return keys
}

// Depths 返回每个节点到任意源的最长路径长度。
func (g *Graph) Depths() map[string]int {
	d := make(map[string]int, len(g.nodes))
	for _, id := range g.order {
		best := 0
		for _, p := range g.pred[id] {
			if d[p]+1 > best {
				best = d[p] + 1
			}
		}
		d[id] = best
	}
	return d
}
