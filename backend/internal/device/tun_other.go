//go:build !linux && !darwin

package device

import "fmt"

func OpenTUN(name string, mtu int) (Device, error) {
	return nil, fmt.Errorf("tun not supported on this platform (name=%s mtu=%d)", name, mtu)
}
