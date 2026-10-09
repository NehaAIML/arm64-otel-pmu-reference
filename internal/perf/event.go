//go:build linux
// +build linux

package perf

import (
    "fmt"
    "golang.org/x/sys/unix"
)

func OpenAttr(eventType, config uint64, pid int, cpu int, groupFd int, flags int) (int, error) {
    attr := &unix.PerfEventAttr{
        Type:       uint32(eventType),
        Config:     config,
        Size:       uint32(unix.SizeOfPerfEventAttr),
        SampleFreq: 99,
        SampleType: unix.PERF_SAMPLE_READ,
    }
    attr.SetFreq(1)
    attr.SetInherit(0)

    fd, err := unix.PerfEventOpen(attr, pid, cpu, groupFd, flags)
    if err != nil {
        return -1, fmt.Errorf("perf_event_open failed: %w", err)
    }
    return fd, nil
}
