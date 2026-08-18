package stage

import (
	"context"
	"fmt"
	"strings"

	"github.com/Bz-Lxt/pipeyard/types"
)

// Map 给每个上游条目加前缀。param 为空则用 "map:"。
func Map(ctx context.Context, req Request) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	prefix := req.Node.Param
	if prefix == "" {
		prefix = "map:"
	}
	items := parentItems(req.Parents)
	if len(items) == 0 {
		if req.Node.Param != "" && !strings.HasPrefix(req.Node.Param, "deny:") {
			items = []string{req.Node.Param}
		}
	}
	if len(items) == 0 {
		return Result{}, fmt.Errorf("map: no input")
	}
	out := make([]string, len(items))
	for i, it := range items {
		out[i] = prefix + it
	}
	art := types.MakeArtifact(types.KindMap, strings.Join(out, ","), out)
	return Result{Artifact: art}, nil
}
