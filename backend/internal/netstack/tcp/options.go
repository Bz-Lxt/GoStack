package tcp

const (
	optEOL  = 0
	optNOP  = 1
	optMSS  = 2
	optWS   = 3
	optSACK = 4
	optTS   = 8
)

type Options struct {
	MSS           uint16
	HasMSS        bool
	WScale        uint8
	HasWScale     bool
	SACKPermitted bool
	HasTS         bool
	TSVal         uint32
	TSEcr         uint32
}

func ParseOptions(b []byte) Options {
	var o Options
	for i := 0; i < len(b); {
		kind := b[i]
		if kind == optEOL {
			break
		}
		if kind == optNOP {
			i++
			continue
		}
		if i+1 >= len(b) {
			break
		}
		n := int(b[i+1])
		if n < 2 || i+n > len(b) {
			break
		}
		body := b[i+2 : i+n]
		switch kind {
		case optMSS:
			if len(body) == 2 {
				o.MSS = uint16(body[0])<<8 | uint16(body[1])
				o.HasMSS = true
			}
		case optWS:
			if len(body) == 1 {
				o.WScale = body[0]
				o.HasWScale = true
			}
		case optSACK:
			o.SACKPermitted = true
		case optTS:
			if len(body) == 8 {
				o.HasTS = true
				o.TSVal = uint32(body[0])<<24 | uint32(body[1])<<16 | uint32(body[2])<<8 | uint32(body[3])
				o.TSEcr = uint32(body[4])<<24 | uint32(body[5])<<16 | uint32(body[6])<<8 | uint32(body[7])
			}
		}
		i += n
	}
	return o
}

func EncodeMSS(mss uint16) []byte {
	return []byte{optMSS, 4, byte(mss >> 8), byte(mss)}
}
