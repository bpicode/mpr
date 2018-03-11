package head

import (
	"fmt"
	"io"
)

// Lead is the first part of an RPM package file. In previous versions of RPM, it was used to store information used
// internally by RPM. Today, however, the lead's sole purpose is to make it easy to identify an RPM package file.
// For example, the file(1) command uses the lead. All the information contained in the lead has been duplicated or
// superseded by information contained in the header.
// See also:
// * http://rpm5.org/docs/api/structrpmlead.html
// * http://ftp.rpm.org/max-rpm/s1-rpm-file-format-rpm-file-format.html
type Lead struct {
	magic       magic
	leadVersion leadVersion
	pkgType     pkgType
	arch        arch
	pkg         pkg
	os          osys
	sigtype     sigtype
	reserved    reserved
}

// WriteTo writes the data of the receiver Lead to an io.Writer.
func (l *Lead) WriteTo(to io.Writer) (int64, error) {
	ws := []io.WriterTo{l.magic, l.leadVersion, l.pkgType, l.arch, l.pkg, l.os, l.sigtype, l.reserved}
	var n int64
	for _, w := range ws {
		num, err := w.WriteTo(to)
		n += num
		if err != nil {
			return n, fmt.Errorf("failed to write lead: %v", err)
		}
	}
	return n, nil
}

// NewLead creates a *Lead with defaults, which may be overridden by the passed Options.
func NewLead(opts ...Option) *Lead {
	var l Lead
	l.magic = []byte{0xed, 0xab, 0xee, 0xdb}
	l.leadVersion = leadVersion{major: 3, minor: 0}
	l.pkgType = Binary
	l.arch = I386x8664
	l.os = LINUX
	l.sigtype = HeaderStyle
	l.reserved = make([]byte, 16)
	for _, o := range opts {
		o(&l)
	}
	return &l
}

// Option is an action that reconfigures the Lead section.
type Option func(lead *Lead)

// Package configures the package metadata, in particular the name of the package, its version and release.
// For package like myapp-2.2.1-1.i386.rpm, name = "myapp", version = "2.2.1", release = "1".
func Package(name, version, release string) Option {
	return func(l *Lead) {
		var p pkg
		p.name = name
		p.version = version
		p.release = release
		l.pkg = p
	}
}

// Bin mark an rpm as a binary package.
func Bin() Option {
	return func(l *Lead) {
		l.pkgType = Binary
	}
}

// Src mark an rpm as a source package.
func Src() Option {
	return func(l *Lead) {
		l.pkgType = Source
	}
}

// Arch sets the target architecture.
func Arch(a arch) Option {
	return func(l *Lead) {
		l.arch = a
	}
}

type leadVersion struct {
	major byte
	minor byte
}

func (v leadVersion) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write([]byte{v.major, v.minor})
	if err != nil {
		return int64(n), fmt.Errorf("cannot write lead leadVersion: %v", err)
	}
	return int64(n), nil
}

type pkgType uint16

// Types of packages.
const (
	Binary pkgType = 0
	Source pkgType = 1
)

func (p pkgType) WriteTo(w io.Writer) (int64, error) {
	var h, l = uint8(p >> 8), uint8(p & 0xff)
	n, err := w.Write([]byte{h, l})
	if err != nil {
		return int64(n), fmt.Errorf("cannot write package type: %v", err)
	}
	return int64(n), nil
}

type arch uint16

// Supported architectures.
const (
	_              = iota
	I386x8664 arch = iota
	AlphaSparc64
	SPARC
	MIPS
	PowerPC
	M68K
	SGI
	RS6000
	IA64
	Sparc64
	MIPSel
	ARM
	MiNT
	S390
	S390x
	PowerPC64
	SuperH
	Xtensa
	NOARCH arch = 0xff
)

func (a arch) WriteTo(w io.Writer) (int64, error) {
	h, l := uint8(a>>8), uint8(a&0xff) //noformat
	n, err := w.Write([]byte{h, l})
	if err != nil {
		return int64(n), fmt.Errorf("cannot write arch: %v", err)
	}
	return int64(n), nil
}

type pkg struct {
	name    string
	version string
	release string
}

func (p pkg) WriteTo(w io.Writer) (int64, error) {
	nameVersionRelease := []byte(p.name + "-" + p.version + "-" + p.release)
	numZeroesToPad := 66 - len(nameVersionRelease)
	if numZeroesToPad < 0 {
		return 0, fmt.Errorf("unable to write name-version-release for '%s', it may not exceed 65 bytes", nameVersionRelease)
	}
	n, err := w.Write(nameVersionRelease)
	if err != nil {
		return int64(n), fmt.Errorf("failed to write name-version-release: %v", err)
	}
	zeroes := make([]byte, numZeroesToPad)
	nz, err := w.Write(zeroes)
	if err != nil {
		return int64(n + nz), fmt.Errorf("faild to write %d zeroes after name-version-release: %v", numZeroesToPad, err)
	}
	return int64(n + nz), nil
}

type osys uint16

// Supported operating systems.
const (
	UNKNOWN osys = iota
	LINUX
	IRIX
	SOLARIS
	SUNOS
	AMIGAOS
	AIX
	HPUX10
	OSF1
	FREEBSD
	SCO
	IRIX64
	NEXTSTEP
	BSDI
	MACHTEN
	CYGWINNT
	CYGWIN95
	UNIXSV
	MINT
	OS390
	VMESA
	LINUX390
	MACOSX
)

func (o osys) WriteTo(w io.Writer) (int64, error) {
	var h, l = uint8(o >> 8), uint8(o & 0xff)
	n, err := w.Write([]byte{h, l})
	if err != nil {
		return int64(n), fmt.Errorf("cannot write os: %v", err)
	}
	return int64(n), nil
}

type sigtype uint16

const (
	_ = iota
	_ = iota
	_ = iota
	_ = iota
	_ = iota
	// HeaderStyle indicates "Header-style" signatures, for version 3.0 packages.
	HeaderStyle = iota
)

func (s sigtype) WriteTo(w io.Writer) (int64, error) {
	var h, l = uint8(s >> 8), uint8(s &
		0xff)
	n, err := w.Write([]byte{h, l})
	if err != nil {
		return int64(n), fmt.Errorf("cannot write signature type: %v", err)
	}
	return int64(n), nil
}

type reserved []byte

func (r reserved) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(r)
	if err != nil {
		return int64(n), fmt.Errorf("cannot reserved %d bytes: %v", len(r), err)
	}
	return int64(n), nil
}

type magic []byte

func (m magic) WriteTo(w io.Writer) (int64, error) {
	n, err := w.Write(m)
	if err != nil {
		return int64(n), fmt.Errorf("cannot reserved %d magic bytes: %v", len(m), err)
	}
	return int64(n), nil
}
