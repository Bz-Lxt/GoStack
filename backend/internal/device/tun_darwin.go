//go:build darwin

package device

import (
	"encoding/binary"
	"fmt"
	"os"
	"syscall"

	"golang.org/x/sys/unix"
)

const (
	utunControlName  = "com.apple.net.utun_control"
	sysprotoControl  = 2
	utunOptIfname    = 2
)

type TUN struct {
	f    *os.File
	name string
	mtu  int
}

func OpenTUN(name string, mtu int) (Device, error) {
	if mtu <= 0 {
		mtu = 1500
	}
	fd, err := unix.Socket(unix.AF_SYSTEM, unix.SOCK_DGRAM, sysprotoControl)
	if err != nil {
		return nil, fmt.Errorf("utun socket: %w", err)
	}
	info := unix.CtlInfo{}
	copy(info.Name[:], utunControlName)
	if err := unix.IoctlCtlInfo(fd, &info); err != nil {
		_ = unix.Close(fd)
		return nil, fmt.Errorf("utun ctlinfo: %w", err)
	}
	sa := &unix.SockaddrCtl{ID: info.Id, Unit: 0}
	if err := unix.Connect(fd, sa); err != nil {
		_ = unix.Close(fd)
		return nil, fmt.Errorf("utun connect: %w", err)
	}
	ifn, err := unix.GetsockoptString(fd, sysprotoControl, utunOptIfname)
	if err != nil {
		ifn = name
		if ifn == "" {
			ifn = "utun"
		}
	}
	if err := unix.SetNonblock(fd, false); err != nil {
		_ = unix.Close(fd)
		return nil, err
	}
	f := os.NewFile(uintptr(fd), ifn)
	return &TUN{f: f, name: ifn, mtu: mtu}, nil
}

func (t *TUN) Name() string { return t.name }
func (t *TUN) MTU() int     { return t.mtu }

func (t *TUN) Read(p []byte) (int, error) {
	buf := make([]byte, len(p)+4)
	n, err := t.f.Read(buf)
	if err != nil {
		return 0, err
	}
	if n < 4 {
		return 0, syscall.EIO
	}
	return copy(p, buf[4:n]), nil
}

func (t *TUN) Write(p []byte) (int, error) {
	buf := make([]byte, len(p)+4)
	binary.BigEndian.PutUint32(buf[:4], unix.AF_INET)
	copy(buf[4:], p)
	n, err := t.f.Write(buf)
	if err != nil {
		return 0, err
	}
	if n < 4 {
		return 0, syscall.EIO
	}
	return n - 4, nil
}

func (t *TUN) Close() error { return t.f.Close() }
