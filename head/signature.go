package head

import (
	"io"
)

type signature struct {
	header *Header
}

func newSignature() *signature {
	s := &signature{}
	s.header = new(Header)
	s.header.dbHeader = newDbHeader()
	s.put(binaryEntry(0x3f0, make([]byte, 4096)))                         // Reserved, lol.
	s.put(stringEntry(0x10d, "xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx")) // SHA1, needs to be re-written later.
	s.put(uint32Entry(0x3e8, 0))                                          // 1000 = SIZE/RPMSIGTAG_SIZE is the byte count for the header section plus the compressed payload. Needs to be re-written later.
	s.put(binaryEntry(0x3ec, make([]byte, 16)))                           // MD5, needs to be re-written later.
	s.put(uint32Entry(0x3ef, 0))                                          // PAYLOADSIZE/RPMSIGTAG_PAYLOADSIZE is the byte count for the uncompressed payload. Needs to be re-written later.
	s.put(binaryEntry(0x3e, headerSignatures(len(s.header.entries)+1)))
	s.header.normalize()
	return s
}

func headerSignatures(l int) []byte {
	hs := make([]byte, 0)
	hs = append(hs, uint32ToBytes(62)...)
	hs = append(hs, uint32ToBytes(7)...)
	x := l * 16
	r := x ^ 0xffffffff
	hs = append(hs, uint32ToBytes(uint32(r+1))...)
	hs = append(hs, uint32ToBytes(16)...)
	return hs
}

func (s *signature) put(e *dbEntry) {

	es := s.header.entries[:]
	for i, en := range es {
		if en.index.tag == e.index.tag {
			if en.index.count != e.index.count {
				panic("cannot overwrite entry, count of index does not match")
			}
			if en.index.dataType != e.index.dataType {
				panic("cannot overwrite entry, data type of index does not match")
			}
			if en.payload.size() != e.payload.size() {
				panic("cannot overwrite entry, sizes of payload data do not match")
			}
			s.header.entries[i] = e
			return
		}
	}
	s.header.dbHeader.nIndex++
	s.header.dbHeader.hSize += e.payload.size()
	s.header.entries = append(s.header.entries, e)
}

func (s *signature) WriteTo(w io.Writer) (int64, error) {
	total := int64(0)
	n, err := s.header.dbHeader.WriteTo(w)
	total += n
	if err != nil {
		return total, err
	}

	for _, e := range s.header.entries {
		n, err := w.Write(uint32ToBytes(e.index.tag))
		total += int64(n)
		if err != nil {
			return total, err
		}
		n, err = w.Write(uint32ToBytes(e.index.dataType))
		total += int64(n)
		if err != nil {
			return total, err
		}
		n, err = w.Write(uint32ToBytes(e.index.offset))
		total += int64(n)
		if err != nil {
			return total, err
		}
		n, err = w.Write(uint32ToBytes(e.index.count))
		total += int64(n)
		if err != nil {
			return total, err
		}
	}

	for _, e := range s.header.entries {
		n, err := e.payload.WriteTo(w)
		total += int64(n)
		if err != nil {
			return total, err
		}
	}

	return total, nil
}
