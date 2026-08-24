package netstack

import (
	"fmt"
	"os/exec"
	"runtime"
)

func ConfigureTUN(name, hostIP string, prefix int) error {
	if runtime.GOOS != "linux" {
		return nil
	}
	cmds := [][]string{
		{"ip", "addr", "replace", fmt.Sprintf("%s/%d", hostIP, prefix), "dev", name},
		{"ip", "link", "set", name, "up"},
		{"ip", "route", "replace", "10.0.0.0/24", "dev", name},
	}
	for _, c := range cmds {
		out, err := exec.Command(c[0], c[1:]...).CombinedOutput()
		if err != nil {
			return fmt.Errorf("%v: %s: %w", c, string(out), err)
		}
	}
	return nil
}
