package types

import "github.com/Bz-Lxt/pipeyard/digest"

// Job 是作业的完整视图，Get / List 必须返回副本。
type Job struct {
	ID        digest.Digest
	Name      string
	Status    JobStatus
	CreatedAt string
	UpdatedAt string
	Nodes     []Node
	Edges     []Edge
	Quota     int
	Error     string
}

func (j Job) Clone() Job {
	out := j
	out.Nodes = CloneNodes(j.Nodes)
	if j.Edges != nil {
		out.Edges = append([]Edge(nil), j.Edges...)
	}
	return out
}

func CloneJobs(in []Job) []Job {
	if in == nil {
		return nil
	}
	out := make([]Job, len(in))
	for i := range in {
		out[i] = in[i].Clone()
	}
	return out
}

func (j Job) Node(id string) (Node, bool) {
	for _, n := range j.Nodes {
		if n.ID == id {
			return n, true
		}
	}
	return Node{}, false
}

func (j Job) NodeStatuses() []NodeStatus {
	out := make([]NodeStatus, len(j.Nodes))
	for i, n := range j.Nodes {
		out[i] = n.Status
	}
	return out
}

func (j *Job) RefreshStatus() {
	j.Status = Rollup(j.NodeStatuses())
}

// Summary 给面板用的短信息。
type Summary struct {
	ID     digest.Digest
	Name   string
	Status JobStatus
	Ready  int
	Done   int
	Total  int
}

func (j Job) Summary() Summary {
	var ready, done int
	for _, n := range j.Nodes {
		if n.Status.Runnable() {
			ready++
		}
		if n.Status.Terminal() {
			done++
		}
	}
	return Summary{ID: j.ID, Name: j.Name, Status: j.Status, Ready: ready, Done: done, Total: len(j.Nodes)}
}
