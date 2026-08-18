package types

import (
	"encoding/json"
	"fmt"

	"github.com/Bz-Lxt/pipeyard/digest"
)

// Artifact 是阶段产物。Digest 为空视为无效，不得被当成成功。
type Artifact struct {
	Digest  digest.Digest
	Kind    Kind
	Body    string
	Items   []string
	Bytes   int
	Parents []digest.Digest
}

func (a Artifact) Clone() Artifact {
	out := a
	if a.Items != nil {
		out.Items = append([]string(nil), a.Items...)
	}
	if a.Parents != nil {
		out.Parents = append([]digest.Digest(nil), a.Parents...)
	}
	return out
}

func (a Artifact) Valid() bool {
	return a.Digest != "" && digest.Equal(a.Digest, a.Digest)
}

func (a Artifact) Marshal() ([]byte, error) {
	return json.Marshal(a)
}

func UnmarshalArtifact(raw []byte) (Artifact, error) {
	if len(raw) == 0 {
		return Artifact{}, fmt.Errorf("empty artifact")
	}
	var a Artifact
	if err := json.Unmarshal(raw, &a); err != nil {
		return Artifact{}, err
	}
	if !a.Valid() {
		return Artifact{}, fmt.Errorf("artifact missing digest")
	}
	return a, nil
}

// Merge 把多个上游产物拼成 join 输入。空指针被跳过，不得当成成功产物。
func Merge(kind Kind, parts []*Artifact) Artifact {
	var items []string
	var parents []digest.Digest
	var n int
	for _, p := range parts {
		if p == nil || !p.Valid() {
			continue
		}
		items = append(items, p.Items...)
		if p.Body != "" {
			items = append(items, p.Body)
		}
		parents = append(parents, p.Digest)
		n += p.Bytes
	}
	body := ""
	if len(items) > 0 {
		body = items[0]
	}
	a := Artifact{Kind: kind, Body: body, Items: items, Bytes: n, Parents: parents}
	a.Digest = digest.Pair(string(kind), []byte(body+"|"+fmt.Sprintf("%d", len(items))))
	return a
}

func MakeArtifact(kind Kind, body string, items []string) Artifact {
	if items == nil {
		items = []string{}
	}
	a := Artifact{Kind: kind, Body: body, Items: append([]string(nil), items...), Bytes: len(body)}
	a.Digest = digest.Pair(string(kind)+"|"+body, []byte(fmt.Sprintf("%d", len(items))))
	return a
}
