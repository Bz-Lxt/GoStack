package device

import "io"

type Device interface {
	Name() string
	Read(p []byte) (int, error)
	Write(p []byte) (int, error)
	Close() error
	MTU() int
}

func CopyBuf(p []byte) []byte {
	c := make([]byte, len(p))
	copy(c, p)
	return c
}

var _ io.ReadWriteCloser = (Device)(nil)
