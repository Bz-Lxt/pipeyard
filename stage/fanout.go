package stage

import (
	"context"
	"fmt"
	"strings"

	"github.com/Bz-Lxt/pipeyard/types"
)

// Fanout 按逗号或 param 指定分隔符切开正文。
func Fanout(ctx context.Context, req Request) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	body := parentBody(req.Parents)
	if body == "" {
		body = req.Node.Param
	}
	if body == "" {
		return Result{}, fmt.Errorf("fanout: empty")
	}
	sep := ","
	src := body
	if req.Node.Param != "" && parentBody(req.Parents) != "" {
		sep = req.Node.Param
	}
	parts := strings.Split(src, sep)
	clean := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			clean = append(clean, p)
		}
	}
	if len(clean) == 0 {
		return Result{}, fmt.Errorf("fanout: no parts")
	}
	art := types.MakeArtifact(types.KindFanout, strings.Join(clean, ","), clean)
	return Result{Artifact: art}, nil
}
