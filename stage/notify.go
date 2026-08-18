package stage

import (
	"context"
	"fmt"

	"github.com/Bz-Lxt/pipeyard/types"
)

// Notify 生成一条通知正文，内容来自上游或 param。
func Notify(ctx context.Context, req Request) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	body := parentBody(req.Parents)
	if body == "" {
		body = req.Node.Param
	}
	if body == "" {
		return Result{}, fmt.Errorf("notify: empty")
	}
	msg := "notify:" + body
	if req.Node.Param != "" && parentBody(req.Parents) != "" {
		msg = req.Node.Param + ":" + body
	}
	art := types.MakeArtifact(types.KindNotify, msg, []string{msg})
	return Result{Artifact: art}, nil
}
