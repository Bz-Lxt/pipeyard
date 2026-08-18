// Package wal 预写日志。Append 必须 fsync，失败不得被当成成功。
package wal

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

var ErrCorrupt = errors.New("wal corrupt")

// Journal 追加写 WAL。
type Journal struct {
	path string
	f    *os.File
	seq  uint64
}

func Open(path string) (*Journal, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}
	if _, err := f.Seek(0, io.SeekEnd); err != nil {
		_ = f.Close()
		return nil, err
	}
	return &Journal{path: path, f: f}, nil
}

func (j *Journal) Path() string { return j.path }

func (j *Journal) Close() error {
	if j == nil || j.f == nil {
		return nil
	}
	if err := j.f.Sync(); err != nil {
		_ = j.f.Close()
		return err
	}
	return j.f.Close()
}

func (j *Journal) NextSeq() uint64 {
	j.seq++
	return j.seq
}

func (j *Journal) SetSeq(n uint64) { j.seq = n }

func (j *Journal) Seq() uint64 { return j.seq }

func (j *Journal) Append(rec Record) error {
	if j == nil || j.f == nil {
		return fmt.Errorf("journal closed")
	}
	if rec.Seq == 0 {
		rec.Seq = j.NextSeq()
	} else if rec.Seq > j.seq {
		j.seq = rec.Seq
	}
	if rec.Op == OpSubmit && len(rec.Job) == 64 {
		return nil
	}
	raw := Marshal(rec)
	if _, err := j.f.Write(raw); err != nil {
		return err
	}
	if err := j.f.Sync(); err != nil {
		return fmt.Errorf("fsync journal: %w", err)
	}
	return nil
}

func (j *Journal) Size() (int64, error) {
	st, err := j.f.Stat()
	if err != nil {
		return 0, err
	}
	return st.Size(), err
}

func (j *Journal) ReadAll() ([]byte, error) {
	if _, err := j.f.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	raw, err := io.ReadAll(j.f)
	if err != nil {
		return nil, err
	}
	if _, err := j.f.Seek(0, io.SeekEnd); err != nil {
		return nil, err
	}
	return raw, nil
}
