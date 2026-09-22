package steam

import (
	"encoding/binary"
	"fmt"
	"io"
)

type pbField struct {
	num  int
	wire int
	u    uint64
	b    []byte
}

func pbAppendVarintField(dst []byte, num int, v uint64) []byte {
	dst = pbAppendTag(dst, num, 0)
	return pbAppendUvarint(dst, v)
}

func pbAppendFixed64Field(dst []byte, num int, v uint64) []byte {
	dst = pbAppendTag(dst, num, 1)
	var buf [8]byte
	binary.LittleEndian.PutUint64(buf[:], v)
	return append(dst, buf[:]...)
}

func pbAppendBytesField(dst []byte, num int, b []byte) []byte {
	dst = pbAppendTag(dst, num, 2)
	dst = pbAppendUvarint(dst, uint64(len(b)))
	return append(dst, b...)
}

func pbAppendStringField(dst []byte, num int, s string) []byte {
	return pbAppendBytesField(dst, num, []byte(s))
}

func pbAppendTag(dst []byte, num, wire int) []byte {
	return pbAppendUvarint(dst, uint64(num<<3|wire))
}

func pbAppendUvarint(dst []byte, v uint64) []byte {
	var buf [binary.MaxVarintLen64]byte
	n := binary.PutUvarint(buf[:], v)
	return append(dst, buf[:n]...)
}

func pbParse(b []byte) ([]pbField, error) {
	var fields []pbField
	for len(b) > 0 {
		key, n := binary.Uvarint(b)
		if n <= 0 {
			return nil, fmt.Errorf("pbwire: bad field key")
		}
		b = b[n:]
		f := pbField{num: int(key >> 3), wire: int(key & 7)}
		switch f.wire {
		case 0:
			v, m := binary.Uvarint(b)
			if m <= 0 {
				return nil, fmt.Errorf("pbwire: bad varint for field %d", f.num)
			}
			f.u = v
			b = b[m:]
		case 1:
			if len(b) < 8 {
				return nil, io.ErrUnexpectedEOF
			}
			f.u = binary.LittleEndian.Uint64(b)
			b = b[8:]
		case 2:
			l, m := binary.Uvarint(b)
			if m <= 0 {
				return nil, fmt.Errorf("pbwire: bad length for field %d", f.num)
			}
			b = b[m:]
			if uint64(len(b)) < l {
				return nil, io.ErrUnexpectedEOF
			}
			f.b = b[:l]
			b = b[l:]
		case 5:
			if len(b) < 4 {
				return nil, io.ErrUnexpectedEOF
			}
			f.u = uint64(binary.LittleEndian.Uint32(b))
			b = b[4:]
		default:
			return nil, fmt.Errorf("pbwire: unsupported wire type %d", f.wire)
		}
		fields = append(fields, f)
	}
	return fields, nil
}

func pbFindVarint(fields []pbField, num int) (uint64, bool) {
	for _, f := range fields {
		if f.num == num && (f.wire == 0 || f.wire == 1 || f.wire == 5) {
			return f.u, true
		}
	}
	return 0, false
}

func pbFindBytes(fields []pbField, num int) ([]byte, bool) {
	for _, f := range fields {
		if f.num == num && f.wire == 2 {
			return f.b, true
		}
	}
	return nil, false
}
