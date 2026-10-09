package pmuinfo

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type MIDRInfo struct {
	Implementer, Variant, Revision uint8
	PartNum                        uint16
}

func ReadMIDR() (MIDRInfo, error) {
	f, _ := os.Open("/proc/cpuinfo")
	defer f.Close()
	var info MIDRInfo
	s := bufio.NewScanner(f)
	for s.Scan() {
		p := strings.SplitN(s.Text(), ":", 2)
		if len(p) != 2 {
			continue
		}
		k, v := strings.TrimSpace(p[0]), strings.TrimSpace(p[1])
		var u uint64
		var e error
		switch k {
		case "CPU implementer":
			if u, e = strconv.ParseUint(v, 16, 8); e == nil {
				info.Implementer = uint8(u)
			}
		case "CPU part":
			if u, e = strconv.ParseUint(v, 16, 16); e == nil {
				info.PartNum = uint16(u)
			}
		}
	}
	return info, nil
}
