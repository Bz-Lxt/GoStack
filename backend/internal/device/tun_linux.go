//go:build linux

package device

import (
	"os"
	"strings"
	"unsafe"

	"golang.org/x/sys/unix"
)

const (
	iffTun    = 0x0001
	iffNoPI   = 0x1000
	tunsetiff = 0x400454ca
)

type ifreq struct {
	name  [16]byte
	flags uint16
	_     [22]byte
}

type TUN struct {
	f    *os.File
	name string
	mtu  int
}

func OpenTUN(name string, mtu int) (Device, error) {
	if mtu <= 0 {
		mtu = 1500
	}
	f, err := os.OpenFile("/dev/net/tun", os.O_RDWR, 0)
	if err != nil {
		return nil, err
	}
	var req ifreq
	copy(req.name[:], name)
	req.flags = iffTun | iffNoPI
	_, _, errno := unix.Syscall(unix.SYS_IOCTL, f.Fd(), tunsetiff, uintptr(unsafe.Pointer(&req)))
	if errno != 0 {
		_ = f.Close()
		return nil, errno
	}
	got := strings.TrimRight(string(req.name[:]), "\x00")
	return &TUN{f: f, name: got, mtu: mtu}, nil
}

func (t *TUN) Name() string                 { return t.name }
func (t *TUN) MTU() int                     { return t.mtu }
func (t *TUN) Read(p []byte) (int, error)   { return t.f.Read(p) }
func (t *TUN) Write(p []byte) (int, error)  { return t.f.Write(p) }
func (t *TUN) Close() error                 { return t.f.Close() }
