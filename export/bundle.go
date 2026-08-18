// Package export 把作业图与产物打成可下载的文本包。
package export

import (
	"bytes"
	"fmt"
	"strings"

	"github.com/Bz-Lxt/pipeyard/types"
)

func JobText(j types.Job) string {
	var b bytes.Buffer
	fmt.Fprintf(&b, "job %s name=%s status=%s\n", j.ID, j.Name, j.Status)
	for _, n := range j.Nodes {
		fmt.Fprintf(&b, "  node %s kind=%s status=%s\n", n.ID, n.Kind, n.Status)
		if n.Artifact != nil {
			fmt.Fprintf(&b, "    artifact %s bytes=%d\n", n.Artifact.Digest, n.Artifact.Bytes)
		}
		if n.Error != "" {
			fmt.Fprintf(&b, "    error %s\n", n.Error)
		}
	}
	for _, e := range j.Edges {
		fmt.Fprintf(&b, "  edge %s -> %s\n", e.From, e.To)
	}
	return b.String()
}

func JobsText(jobs []types.Job) string {
	parts := make([]string, 0, len(jobs))
	for _, j := range jobs {
		parts = append(parts, JobText(j))
	}
	return strings.Join(parts, "\n")
}

func SummaryLine(s types.Summary) string {
	return fmt.Sprintf("%s %s %s ready=%d done=%d/%d", s.ID, s.Name, s.Status, s.Ready, s.Done, s.Total)
}
