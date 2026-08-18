package stage

import (
	"fmt"
	"strings"

	"github.com/Bz-Lxt/pipeyard/types"
)

func Trace(req Request, res Result, err error) string {
	var b strings.Builder
	fmt.Fprintf(&b, "job=%s node=%s kind=%s parents=%d", req.JobID, req.Node.ID, req.Node.Kind, len(req.Parents))
	if err != nil {
		fmt.Fprintf(&b, " err=%v", err)
		return b.String()
	}
	fmt.Fprintf(&b, " digest=%s bytes=%d items=%d", res.Artifact.Digest, res.Artifact.Bytes, len(res.Artifact.Items))
	return b.String()
}

func ParentKinds(parents []*types.Artifact) []types.Kind {
	var out []types.Kind
	for _, p := range ValidParents(parents) {
		out = append(out, p.Kind)
	}
	return out
}

func SameKindParents(parents []*types.Artifact, k types.Kind) bool {
	n := 0
	for _, p := range ValidParents(parents) {
		if p.Kind != k {
			return false
		}
		n++
	}
	return n > 0
}
