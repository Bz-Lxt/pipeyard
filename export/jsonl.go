package export

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/Bz-Lxt/pipeyard/types"
)

func JobJSONL(jobs []types.Job) string {
	var b bytes.Buffer
	for _, j := range jobs {
		row := map[string]any{
			"id":     string(j.ID),
			"name":   j.Name,
			"status": string(j.Status),
			"nodes":  len(j.Nodes),
			"edges":  len(j.Edges),
		}
		raw, err := json.Marshal(row)
		if err != nil {
			fmt.Fprintf(&b, "{\"error\":%q}\n", err.Error())
			continue
		}
		b.Write(raw)
		b.WriteByte('\n')
	}
	return b.String()
}

func NodeJSONL(j types.Job) string {
	var b bytes.Buffer
	for _, n := range j.Nodes {
		row := map[string]any{
			"job":    string(j.ID),
			"node":   n.ID,
			"kind":   string(n.Kind),
			"status": string(n.Status),
			"error":  n.Error,
		}
		if n.Artifact != nil {
			row["digest"] = string(n.Artifact.Digest)
		}
		raw, _ := json.Marshal(row)
		b.Write(raw)
		b.WriteByte('\n')
	}
	return b.String()
}

func SummariesJSON(ss []types.Summary) ([]byte, error) {
	return json.Marshal(ss)
}
