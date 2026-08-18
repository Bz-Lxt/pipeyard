package wal

import (
	"path/filepath"
	"testing"
)

func TestAppendReplay(t *testing.T) {
	p := filepath.Join(t.TempDir(), "j.wal")
	j, err := Open(p)
	if err != nil {
		t.Fatal(err)
	}
	if err := j.Append(Record{Op: OpSubmit, Job: "j1", Payload: []byte("a")}); err != nil {
		t.Fatal(err)
	}
	if err := j.Append(Record{Op: OpComplete, Job: "j1", Node: "n"}); err != nil {
		t.Fatal(err)
	}
	raw, err := j.ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	_ = j.Close()
	recs, err := Replay(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 2 {
		t.Fatalf("got %d", len(recs))
	}
	if recs[1].Op != OpComplete {
		t.Fatalf("op %d", recs[1].Op)
	}
}

func TestPartialTailDropped(t *testing.T) {
	frame := Marshal(Record{Op: OpSubmit, Seq: 1, Job: "x"})
	raw := append(frame, frame[:6]...)
	recs, err := Replay(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(recs) != 1 {
		t.Fatalf("got %d", len(recs))
	}
}
