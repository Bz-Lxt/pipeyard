package types

import (
	"encoding/json"
	"fmt"
)

func (j Job) MarshalJSON() ([]byte, error) {
	type alias Job
	return json.Marshal(alias(j))
}

func DecodeSpecJSON(raw []byte) (Spec, error) {
	var s Spec
	if err := json.Unmarshal(raw, &s); err != nil {
		return Spec{}, fmt.Errorf("%w: %v", ErrBadSpec, err)
	}
	return s.Normalize()
}

func EncodeJob(j Job) ([]byte, error) {
	return json.Marshal(j.Clone())
}

func DecodeJob(raw []byte) (Job, error) {
	var j Job
	if err := json.Unmarshal(raw, &j); err != nil {
		return Job{}, err
	}
	return j.Clone(), nil
}

type NodeView struct {
	ID     string     `json:"id"`
	Kind   Kind       `json:"kind"`
	Status NodeStatus `json:"status"`
	Digest string     `json:"digest,omitempty"`
	Error  string     `json:"error,omitempty"`
}

func (j Job) Views() []NodeView {
	out := make([]NodeView, 0, len(j.Nodes))
	for _, n := range j.Nodes {
		v := NodeView{ID: n.ID, Kind: n.Kind, Status: n.Status, Error: n.Error}
		if n.Artifact != nil {
			v.Digest = string(n.Artifact.Digest)
		}
		out = append(out, v)
	}
	return out
}
