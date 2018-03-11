package head

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSignature(t *testing.T) {
	s := newSignature()
	b := bytes.NewBuffer(nil)

	n, err := s.WriteTo(b)
	assert.NoError(t, err)
	assert.True(t, n > 0)

	bs := b.Bytes()
	assert.True(t, len(bs) > 0)
	fmt.Println(hex.Dump(bs))
}
