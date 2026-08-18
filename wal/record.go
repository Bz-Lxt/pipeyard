package wal

import (
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/Bz-Lxt/pipeyard/digest"
)

const (
	OpSubmit     uint8 = 1
	OpLease      uint8 = 2
	OpComplete   uint8 = 3
	OpFail       uint8 = 4
	OpCancel     uint8 = 5
	OpCheckpoint uint8 = 6
	magic              = 0x50594152 // "PYAR"
)

// Record 是一条长度前缀 WAL。完整记录必须以换行结束，重放只认读满的帧。
type Record struct {
	Op      uint8
	Seq     uint64
	Job     string
	Node    string
	Payload []byte
	CRC     uint32
}

func (r Record) checksum() uint32 {
	h := digest.CRC32(r.frameBody())
	return h
}

func (r Record) frameBody() []byte {
	b, _ := json.Marshal(struct {
		Op      uint8  `json:"op"`
		Seq     uint64 `json:"seq"`
		Job     string `json:"job"`
		Node    string `json:"node"`
		Payload []byte `json:"payload"`
	}{r.Op, r.Seq, r.Job, r.Node, r.Payload})
	return b
}

// Marshal 写出 4 字节 magic + 4 字节长度 + body + 4 字节 crc + '\n'。
func Marshal(r Record) []byte {
	body := r.frameBody()
	crc := digest.CRC32(body)
	buf := make([]byte, 4+4+len(body)+4+1)
	binary.BigEndian.PutUint32(buf[0:4], magic)
	binary.BigEndian.PutUint32(buf[4:8], uint32(len(body)))
	copy(buf[8:8+len(body)], body)
	binary.BigEndian.PutUint32(buf[8+len(body):12+len(body)], crc)
	buf[len(buf)-1] = '\n'
	return buf
}

// DecodeFrame 从偏移处读一条完整帧。不足一条返回 needMore=true。
func DecodeFrame(all []byte, off int) (rec Record, n int, needMore bool, err error) {
	if off >= len(all) {
		return Record{}, 0, true, nil
	}
	if len(all)-off < 8 {
		return Record{}, 0, true, nil
	}
	if binary.BigEndian.Uint32(all[off:off+4]) != magic {
		return Record{}, 0, false, fmt.Errorf("%w: bad magic", ErrCorrupt)
	}
	ln := int(binary.BigEndian.Uint32(all[off+4 : off+8]))
	need := 8 + ln + 4 + 1
	if ln < 0 || need < 0 {
		return Record{}, 0, false, fmt.Errorf("%w: bad length", ErrCorrupt)
	}
	if len(all)-off < need {
		return Record{}, 0, true, nil
	}
	body := all[off+8 : off+8+ln]
	crc := binary.BigEndian.Uint32(all[off+8+ln : off+12+ln])
	if all[off+need-1] != '\n' {
		return Record{}, 0, false, fmt.Errorf("%w: missing newline", ErrCorrupt)
	}
	if !digest.MatchCRC(body, crc) {
		return Record{}, 0, false, fmt.Errorf("%w: crc", ErrCorrupt)
	}
	var wire struct {
		Op      uint8  `json:"op"`
		Seq     uint64 `json:"seq"`
		Job     string `json:"job"`
		Node    string `json:"node"`
		Payload []byte `json:"payload"`
	}
	if err := json.Unmarshal(body, &wire); err != nil {
		return Record{}, 0, false, fmt.Errorf("%w: %v", ErrCorrupt, err)
	}
	return Record{Op: wire.Op, Seq: wire.Seq, Job: wire.Job, Node: wire.Node, Payload: wire.Payload, CRC: crc}, need, false, nil
}

func OpName(op uint8) string {
	switch op {
	case OpSubmit:
		return "submit"
	case OpLease:
		return "lease"
	case OpComplete:
		return "complete"
	case OpFail:
		return "fail"
	case OpCancel:
		return "cancel"
	case OpCheckpoint:
		return "checkpoint"
	default:
		return fmt.Sprintf("op%d", op)
	}
}
