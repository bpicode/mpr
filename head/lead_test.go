package head

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteLeadVsGnuFile(t *testing.T) {
	l := NewLead(
		Package("myapp", "1.0.3", "1"),
		Bin(),
		Arch(I386x8664),
	)
	f, err := ioutil.TempFile("", "myapp")
	assert.NoError(t, err)

	n, err := l.WriteTo(f)
	assert.NoError(t, err)
	assert.Equal(t, int64(95), n)

	info, err := f.Stat()
	assert.NoError(t, err)
	assert.Equal(t, int64(95), info.Size())

	c := exec.Command("file", f.Name())
	stdOutCapture := bytes.NewBuffer(nil)
	c.Stdout = stdOutCapture
	c.Stderr = os.Stderr
	err = c.Run()
	assert.NoError(t, err)

	capture := stdOutCapture.String()
	assert.Contains(t, capture, "RPM")
	assert.Contains(t, capture, "x86_64")
	assert.Contains(t, capture, "bin")

	defer f.Close()
	defer os.RemoveAll(f.Name())
}
