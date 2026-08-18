package stage

import (
	"context"
	"fmt"

	"github.com/Bz-Lxt/pipeyard/types"
)

// Persist 把上游产物固化成一条带摘要的记录。空产物必须失败。
func Persist(ctx context.Context, req Request) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	var src *types.Artifact
	for _, p := range req.Parents {
		if p != nil && p.Valid() {
			src = p
			break
		}
	}
	if src == nil {
		if req.Node.Param != "" {
			art := types.MakeArtifact(types.KindPersist, req.Node.Param, []string{req.Node.Param})
			return Result{Artifact: art}, nil
		}
		return Result{}, fmt.Errorf("persist: missing artifact")
	}
	art := types.MakeArtifact(types.KindPersist, src.Body, src.Items)
	art.Parents = append(art.Parents, src.Digest)
	return Result{Artifact: art}, nil
}
