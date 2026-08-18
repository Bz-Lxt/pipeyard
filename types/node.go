package types

import "github.com/Bz-Lxt/pipeyard/digest"

// Node 是作业内一个阶段的运行时视图。
type Node struct {
	ID        string
	Kind      Kind
	Param     string
	Weight    int
	Status    NodeStatus
	LeaseBy   string
	LeaseUntil string
	Error     string
	Artifact  *Artifact
	Attempt   int
}

func (n Node) Clone() Node {
	out := n
	if n.Artifact != nil {
		a := n.Artifact.Clone()
		out.Artifact = &a
	}
	return out
}

func (n Node) Digest() digest.Digest {
	payload := n.Param
	if n.Artifact != nil {
		payload += "|" + string(n.Artifact.Digest)
	}
	return digest.Pair(n.ID+"|"+string(n.Kind)+"|"+string(n.Status), []byte(payload))
}

func (n Node) HasArtifact() bool {
	return n.Artifact != nil && n.Artifact.Digest != ""
}

func CloneNodes(in []Node) []Node {
	if in == nil {
		return nil
	}
	out := make([]Node, len(in))
	for i := range in {
		out[i] = in[i].Clone()
	}
	return out
}
