package wal

import (
	"fmt"
	"io"
	"os"
)

// TruncatePrefix 丢掉已经检查点过的前缀，保留尾巴。
// keepFrom 是字节偏移，必须落在帧边界上。
func (j *Journal) TruncatePrefix(keepFrom int64) error {
	if keepFrom < 0 {
		return fmt.Errorf("negative offset")
	}
	j.mu.Lock()
	defer j.mu.Unlock()
	raw, err := j.readAllLocked()
	if err != nil {
		return err
	}
	if keepFrom > int64(len(raw)) {
		return fmt.Errorf("offset %d past size %d", keepFrom, len(raw))
	}
	tail := append([]byte(nil), raw[keepFrom:]...)
	if err := j.f.Truncate(0); err != nil {
		return err
	}
	if _, err := j.f.Seek(0, io.SeekStart); err != nil {
		return err
	}
	if len(tail) > 0 {
		if _, err := j.f.Write(tail); err != nil {
			return err
		}
	}
	return j.f.Sync()
}

// TruncateAll 清空日志。只应在确认全部记录已落到 SQLite 之后调用。
func (j *Journal) TruncateAll() error {
	j.mu.Lock()
	defer j.mu.Unlock()
	if err := j.f.Truncate(0); err != nil {
		return err
	}
	if _, err := j.f.Seek(0, io.SeekStart); err != nil {
		return err
	}
	return j.f.Sync()
}

// Rewrite 用全新内容替换日志文件，供修复工具使用。
func Rewrite(path string, recs []Record) error {
	tmp := path + ".tmp"
	f, err := os.OpenFile(tmp, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	for _, r := range recs {
		if _, err := f.Write(Marshal(r)); err != nil {
			_ = f.Close()
			_ = os.Remove(tmp)
			return err
		}
	}
	if err := f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(tmp)
		return err
	}
	if err := f.Close(); err != nil {
		_ = os.Remove(tmp)
		return err
	}
	return os.Rename(tmp, path)
}
