//go:build linux

package bpf

import (
	"encoding/binary"
	"fmt"
	"os"
	"strconv"
	"syscall"
	"unsafe"

	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/rlimit"
)

type perfEventAttr struct {
	Type, Size     uint32
	Config         uint64
	SamplePeriod   uint64
	SampleType     uint64
	ReadFormat     uint64
	Disabled       uint64
	Inherit        uint64
	Pinned         uint64
	Exclusive      uint64
	ExcludeUser    uint64
	ExcludeKernel  uint64
	ExcludeHypervisor uint64
	ExcludeIdle    uint64
	Mmap           uint64
	Comm           uint64
	Freq           uint64
}

const (
	PERF_TYPE_HARDWARE = 0; PERF_ATTR_SIZE_VER7 = 112
	PERF_COUNT_HW_CPU_CYCLES = 0; PERF_COUNT_HW_INSTRUCTIONS = 1; PERF_COUNT_HW_CACHE_MISSES = 3
	PERF_FORMAT_TOTAL_TIMES = 1; PERF_FORMAT_ID = 4
	PERF_EVENT_IOC_SET_BPF = 0x40042408; PERF_EVENT_IOC_ENABLE = 0x20002400
)

type perfEvent struct { FD, CPU int }

func getNumCPUs() (int, error) {
	data, _ := os.ReadFile("/sys/devices/system/cpu/online")
	var s, e int; if n, _ := fmt.Sscanf(string(data), "%d-%d", &s, &e); n == 2 { return e - s + 1, nil }
	return 1, nil
}

func GetNumCPUs() int { n, _ := getNumCPUs(); return n }

func openPerf(config, freq uint64) ([]*perfEvent, error) {
	ncpu, _ := getNumCPUs()
	attr := &perfEventAttr{Type: PERF_TYPE_HARDWARE, Size: PERF_ATTR_SIZE_VER7, Config: config, SamplePeriod: freq, Freq: 1, ReadFormat: PERF_FORMAT_TOTAL_TIMES | PERF_FORMAT_ID, Disabled: 1, Inherit: 1}
	var evts []*perfEvent
	for cpu := 0; cpu < ncpu; cpu++ {
		fd, _, errno := syscall.Syscall6(syscall.SYS_PERF_EVENT_OPEN, uintptr(unsafe.Pointer(attr)), ^uintptr(0), uintptr(cpu), 0, 0, 0)
		if errno != 0 { for _, e := range evts { syscall.Close(e.FD) }; return nil, errno }
		evts = append(evts, &perfEvent{FD: int(fd), CPU: cpu})
	}
	return evts, nil
}

type PMUCollector struct { Coll *ebpf.Collection; Ticks, Counts, Misses []*perfEvent }

func NewPMUCollector(path string) (*PMUCollector, error) {
	_ = rlimit.RemoveMemlock()
	spec, err := ebpf.LoadCollectionSpec(path); if err != nil { return nil, err }
	coll, err := ebpf.NewCollection(spec); if err != nil { return nil, err }
	pc := &PMUCollector{Coll: coll}
	if pc.Ticks, err = openPerf(PERF_COUNT_HW_CPU_CYCLES, 99); err != nil { pc.Close(); return nil, err }
	if pc.Counts, err = openPerf(PERF_COUNT_HW_INSTRUCTIONS, 99); err != nil { pc.Close(); return nil, err }
	if pc.Misses, err = openPerf(PERF_COUNT_HW_CACHE_MISSES, 99); err != nil { pc.Close(); return nil, err }
	
	fill := func(m *ebpf.Map, evts []*perfEvent) error {
		for _, e := range evts { if err := m.Update(uint32(e.CPU), uint32(e.FD), ebpf.UpdateAny); err != nil { return err } }; return nil
	}
	if err := fill(coll.Maps["cycles_events"], pc.Ticks); err != nil { pc.Close(); return nil, err }
	if err := fill(coll.Maps["instr_events"], pc.Counts); err != nil { pc.Close(); return nil, err }
	if err := fill(coll.Maps["miss_events"], pc.Misses); err != nil { pc.Close(); return nil, err }
	
	prog := coll.Programs["collect_pmu_counters"]; pfd := prog.FD()
	for _, e := range pc.Ticks { 
		syscall.Syscall(syscall.SYS_IOCTL, uintptr(e.FD), PERF_EVENT_IOC_SET_BPF, uintptr(pfd))
		syscall.Syscall(syscall.SYS_IOCTL, uintptr(e.FD), PERF_EVENT_IOC_ENABLE, 0)
	}
	return pc, nil
}

func (p *PMUCollector) Snapshots() *ebpf.Map { return p.Coll.Maps["pmu_snapshots"] }
func (p *PMUCollector) Close() {
	for _, evs := range [][]*perfEvent{p.Ticks, p.Counts, p.Misses} { for _, e := range evs { if e.FD > 0 { syscall.Close(e.FD) } } }
	if p.Coll != nil { p.Coll.Close() }
}

type MapReader struct { m *ebpf.Map; keys []uint32 }
func NewMapReader(m *ebpf.Map, ncpu int) *MapReader {
	keys := make([]uint32, ncpu); for i := 0; i < ncpu; i++ { keys[i] = uint32(i) }; return &MapReader{m: m, keys: keys}
}

type Snapshot struct { 
	CPU uint32; CgroupID uint64; 
	Cycles, CyclesEn, CyclesRun, Instr, InstrEn, InstrRun, Miss, MissEn, MissRun uint64 
}

func (r *MapReader) ForEach(fn func(Snapshot) error) error {
	var key struct{ CPU uint32; CgroupID uint64 }; var val [72]byte
	for _, cpu := range r.keys {
		key.CPU = cpu; if err := r.m.Lookup(&key, &val); err != nil { continue }
		fn(Snapshot{CPU: key.CPU, CgroupID: key.CgroupID, 
			Cycles: binary.NativeEndian.Uint64(val[0:8]), CyclesEn: binary.NativeEndian.Uint64(val[8:16]), CyclesRun: binary.NativeEndian.Uint64(val[16:24]),
			Instr: binary.NativeEndian.Uint64(val[24:32]), InstrEn: binary.NativeEndian.Uint64(val[32:40]), InstrRun: binary.NativeEndian.Uint64(val[40:48]),
			Miss: binary.NativeEndian.Uint64(val[48:56]), MissEn: binary.NativeEndian.Uint64(val[56:64]), MissRun: binary.NativeEndian.Uint64(val[64:72])})
	}
	return nil
}
func (r *MapReader) Delete(s Snapshot) error { key := struct{ CPU uint32; CgroupID uint64 }{s.CPU, s.CgroupID}; return r.m.Delete(&key) }
