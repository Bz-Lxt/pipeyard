package pipeyard_test

import (
	"testing"

	"github.com/Bz-Lxt/pipeyard/wal"
)

func TestReplayDropsPartialFrame(t *testing.T) {
	frame := wal.Marshal(wal.Record{
		Op:      wal.OpSubmit,
		Seq:     1,
		Job:     "origjob",
		Payload: []byte(`{"Name":"orig","Nodes":[{"ID":"a","Kind":"validate","Param":"x"}]}`),
	})
	raw := append(append([]byte{}, frame...), []byte("PYAR")...)
	recs, err := wal.Replay(raw)
	if err != nil {
		t.Fatal(err)
	}
	if err := wal.Verify(recs); err != nil {
		t.Fatalf("verify: %v", err)
	}
	if len(recs) != 1 {
		t.Fatalf("replay kept extra submit: n=%d", len(recs))
	}
	if recs[0].Job != "origjob" {
		t.Fatalf("job %s", recs[0].Job)
	}
}
