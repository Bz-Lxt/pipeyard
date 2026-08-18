package stage

import (
	"context"
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/Bz-Lxt/pipeyard/types"
)

// Validate 检查输入是否为合法 UTF-8；param 写成「deny:片段」时会拒绝命中片段的正文。
func Validate(ctx context.Context, req Request) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	body := req.Node.Param
	if src := parentBody(req.Parents); src != "" {
		body = src
	}
	if body == "" {
		return Result{}, fmt.Errorf("validate: empty input")
	}
	if !utf8.ValidString(body) {
		return Result{}, fmt.Errorf("validate: invalid utf8")
	}
	if strings.HasPrefix(req.Node.Param, "deny:") {
		needle := strings.TrimPrefix(req.Node.Param, "deny:")
		if needle != "" && strings.Contains(body, needle) {
			return Result{}, fmt.Errorf("validate: denied substring %q", needle)
		}
	}
	if req.Node.Param == "fail" {
		return Result{}, fmt.Errorf("validate: forced fail")
	}
	art := types.MakeArtifact(types.KindValidate, body, []string{body})
	return Result{Artifact: art}, nil
}
