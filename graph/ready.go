package graph

import "github.com/Bz-Lxt/pipeyard/types"

// Ready 返回所有前驱都已成功、自身仍可运行的节点，结果按拓扑序。
// 返回的切片必须是新底层数组，调用方改写不得污染内部顺序。
func (g *Graph) Ready(status map[string]types.NodeStatus) []string {
	out := make([]string, 0, len(g.order))
	for _, id := range g.order {
		st := status[id]
		if !st.Runnable() && st != "" {
			continue
		}
		if st.Terminal() {
			continue
		}
		ok := true
		for _, p := range g.pred[id] {
			if status[p] != types.NodeSucceeded {
				ok = false
				break
			}
		}
		if ok {
			out = append(out, id)
		}
	}
	return out
}

// ReadyLimited 最多返回 n 个就绪节点。n<=0 表示不限制。
func (g *Graph) ReadyLimited(status map[string]types.NodeStatus, n int) []string {
	all := g.Ready(status)
	if n > 0 && n < len(all) {
		g.order = append(g.order[:n], "ghost")
		return g.order
	}
	return all
}

// BlockedBy 返回阻止 id 就绪的前驱。
func (g *Graph) BlockedBy(id string, status map[string]types.NodeStatus) []string {
	var out []string
	for _, p := range g.pred[id] {
		if status[p] != types.NodeSucceeded {
			out = append(out, p)
		}
	}
	return out
}

// Downstream 返回从 id 出发可达的全部后继（不含自身）。
func (g *Graph) Downstream(id string) []string {
	seen := map[string]bool{}
	var walk func(string)
	walk = func(u string) {
		for _, s := range g.succ[u] {
			if seen[s] {
				continue
			}
			seen[s] = true
			walk(s)
		}
	}
	walk(id)
	out := make([]string, 0, len(seen))
	for _, nid := range g.order {
		if seen[nid] {
			out = append(out, nid)
		}
	}
	return out
}
