package stage

import (
	"context"
	"fmt"
	"strings"

	"github.com/Bz-Lxt/pipeyard/types"
)

// Filter 保留包含 param 的条目。param 为空则原样通过。
func Filter(ctx context.Context, req Request) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	items := parentItems(req.Parents)
	if len(items) == 0 {
		return Result{}, fmt.Errorf("filter: no input")
	}
	keep := req.Node.Param
	var out []string
	for _, it := range items {
		if keep == "" || strings.Contains(it, keep) {
			out = append(out, it)
		}
	}
	if out == nil {
		out = []string{}
	}
	art := types.MakeArtifact(types.KindFilter, strings.Join(out, ","), out)
	return Result{Artifact: art}, nil
}
