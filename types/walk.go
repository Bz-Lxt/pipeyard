package types

func WalkReady(nodes []Node, preds map[string][]string) []string {
	st := map[string]NodeStatus{}
	for _, n := range nodes {
		st[n.ID] = n.Status
	}
	var out []string
	for _, n := range nodes {
		if n.Status.Terminal() || n.Status == NodeLeased {
			continue
		}
		ok := true
		for _, p := range preds[n.ID] {
			if st[p] != NodeSucceeded {
				ok = false
				break
			}
		}
		if ok {
			out = append(out, n.ID)
		}
	}
	return out
}

func IndexNodes(nodes []Node) map[string]Node {
	m := make(map[string]Node, len(nodes))
	for _, n := range nodes {
		m[n.ID] = n
	}
	return m
}

func PredMap(edges []Edge) map[string][]string {
	m := map[string][]string{}
	for _, e := range edges {
		m[e.To] = append(m[e.To], e.From)
	}
	return m
}

func SuccMap(edges []Edge) map[string][]string {
	m := map[string][]string{}
	for _, e := range edges {
		m[e.From] = append(m[e.From], e.To)
	}
	return m
}
