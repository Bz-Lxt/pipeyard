package stage

import (
	"context"
	"fmt"

	"github.com/Bz-Lxt/pipeyard/types"
)

// Join 合并全部有效上游产物。若没有任何有效产物则失败，不得返回空指针当成功。
func Join(ctx context.Context, req Request) (Result, error) {
	art := types.Merge(types.KindJoin, req.Parents)
	if !art.Valid() || len(art.Parents) == 0 {
		return Result{}, fmt.Errorf("join: no valid parents")
	}
	if req.Node.Param == "require" && len(art.Parents) < 2 {
		return Result{}, fmt.Errorf("join: need at least two parents")
	}
	return Result{Artifact: art}, nil
}
