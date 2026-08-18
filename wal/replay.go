package wal

// Replay 解码文件中所有完整帧。尾部半截记录被丢弃，不得把半截当成功。
func Replay(raw []byte) ([]Record, error) {
	var out []Record
	off := 0
	for off < len(raw) {
		rec, n, needMore, err := DecodeFrame(raw, off)
		if err != nil {
			return out, err
		}
		if needMore {
			break
		}
		out = append(out, rec)
		off += n
	}
	return out, nil
}

// ReplayFile 打开只读文件重放。
func ReplayFile(path string) ([]Record, error) {
	j, err := Open(path)
	if err != nil {
		return nil, err
	}
	defer j.Close()
	raw, err := j.ReadAll()
	if err != nil {
		return nil, err
	}
	return Replay(raw)
}

// LastSeq 返回重放结果中最大序号。
func LastSeq(recs []Record) uint64 {
	var max uint64
	for _, r := range recs {
		if r.Seq > max {
			max = r.Seq
		}
	}
	return max
}

// AppliedPrefix 返回序号 <= applied 的记录条数对应的字节前缀长度。
func AppliedPrefix(raw []byte, applied uint64) (int, error) {
	off := 0
	for off < len(raw) {
		rec, n, needMore, err := DecodeFrame(raw, off)
		if err != nil {
			return off, err
		}
		if needMore {
			break
		}
		if rec.Seq > applied {
			return off, nil
		}
		off += n
	}
	return off, nil
}
