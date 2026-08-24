package timeutil

import (
	"time"
)

// Beijing is GMT+8.
var Beijing = time.FixedZone("CST", 8*3600)

func Now() time.Time {
	return time.Now().In(Beijing)
}

func NowNaive() time.Time {
	return Now().Truncate(time.Millisecond)
}

func Format(t time.Time) string {
	if t.IsZero() {
		t = Now()
	}
	return t.In(Beijing).Format("2006-01-02 15:04:05.000")
}

func FormatRFC3339(t time.Time) string {
	if t.IsZero() {
		t = Now()
	}
	return t.In(Beijing).Format(time.RFC3339Nano)
}
