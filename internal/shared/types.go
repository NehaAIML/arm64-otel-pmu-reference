package shared

type Snapshot struct {
	CPU       uint32
	CgroupID  uint64
	Cycles    uint64
	CyclesEn  uint64
	CyclesRun uint64
	Instr     uint64
	InstrEn   uint64
	InstrRun  uint64
	Miss      uint64
	MissEn    uint64
	MissRun   uint64
}

type MapReader interface {
	ForEach(fn func(Snapshot) error) error
	Delete(s Snapshot) error
}
