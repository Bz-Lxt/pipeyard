// Package schedule 从就绪集里按权重与公平性挑节点。
package schedule

import "github.com/Bz-Lxt/pipeyard/types"

type Candidate struct {
	Job   string
	Node  string
	Weight int
	Wait  int
}

func Pick(cands []Candidate, n int) []Candidate {
	if n <= 0 || n >= len(cands) {
		return append([]Candidate(nil), cands...)
	}
	ranked := append([]Candidate(nil), cands...)
	for i := 0; i < len(ranked); i++ {
		for j := i + 1; j < len(ranked); j++ {
			if less(ranked[j], ranked[i]) {
				ranked[i], ranked[j] = ranked[j], ranked[i]
			}
		}
	}
	return append([]Candidate(nil), ranked[:n]...)
}

func less(a, b Candidate) bool {
	if a.Wait != b.Wait {
		return a.Wait > b.Wait
	}
	if a.Weight != b.Weight {
		return a.Weight > b.Weight
	}
	if a.Job != b.Job {
		return a.Job < b.Job
	}
	return a.Node < b.Node
}

func FromJob(job types.Job, ready []string) []Candidate {
	var out []Candidate
	for _, id := range ready {
		n, ok := job.Node(id)
		if !ok {
			continue
		}
		out = append(out, Candidate{Job: string(job.ID), Node: id, Weight: n.Weight, Wait: n.Attempt})
	}
	return out
}
