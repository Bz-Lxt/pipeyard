package stage

import (
	"context"
	"fmt"
	"strings"

	"github.com/Bz-Lxt/pipeyard/types"
)

// Reduce 用 param 作为分隔符拼接条目，默认 "|"。
func Reduce(ctx context.Context, req Request) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	items := parentItems(req.Parents)
	if len(items) == 0 {
		return Result{}, fmt.Errorf("reduce: no input")
	}
	sep := req.Node.Param
	if sep == "" {
		sep = "|"
	}
	body := strings.Join(items, sep)
	art := types.MakeArtifact(types.KindReduce, body, []string{body})
	return Result{Artifact: art}, nil
}
