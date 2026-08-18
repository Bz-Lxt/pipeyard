package repair

import (
	"fmt"

	"github.com/Bz-Lxt/pipeyard/wal"
)

func Gap(applied, walSeq uint64) int64 {
	if walSeq >= applied {
		return int64(walSeq - applied)
	}
	return -int64(applied - walSeq)
}

func NeedsReplay(applied, walSeq uint64) bool {
	return walSeq > applied
}

func Summarize(recs []wal.Record, applied uint64) string {
	return fmt.Sprintf("records=%d last=%d applied=%d gap=%d",
		len(recs), wal.LastSeq(recs), applied, Gap(applied, wal.LastSeq(recs)))
}
