package catalog

import (
	"fmt"
	"strings"

	"github.com/Bz-Lxt/pipeyard/types"
)

func Describe(k types.Kind) string {
	info, ok := Lookup(k)
	if !ok {
		return string(k)
	}
	need := "无上游也可跑"
	if info.NeedsParent {
		need = "需要有效上游产物"
	}
	return fmt.Sprintf("%s（%s），产出 %s，%s", info.Title, info.Kind, info.Produces, need)
}

func DescribeSpec(spec types.Spec) string {
	var b strings.Builder
	fmt.Fprintf(&b, "作业 %s，%d 个节点，%d 条边", spec.Name, len(spec.Nodes), len(spec.Edges))
	for _, n := range spec.Nodes {
		fmt.Fprintf(&b, "\n- %s: %s", n.ID, Describe(n.Kind))
	}
	return b.String()
}

func Known(k string) bool {
	return types.KnownKind(types.Kind(k))
}
