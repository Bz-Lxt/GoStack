package tcp

import (
	"net"
	"net/netip"
	"strconv"
)

type Addr struct {
	IP   netip.Addr
	Port uint16
}

func (a Addr) Network() string { return "tcp" }

func (a Addr) String() string {
	return net.JoinHostPort(a.IP.String(), strconv.Itoa(int(a.Port)))
}

func ParseListen(s string) (netip.Addr, uint16, error) {
	host, port, err := net.SplitHostPort(s)
	if err != nil {
		return netip.Addr{}, 0, err
	}
	ip, err := netip.ParseAddr(host)
	if err != nil {
		return netip.Addr{}, 0, err
	}
	p, err := strconv.Atoi(port)
	if err != nil || p < 0 || p > 65535 {
		return netip.Addr{}, 0, err
	}
	return ip, uint16(p), nil
}
