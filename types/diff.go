package types

func DiffStatus(a, b Job) []string {
	var out []string
	if a.Status != b.Status {
		out = append(out, "job:"+string(a.Status)+"->"+string(b.Status))
	}
	am := map[string]NodeStatus{}
	for _, n := range a.Nodes {
		am[n.ID] = n.Status
	}
	for _, n := range b.Nodes {
		if am[n.ID] != n.Status {
			out = append(out, n.ID+":"+string(am[n.ID])+"->"+string(n.Status))
		}
	}
	return out
}

func SameGraph(a, b Job) bool {
	if len(a.Nodes) != len(b.Nodes) || len(a.Edges) != len(b.Edges) {
		return false
	}
	for i := range a.Nodes {
		if a.Nodes[i].ID != b.Nodes[i].ID || a.Nodes[i].Kind != b.Nodes[i].Kind {
			return false
		}
	}
	for i := range a.Edges {
		if a.Edges[i] != b.Edges[i] {
			return false
		}
	}
	return true
}

func FailedNodes(j Job) []string {
	var out []string
	for _, n := range j.Nodes {
		if n.Status == NodeFailed {
			out = append(out, n.ID)
		}
	}
	return out
}

func SucceededNodes(j Job) []string {
	var out []string
	for _, n := range j.Nodes {
		if n.Status == NodeSucceeded {
			out = append(out, n.ID)
		}
	}
	return out
}
