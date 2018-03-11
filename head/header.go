package head

import (
	"fmt"
	"io"
	"sort"
)

type dbHeader struct {
	magic    []byte // must be "\216\255\350" = "0x8e, 0xad, 0xe8"
	version  uint8  // usually 1
	reserved []byte // must be "\0\0\0\0"
	nIndex   uint32 // number of index records
	hSize    uint32 // size of storage area for data
}

func newDbHeader() *dbHeader {
	h := dbHeader{
		magic:    []byte{0x8e, 0xad, 0xe8},
		version:  1,
		reserved: []byte{0x0, 0x0, 0x0, 0x0},
	}
	return &h
}

func (h *dbHeader) WriteTo(w io.Writer) (int64, error) {
	total := int64(0)
	n, err := w.Write(h.magic)
	total += int64(n)
	if err != nil {
		return total, err
	}
	n, err = w.Write([]byte{h.version})
	total += int64(n)
	if err != nil {
		return total, err
	}
	n, err = w.Write(h.reserved)
	total += int64(n)
	if err != nil {
		return total, err
	}

	n, err = w.Write(uint32ToBytes(h.nIndex))
	total += int64(n)
	if err != nil {
		return total, err
	}

	n, err = w.Write(uint32ToBytes(h.hSize))
	total += int64(n)
	return total, err
}

type dbHeaderIndex struct {
	tag      uint32 // the key
	dataType uint32 // data type
	offset   uint32 // where to find the data in the storage area
	count    uint32 // how many data items are stored in this key
}

type dbPayload interface {
	io.WriterTo
	size() uint32
}

type dbEntry struct {
	index   *dbHeaderIndex
	payload dbPayload
}

// Header represents a header section "in the rpm-sense". It is basically an embedded database.
type Header struct {
	dbHeader *dbHeader
	entries  []*dbEntry
}

func (h *Header) sort() {
	sort.Slice(h.entries, func(i, j int) bool {
		one := h.entries[i]
		two := h.entries[j]
		return one.index.tag < two.index.tag
	})
}

func (h *Header) normalize() {
	h.sort()
	currentOffset := uint32(0)
	for _, e := range h.entries {
		e.index.offset = currentOffset
		currentOffset += e.payload.size()
	}
	if h.dbHeader.hSize != currentOffset {
		panic(fmt.Sprintf("mismatch: header size %d  vs. sum of offsets %d", h.dbHeader.hSize, currentOffset))
	}
}

type binaryPayload struct {
	data []byte
}

func (b *binaryPayload) size() uint32 {
	return uint32(len(b.data))
}

func (b *binaryPayload) WriteTo(r io.Writer) (int64, error) {
	n, err := r.Write(b.data)
	return int64(n), err
}

func binaryEntry(tag uint32, data []byte) *dbEntry {
	e := dbEntry{}
	e.index = &dbHeaderIndex{
		tag:      tag,
		count:    uint32(len(data)),
		dataType: 7,
	}
	e.payload = &binaryPayload{data: data}
	return &e
}

func uint32Entry(tag uint32, data uint32) *dbEntry {
	e := dbEntry{}
	e.index = &dbHeaderIndex{
		tag:      tag,
		count:    1,
		dataType: 4,
	}
	e.payload = &uint32Payload{data: data}
	return &e
}

type uint32Payload struct {
	data uint32
}

func (u *uint32Payload) size() uint32 {
	return 4
}

func (u *uint32Payload) WriteTo(r io.Writer) (int64, error) {
	n, err := r.Write(uint32ToBytes(u.data))
	return int64(n), err
}

func stringEntry(tag uint32, data string) *dbEntry {
	e := dbEntry{}
	e.index = &dbHeaderIndex{
		tag:      tag,
		dataType: 6,
		count:    1,
	}
	e.payload = &stringPayload{data: data}
	return &e
}

type stringPayload struct {
	data string
}

func (s *stringPayload) size() uint32 {
	return uint32(len(s.data))
}

func (s *stringPayload) WriteTo(r io.Writer) (int64, error) {
	n, err := r.Write([]byte(s.data))
	return int64(n), err
}
