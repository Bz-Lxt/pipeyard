package stage

import (
	"fmt"

	"github.com/Bz-Lxt/pipeyard/textutil"
	"github.com/Bz-Lxt/pipeyard/types"
)

func ValidParents(parents []*types.Artifact) []*types.Artifact {
	var out []*types.Artifact
	for _, p := range parents {
		if p != nil && p.Valid() {
			out = append(out, p)
		}
	}
	return out
}

func FirstBody(parents []*types.Artifact, fallback string) string {
	if b := parentBody(parents); b != "" {
		return textutil.Compact(b)
	}
	return textutil.Compact(fallback)
}

func RequireParents(parents []*types.Artifact, n int) error {
	got := len(ValidParents(parents))
	if got < n {
		return fmt.Errorf("need %d valid parents, got %d", n, got)
	}
	return nil
}
