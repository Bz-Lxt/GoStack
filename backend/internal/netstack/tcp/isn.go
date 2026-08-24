package tcp

import (
	"encoding/binary"
	"hash/fnv"
	"time"
)

func GenerateISN(q Quad) uint32 {
	h := fnv.New32a()
	lb, _ := q.LocalIP.MarshalBinary()
	rb, _ := q.RemoteIP.MarshalBinary()
	_, _ = h.Write(lb)
	_, _ = h.Write(rb)
	var ports [4]byte
	binary.BigEndian.PutUint16(ports[0:2], q.LocalPort)
	binary.BigEndian.PutUint16(ports[2:4], q.RemotePort)
	_, _ = h.Write(ports[:])
	return uint32(time.Now().UnixMicro()) + h.Sum32()
}
